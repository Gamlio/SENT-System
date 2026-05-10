package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	assetSvc "SENT_backend/service/assets/internal/service"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func generateSecureToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetActiveEnrollmentToken (GATE): Lấy mã token hiện tại (nếu còn hạn)
func GetActiveEnrollmentToken(c *gin.Context) {
	orgID := c.GetUint("org_id")

	// Lấy tên người dùng từ token (nếu có) để ghi log ai là người tạo
	usernameVal, _ := c.Get("username")
	creator := "SYSTEM_AUTO"
	if usernameVal != nil {
		creator = usernameVal.(string)
	}

	var tokenRecord models.EnrollmentToken

	// 1. Tìm mã token còn hạn
	err := database.DB.Select("token", "expires_at").
		Where("org_id = ? AND expires_at > ?", orgID, time.Now()).
		Order("expires_at desc").
		First(&tokenRecord).Error

	// 2. NẾU KHÔNG TÌM THẤY (Hoặc đã hết hạn) -> TỰ ĐỘNG SINH MÃ MỚI
	if err != nil {
		randomPart := strings.ToUpper(generateSecureToken(4))
		tokenString := fmt.Sprintf("SENT-%d-%s", orgID, randomPart)

		tokenRecord = models.EnrollmentToken{
			Token:     tokenString,
			OrgID:     orgID,
			CreatedBy: creator,
			ExpiresAt: time.Now().Add(15 * time.Minute), // Mã có thời hạn 15 phút
		}

		// Lưu vào Database
		if err := database.DB.Create(&tokenRecord).Error; err != nil {
			c.JSON(500, gin.H{"error": "Không thể tự động sinh mã cài đặt"})
			return
		}
	}

	// 3. Trả về mã (Dù là mã cũ còn hạn hay mã vừa mới tạo)
	c.JSON(200, gin.H{
		"token":      tokenRecord.Token,
		"expires_at": tokenRecord.ExpiresAt,
		"expires_in": int(time.Until(tokenRecord.ExpiresAt).Seconds()),
	})
}

// GenerateEnrollmentToken (GATE): Sinh mã cài đặt 24h
func GenerateEnrollmentToken(c *gin.Context) {
	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	// 1. Kiểm tra xem đã có mã nào còn hạn không
	var existingToken models.EnrollmentToken
	if err := database.DB.Where("org_id = ? AND expires_at > ?", orgID, time.Now()).First(&existingToken).Error; err == nil {
		c.JSON(400, gin.H{"error": "Mã cũ vẫn còn hiệu lực. Vui lòng đợi mã cũ hết hạn!"})
		return
	}

	randomPart := strings.ToUpper(generateSecureToken(4))
	tokenString := fmt.Sprintf("SENT-%d-%s", orgID, randomPart)

	tokenRecord := models.EnrollmentToken{
		Token:     tokenString,
		OrgID:     orgID,
		CreatedBy: username.(string),
		ExpiresAt: time.Now().Add(15 * time.Minute), // Cập nhật thành 15 phút
	}

	if err := database.DB.Create(&tokenRecord).Error; err != nil {
		c.JSON(500, gin.H{"error": "Không thể sinh mã"})
		return
	}

	// Trả về thêm expires_in tương tự
	c.JSON(200, gin.H{
		"token":      tokenRecord.Token,
		"expires_at": tokenRecord.ExpiresAt,
		"expires_in": int(time.Until(tokenRecord.ExpiresAt).Seconds()),
	})
}

// EnrollAsset (GATE): Tiếp nhận đăng ký từ máy trạm
func EnrollAsset(c *gin.Context) {
	var req models.EnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	var tokenRecord models.EnrollmentToken
	// TỐI ƯU: Chỉ cần kiểm tra sự tồn tại (Exists), không cần lấy cả record nếu chỉ để check
	err := database.DB.Select("org_id").
		Where("token = ? AND expires_at > ?", req.Token, time.Now()).
		First(&tokenRecord).Error

	if err != nil {
		c.JSON(401, gin.H{"error": "Mã cài đặt không hợp lệ hoặc đã hết hạn"})
		return
	}

	newSecretKey := generateSecureToken(16)
	svc := &assetSvc.AssetLifecycleService{}

	// Sử dụng AssetHWID đã chuẩn hóa trong request
	if err := svc.EnrollWithKey(req, tokenRecord.OrgID, newSecretKey); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":    "Đăng ký thành công",
		"secret_key": newSecretKey,
	})
}
