package agents

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	agentSvc "sent_backend/internal/service/agents" // Alias cho service
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func generateSecureToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateEnrollmentToken (GATE): Sinh mã cài đặt 24h
func GenerateEnrollmentToken(c *gin.Context) {
	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	randomPart := strings.ToUpper(generateSecureToken(4))
	tokenString := fmt.Sprintf("SENT-%d-%s", orgID, randomPart)

	tokenRecord := models.EnrollmentToken{
		Token:     tokenString,
		OrgID:     orgID,
		CreatedBy: username.(string),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := database.DB.Create(&tokenRecord).Error; err != nil {
		c.JSON(500, gin.H{"error": "Không thể sinh mã"})
		return
	}
	c.JSON(200, tokenRecord)
}

// EnrollAgent (GATE): Tiếp nhận đăng ký từ máy trạm
func EnrollAgent(c *gin.Context) {
	var req models.EnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// 1. Xác thực Token (Bước bảo vệ cửa ngõ)
	var tokenRecord models.EnrollmentToken
	if err := database.DB.Where("token = ? AND expires_at > ?", req.Token, time.Now()).First(&tokenRecord).Error; err != nil {
		c.JSON(401, gin.H{"error": "Mã cài đặt không hợp lệ hoặc đã hết hạn"})
		return
	}

	// 2. Gọi Service Brain xử lý đăng ký và tạo Ticket
	svc := &agentSvc.AgentLifecycleService{}
	if err := svc.EnrollAgentRequest(req, tokenRecord.OrgID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Đăng ký thành công, vui lòng chờ Admin phê duyệt."})
}
