package main

import (
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
	database.Migrate()
	database.InitMongoDB()
	database.InitRedis()

	r := gin.Default()

	assetsAPI := r.Group("/api/v1/assets")
	{
		assetsAPI.POST("/enroll", middleware.AssetEnrollRateLimitMiddleware(), handlers.EnrollAsset)

		assetsAPI.POST("/internal/cleanup", handlers.InternalCleanupHandler)

		agentProtected := assetsAPI.Group("")
		agentProtected.Use(middleware.AssetFloodProtectionMiddleware())
		{
			agentProtected.POST("/push", middleware.AssetHMACAuth(), handlers.PushDataHandler)
		}

		admin := assetsAPI.Group("")
		admin.Use(middleware.AuthRequired())
		{
			// Static routes first
			admin.GET("/stats", handlers.GetStats)
			admin.GET("/types", handlers.GetAssetTypes)
			admin.GET("/active-token", handlers.GetActiveEnrollmentToken)
			admin.PUT("/types/:id", middleware.RequirePermission("system_config"), handlers.UpdateAssetType)
			admin.POST("/types", middleware.RequirePermission("system_config"), handlers.CreateAssetType)

			// Asset list and general queries
			admin.GET("", middleware.RequirePermission("asset_view"), handlers.GetAssets)
			admin.POST("/bulk-delete", middleware.RequirePermission("asset_delete"), handlers.RequestBulkDeleteAssets)
			admin.POST("/bulk-assign", middleware.RequirePermission("asset_move"), handlers.BulkAssignManager)

			// Asset by HWID routes
			admin.GET("/:hwid/software", handlers.GetAssetSoftware)
			admin.GET("/:hwid/usb", handlers.GetAssetUSB)
			admin.GET("/:hwid/ports", handlers.GetAssetPorts)
			admin.GET("/:hwid/logs", handlers.GetAssetLogs)
			admin.GET("/:hwid/type", middleware.RequirePermission("asset_view"), handlers.GetAssetType)

			admin.GET("/:hwid", middleware.RequirePermission("asset_view"), handlers.GetAssetDetail)
			admin.PUT("/:hwid/assign", middleware.RequirePermission("asset_move"), handlers.AssignManager)
			admin.PUT("/:hwid/department", middleware.RequirePermission("asset_move"), handlers.UpdateDepartment)
			admin.PUT("/:hwid/type", middleware.RequirePermission("asset_move"), handlers.UpdateDeviceType)
			admin.POST("/:hwid/request-delete", middleware.RequirePermission("asset_delete"), handlers.RequestDeleteAsset)
			admin.DELETE("/:hwid", middleware.RequirePermission("asset_delete"), handlers.RequestDeleteAsset)
			admin.PUT("/:hwid/group", middleware.RequirePermission("asset_move"), handlers.UpdateAssetGroup)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
