package main

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/users/internal/handlers"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Khởi tạo môi trường và hạ tầng
	_ = godotenv.Load()
	database.InitPostgres()

	cache.InitRedis(
		os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_PASSWORD"),
		0,
	)

	r := gin.Default()

	// 2. Định nghĩa API cho User Management
	usersAPI := r.Group("/api/v1/users")
	usersAPI.Use(middleware.AuthRequired())
	{
		// Quyền xem danh sách nhân sự
		viewPerm := middleware.RequirePermission("user_view")
		usersAPI.GET("", viewPerm, handlers.GetUsers)
		usersAPI.GET("/:id", viewPerm, handlers.GetUserDetail)

		// Quyền quản trị nhân sự (Thêm/Sửa/Xóa/Phân quyền)
		managePerm := middleware.RequirePermission("user_manage")
		usersAPI.POST("", managePerm, handlers.CreateUser)
		usersAPI.PUT("/:id", managePerm, handlers.UpdateUser)
		usersAPI.DELETE("/:id", managePerm, handlers.DeleteUser)
		usersAPI.PUT("/:id/permissions", managePerm, handlers.UpdatePermissions)
	}

	// 3. Khởi chạy trên cổng 8011
	port := os.Getenv("PORT")
	if port == "" {
		port = "8011"
	}
	r.Run(":" + port)
}
