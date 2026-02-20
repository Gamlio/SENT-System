package main

import (
	"os"
	"strings"
	"time"

	v1 "sent_backend/internal/api/v1"
	"sent_backend/internal/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load biến môi trường TRƯỚC TIÊN
	// Nếu không load được thì cũng không sao (có thể chạy bằng biến hệ thống)
	_ = godotenv.Load()

	// 2. Khởi tạo DB (Trong db.go bạn nhớ dùng os.Getenv("DATABASE_URL"))
	database.InitDB()

	// 3. Cấu hình chế độ Gin (Debug/Release)
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	r := gin.Default()

	// 4. Xử lý CORS từ biến môi trường
	originsEnv := os.Getenv("ALLOWED_ORIGINS")
	var allowOrigins []string
	if originsEnv != "" {
		allowOrigins = strings.Split(originsEnv, ",")
	}

	// Cấu hình CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins, // Dùng danh sách từ .env
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Static("/uploads", "./uploads")
	v1Group := r.Group("/api/v1")
	{
		auth := v1Group.Group("/auth")
		{
			auth.POST("/login", v1.LoginHandler)
			auth.POST("/register", v1.RegisterSMEHandler)
		}

		agents := v1Group.Group("/agents")
		{
			agents.POST("/push", v1.PushDataHandler)
			agents.GET("/stats", v1.GetStats)
			agents.GET("", v1.GetAgents)

			// --- ĐĂNG KÝ ROUTE API Ở ĐÂY ---
			agents.POST("/bulk-whitelist", v1.AddBulkWhitelist) // Phải để trên /:hwid

			agents.GET("/:hwid", v1.GetAgentDetail)
			agents.GET("/:hwid/logs", v1.GetAgentLogs)

			agents.GET("/:hwid/whitelist", v1.GetAgentWhitelist)
			agents.POST("/:hwid/whitelist", v1.AddAgentWhitelist)
			agents.DELETE("/whitelist/:id", v1.DeleteAgentWhitelist)
		}

		aiDocs := v1Group.Group("/docs")
		{
			aiDocs.GET("", v1.GetPoliciesHandler)          // Lấy danh sách file PDF/Word
			aiDocs.POST("/upload", v1.UploadPolicyHandler) // Upload file mới
			aiDocs.DELETE("/:id", v1.DeletePolicyHandler)  // Xóa file vật lý và DB
			aiDocs.PUT("/:id", v1.UpdatePolicyHandler)     // Sửa thông tin tài liệu
		}

		// --- 2. TRUNG TÂM CHÍNH SÁCH (QUẢN LÝ QUY TẮC KỸ THUẬT) ---
		// Gom tất cả software, usb, network vào đây
		policies := v1Group.Group("/policies")
		{
			policies.GET("", v1.GetPoliciesByCategory) // Lấy luật theo category (SOFTWARE/USB/...)
			policies.POST("", v1.AddUniversalPolicy)   // Thêm luật kỹ thuật mới
			policies.DELETE("/:id", v1.DeletePolicy)   // Xóa luật kỹ thuật
		}
	}

	// 5. Chạy Server theo PORT trong .env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Mặc định nếu thiếu
	}
	r.Run(":" + port)
}
