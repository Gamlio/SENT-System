package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sent_backend/internal/auth"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/utils"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct{}

// RegisterSME: Logic tạo Tổ chức và Admin gốc
func (s *AuthService) RegisterSME(companyName, email, username, password string) (string, string, error) {
	newCompanyCode := s.generateUniqueCompanyCode()

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Tạo Tổ chức
		org := models.Organization{
			Name:              companyName,
			CompanyCode:       newCompanyCode,
			EnrollTokenPrefix: "SENT-" + newCompanyCode,
		}
		if err := tx.Create(&org).Error; err != nil {
			return err
		}

		// 2. Kiểm tra User trùng
		var existingUser models.User
		if err := tx.Where("username = ? AND org_id = ?", username, org.ID).First(&existingUser).Error; err == nil {
			return fmt.Errorf("tên đăng nhập đã tồn tại trong tổ chức này")
		}

		// 3. Tạo Admin gốc với đầy đủ quyền hạn hệ thống
		hashed, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
		user := models.User{
			Username:           username,
			PasswordHash:       string(hashed),
			Email:              email,
			OrgID:              &org.ID,
			PermAssetView:      true,
			PermAssetAction:    true,
			PermAssetDelete:    true,
			PermAssetMove:      true,
			PermPolicyView:     true,
			PermPolicyManage:   true,
			PermIncidentView:   true,
			PermIncidentAction: true,
			PermDocView:        true,
			PermDocManage:      true,
			PermUserView:       true,
			PermUserManage:     true,
			PermGroupManage:    true,
			PermApprovalView:   true,
			PermApprovalFinal:  true,
			PermSystemConfig:   true,
			ApprovalStatus:     "APPROVED",
		}
		return tx.Create(&user).Error
	})

	if err != nil {
		return "", "", err
	}

	// 4. Gửi Email thông báo (Async)
	loginURL := fmt.Sprintf("%s/login/%s", os.Getenv("APP_URL"), newCompanyCode)
	go s.sendWelcomeEmail(companyName, email, newCompanyCode, username, loginURL)

	return newCompanyCode, loginURL, nil
}

// Login: Xác thực và trả về Token cùng ma trận quyền
func (s *AuthService) Login(companyCode, username, password string) (string, *models.User, *models.Organization, error) {
	var org models.Organization
	if err := database.DB.Where("company_code = ?", companyCode).First(&org).Error; err != nil {
		return "", nil, nil, fmt.Errorf("không gian làm việc không tồn tại")
	}

	var user models.User
	if err := database.DB.Where("username = ? AND org_id = ?", username, org.ID).First(&user).Error; err != nil {
		return "", nil, nil, fmt.Errorf("tài khoản hoặc mật khẩu không chính xác")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, nil, fmt.Errorf("tài khoản hoặc mật khẩu không chính xác")
	}

	token, err := auth.GenerateToken(user.Username, *user.OrgID, user.ID)
	return token, &user, &org, err
}

// Helper: Sinh mã công ty[cite: 45]
func (s *AuthService) generateUniqueCompanyCode() string {
	for {
		b := make([]byte, 3)
		rand.Read(b)
		code := "SME-" + strings.ToUpper(hex.EncodeToString(b))
		var existing models.Organization
		if err := database.DB.Where("company_code = ?", code).First(&existing).Error; err != nil {
			return code
		}
	}
}

func (s *AuthService) sendWelcomeEmail(comp, email, code, user, url string) {
	subject := "Khởi tạo thành công Không gian SOC - SENT System"
	body := fmt.Sprintf("<h2>Chào mừng %s!</h2> Mã Workspace: <b>%s</b>", comp, code)
	_ = utils.SendEmail([]string{email}, subject, body)
}
