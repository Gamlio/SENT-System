package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type AssetEnrollmentService struct{}

func (s *AssetEnrollmentService) generateSecureToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetOrCreateActiveToken: Logic lấy hoặc tự sinh mã mới
func (s *AssetEnrollmentService) GetOrCreateActiveToken(orgID uint, creator string) (models.EnrollmentToken, error) {
	var tokenRecord models.EnrollmentToken

	// 1. Tìm mã token còn hạn
	err := database.DB.Select("token", "expires_at").
		Where("org_id = ? AND expires_at > ?", orgID, time.Now()).
		Order("expires_at desc").
		First(&tokenRecord).Error

	// 2. Nếu không có mã hợp lệ -> Sinh mã mới
	if err != nil {
		randomPart := strings.ToUpper(s.generateSecureToken(4))
		tokenString := fmt.Sprintf("SENT-%d-%s", orgID, randomPart)

		tokenRecord = models.EnrollmentToken{
			Token:     tokenString,
			OrgID:     orgID,
			CreatedBy: creator,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}

		if err := database.DB.Create(&tokenRecord).Error; err != nil {
			return tokenRecord, err
		}
	}
	return tokenRecord, nil
}

// ValidateToken: Kiểm tra mã cài đặt khi máy trạm gửi lên
func (s *AssetEnrollmentService) ValidateToken(token string) (uint, error) {
	var tokenRecord models.EnrollmentToken
	err := database.DB.Select("org_id").
		Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&tokenRecord).Error

	if err != nil {
		return 0, fmt.Errorf("mã cài đặt không hợp lệ hoặc đã hết hạn")
	}
	return tokenRecord.OrgID, nil
}
