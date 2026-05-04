package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"time"

	"sent_backend/internal/cache"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetAssetSecretFromCache lấy SecretKey từ Cache, nếu không có thì query DB và lưu lại
func GetAssetSecretFromCache(hwid string) (string, uint, error) {
	// 1. Ưu tiên lấy từ Distributed Redis Cache
	secret, orgID, err := cache.GetAssetCache(hwid)
	if err == nil {
		return secret, orgID, nil
	}

	// Cache miss, truy vấn DB
	var asset models.Asset
	if err := database.DB.Where("asset_hwid = ?", hwid).First(&asset).Error; err != nil {
		return "", 0, err
	}

	// Lưu JSON string vào Redis Cache với TTL 1 giờ để đồng bộ trạng thái cluster
	cache.SetAssetCache(hwid, asset.SecretKey, asset.OrgID, 1*time.Hour)

	return asset.SecretKey, asset.OrgID, nil
}

// AssetHMACAuth kiểm tra chữ ký điện tử của asset
func AssetHMACAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy HWID từ Header (Lazy-Parsing triệt để, không đọc body)
		hwid := c.GetHeader("X-Asset-HWID")
		if hwid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu HWID (X-Asset-HWID) trên Header"})
			c.Abort()
			return
		}

		// 2. Lấy chữ ký, timestamp và sequence từ Header
		sig := c.GetHeader("X-Sent-Signature")
		ts := c.GetHeader("X-Sent-Timestamp")    // Chống Replay
		seqStr := c.GetHeader("X-Sent-Sequence") // Chống Replay tuyệt đối

		if sig == "" || ts == "" || seqStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Từ chối truy cập: Thiếu chữ ký, timestamp hoặc sequence number"})
			c.Abort()
			return
		}

		// 3. Chống Replay Attack (Kiểm tra timestamp)
		reqTimestamp, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Timestamp không hợp lệ"})
			c.Abort()
			return
		}

		now := time.Now().Unix()
		// Cho phép lệch tối đa 60 giây để giảm thiểu rủi ro
		if now-reqTimestamp > 60 || reqTimestamp-now > 60 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Yêu cầu đã hết hạn (Replay Attack Protection)"})
			c.Abort()
			return
		}

		// Kiểm tra Sequence Number qua Redis (Chống Replay tuyệt đối)
		seq, err := strconv.ParseInt(seqStr, 10, 64)
		if err != nil || !cache.CheckSequenceNumber(hwid, seq) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Yêu cầu trùng lặp hoặc Sequence không hợp lệ (Replay Attack Detected)"})
			c.Abort()
			return
		}

		// 4. Lấy SecretKey từ Cache hoặc DB TRƯỚC KHI đọc body
		secretKey, orgID, err := GetAssetSecretFromCache(hwid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Asset không tồn tại trong hệ thống"})
			c.Abort()
			return
		}

		// 5. Streaming HMAC: Tính toán mã băm ĐỒNG THỜI khi đọc Body (Zero-Copy memory optimization)
		mac := hmac.New(sha256.New, []byte(secretKey))

		// Dùng io.TeeReader để vừa đọc request body vừa đẩy dữ liệu vào hasher
		tee := io.TeeReader(c.Request.Body, mac)
		bodyBytes, err := io.ReadAll(tee)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể đọc dữ liệu payload"})
			c.Abort()
			return
		}

		// Chuyển raw body vào context cho các middleware phía sau
		c.Set("raw_body", bodyBytes)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 6. Tính mã băm mong đợi
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		// 7. So sánh an toàn (Chống Timing Attack)
		if subtle.ConstantTimeCompare([]byte(sig), []byte(expectedSignature)) != 1 {
			// Đây là Hacker giả mạo!
			c.JSON(http.StatusForbidden, gin.H{"error": "CẢNH BÁO BẢO MẬT: Chữ ký giả mạo hoặc dữ liệu đã bị sửa đổi trên đường truyền!"})
			c.Abort()
			return
		}

		// Bơm thông tin asset vào context để các handler sau sử dụng
		c.Set("asset_hwid", hwid)
		c.Set("org_id", orgID)

		c.Next()
	}
}
