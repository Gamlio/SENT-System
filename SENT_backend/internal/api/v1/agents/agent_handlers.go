package agents

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	// Import các package mới chia
	"sent_backend/internal/service/agent_data"
	"sent_backend/internal/service/security"
	"time"

	"github.com/gin-gonic/gin"
)

// PushDataHandler: Tiếp nhận dữ liệu từ Agent v3.1
func PushDataHandler(c *gin.Context) {
	var req struct {
		LogType     string      `json:"log_type"`
		CompanyCode string      `json:"company_code"`
		HWID        string      `json:"hwid"`
		Hostname    string      `json:"hostname"`
		Data        interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// 1. XÁC THỰC CÔNG TY
	var org models.Organization
	if err := database.DB.Where("company_code = ?", req.CompanyCode).First(&org).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mã công ty không tồn tại hoặc sai"})
		return
	}

	// 1.5. TÌM VÙNG (REGION) MẶC ĐỊNH
	var region models.Region
	if err := database.DB.Where("org_id = ?", org.ID).First(&region).Error; err != nil {
		region = models.Region{
			OrgID:       org.ID,
			Name:        "Trụ sở chính",
			EnrollToken: "AUTO-" + req.CompanyCode,
		}
		database.DB.Create(&region)
	}

	// 2. TÌM HOẶC TẠO AGENT
	var agent models.Agent
	result := database.DB.Where("hw_id = ?", req.HWID).First(&agent)

	if result.Error != nil {
		// Máy mới -> Tạo mới
		agent = models.Agent{
			HWID:      req.HWID,
			OrgID:     org.ID,
			RegionID:  region.ID,
			Hostname:  req.Hostname,
			IPAddress: c.ClientIP(),
			Status:    "online",
			LastSeen:  time.Now(),
		}
		if err := database.DB.Create(&agent).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lưu Agent: " + err.Error()})
			return
		}
	} else {
		// Máy cũ -> Cập nhật trạng thái
		updates := map[string]interface{}{
			"ip_address": c.ClientIP(),
			"last_seen":  time.Now(),
			"status":     "online",
		}
		// Tự sửa lỗi mất OrgID/RegionID
		if agent.OrgID == 0 {
			updates["org_id"] = org.ID
		}
		if agent.RegionID == 0 {
			updates["region_id"] = region.ID
		}

		database.DB.Model(&agent).Updates(updates)
	}

	// 3. XỬ LÝ DỮ LIỆU LOG (Dùng các package mới chia)
	switch req.LogType {
	case "inventory":
		agent_data.ProcessInventory(agent, req.Data)
	case "telemetry":
		agent_data.ProcessTelemetry(agent, req.Data)
		go security.AnalyzeBehaviorAI(agent.HWID, "telemetry", req.Data)
	case "software":
		agent_data.ProcessSoftware(agent, req.Data)
		// KIỂM TRA CHÍNH SÁCH NGAY LẬP TỨC
		go security.CheckSoftwareCompliance(agent, req.Data)
	case "usb":
		agent_data.ProcessUSB(agent, req.Data)
		// KIỂM TRA CHÍNH SÁCH USB
		go security.AnalyzeBehaviorAI(agent.HWID, "usb", req.Data)
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed", "type": req.LogType})
}

// GetAgents: Lấy danh sách máy trạm
func GetAgents(c *gin.Context) {
	var agents []models.Agent
	// Preload Manager để hiển thị người quản lý
	if err := database.DB.Preload("Manager").Find(&agents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy dữ liệu"})
		return
	}

	threshold := time.Now().Add(-2 * time.Minute)
	for i := range agents {
		if agents[i].LastSeen.After(threshold) {
			agents[i].Status = "online"
		} else {
			agents[i].Status = "offline"
		}
	}
	c.JSON(http.StatusOK, agents)
}

// GetAgentDetail: Lấy chi tiết kèm USB, Software, Inventory
func GetAgentDetail(c *gin.Context) {
	hwid := c.Param("hwid")
	var agent models.Agent

	err := database.DB.Preload("Inventory").
		Preload("Software").
		Preload("Alerts").
		Preload("USBLogs"). // <--- Đã thêm USBLogs
		Where("hw_id = ?", hwid).
		First(&agent).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy thiết bị"})
		return
	}

	threshold := time.Now().Add(-2 * time.Minute)
	if agent.LastSeen.After(threshold) {
		agent.Status = "online"
	} else {
		agent.Status = "offline"
	}

	c.JSON(http.StatusOK, agent)
}

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

func AssignManager(c *gin.Context) {
	hwid := c.Param("hwid")
	var req struct {
		UserID uint `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}
	// Fix lỗi phân bổ
	if err := database.DB.Model(&models.Agent{}).Where("hw_id = ?", hwid).Update("user_id", req.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi DB"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã phân bổ thành công"})
}
