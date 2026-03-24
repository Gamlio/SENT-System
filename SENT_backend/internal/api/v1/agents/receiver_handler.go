package agents

import (
	"encoding/json"
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	agentService "sent_backend/internal/service/agents"

	"github.com/gin-gonic/gin"
)

// Payload đón dữ liệu thô
type AgentPayload struct {
	LogType  string          `json:"log_type"`
	HWID     string          `json:"hwid"`
	Hostname string          `json:"hostname"`
	Data     json.RawMessage `json:"data"`
}

// PushDataHandler: Cổng tiếp nhận duy nhất
func PushDataHandler(c *gin.Context) {
	// 1. Bind dữ liệu
	var req agentService.AgentPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// 2. Xác thực danh tính thiết bị
	var agent models.Agent
	if err := database.DB.Where("hw_id = ?", req.HWID).First(&agent).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Thiết bị không tồn tại"})
		return
	}

	// Tại đây bạn có thể thêm logic kiểm tra HMAC Signature nếu cần
	// signature := c.GetHeader("X-Sent-Signature")
	// verifyHMAC(req.Data, agent.SecretKey, signature)

	// 3. Quăng dữ liệu vào hàng đợi bất đồng bộ (Goroutine) cho Bộ não xử lý
	go agentService.ProcessAgentData(agentService.AgentPayload{
		Type:     "DATA",
		LogType:  req.LogType,
		HWID:     req.HWID,
		Hostname: req.Hostname,
		Data:     req.Data,
	})

	// 4. Phản hồi ngay lập tức cho Agent
	c.JSON(http.StatusOK, gin.H{"message": "Dữ liệu đã được tiếp nhận", "status": agent.Status})
}
