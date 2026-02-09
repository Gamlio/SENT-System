package main

import (
	v1 "sent_backend/internal/api/v1" // Import các handler
	"sent_backend/internal/database"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	database.InitDB() // AutoMigrate 18 bảng

	r := gin.Default()

	// CORS cho React Frontend
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		MaxAge:       12 * time.Hour,
	}))

	v1Group := r.Group("/api/v1")
	{
		// Nhóm xác thực
		auth := v1Group.Group("/auth")
		{
			auth.POST("/login", v1.LoginHandler)
			auth.POST("/register-sme", v1.RegisterSMEHandler)
		}

		// Nhóm tiếp nhận dữ liệu từ Agent
		agents := v1Group.Group("/agents")
		{
			agents.POST("/push", v1.PushDataHandler) // Điểm tiếp nhận chính
		}
	}

	r.Run(":8000")
}
