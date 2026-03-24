package agents

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/scoring"
	"sent_backend/internal/websocket"
	"time"

	"gorm.io/gorm"
)

type AgentLifecycleService struct{}

// EnrollAgentRequest: Xử lý đăng ký máy và tạo vé duyệt
func (s *AgentLifecycleService) EnrollAgentRequest(req models.EnrollRequest, orgID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var agent models.Agent
		err := tx.Where("hw_id = ?", req.HWID).First(&agent).Error

		if err == nil {
			// Máy cũ: Chuyển về PENDING để kiểm tra lại
			tx.Model(&agent).Updates(map[string]interface{}{
				"status": "PENDING", "hostname": req.Hostname, "ip_address": req.IPAddress,
			})
		} else {
			// Máy mới: Tạo mới hoàn toàn
			agent = models.Agent{
				HWID: req.HWID, Hostname: req.Hostname, IPAddress: req.IPAddress,
				OrgID: orgID, Status: "PENDING", LastSeen: time.Now(),
			}
			tx.Create(&agent)
		}

		// Tạo đơn phê duyệt
		ticket := models.ApprovalTicket{
			OrgID: orgID, ModuleType: "AGENT_ENROLL", ActionType: "ENROLL",
			TargetName: agent.HWID, Status: "PENDING", RequestedBy: "System_Enroll",
		}
		return tx.Create(&ticket).Error
	})
}

// CreateBulkDeleteRequest: Logic Maker-Checker cho việc xóa máy
func (s *AgentLifecycleService) CreateBulkDeleteRequest(hwids []string, orgID uint, reason string, requester string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Chuyển trạng thái máy sang PENDING_DELETE trên giao diện
		tx.Model(&models.Agent{}).Where("hw_id IN ? AND org_id = ?", hwids, orgID).Update("status", "PENDING_DELETE")

		// 2. Tạo Ticket
		snap, _ := json.Marshal(map[string]interface{}{"hwids": hwids, "reason": reason})
		ticket := models.ApprovalTicket{
			OrgID: orgID, ModuleType: "AGENT_BULK_DELETE", ActionType: "DELETE",
			TargetName: fmt.Sprintf("Xóa %d máy trạm", len(hwids)),
			Status:     "PENDING", RequestedBy: requester, SnapshotData: string(snap),
		}

		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}

		// 3. Thông báo Real-time cho Admin qua WebSocket
		websocket.GlobalHub.Broadcast(map[string]interface{}{"type": "REFRESH_AGENT_LIST"})
		return nil
	})
}
func (s *AgentLifecycleService) UpdateDeviceType(hwid string, deviceType string) error {
	// 1. Cập nhật thông tin trong Database
	err := database.DB.Model(&models.Agent{}).
		Where("hw_id = ?", hwid).
		Update("device_type", deviceType).Error

	if err != nil {
		return err
	}

	// 2. TỰ ĐỘNG: Tính lại điểm rủi ro ngay vì DeviceType làm thay đổi trọng số tài sản
	// Chúng ta dùng Goroutine để không làm chậm phản hồi của API
	go scoring.RecalculateRiskScore(hwid)

	return nil
}

// AssignManager: Gán nhân sự phụ trách máy
func (s *AgentLifecycleService) AssignManager(hwid string, orgID uint, userID uint) error {
	return database.DB.Model(&models.Agent{}).
		Where("hw_id = ? AND org_id = ?", hwid, orgID).
		Update("user_id", userID).Error
}
