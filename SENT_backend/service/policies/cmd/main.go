package main

import (
	"log"
	"os"

	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/policies/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	database.InitPostgres()
	database.InitMongoDB()

	if err := cache.InitRedis(
		os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_PASSWORD"),
		0,
	); err != nil {
		log.Fatalf("failed to init redis: %v", err)
	}

	r := gin.Default()

	policyAPI := r.Group("/api/v1/policies")
	{
		// 1. CỔNG DÀNH RIÊNG CHO AGENT (KIỂM TRA CHỮ KÝ HMAC - BẢO MẬT TUYỆT ĐỐI)
		policyAPI.POST("/internal/baseline", handlers.InternalSaveBaseline)
		policyAPI.POST("/internal/check-violation", handlers.CheckPolicyViolation)

		// Tuyến đường xử lý kéo luật thời gian thực cho Agent (Sửa lỗi dứt điểm HTTP 404)
		policyAPI.POST("/download", middleware.AssetHMACAuth(), handlers.DownloadAgentPolicies)

		// 2. CỔNG QUẢN TRỊ DÀNH CHO WEB ADMIN (XÁC THỰC SESSION / JWT TOKEN)
		adminAPI := policyAPI.Group("")
		adminAPI.Use(middleware.AuthRequired())
		{
			viewPerm := middleware.RequirePermission("policy_view")
			adminAPI.GET("", viewPerm, handlers.GetPolicies)
			adminAPI.GET("/:id", viewPerm, handlers.GetPolicyDetail)

			managePerm := middleware.RequirePermission("policy_manage")
			adminAPI.POST("", managePerm, handlers.CreatePolicy)
			adminAPI.PUT("/:id", managePerm, handlers.UpdatePolicy)
			adminAPI.POST("/bulk-delete", managePerm, handlers.BulkDeletePolicy)
			adminAPI.POST("/bulk-approve", middleware.RequirePermission("approval_final"), handlers.BulkApprovePolicy)
			adminAPI.POST("/approve-by-asset", middleware.RequirePermission("approval_final"), handlers.ApproveAssetBaseline)
			adminAPI.DELETE("/:id", managePerm, handlers.DeletePolicy)

			adminAPI.PUT("/:id/approve", middleware.RequirePermission("approval_final"), handlers.ApprovePolicy)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
