package main

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	handlers "SENT_backend/service/approvals/internal/handlers" // Giả định thư mục handlers
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()
	database.InitPostgres()
	cache.InitRedis(
		os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_PASSWORD"),
		0,
	)

	r := gin.Default()

	approvalsAPI := r.Group("/api/v1/approvals")
	approvalsAPI.Use(middleware.AuthRequired())
	{

		approvalsAPI.GET("", middleware.RequirePermission("approval_view"), handlers.GetTickets)
		approvalsAPI.PUT("/:id/review", middleware.RequirePermission("approval_final"), handlers.ReviewTicket)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8013"
	}
	r.Run(":" + port)
}
