package main

import (
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/ai/internal/handlers"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	database.InitPostgres()
	database.InitMongoDB()

	r := gin.Default()

	aiAPI := r.Group("/api/v1/ai")
	aiAPI.Use(middleware.AuthRequired())
	{
		aiAPI.POST("/chat", handlers.ChatHandler)
		aiAPI.PUT("/sessions/:id", handlers.RenameSession)
		aiAPI.DELETE("/sessions/:id", handlers.DeleteSession)
		aiAPI.GET("/chat/:session_id", handlers.GetChatHistory)
		aiAPI.GET("/sessions", handlers.GetSessions)
		aiAPI.POST("/sessions", handlers.CreateSession)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8008"
	}
	r.Run(":" + port)
}
