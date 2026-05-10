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
	// 1. Khởi tạo môi trường và hạ tầng
	_ = godotenv.Load()
	database.InitPostgres() // Quản lý SentSoftware (Postgres)
	database.InitMongoDB()  // Quản lý SoftwareItem telemetry (MongoDB)

	r := gin.Default()

	// 2. Định nghĩa API cho Software & Software Management
	softwareAPI := r.Group("/api/v1/software")
	softwareAPI.Use(middleware.AuthRequired())
	{
		// Quản lý phiên bản Agent (Trang Softwares cũ)
		softwareAPI.GET("/softwares", handlers.GetSoftwares)
		softwareAPI.POST("/softwares", middleware.RequirePermission("system_config"), handlers.CreateSoftware)
		softwareAPI.GET("/softwares/latest", handlers.GetLatestSoftware) // Agent gọi để tự động update

		// Quản lý kho phần mềm của thiết bị (Software Inventory)
		softwareAPI.GET("/inventory/:hwid", middleware.RequirePermission("asset_view"), handlers.GetSoftwareInventory)
	}

	// 3. Khởi chạy trên cổng 8012
	port := os.Getenv("PORT")
	if port == "" {
		port = "8012"
	}
	r.Run(":" + port)
}
