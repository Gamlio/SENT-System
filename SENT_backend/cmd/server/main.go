package main

import (
	"os"
	"strings"
	"time"

	"sent_backend/internal/database"
	// IMPORT CÁC PACKAGE ĐÃ CHIA NHỎ
	"sent_backend/internal/api/v1/agents"
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

	r := gin.Default()

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
			authGroup.POST("/login", auth.LoginHandler)
			authGroup.POST("/register", auth.RegisterSMEHandler)
		}
		agentPublicGroup := v1Group.Group("/agents")
		{
			agentPublicGroup.POST("/push", agents.PushDataHandler)
		}
		protected := v1Group.Group("")
		protected.Use(middleware.AuthRequired()) // <--- CHỐT BẢO VỆ NẰM Ở ĐÂY
		{
			usersGroup := protected.Group("/users")
			{
				usersGroup.GET("", users.GetUsers)
				usersGroup.POST("", users.CreateUser)
				usersGroup.PUT("/:id", users.UpdateUser)
				usersGroup.DELETE("/:id", users.DeleteUser)
			}

			agentsGroup := protected.Group("/agents")
			{
				agentsGroup.GET("/stats", agents.GetStats)
				agentsGroup.GET("", agents.GetAgents)
				agentsGroup.GET("/:hwid", agents.GetAgentDetail)
				agentsGroup.GET("/:hwid/logs", agents.GetAgentLogs)

				agentsGroup.PUT("/:hwid/assign", agents.AssignManager)
				agentPublicGroup.GET("/sync-policies", policies.SyncPoliciesForAgent)
			}

			aiDocs := protected.Group("/docs")
			{
				aiDocs.GET("", docs.GetDocuments)           // Lấy danh sách tài liệu
				aiDocs.POST("/upload", docs.UploadDocument) // Upload PDF/Word
				aiDocs.DELETE("/:id", docs.DeleteDocument)  // Xóa tài liệu
				aiDocs.PUT("/:id", docs.UpdateDocument)     // Sửa tên tài liệu
			}

			// 2. GROUP POLICIES (CHÍNH SÁCH KỸ THUẬT CHO AGENT)
			policiesGroup := protected.Group("/policies")
			{
				policiesGroup.GET("", policies.GetPoliciesByCategory) // Lấy luật JSON
				policiesGroup.POST("/bulk", policies.AddBulkPolicies) // Thêm nhiều luật cùng lúc (Dành cho import Excel)
				policiesGroup.POST("", policies.AddUniversalPolicy)   // Thêm luật JSON
				policiesGroup.DELETE("/:id", policies.DeletePolicy)   // Xóa luật JSON
			}
			dashGroup := protected.Group("/dashboard")
			{
				dashGroup.GET("/stats", dashboard.GetDashboardStats)

			}

			incidentsGroup := protected.Group("/incidents")
			{
				dashGroup.GET("/incidents", incidents.GetIncidents)
				incidentsGroup.GET("", incidents.GetIncidents)
				incidentsGroup.GET("/:id", incidents.GetIncidentDetail)
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
