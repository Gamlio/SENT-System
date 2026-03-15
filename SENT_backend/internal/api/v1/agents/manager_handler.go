package agents

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/scoring"
	"time"

	"github.com/gin-gonic/gin"
)

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
		Preload("Manager"). // <--- BẮT BUỘC THÊM DÒNG NÀY
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

// [MỚI] UpdateDeviceType: Cập nhật phân loại thiết bị và tính lại điểm
func UpdateDeviceType(c *gin.Context) {
	hwid := c.Param("hwid")
	var req struct {
		DeviceType string `json:"device_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Danh sách các loại hợp lệ
	validTypes := map[string]bool{"SERVER": true, "IT_ADMIN": true, "OFFICE": true, "GUEST": true}
	if !validTypes[req.DeviceType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Loại thiết bị không hợp lệ"})
		return
	}

	// Cập nhật Database
	if err := database.DB.Model(&models.Agent{}).Where("hw_id = ?", hwid).Update("device_type", req.DeviceType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi DB khi cập nhật"})
		return
	}

	// [QUAN TRỌNG] Phải tính lại điểm rủi ro ngay lập tức vì hệ số W_asset đã thay đổi
	go scoring.RecalculateRiskScore(hwid)

	c.JSON(http.StatusOK, gin.H{"message": "Đã cập nhật loại thiết bị thành công", "device_type": req.DeviceType})
}
