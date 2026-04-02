package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sent_backend/internal/auth"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/utils" // Import thư mục tiện ích vừa tạo
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Hàm sinh mã công ty ngẫu nhiên (VD: SME-A1B2C3)
func generateCompanyCode() string {
	b := make([]byte, 3)
	rand.Read(b)
	return "SME-" + strings.ToUpper(hex.EncodeToString(b))
}

func RegisterSMEHandler(c *gin.Context) {
	var req struct {
		CompanyName string `json:"company_name" binding:"required,min=3,max=100"`
		Email       string `json:"email" binding:"required,email"` // [MỚI] Bắt buộc có Email để gửi link
		Username    string `json:"username" binding:"required,min=3,max=50,alphanum"`
		Password    string `json:"password" binding:"required,min=8,max=128"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu sai định dạng: " + err.Error()})
		return
	}

	newCompanyCode := generateCompanyCode()
	var existingOrg models.Organization
	for {
		if err := database.DB.Where("company_code = ?", newCompanyCode).First(&existingOrg).Error; err != nil {
			break
		}
		newCompanyCode = generateCompanyCode()
	}

	// 1. Tạo Tổ chức
	org := models.Organization{
		Name:              req.CompanyName,
		CompanyCode:       newCompanyCode,
		EnrollTokenPrefix: "SENT-" + newCompanyCode,
	}
	database.DB.Create(&org)

	// 2. Tạo User Admin (Bây giờ Username chỉ cần duy nhất trong nội bộ Công ty)
	var existingUser models.User
	if err := database.DB.Where("username = ? AND org_id = ?", req.Username, org.ID).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tên đăng nhập này đã có người sử dụng. Vui lòng chọn tên khác!"})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	user := models.User{
		Username:           req.Username,
		PasswordHash:       string(hashed),
		Email:              req.Email,
		OrgID:              &org.ID,
		PermAgentView:      true,
		PermAgentAction:    true,
		PermAgentDelete:    true,
		PermPolicyView:     true,
		PermPolicyAction:   true,
		PermIncidentView:   true,
		PermIncidentAction: true,
		PermDocView:        true,
		PermDocManage:      true,
		PermUserManage:     true,
		PermApprovalManage: true,
		ApprovalStatus:     "APPROVED",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi tạo tài khoản"})
		return
	}

	// 3. Gửi Email thông báo (Dùng goroutine chạy ngầm để API phản hồi nhanh)
	loginURL := fmt.Sprintf("http://localhost:3000/login/%s", newCompanyCode)
	go func() {
		subject := "Khởi tạo thành công Không gian SOC - SENT System"
		body := fmt.Sprintf(`
			<div style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
				<h2 style="color: #10b981;">Chào mừng %s đến với SENT System!</h2>
				<p>Không gian làm việc bảo mật của công ty bạn đã được khởi tạo thành công.</p>
				<div style="background-color: #f8fafc; padding: 15px; border-radius: 8px; border-left: 4px solid #10b981; margin: 20px 0;">
					<ul style="list-style-type: none; padding: 0; margin: 0;">
						<li style="margin-bottom: 10px;"><b>🏢 Mã Công ty (Workspace):</b> <span style="font-family: monospace; font-size: 16px; font-weight: bold; color: #0f172a;">%s</span></li>
						<li><b>👤 Tài khoản Admin:</b> %s</li>
					</ul>
				</div>
				<p>Vui lòng đăng nhập qua đường dẫn an toàn dành riêng cho tổ chức của bạn:</p>
				<a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #10b981; color: white; text-decoration: none; border-radius: 6px; font-weight: bold; margin-top: 10px;">Truy cập Bảng điều khiển</a>
			</div>
		`, req.CompanyName, newCompanyCode, req.Username, loginURL)

		_ = utils.SendEmail([]string{req.Email}, subject, body)
	}()

	// 4. Phản hồi cho React
	c.JSON(http.StatusOK, gin.H{
		"message":      "Đăng ký thành công",
		"company_code": newCompanyCode,
		"login_url":    loginURL, // Trả link để giao diện Web tự động chuyển hướng
	})
}

func LoginHandler(c *gin.Context) {
	var req struct {
		CompanyCode string `json:"company_code" binding:"required,min=7,max=20,alphanum-"` // [BẢO MẬT] BẮT BUỘC PHẢI CÓ
		Username    string `json:"username" binding:"required,min=3,max=50"`
		Password    string `json:"password" binding:"required,min=1,max=128"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	if req.CompanyCode == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Yêu cầu mã Workspace (Company Code) để đăng nhập!"})
		return
	}

	// 1. CHỐT CHẶN BẢO MẬT 1: Tìm công ty trước
	var org models.Organization
	if err := database.DB.Where("company_code = ?", req.CompanyCode).First(&org).Error; err != nil {
		// Dùng thông báo chung chung để chống dò quét
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Thông tin đăng nhập không chính xác hoặc không gian làm việc không tồn tại!"})
		return
	}

	// 2. CHỐT CHẶN BẢO MẬT 2: Tìm User TRONG NỘI BỘ công ty đó
	var user models.User
	if err := database.DB.Where("username = ? AND org_id = ?", req.Username, org.ID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Thông tin đăng nhập không chính xác hoặc không gian làm việc không tồn tại!"})
		return
	}

	// 3. Kiểm tra mật khẩu
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Thông tin đăng nhập không chính xác hoặc không gian làm việc không tồn tại!"})
		return
	}

	// 4. Cập nhật thành công, cấp Token
	token, _ := auth.GenerateToken(user.Username, *user.OrgID)

	c.JSON(http.StatusOK, gin.H{
		"token":        token,
		"username":     user.Username,
		"company_code": org.CompanyCode,
		"permissions": map[string]bool{
			"agent_view":      user.PermAgentView,
			"agent_action":    user.PermAgentAction,
			"agent_delete":    user.PermAgentDelete,
			"policy_view":     user.PermPolicyView,
			"policy_action":   user.PermPolicyAction,
			"incident_view":   user.PermIncidentView,
			"incident_action": user.PermIncidentAction,
			"doc_view":        user.PermDocView,
			"doc_manage":      user.PermDocManage,
			"user_manage":     user.PermUserManage,
			"approval_manage": user.PermApprovalManage,
		},
	})
}
