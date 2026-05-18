package main

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/policies/internal/handlers"
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

	policyAPI := r.Group("/api/v1/policies")
	{
		policyAPI.POST("/internal/baseline", handlers.InternalSaveBaseline)
		policyAPI.Use(middleware.AuthRequired())
		{
			viewPerm := middleware.RequirePermission("policy_view")
			policyAPI.GET("", viewPerm, handlers.GetPolicies)
			policyAPI.GET("/:id", viewPerm, handlers.GetPolicyDetail)

			managePerm := middleware.RequirePermission("policy_manage")
			policyAPI.POST("", managePerm, handlers.CreatePolicy)
			policyAPI.PUT("/:id", managePerm, handlers.UpdatePolicy)
			policyAPI.POST("/bulk-delete", managePerm, handlers.BulkDeletePolicy)
			policyAPI.DELETE("/:id", managePerm, handlers.DeletePolicy)

			policyAPI.PUT("/:id/approve", middleware.RequirePermission("approval_final"), handlers.ApprovePolicy)
			whitelist := policyAPI.Group("/whitelist")
			{
				whitelist.GET("/:hwid", viewPerm, handlers.GetAssetWhitelist)
				whitelist.DELETE("/item/:item_id", managePerm, handlers.RemoveWhitelistItem)
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
