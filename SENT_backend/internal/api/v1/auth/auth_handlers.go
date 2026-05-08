package auth

import (
	"net/http"
	"sent_backend/internal/service/auth"

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

	svc := auth.AuthService{}
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

	svc := auth.AuthService{}
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
