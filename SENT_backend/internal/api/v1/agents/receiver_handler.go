package agents

import (
	"encoding/json"
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	agentService "sent_backend/internal/service/agents"
	"sent_backend/internal/service/scoring"

	"github.com/gin-gonic/gin"
)

// Payload đón dữ liệu thô
type AgentPayload struct {
	LogType  string          `json:"log_type" binding:"required,max=50,alphanum"`
	HWID     string          `json:"hwid" binding:"required,max=64,alphanum"`
	Hostname string          `json:"hostname" binding:"required,min=1,max=255"`
	Data     json.RawMessage `json:"data" binding:"required"`
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
func UpdateDepartment(c *gin.Context) {
	hwid := c.Param("hwid")
	var req struct {
		DepartmentTag string `json:"department_tag" binding:"required,min=1,max=50,alphanum-"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	// Lưu xuống DB
	if err := database.DB.Model(&models.Agent{}).Where("hw_id = ?", hwid).Update("department_tag", req.DepartmentTag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật phòng ban"})
		return
	}

	// [TÙY CHỌN] Tính lại điểm rủi ro ngay lập tức vì đổi ngữ cảnh có thể làm thay đổi P1/P4
	scoring.RecalculateRiskScore(hwid)

	c.JSON(http.StatusOK, gin.H{"message": "Đã cập nhật Nhãn bảo mật (Department)"})
}
