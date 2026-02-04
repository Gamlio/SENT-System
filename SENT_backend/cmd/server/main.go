package main

import (
	"net/http"
	"sent_backend/internal/auth"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	godotenv.Load()
	database.InitDB() // Khởi tạo DB và AutoMigrate

	r := gin.Default()

	// 1. SỬA LỖI CORS: Cho phép React (3000) gọi API (8000)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Nhánh API v1
	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", LoginHandler)
			authGroup.POST("/register", RegisterHandler) // Đăng ký SME
		}

		assets := v1.Group("/assets")
		{
			// Sau này thêm Middleware AuthRequired ở đây để bảo mật
			assets.GET("/", GetAssetsHandler)
		}
	}

	r.Run(":8000")
}

// --- LOGIC HANDLERS ---

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

	// 1. Tạo Organization mới cho SME
	newOrg := models.Organization{Name: req.CompanyName}
	if err := database.DB.Create(&newOrg).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tên công ty đã tồn tại"})
		return
	}

	// 2. Băm mật khẩu (Bcrypt) - Cực kỳ quan trọng cho Security Foxconn
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)

	// 3. Tạo User Admin cho SME đó (Level 3)
	newUser := models.User{
		Username:       req.Username,
		HashedPassword: string(hashedPassword),
		Level:          3,
		OrgID:          &newOrg.ID,
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo tài khoản Admin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đăng ký SME thành công!"})
}

func LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin đăng nhập"})
		return
	}

	// 1. Tìm User trong DB
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai tài khoản hoặc mật khẩu"})
		return
	}

	// 2. Kiểm tra mật khẩu Bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sai tài khoản hoặc mật khẩu"})
		return
	}

	// 3. Tạo JWT Token
	token, err := auth.GenerateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi tạo Token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":  token,
		"level":  user.Level,
		"org_id": user.OrgID,
	})
}

func GetAssetsHandler(c *gin.Context) {
	// Tạm thời lấy tất cả, sau này sẽ dùng Middleware để lấy org_id từ Token
	var assets []models.Agent
	database.DB.Find(&assets)
	c.JSON(http.StatusOK, assets)
}
