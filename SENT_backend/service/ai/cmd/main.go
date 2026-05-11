package main

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/ai/internal/handlers"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	fmt.Println("Starting AI Service...")
	database.InitPostgres()
	database.InitMongoDB()
	database.InitRedis()

	redisAddr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	if os.Getenv("REDIS_HOST") == "" {
		redisAddr = "redis:6379"
	}

	err := cache.InitRedis(redisAddr, os.Getenv("REDIS_PASSWORD"), 0)
	if err != nil {
		fmt.Printf("❌ Lỗi kết nối Redis: %v\n", err)
	} else {
		fmt.Println("✅ Đã kết nối Redis cho Cache")
	}
	r := gin.Default()

	aiAPI := r.Group("/api/v1/ai")
	aiAPI.Use(middleware.AuthRequired())
	{
		aiAPI.POST("/chat/:session_id", handlers.ChatHandler)

		aiAPI.PUT("/sessions/:id", handlers.RenameSession)
		aiAPI.DELETE("/sessions/:id", handlers.DeleteSession)
		aiAPI.GET("/chat/:session_id", handlers.GetChatHistory)
		aiAPI.GET("/sessions", handlers.GetSessions)
		aiAPI.POST("/sessions", handlers.CreateSession)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
