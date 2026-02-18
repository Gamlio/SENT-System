package v1

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
)

func GetDashboardStats(c *gin.Context) {
	// 1. Lấy OrgID từ người dùng (R1/R2 xem tất cả, R3/R4 xem theo công ty)
	orgID := c.MustGet("org_id").(uint)

	var stats struct {
		TotalAgents  int64 `json:"total_agents"`
		OnlineAgents int64 `json:"online_agents"`
		TotalAlerts  int64 `json:"total_alerts"`
		TotalRegions int64 `json:"total_regions"`
	}

	// 2. Truy vấn số liệu thực tế từ Database
	database.DB.Model(&models.Agent{}).Where("org_id = ?", orgID).Count(&stats.TotalAgents)
	database.DB.Model(&models.Agent{}).Where("org_id = ? AND status = ?", orgID, "online").Count(&stats.OnlineAgents)
	database.DB.Model(&models.SecurityAlert{}).Where("org_id = ?", orgID).Count(&stats.TotalAlerts)
	database.DB.Model(&models.Region{}).Where("org_id = ?", orgID).Count(&stats.TotalRegions)

	c.JSON(http.StatusOK, stats)
}
