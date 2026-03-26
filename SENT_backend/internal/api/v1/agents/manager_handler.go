package agents

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	agentSvc "sent_backend/internal/service/agents" // Brain
	"sent_backend/internal/websocket"

	// Để tính điểm
	"github.com/gin-gonic/gin"
)

// GetStats (GATE)
func GetStats(c *gin.Context) {
	svc := &agentSvc.AgentDataService{}
	stats, _ := svc.GetAgentStats()
	c.JSON(http.StatusOK, stats)
}

// GetAgents (GATE)
func GetAgents(c *gin.Context) {
	orgID := c.GetUint("org_id")
	svc := &agentSvc.AgentDataService{}
	c.JSON(http.StatusOK, svc.GetAgentList(orgID))
}

// GetAgentDetail (GATE)
func GetAgentDetail(c *gin.Context) {
	hwid := c.Param("hwid")
	svc := &agentSvc.AgentDataService{}
	agent, err := svc.GetAgentDetail(hwid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy thiết bị"})
		return
	}
	c.JSON(http.StatusOK, agent)
}

// GetAgentLogs (GATE)
func GetAgentLogs(c *gin.Context) {
	hwid := c.Param("hwid")
	svc := &agentSvc.AgentDataService{}
	c.JSON(http.StatusOK, svc.GetAgentLogs(hwid))
}

// --- CÁC HÀM THAY ĐỔI TRẠNG THÁI GỌI LIFECYCLE SERVICE ---

func AssignManager(c *gin.Context) {
	hwid := c.Param("hwid")
	var req struct {
		UserID uint `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return
	}

	svc := &agentSvc.AgentLifecycleService{}
	if err := svc.AssignManager(hwid, c.GetUint("org_id"), req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã phân công nhân sự"})
}

func RequestDeleteAgent(c *gin.Context) {
	hwid := c.Param("hwid")
	username, _ := c.Get("username")
	orgID := c.GetUint("org_id")

	svc := &agentSvc.AgentLifecycleService{}
	if err := svc.CreateBulkDeleteRequest([]string{hwid}, orgID, "Admin yêu cầu gỡ bỏ", username.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu gỡ bỏ"})
}

// RequestBulkDeleteAgents (GATE): Xóa nhiều máy
func RequestBulkDeleteAgents(c *gin.Context) {
	var req struct {
		HWIDs  []string `json:"hwids"`
		Reason string   `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return
	}

	username, _ := c.Get("username")
	svc := &agentSvc.AgentLifecycleService{}
	err := svc.CreateBulkDeleteRequest(req.HWIDs, c.GetUint("org_id"), req.Reason, username.(string))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Đã gửi yêu cầu gỡ bỏ hàng loạt"})
}
func UpdateDeviceType(c *gin.Context) {
	hwid := c.Param("hwid")
	var req struct {
		DeviceType string `json:"device_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Gọi Brain xử lý cập nhật và tính lại điểm rủi ro
	svc := &agentSvc.AgentLifecycleService{}
	if err := svc.UpdateDeviceType(hwid, req.DeviceType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật loại thiết bị thành công"})
}

// TriggerBaseline: Ra lệnh cho Agent quét sạch hệ thống (Zero Trust)
func TriggerBaseline(c *gin.Context) {
	hwid := c.Param("hwid")

	// 1. Kiểm tra máy có Online không qua Hub
	websocket.GlobalHub.Mu.Lock()
	_, isOnline := websocket.GlobalHub.Clients[hwid]
	websocket.GlobalHub.Mu.Unlock()

	if !isOnline {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Máy trạm hiện đang Offline, không thể nhận lệnh."})
		return
	}

	// 2. Đẩy lệnh xuống Agent qua WebSocket
	websocket.GlobalHub.PushCommand(hwid, gin.H{
		"type": "TRIGGER_BASELINE",
		"data": gin.H{"requester": "Admin"},
	})

	// 3. Cập nhật trạng thái Baseline trong DB
	database.DB.Model(&models.Agent{}).Where("hw_id = ?", hwid).Update("baseline_status", "SCANNING")

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi lệnh quét Baseline tới máy trạm"})
}
