package main

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/groups/internal/handlers"
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

	groupsAPI := r.Group("/api/v1/groups")
	groupsAPI.Use(middleware.AuthRequired())
	groupsAPI.Use(middleware.RequirePermission("group_manage"))
	{
		groupsAPI.GET("", handlers.HandleGetGroups)
		groupsAPI.POST("", handlers.HandleCreateGroup)
		groupsAPI.GET("/:id", handlers.HandleGetGroupDetail)
		groupsAPI.PUT("/:id", handlers.HandleUpdateGroup)
		groupsAPI.DELETE("/:id", handlers.HandleDeleteGroup)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
