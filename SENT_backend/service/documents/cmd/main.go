package main

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/middleware"
	"SENT_backend/pkg/models/database"
	"SENT_backend/service/documents/internal/handlers"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()
	database.InitPostgres()
	database.InitRedis()
	database.InitMongoDB()
	cache.InitRedis(
		os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_PASSWORD"),
		0,
	)
	r := gin.Default()
	r.Static("/uploads", "./uploads")
	docAPI := r.Group("/api/v1/docs")
	docAPI.Use(middleware.AuthRequired())
	{
		docAPI.GET("", middleware.RequirePermission("doc_view"), handlers.GetDocuments)
		docAPI.GET("/:id/download", middleware.RequirePermission("doc_view"), handlers.DownloadDocument)

		docAPI.POST("/upload", middleware.RequirePermission("doc_manage"), handlers.UploadDocument)
		docAPI.PUT("/:id", middleware.RequirePermission("doc_manage"), handlers.UpdateDocument)
		docAPI.DELETE("/:id", middleware.RequirePermission("doc_manage"), handlers.DeleteDocument)
		docAPI.PUT("/:id/approve", middleware.RequirePermission("approval_final"), handlers.ApproveDocument)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
