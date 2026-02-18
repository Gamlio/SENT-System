package main

import (
	"os"
	"strings"
	"time"

	v1 "sent_backend/internal/api/v1"
	"sent_backend/internal/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load biến môi trường TRƯỚC TIÊN
	// Nếu không load được thì cũng không sao (có thể chạy bằng biến hệ thống)
	_ = godotenv.Load()

	// 2. Khởi tạo DB (Trong db.go bạn nhớ dùng os.Getenv("DATABASE_URL"))
	database.InitDB()

	// 3. Cấu hình chế độ Gin (Debug/Release)
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	r := gin.Default()

	// 4. Xử lý CORS từ biến môi trường
	originsEnv := os.Getenv("ALLOWED_ORIGINS")
	var allowOrigins []string
	if originsEnv != "" {
		allowOrigins = strings.Split(originsEnv, ",")
	}

	// Cấu hình CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins, // Dùng danh sách từ .env
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	v1Group := r.Group("/api/v1")
	{
		auth := v1Group.Group("/auth")
		{
			auth.POST("/login", v1.LoginHandler)
			auth.POST("/register", v1.RegisterSMEHandler)
		}

		agents := v1Group.Group("/agents")
		{
			agents.POST("/push", v1.PushDataHandler)
			agents.GET("/stats", v1.GetStats)
			agents.GET("", v1.GetAgents)
			agents.GET("/:hwid", v1.GetAgentDetail)
		}
	}

	// 5. Chạy Server theo PORT trong .env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Mặc định nếu thiếu
	}
	r.Run(":" + port)
}
