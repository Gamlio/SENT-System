package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Kiểm tra chữ ký điện tử của asset
func assetHMACAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy chữ ký từ Header
		signatureHeader := c.GetHeader("X-Sent-Signature")
		if signatureHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Từ chối truy cập: Thiếu chữ ký (X-Sent-Signature)"})
			c.Abort()
			return
		}

		bodyBytes, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể đọc dữ liệu"})
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 3. Lấy HWID từ Body để tìm asset trong DB
		var payload struct {
			HWID string `json:"asset_hwid"`
		}
		if err := json.Unmarshal(bodyBytes, &payload); err != nil || payload.HWID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payload không hợp lệ (Thiếu HWID)"})
			c.Abort()
			return
		}

		// 4. Tìm asset và lấy SecretKey
		var asset models.Asset
		if err := database.DB.Where("asset_hwid = ?", payload.HWID).First(&asset).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "asset không tồn tại trong hệ thống"})
			c.Abort()
			return
		}

		// 5. Tính toán lại mã băm HMAC-SHA256
		mac := hmac.New(sha256.New, []byte(asset.SecretKey))
		mac.Write(bodyBytes)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		// 6. So sánh chữ ký
		if signatureHeader != expectedSignature {
			// Đây là Hacker giả mạo!
			c.JSON(http.StatusForbidden, gin.H{"error": "CẢNH BÁO BẢO MẬT: Chữ ký giả mạo hoặc dữ liệu đã bị sửa đổi trên đường truyền!"})
			c.Abort()
			return
		}

		// Bơm thông tin asset vào context để các handler sau sử dụng
		c.Set("asset_hwid", asset.AssetHWID)
		c.Set("org_id", asset.OrgID)

		c.Next()
	}
}
