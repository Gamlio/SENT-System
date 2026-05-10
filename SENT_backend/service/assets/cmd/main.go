package main

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	handlers "SENT_backend/service/assets/internal/handlers"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()
	database.InitPostgres()
	database.InitMongoDB()
	database.InitRedis()
	cache.InitRedis(
		os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_PASSWORD"),
		0,
	)

	r := gin.Default()
	agentAPI := r.Group("/api/v1/assets")
	agentAPI.Use(middleware.AssetFloodProtectionMiddleware())
	{
		agentAPI.POST("/enroll", handlers.EnrollAsset)

		agentAPI.POST("/push", middleware.AssetHMACAuth(), handlers.PushDataHandler)
	}

	adminAPI := r.Group("/api/v1/assets")
	adminAPI.Use(middleware.AuthRequired())
	{
		adminAPI.GET("/stats", handlers.GetStats)
		adminAPI.GET("", middleware.RequirePermission("asset_view"), handlers.GetAssets)
		adminAPI.GET("/:hwid", middleware.RequirePermission("asset_view"), handlers.GetAssetDetail)

		adminAPI.GET("/:hwid/software", handlers.GetAssetSoftware)
		adminAPI.GET("/:hwid/usb", handlers.GetAssetUSB)
		adminAPI.GET("/:hwid/ports", handlers.GetAssetPorts)

		adminAPI.GET("/types", handlers.GetAssetTypes)
		adminAPI.PUT("/types/:id", middleware.RequirePermission("system_config"), handlers.UpdateAssetType)
		adminAPI.PUT("/:hwid/assign", middleware.RequirePermission("asset_move"), handlers.AssignManager)
		adminAPI.PUT("/:hwid/type", middleware.RequirePermission("asset_move"), handlers.UpdateDeviceType)

		adminAPI.DELETE("/:hwid", middleware.RequirePermission("asset_delete"), handlers.RequestDeleteAsset)
		adminAPI.POST("/bulk-delete", middleware.RequirePermission("asset_delete"), handlers.RequestBulkDeleteAssets)

		adminAPI.GET("/active-token", handlers.GetActiveEnrollmentToken)
	}

	// 4. Chạy Service
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}
	r.Run(":" + port)
}
