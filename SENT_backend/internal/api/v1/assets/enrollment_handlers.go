package assets

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	assetSvc "sent_backend/internal/service/assets"
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
	username, _ := c.Get("username")

	var tokenRecord models.EnrollmentToken

	// 1. Tìm token gần nhất còn hạn
	err := database.DB.Where("org_id = ? AND expires_at > ?", orgID, time.Now()).
		Order("expires_at desc").
		First(&tokenRecord).Error

	// 2. Nếu không tìm thấy mã nào còn hạn, tiến hành tạo tự động ngay lập tức
	if err != nil {
		randomPart := strings.ToUpper(generateSecureToken(4))
		tokenString := fmt.Sprintf("SENT-%d-%s", orgID, randomPart)

		tokenRecord = models.EnrollmentToken{
			Token:     tokenString,
			OrgID:     orgID,
			CreatedBy: username.(string),
			ExpiresAt: time.Now().Add(15 * time.Minute), // Thời hạn 15 phút
		}

		if err := database.DB.Create(&tokenRecord).Error; err != nil {
			c.JSON(500, gin.H{"error": "Không thể tự động sinh mã cài đặt"})
			return
		}
	}

	// 3. Trả về thông tin mã (dù là mã cũ đang chạy hay mã vừa mới tạo)
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

// Enrollasset (GATE): Tiếp nhận đăng ký từ máy trạm
func Enrollasset(c *gin.Context) {
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

	// 2. Sinh Secret Key ngẫu nhiên cho máy này (Ví dụ dùng hàm có sẵn)
	newSecretKey := generateSecureToken(16)

	// 3. Gọi Service để lưu asset kèm SecretKey vào DB
	svc := &assetSvc.AssetLifecycleService{}
	// Bạn cần sửa hàm này trong lifecycle_service.go để nhận thêm tham số secretKey
	if err := svc.EnrollWithKey(req, tokenRecord.OrgID, newSecretKey); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// [QUAN TRỌNG]: Trả về Key cho asset
	c.JSON(200, gin.H{
		"message":    "Đăng ký thành công",
		"secret_key": newSecretKey,
	})
}
