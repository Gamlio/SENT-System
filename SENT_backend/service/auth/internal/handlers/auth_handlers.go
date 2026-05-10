package handlers

import (
	"SENT_backend/service/auth/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterSMEHandler(c *gin.Context) {
	var req struct {
		CompanyName string `json:"company_name" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := service.AuthService{}
	code, url, err := svc.RegisterSME(req.CompanyName, req.Email, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"company_code": code, "login_url": url})
}

func LoginHandler(c *gin.Context) {
	var req struct {
		CompanyCode string `json:"company_code" binding:"required"`
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := service.AuthService{}
	token, user, org, err := svc.Login(req.CompanyCode, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Trả về ma trận quyền đồng bộ với DB[cite: 45]
	c.JSON(http.StatusOK, gin.H{
		"token":        token,
		"username":     user.Username,
		"company_code": org.CompanyCode,
		"permissions": map[string]bool{
			"asset_view":      user.PermAssetView,
			"asset_delete":    user.PermAssetDelete,
			"asset_move":      user.PermAssetMove,
			"incident_action": user.PermIncidentAction,
			"policy_view":     user.PermPolicyView,
			"doc_view":        user.PermDocView,
			"user_view":       user.PermUserView,
			"incident_view":   user.PermIncidentView,
			"approval_view":   user.PermApprovalView,
			"asset_action":    user.PermAssetAction,
			"group_manage":    user.PermGroupManage,
			"policy_manage":   user.PermPolicyManage,
			"user_manage":     user.PermUserManage,
			"doc_manage":      user.PermDocManage,
			"approval_final":  user.PermApprovalFinal,
			"system_config":   user.PermSystemConfig,
		},
	})
}
func ForgotPasswordHandler(c *gin.Context) {
	var req struct {
		CompanyCode string `json:"company_code" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	svc := service.AuthService{}
	if err := svc.ForgotPassword(req.CompanyCode, req.Email); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Link khôi phục đã được gửi tới Email của bạn"})
}

func ResetPasswordHandler(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Mật khẩu tối thiểu 8 ký tự"})
		return
	}

	svc := service.AuthService{}
	if err := svc.ResetPassword(req.Token, req.NewPassword); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Đổi mật khẩu thành công. Vui lòng đăng nhập lại"})
}
