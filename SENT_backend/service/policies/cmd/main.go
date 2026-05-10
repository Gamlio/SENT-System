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

	// 2. Định nghĩa API cho Policy Management
	policyAPI := r.Group("/api/v1/policies")
	policyAPI.Use(middleware.AuthRequired()) // Bắt buộc xác thực JWT
	{
		// Quyền xem chính sách
		viewPerm := middleware.RequirePermission("policy_view")
		policyAPI.GET("", viewPerm, handlers.GetPolicies)
		policyAPI.GET("/:id", viewPerm, handlers.GetPolicyDetail)

		// Quyền quản lý và phê duyệt
		managePerm := middleware.RequirePermission("policy_manage")
		policyAPI.POST("", managePerm, handlers.CreatePolicy)
		policyAPI.PUT("/:id", managePerm, handlers.UpdatePolicy)
		policyAPI.DELETE("/:id", managePerm, handlers.DeletePolicy)

		// Luồng phê duyệt (Approval Workflow)
		policyAPI.PUT("/:id/approve", middleware.RequirePermission("approval_final"), handlers.ApprovePolicy)
	}

	// 3. Khởi chạy trên cổng 8006
	port := os.Getenv("PORT")
	if port == "" {
		port = "8006"
	}
	r.Run(":" + port)
}
