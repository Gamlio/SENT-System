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
	{
		groupsAPI.GET("", handlers.HandleGetGroups)
		groupsAPI.GET("/:id", handlers.HandleGetGroupDetail)

		adminAPI := groupsAPI.Group("")
		adminAPI.Use(middleware.RequirePermission("group_manage"))
		{
			adminAPI.POST("", handlers.HandleCreateGroup)
			adminAPI.PUT("/:id", handlers.HandleUpdateGroup)
			adminAPI.DELETE("/:id", handlers.HandleDeleteGroup)
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
