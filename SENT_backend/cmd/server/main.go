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
				agentsGroup.POST("/bulk-whitelist", agents.AddBulkWhitelist)
				agentsGroup.GET("/:hwid", agents.GetAgentDetail)
				agentsGroup.GET("/:hwid/logs", agents.GetAgentLogs)
				agentsGroup.GET("/:hwid/whitelist", agents.GetAgentWhitelist)
				agentsGroup.POST("/:hwid/whitelist", agents.AddAgentWhitelist)
				agentsGroup.DELETE("/whitelist/:id", agents.DeleteAgentWhitelist)
				agentsGroup.PUT("/:hwid/assign", agents.AssignManager)
			}

			aiDocs := protected.Group("/docs")
			{
				aiDocs.GET("", policies.GetPoliciesHandler)
				aiDocs.POST("/upload", policies.UploadPolicyHandler)
				aiDocs.DELETE("/:id", policies.DeletePolicyHandler)
				aiDocs.PUT("/:id", policies.UpdatePolicyHandler)
			}

			policiesGroup := protected.Group("/policies")
			{
				policiesGroup.GET("", policies.GetPoliciesByCategory)
				policiesGroup.POST("", policies.AddUniversalPolicy)
				policiesGroup.DELETE("/:id", policies.DeletePolicy)
			}

			dashGroup := protected.Group("/dashboard")
			{
				dashGroup.GET("/stats", dashboard.GetDashboardStats)
				dashGroup.GET("/incidents", dashboard.GetIncidents)
			}

			incidentsGroup := protected.Group("/incidents")
			{
				incidentsGroup.GET("", dashboard.GetIncidents)
				incidentsGroup.GET("/:id", dashboard.GetIncidentDetail)
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
