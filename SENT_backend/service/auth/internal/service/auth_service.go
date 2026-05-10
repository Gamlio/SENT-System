package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"SENT_backend/pkg/utils"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

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

	token, err := GenerateToken(user.Username, *user.OrgID, user.ID)
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
func (s *AuthService) ForgotPassword(companyCode, email string) error {
	var org models.Organization
	if err := database.DB.Where("company_code = ?", companyCode).First(&org).Error; err != nil {
		return fmt.Errorf("không gian làm việc không tồn tại")
	}

	var user models.User
	if err := database.DB.Where("email = ? AND org_id = ?", email, org.ID).First(&user).Error; err != nil {
		return fmt.Errorf("không tìm thấy tài khoản với email này trong tổ chức")
	}

	// 1. Tạo token ngẫu nhiên
	b := make([]byte, 20)
	rand.Read(b)
	token := hex.EncodeToString(b)
	expiry := time.Now().Add(1 * time.Hour) // Hết hạn sau 1 giờ

	// 2. Lưu vào DB
	user.ResetToken = token
	user.ResetTokenExpiry = &expiry
	if err := database.DB.Save(&user).Error; err != nil {
		return err
	}

	// 3. Gửi Email
	resetURL := fmt.Sprintf("%s/reset-password/%s", os.Getenv("APP_URL"), token)
	subject := "Yêu cầu khôi phục mật khẩu - SENT SOC"
	body := fmt.Sprintf(`
        <h3>Yêu cầu khôi phục mật khẩu</h3>
        <p>Bạn đã yêu cầu đặt lại mật khẩu cho tài khoản <b>%s</b> tại workspace <b>%s</b>.</p>
        <p>Vui lòng nhấn vào link bên dưới để tạo mật khẩu mới (có hiệu lực trong 60 phút):</p>
        <a href="%s" style="padding:10px 20px; background:#4f46e5; color:white; text-decoration:none; border-radius:5px;">Đặt lại mật khẩu</a>
    `, user.Username, companyCode, resetURL)

	return utils.SendEmail([]string{email}, subject, body)
}

// ResetPassword: Xác thực token và cập nhật mật khẩu mới
func (s *AuthService) ResetPassword(token, newPassword string) error {
	var user models.User
	now := time.Now()

	// Kiểm tra token và thời hạn
	err := database.DB.Where("reset_token = ? AND reset_token_expiry > ?", token, now).First(&user).Error
	if err != nil {
		return fmt.Errorf("mã khôi phục không hợp lệ hoặc đã hết hạn")
	}

	// Hash mật khẩu mới
	hashed, _ := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	user.PasswordHash = string(hashed)

	// Xóa token sau khi dùng xong
	user.ResetToken = ""
	user.ResetTokenExpiry = nil

	return database.DB.Save(&user).Error
}
