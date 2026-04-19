package main

import (
	"os"
	"strings"
	"time"

	"sent_backend/internal/database"
	"sent_backend/internal/websocket"

	// IMPORT CÁC PACKAGE ĐÃ CHIA NHỎ
	"sent_backend/internal/api/v1/ai"
	"sent_backend/internal/api/v1/approvals"
	"sent_backend/internal/api/v1/assets"
	"sent_backend/internal/api/v1/auth"
	"sent_backend/internal/api/v1/dashboard"
	"sent_backend/internal/api/v1/docs"
	"sent_backend/internal/api/v1/incidents"
	"sent_backend/internal/api/v1/policies"
	"sent_backend/internal/api/v1/users"
	"sent_backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	database.InitDB()

	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	r := gin.New()

	r.RedirectTrailingSlash = false
	// ===== BẢNG MÔNG BẢO MẬT TẦNG GLOBAL =====
	// 1. Kiểm tra Content-Type, Body Size -> Chặn DoS attacks
	r.Use(middleware.SecurityValidationMiddleware())
	// 2. Rate limiting chung
	limiter := middleware.RateLimitMiddleware(10, 20)
	r.Use(limiter)
	// 3. IP Blacklist check
	r.Use(middleware.IPBlacklistMiddleware())

	originsEnv := os.Getenv("ALLOWED_ORIGINS")
	var allowOrigins []string
	if originsEnv != "" {
		allowOrigins = strings.Split(originsEnv, ",")
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Static("/uploads", "./uploads")

	v1Group := r.Group("/api/v1")
	{

		authGroup := v1Group.Group("/auth")
		authGroup.Use(middleware.InputSanitizationMiddleware())
		{
			authGroup.POST("/login", middleware.UserLoginRateLimitMiddleware(), auth.LoginHandler)
			authGroup.POST("/register", middleware.ValidateUserCreationMiddleware(), auth.RegisterSMEHandler)
		}

		assetPublicGroup := v1Group.Group("/assets")
		{
			assetPublicGroup.POST("/push", middleware.AssetFloodProtectionMiddleware(), middleware.ValidateAssetPayloadMiddleware(), assets.PushDataHandler)
			assetPublicGroup.POST("/enroll", middleware.AssetEnrollRateLimitMiddleware(), assets.EnrollAsset)
			assetPublicGroup.GET("/sync-policies", assets.GetActiveEnrollmentToken, policies.SyncPoliciesForAsset)
		}

		protected := v1Group.Group("")
		protected.Use(middleware.AuthRequired())
		protected.Use(middleware.InputSanitizationMiddleware())
		{
			protected.GET("/ws", websocket.WsHandler)
			usersGroup := protected.Group("/users")
			{
				usersGroup.GET("", users.GetUsers)
				usersGroup.POST("", users.CreateUser)
				usersGroup.PUT("/:id", users.UpdateUser)
				usersGroup.DELETE("/:id", users.DeleteUser)
			}

			assetsGroup := protected.Group("/assets")
			{
				assetsGroup.GET("/stats", assets.GetStats)
				assetsGroup.GET("", assets.GetAssets)
				assetsGroup.GET("/:hwid", assets.GetAssetDetail)
				assetsGroup.GET("/:hwid/logs", assets.GetAssetLogs)
				assetsGroup.GET("/active-token", assets.GetActiveEnrollmentToken)
				assetsGroup.POST("/generate-token", assets.GenerateEnrollmentToken)
				assetsGroup.POST("/:hwid/trigger-baseline", assets.TriggerBaseline)
				assetsGroup.PUT("/:hwid/assign", assets.AssignManager)
				assetsGroup.PUT("/:hwid/device-type", assets.UpdateDeviceType)
				assetsGroup.PUT("/:hwid/department", assets.UpdateDepartment)
				assetsGroup.POST("/:hwid/request-delete", assets.RequestDeleteAsset)
				assetsGroup.POST("/bulk-request-delete", assets.RequestBulkDeleteAssets)
			}

			aiDocs := protected.Group("/docs")
			{
				aiDocs.GET("", docs.GetDocuments)           // Lấy danh sách tài liệu
				aiDocs.POST("/upload", docs.UploadDocument) // Upload PDF/Word
				aiDocs.DELETE("/:id", docs.DeleteDocument)  // Xóa tài liệu
				aiDocs.PUT("/:id", docs.UpdateDocument)     // Sửa tên tài liệu
			}

			// 2. GROUP POLICIES (CHÍNH SÁCH KỸ THUẬT CHO asset)
			policiesGroup := protected.Group("/policies")
			{
				policiesGroup.GET("", policies.GetPoliciesByCategory)           // Lấy luật JSON
				policiesGroup.POST("/bulk", policies.AddBulkPolicies)           // Thêm nhiều luật cùng lúc (Dành cho import Excel)
				policiesGroup.POST("", policies.AddPolicy)                      // Thêm luật JSON
				policiesGroup.DELETE("/:id", policies.DeletePolicy)             // Xóa luật JSONager)
				policiesGroup.POST("/bulk-delete", policies.DeleteBulkPolicies) // Xóa nhiều luật cùng lúc

			}
			dashGroup := protected.Group("/dashboard")
			{
				// Trỏ về đúng Handler mỏng vừa tạo
				dashGroup.GET("/stats", dashboard.GetDashboardStats)
			}

			incidentsGroup := protected.Group("/incidents")
			{
				incidentsGroup.GET("", incidents.GetIncidents)          // Lấy danh sách (Đã đổi tên khớp Handler)
				incidentsGroup.GET("/:id", incidents.GetIncidentDetail) // Chi tiết sự cố + Timeline Audit
				incidentsGroup.POST("/close", incidents.CloseIncident)  // Đóng Case kèm Baseline Proof
				incidentsGroup.PUT("/:id/playbook", incidents.UpdatePlaybookProgress)

				// --- Dành cho Admin/Auditor (Chuyển giao quyền quản lý) ---
				// Gợi ý: Nên bọc qua một middleware CheckRole("ADMIN") ở đây
				incidentsGroup.PUT("/:id/assign", incidents.AssignIncident)                   // Chỉ Admin mới được phân công người làm
				incidentsGroup.GET("/audit/:audit_id/verify", incidents.VerifyAuditIntegrity) // Kiểm tra tính toàn vẹn của Log

			}
			aiGroup := protected.Group("/ai")
			{
				aiGroup.POST("/chat", ai.ChatHandler)
				aiGroup.POST("/sessions", ai.CreateSession) // Tạo phiên mới
				aiGroup.GET("/sessions", ai.GetSessions)    // Lấy danh sách phiên

				// Chat trong phiên cụ thể
				aiGroup.POST("/chat/:session_id", ai.ChatHandler)
				// Các endpoint khác như xóa phiên, lấy lịch sử chat... có thể thêm sau khi có cơ sở hạ tầng chính. Hiện tập trung vào chức năng chat chính đã.
				aiGroup.DELETE("/sessions/:id", ai.DeleteSession) // Xóa phiên
				aiGroup.PUT("/sessions/:id", ai.RenameSession)
				aiGroup.GET("/chat/:session_id", ai.GetChatHistory) // Lấy lịch sử chat của phiên
			}
			approvalsGroup := protected.Group("/approvals")
			{
				// Lấy danh sách các đơn cần duyệt (có thể lọc theo ModuleType)
				approvalsGroup.GET("", approvals.GetTickets)

				// Admin thao tác: Duyệt hoặc Từ chối
				approvalsGroup.PUT("/:id/review", approvals.ReviewTicket)
			}

		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
