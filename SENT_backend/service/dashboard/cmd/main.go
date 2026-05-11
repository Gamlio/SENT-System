package main

import (
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/dashboard/internal/handlers"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	database.InitPostgres()
	database.InitMongoDB()
	database.InitRedis()

	r := gin.Default()

	dashAPI := r.Group("/api/v1/dashboard")
	dashAPI.Use(middleware.AuthRequired())
	{
		dashAPI.GET("/stats", handlers.GetDashboardStats)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
