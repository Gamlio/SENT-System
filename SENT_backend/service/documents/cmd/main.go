package main

import (
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

	r := gin.Default()

	// 2. Cấu hình Routes cho Document Service
	docAPI := r.Group("/api/v1/docs")
	docAPI.Use(middleware.AuthRequired()) // Bắt buộc đăng nhập [cite: 3]
	{
		// Xem danh sách và tải tài liệu
		docAPI.GET("", middleware.RequirePermission("doc_view"), handlers.GetDocuments)
		docAPI.GET("/:id/download", middleware.RequirePermission("doc_view"), handlers.DownloadDocument)

		// Quản lý tài liệu (Yêu cầu quyền manage)
		docAPI.POST("/upload", middleware.RequirePermission("doc_manage"), handlers.UploadDocument)
		docAPI.DELETE("/:id", middleware.RequirePermission("doc_manage"), handlers.DeleteDocument)
		docAPI.PUT("/:id/approve", middleware.RequirePermission("approval_final"), handlers.ApproveDocument)
	}

	// 3. Chạy Service trên cổng 8007 (Theo cấu hình Docker Compose)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8007"
	}
	r.Run(":" + port)
}
