package main

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/auth/internal/handlers"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	database.InitPostgres()
	database.Migrate()
	cache.InitRedis(
		os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_PASSWORD"),
		0,
	)

	r := gin.Default()

	// 2. Định nghĩa Routes cho Auth Service
	authAPI := r.Group("/api/v1/auth")
	{
		// Đăng ký tổ chức mới (SME)
		authAPI.POST("/register", handlers.RegisterSMEHandler)

		// Đăng nhập hệ thống
		authAPI.POST("/login", handlers.LoginHandler)

		// Quên mật khẩu & Reset
		authAPI.POST("/forgot-password", handlers.ForgotPasswordHandler)
		authAPI.POST("/reset-password", handlers.ResetPasswordHandler)
	}

	// 3. Chạy Service trên cổng 8001
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}
	r.Run(":" + port)
}
