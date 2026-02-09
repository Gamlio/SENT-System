package v1

import (
	"net/http"
	"sent_backend/internal/auth"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func RegisterSMEHandler(c *gin.Context) {
	var req struct {
		CompanyName string `json:"company_name"`
		Username    string `json:"username"`
		Password    string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu sai định dạng"})
		return
	}

	// 1. Tạo Org
	org := models.Organization{
		Name:              req.CompanyName,
		EnrollTokenPrefix: "SENT-" + req.CompanyName,
	}
	database.DB.Create(&org)

	// 2. Tạo Admin Level 3 cho SME
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	user := models.User{
		Username:     req.Username,
		PasswordHash: string(hashed), // Đã sửa từ HashedPassword
		RoleLevel:    3,              // SME Admin
		OrgID:        &org.ID,
	}
	database.DB.Create(&user)

	c.JSON(http.StatusOK, gin.H{"message": "Đăng ký công ty và tài khoản quản trị thành công"})
}

func LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	c.ShouldBindJSON(&req)

	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại"})
		return
	}

	// Kiểm tra PasswordHash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mật khẩu không chính xác"})
		return
	}

	token, _ := auth.GenerateToken(user.Username)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"role":  user.RoleLevel,
		"org":   user.OrgID,
	})
}
