package v1

import (
	"net/http"
	"sent_backend/internal/auth"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterHandler: Đăng ký SME mới và tạo Admin (Level 2)
func RegisterHandler(c *gin.Context) {
	var req struct {
		CompanyName string `json:"company_name"`
		Username    string `json:"username"`
		Password    string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// 1. Tạo Organization
	org := models.Organization{Name: req.CompanyName}
	database.DB.Create(&org)

	// 2. Băm mật khẩu và tạo User Level 2
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	user := models.User{
		Username:       req.Username,
		HashedPassword: string(hashed),
		Level:          2, // SME Admin
		OrgID:          &org.ID,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo tài khoản"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đăng ký SME thành công"})
}

// LoginHandler: Xác thực và trả về JWT
func LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	c.ShouldBindJSON(&req)

	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai tài khoản hoặc mật khẩu"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.HashedPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai tài khoản hoặc mật khẩu"})
		return
	}

	token, _ := auth.GenerateToken(user.Username)
	c.JSON(http.StatusOK, gin.H{
		"token":  token,
		"level":  user.Level,
		"org_id": user.OrgID,
	})
}
