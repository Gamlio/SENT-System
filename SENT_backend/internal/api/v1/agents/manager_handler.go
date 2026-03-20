package agents

import (
	"encoding/json"
	"fmt"
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
	if err := database.DB.Preload("Manager").
		Where("status != ?", "RETIRED"). // <--- THÊM DÒNG NÀY
		Find(&agents).Error; err != nil {
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

// POST /api/v1/agents/:hwid/request-delete
func RequestDeleteAgent(c *gin.Context) {
	hwid := c.Param("hwid")

	// 1. ÉP KIỂU AN TOÀN CHO GORM
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không xác định được danh tính tổ chức"})
		return
	}
	orgID := orgIDVal.(uint) // QUAN TRỌNG: Ép về uint

	usernameVal, _ := c.Get("username")
	username := usernameVal.(string)

	// 2. Tìm Agent (Dùng orgID đã ép kiểu)
	var agent models.Agent
	if err := database.DB.Where("hw_id = ? AND org_id = ?", hwid, orgID).First(&agent).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy thiết bị này trong hệ thống!"})
		return
	}

	// 3. Kiểm tra xem đã có đơn xin xóa nào đang chờ chưa
	var existingTicket models.ApprovalTicket
	err := database.DB.Where("module_type = ? AND target_name = ? AND status = ?",
		"AGENT_DELETE", agent.HWID, "PENDING").First(&existingTicket).Error

	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiết bị này đang có đơn xin xóa chờ phê duyệt rồi!"})
		return
	}

	// 4. Tạo Đơn xin phê duyệt (Approval Ticket)
	ticket := models.ApprovalTicket{
		OrgID:        orgID,
		ModuleType:   "AGENT_DELETE",
		ActionType:   "DELETE",
		TargetID:     0,
		TargetName:   agent.HWID,
		Status:       "PENDING",
		RequestedBy:  username,
		SnapshotData: fmt.Sprintf(`{"hostname": "%s", "ip": "%s", "reason": "Admin yêu cầu gỡ bỏ thiết bị"}`, agent.Hostname, agent.IPAddress),
	}

	if err := database.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi tạo đơn yêu cầu"})
		return
	}

	// 5. Đổi trạng thái Agent thành "PENDING_DELETE"
	database.DB.Model(&agent).Update("status", "PENDING_DELETE")

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu gỡ bỏ thiết bị. Đang chờ SOC Admin phê duyệt."})
}
func RequestBulkDeleteAgents(c *gin.Context) {
	var req struct {
		HWIDs  []string `json:"hwids"`
		Reason string   `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu gửi lên không hợp lệ"})
		return
	}

	if len(req.HWIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Danh sách thiết bị trống"})
		return
	}

	orgIDVal, _ := c.Get("org_id")
	orgID := orgIDVal.(uint)
	usernameVal, _ := c.Get("username")

	// 1. Khóa tạm thời các máy này (Chuyển sang PENDING_DELETE)
	database.DB.Model(&models.Agent{}).
		Where("hw_id IN ? AND org_id = ?", req.HWIDs, orgID).
		Update("status", "PENDING_DELETE")

	// 2. Gói gọn danh sách HWID và Lý do vào SnapshotData
	snapshotData, _ := json.Marshal(map[string]interface{}{
		"hwids":  req.HWIDs,
		"reason": req.Reason,
	})

	// 3. Tạo Đơn xin phê duyệt DUY NHẤT
	ticket := models.ApprovalTicket{
		OrgID:        orgID,
		ModuleType:   "AGENT_BULK_DELETE",
		ActionType:   "DELETE",
		TargetID:     0,
		TargetName:   fmt.Sprintf("Hủy %d thiết bị", len(req.HWIDs)), // Hiển thị: "Hủy 299 thiết bị"
		Status:       "PENDING",
		RequestedBy:  usernameVal.(string),
		SnapshotData: string(snapshotData),
	}

	if err := database.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi tạo đơn xin xóa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Đã tạo đơn yêu cầu gỡ bỏ %d thiết bị.", len(req.HWIDs))})
}
