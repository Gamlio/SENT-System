package v1

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

// PushDataHandler: Tiếp nhận dữ liệu từ Agent v3.1
func PushDataHandler(c *gin.Context) {
	var req struct {
		LogType     string      `json:"log_type"`
		EnrollToken string      `json:"enroll_token"`
		HWID        string      `json:"hwid"`
		Hostname    string      `json:"hostname"`
		Data        interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	var agent models.Agent
	result := database.DB.Where("hw_id = ?", req.HWID).First(&agent)

	if result.Error != nil {
		// Silent Enrollment: Đăng ký máy mới qua Token
		var region models.Region
		if err := database.DB.Where("enroll_token = ?", req.EnrollToken).First(&region).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Mã Enrollment Token không tồn tại hoặc sai"})
			return
		}

		agent = models.Agent{
			HWID:     req.HWID,
			OrgID:    region.OrgID,
			RegionID: region.ID,
			Hostname: req.Hostname,
			Status:   "online",
			LastSeen: time.Now(),
		}
		database.DB.Create(&agent)
	}

	// Phân phối dữ liệu vào Service để xử lý logic
	switch req.LogType {
	case "inventory":
		service.ProcessInventory(agent, req.Data)
	case "telemetry":
		service.ProcessTelemetry(agent, req.Data)
	case "software":
		service.ProcessSoftware(agent, req.Data)
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed", "type": req.LogType})
}

// GetAgents: Lấy danh sách máy trạm
func GetAgents(c *gin.Context) {
	var agents []models.Agent
	if err := database.DB.Find(&agents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy dữ liệu"})
		return
	}
	c.JSON(http.StatusOK, agents)
}

// GetAgentDetail: Lấy chi tiết kèm Inventory và Software
func GetAgentDetail(c *gin.Context) {
	hwid := c.Param("hwid")
	var agent models.Agent

	// Thêm .Preload("Alerts") vào chuỗi truy vấn
	err := database.DB.Preload("Inventory").
		Preload("Software").
		Preload("Alerts"). // <--- THÊM CÁI NÀY
		Where("hw_id = ?", hwid).
		First(&agent).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy thiết bị"})
		return
	}
	c.JSON(http.StatusOK, agent)
}

// GetStats: Trả về số liệu cho Dashboard
func GetStats(c *gin.Context) {
	var total, online, alerts, regions int64
	database.DB.Model(&models.Agent{}).Count(&total)
	database.DB.Model(&models.Agent{}).Where("status = ?", "online").Count(&online)
	database.DB.Model(&models.SecurityAlert{}).Where("is_resolved = ?", false).Count(&alerts)
	database.DB.Model(&models.Region{}).Count(&regions)

	c.JSON(http.StatusOK, gin.H{
		"total": total, "online": online, "alerts": alerts, "regions": regions,
	})
}
func GetAgentLogs(c *gin.Context) {
	hwid := c.Param("hwid")
	var alerts []models.SecurityAlert

	// Lấy tất cả cảnh báo của HWID này, sắp xếp mới nhất trước
	result := database.DB.Where("hw_id = ?", hwid).Order("created_at desc").Find(&alerts)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Lỗi truy vấn Log"})
		return
	}

	c.JSON(200, alerts)
}
