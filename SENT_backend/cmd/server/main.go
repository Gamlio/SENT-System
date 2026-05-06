package main

import (
	"os"

	"strings"
	"time"

	"sent_backend/internal/database"
	"sent_backend/internal/websocket"

	"sent_backend/internal/api/v1/ai"
	"sent_backend/internal/api/v1/approvals"
	"sent_backend/internal/api/v1/assets"
	"sent_backend/internal/api/v1/auth"
	"sent_backend/internal/api/v1/group"

	"sent_backend/internal/api/v1/behavior"
	"sent_backend/internal/api/v1/dashboard"
	"sent_backend/internal/api/v1/docs"
	"sent_backend/internal/api/v1/incidents"
	"sent_backend/internal/api/v1/policies"
	"sent_backend/internal/api/v1/users"
	"sent_backend/internal/api/v1/version"
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
	r.Use(gin.Logger())
	r.RedirectTrailingSlash = false
	r.Use(middleware.IPBlacklistMiddleware())
	r.Use(middleware.SecurityValidationMiddleware())
	limiter := middleware.RateLimitMiddleware(10, 20)
	r.Use(limiter)

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
		{
			authGroup.POST("/login", middleware.UserLoginRateLimitMiddleware(), auth.LoginHandler)
			authGroup.POST("/register", middleware.ValidateUserCreationMiddleware(), auth.RegisterSMEHandler)
		}

		assetPublicGroup := v1Group.Group("/assets")
		{
			assetPublicGroup.POST("/push", middleware.AssetHMACAuth(), middleware.AssetFloodProtectionMiddleware(), middleware.ValidateAssetPayloadMiddleware(), assets.PushDataHandler)
			assetPublicGroup.POST("/enroll", middleware.AssetEnrollRateLimitMiddleware(), assets.EnrollAsset)
			assetPublicGroup.GET("/sync-policies", assets.GetActiveEnrollmentToken, policies.SyncPoliciesForAsset)
		}

		v1Group.GET("/versions", version.GetVersions)
		v1Group.POST("/internal/update-version", version.GitHubWebhookHandler)

		protected := v1Group.Group("")
		protected.Use(middleware.AuthRequired())
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
				assetsGroup.GET("/:hwid/software", assets.GetAssetSoftware)
				assetsGroup.GET("/:hwid/usb", assets.GetAssetUSB)
				assetsGroup.GET("/:hwid/ports", assets.GetAssetPorts)
				assetsGroup.GET("/active-token", assets.GetActiveEnrollmentToken)
				assetsGroup.POST("/generate-token", assets.GenerateEnrollmentToken)
				assetsGroup.PUT("/:hwid/assign", assets.AssignManager)
				assetsGroup.PUT("/types/:id", assets.UpdateAssetType)
				assetsGroup.GET("/types", assets.GetAssetTypes)
				assetsGroup.PUT("/:hwid/department", assets.UpdateDepartment)
				assetsGroup.POST("/:hwid/request-delete", assets.RequestDeleteAsset)
				assetsGroup.POST("/bulk-request-delete", assets.RequestBulkDeleteAssets)
			}

			aiDocs := protected.Group("/docs")
			{
				aiDocs.GET("", docs.GetDocuments)
				aiDocs.POST("/upload", docs.UploadDocument)
				aiDocs.DELETE("/:id", docs.DeleteDocument)
				aiDocs.PUT("/:id", docs.UpdateDocument)
				aiDocs.POST("/:id/delete-request", docs.DeleteDocument)
			}

			policiesGroup := protected.Group("/policies")
			{
				policiesGroup.GET("", policies.GetPoliciesByCategory)
				policiesGroup.POST("/bulk", policies.AddBulkPolicies)
				policiesGroup.DELETE("/:id", policies.DeletePolicy)
				policiesGroup.POST("/bulk-delete", policies.DeleteBulkPolicies)
				policiesGroup.GET("/groups", policies.GetPolicyGroups)
			}
			dashGroup := protected.Group("/dashboard")
			{
				dashGroup.GET("/stats", dashboard.GetDashboardStats)
			}

			incidentsGroup := protected.Group("/incidents")
			{
				incidentsGroup.GET("", incidents.GetIncidents)
				incidentsGroup.GET("/:id", incidents.GetIncidentDetail)
				incidentsGroup.POST("/close", incidents.CloseIncident)
				incidentsGroup.PUT("/:id/playbook", incidents.UpdatePlaybookProgress)

				incidentsGroup.PUT("/:id/assign", incidents.AssignIncident)
				incidentsGroup.GET("/audit/:audit_id/verify", incidents.VerifyAuditIntegrity)

			}
			behaviorGroup := protected.Group("/behaviors")
			{
				behaviorGroup.GET("", behavior.GetBehaviors)
				behaviorGroup.GET("/:id", behavior.GetBehaviorDetail)
				behaviorGroup.POST("/create-incident", behavior.CreateIncidentHandler)
			}
			aiGroup := protected.Group("/ai")
			{
				aiGroup.POST("/chat", ai.ChatHandler)
				aiGroup.POST("/sessions", ai.CreateSession)
				aiGroup.GET("/sessions", ai.GetSessions)

				aiGroup.POST("/chat/:session_id", ai.ChatHandler)
				aiGroup.DELETE("/sessions/:id", ai.DeleteSession)
				aiGroup.PUT("/sessions/:id", ai.RenameSession)
				aiGroup.GET("/chat/:session_id", ai.GetChatHistory)
			}
			approvalsGroup := protected.Group("/approvals")
			{
				approvalsGroup.GET("", approvals.GetTickets)
				approvalsGroup.PUT("/:id/review", approvals.ReviewTicket)
			}
			groupsGroup := protected.Group("/groups")
			{
				groupsGroup.GET("", group.HandleGetGroups)
				groupsGroup.POST("", group.HandleCreateGroup)
				groupsGroup.PUT("/:id", group.HandleUpdateGroup)
				groupsGroup.DELETE("/:id", group.HandleDeleteGroup)
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
