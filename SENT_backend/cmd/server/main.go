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

		agentsGroup := v1Group.Group("/agents")
		{
			agentsGroup.POST("/push", agents.PushDataHandler)
			agentsGroup.GET("/stats", agents.GetStats)
			agentsGroup.GET("", agents.GetAgents)
			agentsGroup.POST("/bulk-whitelist", agents.AddBulkWhitelist)
			agentsGroup.GET("/:hwid", agents.GetAgentDetail)
			agentsGroup.GET("/:hwid/logs", agents.GetAgentLogs)
			agentsGroup.GET("/:hwid/whitelist", agents.GetAgentWhitelist)
			agentsGroup.POST("/:hwid/whitelist", agents.AddAgentWhitelist)
			agentsGroup.DELETE("/whitelist/:id", agents.DeleteAgentWhitelist)
		}

		aiDocs := v1Group.Group("/docs")
		{
			aiDocs.GET("", policies.GetPoliciesHandler)
			aiDocs.POST("/upload", policies.UploadPolicyHandler)
			aiDocs.DELETE("/:id", policies.DeletePolicyHandler)
			aiDocs.PUT("/:id", policies.UpdatePolicyHandler)
		}

		policiesGroup := v1Group.Group("/policies")
		{
			policiesGroup.GET("", policies.GetPoliciesByCategory)
			policiesGroup.POST("", policies.AddUniversalPolicy)
			policiesGroup.DELETE("/:id", policies.DeletePolicy)
		}

		usersGroup := v1Group.Group("/users")
		{
			usersGroup.GET("", users.GetUsers)
			usersGroup.POST("", users.CreateUser)
			usersGroup.DELETE("/:id", users.DeleteUser)
		}

		// Nhóm mới dành riêng cho Dashboard
		dashGroup := v1Group.Group("/dashboard")
		{
			dashGroup.GET("/stats", dashboard.GetDashboardStats)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	r.Run(":" + port)
}
