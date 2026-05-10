package main

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/incidents/internal/handlers"
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

	// 2. Định nghĩa API cho Incident Service
	incidentAPI := r.Group("/api/v1/incidents")
	incidentAPI.Use(middleware.AuthRequired()) // Yêu cầu đăng nhập
	{
		// Quyền xem sự cố
		viewPerm := middleware.RequirePermission("incident_view")
		incidentAPI.GET("", viewPerm, handlers.GetIncidents)
		incidentAPI.GET("/:id", viewPerm, handlers.GetIncidentDetail)
		incidentAPI.GET("/:id/activities", viewPerm, handlers.GetIncidentActivities)

		// Quyền xử lý sự cố
		actionPerm := middleware.RequirePermission("incident_action")
		incidentAPI.POST("", actionPerm, handlers.CreateIncident)
		incidentAPI.PUT("/:id/status", actionPerm, handlers.UpdateIncidentStatus)
		incidentAPI.POST("/:id/comments", actionPerm, handlers.AddIncidentComment)
		incidentAPI.PUT("/:id/assign", actionPerm, handlers.AssignIncident)
	}

	// 3. Khởi chạy trên cổng 8005
	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}
	r.Run(":" + port)
}
