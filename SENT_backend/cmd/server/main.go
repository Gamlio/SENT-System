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
	r.Use(gin.Recovery())
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
				usersGroup.GET("", middleware.RequirePermission("user_view"), users.GetUsers)

				userManage := middleware.RequirePermission("user_manage")
				usersGroup.POST("", userManage, users.CreateUser)
				usersGroup.PUT("/:id", userManage, users.UpdateUser)
				usersGroup.DELETE("/:id", userManage, users.DeleteUser)
			}
			assetsGroup := protected.Group("/assets")
			{
				viewAssets := middleware.RequirePermission("asset_view")
				assetsGroup.GET("/stats", viewAssets, assets.GetStats)
				assetsGroup.GET("", viewAssets, assets.GetAssets)
				assetsGroup.GET("/:hwid", viewAssets, assets.GetAssetDetail)
				assetsGroup.GET("/:hwid/logs", viewAssets, assets.GetAssetLogs)
				assetsGroup.GET("/:hwid/software", viewAssets, assets.GetAssetSoftware)
				assetsGroup.GET("/:hwid/usb", viewAssets, assets.GetAssetUSB)
				assetsGroup.GET("/:hwid/ports", viewAssets, assets.GetAssetPorts)
				assetsGroup.GET("/active-token", viewAssets, assets.GetActiveEnrollmentToken)
				assetsGroup.GET("/types", viewAssets, assets.GetAssetTypes)

				actionAssets := middleware.RequirePermission("asset_action")
				assetsGroup.POST("/generate-token", actionAssets, assets.GenerateEnrollmentToken)
				assetsGroup.PUT("/:hwid/assign", actionAssets, assets.AssignManager)
				assetsGroup.PUT("/:hwid/department", actionAssets, assets.UpdateDepartment)

				assetsGroup.PUT("/types/:id", middleware.RequirePermission("asset_move"), assets.UpdateAssetType)

				deleteAssets := middleware.RequirePermission("asset_delete")
				assetsGroup.POST("/:hwid/request-delete", deleteAssets, assets.RequestDeleteAsset)
				assetsGroup.POST("/bulk-request-delete", deleteAssets, assets.RequestBulkDeleteAssets)
			}

			aiDocs := protected.Group("/docs")
			{
				aiDocs.GET("", middleware.RequirePermission("doc_view"), docs.GetDocuments)

				docManage := middleware.RequirePermission("doc_manage")
				aiDocs.POST("/upload", docManage, docs.UploadDocument)
				aiDocs.DELETE("/:id", docManage, docs.DeleteDocument)
				aiDocs.PUT("/:id", docManage, docs.UpdateDocument)
				aiDocs.POST("/:id/delete-request", docManage, docs.DeleteDocument)
			}

			policiesGroup := protected.Group("/policies")
			{
				policyView := middleware.RequirePermission("policy_view")
				policiesGroup.GET("", policyView, policies.GetPoliciesByCategory)
				policiesGroup.GET("/groups", policyView, policies.GetPolicyGroups)

				policyManage := middleware.RequirePermission("policy_manage")
				policiesGroup.POST("/bulk", policyManage, policies.AddBulkPolicies)
				policiesGroup.DELETE("/:id", policyManage, policies.DeletePolicy)
				policiesGroup.POST("/bulk-delete", policyManage, policies.DeleteBulkPolicies)
			}

			dashGroup := protected.Group("/dashboard")
			{
				dashGroup.GET("/stats", middleware.RequirePermission("asset_view"), dashboard.GetDashboardStats)
			}

			incidentsGroup := protected.Group("/incidents")
			{
				incidentView := middleware.RequirePermission("incident_view")
				incidentsGroup.GET("", incidentView, incidents.GetIncidents)
				incidentsGroup.GET("/:id", incidentView, incidents.GetIncidentDetail)

				incidentAction := middleware.RequirePermission("incident_action")
				incidentsGroup.POST("/audit/upload", incidentAction, incidents.AddAuditLogHandler)
				incidentsGroup.PUT("/:id/assign", incidentAction, incidents.AssignIncident)
				incidentsGroup.POST("/close", incidentAction, incidents.CloseIncident)
				incidentsGroup.GET("/audit/:audit_id/verify", incidentAction, incidents.VerifyAuditIntegrity)
			}

			// --- MODULE: BEHAVIORS ---
			behaviorGroup := protected.Group("/behaviors")
			{
				incidentView := middleware.RequirePermission("incident_view")
				behaviorGroup.GET("", incidentView, behavior.GetBehaviors)
				behaviorGroup.GET("/:id", incidentView, behavior.GetBehaviorDetail)

				incidentAction := middleware.RequirePermission("incident_action")
				behaviorGroup.POST("/create-incident", incidentAction, behavior.CreateIncidentHandler)
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
				approvalsGroup.GET("", middleware.RequirePermission("approval_view"), approvals.GetTickets)
				approvalsGroup.PUT("/:id/review", middleware.RequirePermission("approval_final"), approvals.ReviewTicket)
			}

			groupsGroup := protected.Group("/groups")
			groupsGroup.Use(middleware.RequirePermission("group_manage"))
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
