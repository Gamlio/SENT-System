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

// Kiểm tra chữ ký điện tử của Agent
func AgentHMACAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy chữ ký từ Header
		signatureHeader := c.GetHeader("X-Signature")
		if signatureHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Từ chối truy cập: Thiếu chữ ký (Signature)"})
			c.Abort()
			return
		}

		// 2. Đọc Body (Payload) một cách an toàn
		bodyBytes, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Không thể đọc dữ liệu"})
			c.Abort()
			return
		}
		// Đẩy body ngược lại vào Request để các hàm phía sau (ShouldBindJSON) còn đọc được
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 3. Lấy HWID từ Body để tìm Agent trong DB
		var payload struct {
			HWID string `json:"hwid"`
		}
		if err := json.Unmarshal(bodyBytes, &payload); err != nil || payload.HWID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payload không hợp lệ (Thiếu HWID)"})
			c.Abort()
			return
		}

		// 4. Tìm Agent và lấy SecretKey
		var agent models.Agent
		if err := database.DB.Where("hw_id = ?", payload.HWID).First(&agent).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Agent không tồn tại trong hệ thống"})
			c.Abort()
			return
		}

		// Nếu máy chưa có Secret Key (Đang Pending), tạm cho qua để xử lý Zero-Trust ở Processor
		if agent.SecretKey == "" {
			c.Next()
			return
		}

		// 5. Tính toán lại mã băm HMAC-SHA256
		mac := hmac.New(sha256.New, []byte(agent.SecretKey))
		mac.Write(bodyBytes)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		// 6. So sánh chữ ký
		if signatureHeader != expectedSignature {
			// Đây là Hacker giả mạo!
			c.JSON(http.StatusForbidden, gin.H{"error": "CẢNH BÁO BẢO MẬT: Chữ ký giả mạo hoặc dữ liệu đã bị sửa đổi trên đường truyền!"})
			c.Abort()
			return
		}

		c.Next()
	}
}
