package v1

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service" // Import Service vừa viết
	"time"

	"github.com/gin-gonic/gin"
)

func PushDataHandler(c *gin.Context) {
	var req struct {
		LogType     string      `json:"log_type"`
		EnrollToken string      `json:"enroll_token"`
		HWID        string      `json:"hwid"`
		Hostname    string      `json:"hostname"` // Agent v3.1 có gửi kèm hostname
		Data        interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	var agent models.Agent
	// 1. Tìm Agent trong DB
	result := database.DB.Where("hwid = ?", req.HWID).First(&agent)

	if result.Error != nil {
		// 2. Đăng ký máy mới (Silent Enrollment)
		var region models.Region
		if err := database.DB.Where("enroll_token = ?", req.EnrollToken).First(&region).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Enrollment Token không hợp lệ"})
			return
		}

		agent = models.Agent{
			HWID:     req.HWID,
			OrgID:    region.OrgID,
			RegionID: region.ID,
			Hostname: req.Hostname, // Lưu hostname lúc đăng ký
			Status:   "online",
			LastSeen: time.Now(),
		}
		database.DB.Create(&agent)
	}

	// 3. Gọi Service xử lý logic nghiệp vụ
	switch req.LogType {
	case "inventory":
		service.ProcessInventory(agent, req.Data)
	case "telemetry":
		service.ProcessTelemetry(agent, req.Data)
	case "software":
		service.ProcessSoftware(agent, req.Data)
		// case "event_security": service.ProcessEvents(...) // Nếu bạn làm tiếp phần Event
	}

	c.JSON(http.StatusOK, gin.H{"status": "processed", "type": req.LogType})
}
