package main

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/behavior/internal/handlers"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()
	database.InitPostgres()
	database.InitMongoDB()
	cache.InitRedis(
		os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_PASSWORD"),
		0,
	)

	r := gin.Default()

	behaviorAPI := r.Group("/api/v1/behaviors")
	{
		// Route này cho phép gọi nội bộ từ asset-service, không cần AuthRequired
		behaviorAPI.POST("/log", handlers.HandleLogBehavior)

		// Các route dành cho giao diện người dùng mới cần bảo vệ
		protected := behaviorAPI.Group("")
		protected.Use(middleware.AuthRequired())
		{
			protected.GET("", middleware.RequirePermission("incident_view"), handlers.GetBehaviors)
			protected.GET("/:id", middleware.RequirePermission("incident_view"), handlers.GetBehaviorDetail)
			protected.POST("/create-incident", middleware.RequirePermission("incident_action"), handlers.CreateIncidentHandler)
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
