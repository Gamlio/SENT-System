package main

import (
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/scoring/internal/handlers"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Khởi tạo kết nối DB và MongoDB cho Scoring Service
	database.InitPostgres()
	database.InitMongoDB()

	r := gin.Default()

	scoringAPI := r.Group("/api/v1/scoring")
	{
		// 1. ENDPOINT DÀNH CHO CÁC SERVICE KHÁC (Internal API)
		// Không sử dụng AuthRequired nếu gọi trong mạng nội bộ Docker hoặc sử dụng API Key riêng
		scoringAPI.POST("/recalculate/:hwid", handlers.HandleRecalculate)

		// 2. ENDPOINTS DÀNH CHO DASHBOARD (Yêu cầu quyền hạn)
		protected := scoringAPI.Group("")
		protected.Use(middleware.AuthRequired())
		{
			protected.POST("/recalculate/all", middleware.RequirePermission("system_config"), handlers.RecalculateAll)
			protected.POST("/recalculate/type/:type_id", middleware.RequirePermission("system_config"), handlers.RecalculateByType)
			protected.GET("/history/:hwid", middleware.RequirePermission("asset_view"), handlers.GetScoreHistory)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8010" // Cổng mặc định cho Scoring Service
	}
	r.Run(":" + port)
}
