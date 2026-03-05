package agents

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

// GetStats: Số liệu Dashboard
func GetStats(c *gin.Context) {
	var total, online, alerts, regions int64
	database.DB.Model(&models.Agent{}).Count(&total)

	threshold := time.Now().Add(-2 * time.Minute)
	database.DB.Model(&models.Agent{}).Where("last_seen >= ?", threshold).Count(&online)

	database.DB.Model(&models.SecurityAlert{}).Where("is_resolved = ?", false).Count(&alerts)
	database.DB.Model(&models.Region{}).Count(&regions)

	c.JSON(http.StatusOK, gin.H{
		"total": total, "online": online, "alerts": alerts, "regions": regions,
	})
}

func GetAgentLogs(c *gin.Context) {
	hwid := c.Param("hwid")
	var alerts []models.SecurityAlert
	database.DB.Where("hw_id = ?", hwid).Order("created_at desc").Find(&alerts)
	c.JSON(200, alerts)
}
