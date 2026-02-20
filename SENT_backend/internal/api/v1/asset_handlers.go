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
		// Máy mới lần đầu kết nối
		var region models.Region
		if err := database.DB.Where("enroll_token = ?", req.EnrollToken).First(&region).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Mã Enrollment Token không tồn tại hoặc sai"})
			return
		}

		agent = models.Agent{
			HWID:      req.HWID,
			OrgID:     region.OrgID,
			RegionID:  region.ID,
			Hostname:  req.Hostname,
			IPAddress: c.ClientIP(), // <--- Lấy IP mạng của thiết bị
			Status:    "online",
			LastSeen:  time.Now(),
		}
		database.DB.Create(&agent)
	} else {
		// Máy cũ gửi log lại -> Cập nhật IP nhỡ máy mang đi chỗ khác (đổi mạng)
		agent.IPAddress = c.ClientIP()
		agent.LastSeen = time.Now()
		agent.Status = "online"
		database.DB.Save(&agent)
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

// GetAgents: Lấy danh sách máy trạm kèm trạng thái Realtime
func GetAgents(c *gin.Context) {
	var agents []models.Agent
	if err := database.DB.Find(&agents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy dữ liệu"})
		return
	}

	// TÍNH TOÁN TRẠNG THÁI ĐỘNG: Nếu LastSeen cũ hơn 2 phút -> Đánh dấu Offline
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

// GetAgentDetail: Lấy chi tiết kèm Inventory và Software
func GetAgentDetail(c *gin.Context) {
	hwid := c.Param("hwid")
	var agent models.Agent

	err := database.DB.Preload("Inventory").
		Preload("Software").
		Preload("Alerts").
		Where("hw_id = ?", hwid).
		First(&agent).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy thiết bị"})
		return
	}

	// TÍNH TOÁN TRẠNG THÁI ĐỘNG TRƯỚC KHI TRẢ VỀ
	threshold := time.Now().Add(-2 * time.Minute)
	if agent.LastSeen.After(threshold) {
		agent.Status = "online"
	} else {
		agent.Status = "offline"
	}

	c.JSON(http.StatusOK, agent)
}

// GetStats: Trả về số liệu Realtime cho Dashboard
func GetStats(c *gin.Context) {
	var total, online, alerts, regions int64

	// 1. Tổng số máy
	database.DB.Model(&models.Agent{}).Count(&total)

	// 2. TÍNH SỐ MÁY ONLINE: Chỉ đếm những máy có LastSeen trong 2 phút đổ lại đây
	threshold := time.Now().Add(-2 * time.Minute)
	database.DB.Model(&models.Agent{}).Where("last_seen >= ?", threshold).Count(&online)

	// 3. Các thông số khác
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

// Lấy danh sách Whitelist của 1 máy
func GetAgentWhitelist(c *gin.Context) {
	var list []models.AgentWhitelist
	database.DB.Where("hwid = ?", c.Param("hwid")).Find(&list)
	c.JSON(200, list)
}

func AddAgentWhitelist(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}
	item := models.AgentWhitelist{HWID: c.Param("hwid"), SoftwareName: req.Name}
	database.DB.Create(&item)
	c.JSON(200, item)
}

func DeleteAgentWhitelist(c *gin.Context) {
	database.DB.Delete(&models.AgentWhitelist{}, c.Param("id"))
	c.JSON(200, gin.H{"status": "ok"})
}

// Hàm mới: CẤP PHÉP HÀNG LOẠT CHO NHIỀU MÁY CÙNG LÚC
func AddBulkWhitelist(c *gin.Context) {
	var req struct {
		HWIDs        []string `json:"hwids"`
		SoftwareName string   `json:"software_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Lặp qua danh sách các máy được tích chọn và lưu vào DB
	for _, hwid := range req.HWIDs {
		item := models.AgentWhitelist{HWID: hwid, SoftwareName: req.SoftwareName}
		database.DB.Create(&item)
	}
	c.JSON(200, gin.H{"status": "success"})
}
