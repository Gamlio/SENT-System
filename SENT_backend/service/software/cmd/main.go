package main

import (
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/software/internal/handlers"
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

	r.POST("/api/v1/softwares/update-version", handlers.GitHubWebhookHandler)

	softwareAPI := r.Group("/api/v1")
	softwareAPI.Use(middleware.AuthRequired())
	{
		softwareAPI.GET("/softwares", handlers.GetSoftwares)
		softwareAPI.POST("/softwares", middleware.RequirePermission("system_config"), handlers.CreateSoftware)
		softwareAPI.GET("/softwares/latest", handlers.GetLatestSoftware)
		softwareAPI.GET("/inventory/:hwid", middleware.RequirePermission("asset_view"), handlers.GetSoftwareInventory)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
