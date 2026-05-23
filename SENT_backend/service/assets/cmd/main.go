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

	assetsAPI := r.Group("/api/v1/assets")
	{
		assetsAPI.POST("/enroll", middleware.AssetEnrollRateLimitMiddleware(), handlers.EnrollAsset)

		agentProtected := assetsAPI.Group("")
		agentProtected.Use(middleware.AssetFloodProtectionMiddleware())
		{
			agentProtected.POST("/push", middleware.AssetHMACAuth(), handlers.PushDataHandler)
		}

		admin := assetsAPI.Group("")
		admin.Use(middleware.AuthRequired())
		{
			admin.GET("/stats", handlers.GetStats)
			admin.GET("/types", handlers.GetAssetTypes)
			admin.GET("/active-token", handlers.GetActiveEnrollmentToken)
			admin.GET("", middleware.RequirePermission("asset_view"), handlers.GetAssets)
			admin.POST("/bulk-delete", middleware.RequirePermission("asset_delete"), handlers.RequestBulkDeleteAssets)

			admin.GET("/:hwid/software", handlers.GetAssetSoftware)
			admin.GET("/:hwid/usb", handlers.GetAssetUSB)
			admin.GET("/:hwid/ports", handlers.GetAssetPorts)
			admin.GET("/:hwid/logs", handlers.GetAssetLogs)

			admin.GET("/:hwid", middleware.RequirePermission("asset_view"), handlers.GetAssetDetail)
			admin.PUT("/:hwid/assign", middleware.RequirePermission("asset_move"), handlers.AssignManager)
			admin.PUT("/:hwid/type", middleware.RequirePermission("asset_move"), handlers.UpdateDeviceType)
			admin.DELETE("/:hwid", middleware.RequirePermission("asset_delete"), handlers.RequestDeleteAsset)

			admin.PUT("/types/:id", middleware.RequirePermission("system_config"), handlers.UpdateAssetType)
			admin.POST("/types", middleware.RequirePermission("system_config"), handlers.CreateAssetType)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
