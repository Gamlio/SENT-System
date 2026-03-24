package strategies

import (
	"encoding/json"
	"sent_backend/internal/models"
	"time"

	"gorm.io/gorm"
)

// --- CHIẾN LƯỢC: ĐĂNG KÝ AGENT ---
type AgentEnrollStrategy struct{}

func (s *AgentEnrollStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return tx.Model(&models.Agent{}).Where("hw_id = ?", ticket.TargetName).
		Updates(map[string]interface{}{
			"status":    "ACTIVE",
			"last_seen": time.Now(), // <--- BUMP NÓ LÊN TRÊN CÙNG
		}).Error
}
func (s *AgentEnrollStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return tx.Model(&models.Agent{}).Where("hw_id = ?", ticket.TargetName).
		Updates(map[string]interface{}{
			"status":    "REJECTED",
			"last_seen": time.Now(), // <--- BUMP NÓ LÊN TRÊN CÙNG
		}).Error
}

// ==============================================================
// --- CHIẾN LƯỢC: XÓA 1 AGENT (XÓA MỀM / SOFT DELETE) ---
// ==============================================================
type AgentDeleteStrategy struct{}

func (s *AgentDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// XÓA MỀM: Đổi trạng thái thành RETIRED, xóa SecretKey để chặn kết nối vĩnh viễn
	return tx.Model(&models.Agent{}).Where("hw_id = ?", ticket.TargetName).
		Updates(map[string]interface{}{
			"status":     "RETIRED",
			"secret_key": "", // Thu hồi khóa, vứt bỏ quyền truy cập
		}).Error
}

func (s *AgentDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// BỊ TỪ CHỐI XÓA: Khôi phục máy trạm từ PENDING_DELETE về trạng thái ACTIVE
	return tx.Model(&models.Agent{}).Where("hw_id = ?", ticket.TargetName).Update("status", "ACTIVE").Error
}

// ==============================================================
// --- CHIẾN LƯỢC: XÓA AGENT HÀNG LOẠT (SOFT DELETE) ---
// ==============================================================
type AgentBulkDeleteStrategy struct{}

// Struct dùng để giải mã SnapshotData
type BulkDeleteSnapshot struct {
	HWIDs  []string `json:"hwids"`
	Reason string   `json:"reason"`
}

func (s *AgentBulkDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	var snap BulkDeleteSnapshot
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &snap); err != nil {
		return err
	}

	// XÓA MỀM HÀNG LOẠT: Đổi trạng thái và thu hồi khóa của hàng trăm máy trong 1 nốt nhạc
	return tx.Model(&models.Agent{}).Where("hw_id IN ?", snap.HWIDs).
		Updates(map[string]interface{}{
			"status":     "RETIRED",
			"secret_key": "",
		}).Error
}

func (s *AgentBulkDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	var snap BulkDeleteSnapshot
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &snap); err != nil {
		return err
	}

	// BỊ TỪ CHỐI XÓA HÀNG LOẠT: Khôi phục toàn bộ danh sách máy về ACTIVE
	return tx.Model(&models.Agent{}).Where("hw_id IN ?", snap.HWIDs).Update("status", "ACTIVE").Error
}
