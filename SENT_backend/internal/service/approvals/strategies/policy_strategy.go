package strategies

import (
	"encoding/json"
	"errors"
	"sent_backend/internal/models"

	"gorm.io/gorm"
)

// --- 1. CHIẾN LƯỢC TẠO CHÍNH SÁCH ---
type PolicyCreateStrategy struct{}

type PolicyPayload struct {
	Title      string `json:"title"`
	Category   string `json:"category"`
	Value      string `json:"value"`
	PolicyType string `json:"policy_type"`
	GroupID    *uint  `json:"group_id"` // Thay thế hoàn toàn cho TargetType và TargetAssetHWIDs
}

func (s *PolicyCreateStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// 1. Kiểm tra tính toàn vẹn: Đảm bảo Ticket có thông tin người duyệt
	if ticket.ReviewedBy == "" {
		return errors.New("thiếu thông tin người phê duyệt chính sách")
	}

	// 2. Giải mã dữ liệu từ "Đơn" (Snapshot)
	var payload PolicyPayload
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &payload); err != nil {
		return err
	}

	// 3. Bây giờ mới thực sự tạo Policy trong hệ thống
	newPolicy := models.Policy{
		OrgID:          ticket.OrgID,
		Title:          payload.Title,
		Category:       payload.Category,
		Value:          payload.Value,
		PolicyType:     payload.PolicyType,
		GroupID:        payload.GroupID, // Sử dụng GroupID (Nếu null thì tự hiểu là Global Policy)
		IsActive:       true,
		ApprovalStatus: "APPROVED",
		CreatedBy:      ticket.RequestedBy,
		ApprovedBy:     ticket.ReviewedBy,
	}

	if err := tx.Create(&newPolicy).Error; err != nil {
		return err
	}

	return nil
}

func (s *PolicyCreateStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// THEO TƯ DUY MỚI: Hệ thống SẠCH.
	// Lúc tạo đơn (Ticket), ta chưa hề INSERT vào bảng Policy (không tạo rác PENDING).
	// Nên khi bị từ chối, ta KHÔNG CẦN LÀM GÌ trong bảng Policy cả.
	return nil
}

// --- 2. CHIẾN LƯỢC XÓA 1 CHÍNH SÁCH ---
type PolicyDeleteStrategy struct{}

func (s *PolicyDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Thực hiện xóa thật bản ghi trong DB khi được duyệt
	return tx.Delete(&models.Policy{}, ticket.TargetID).Error
}

func (s *PolicyDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Không làm gì cả, chính sách vẫn hoạt động bình thường
	return nil
}

// --- 3. CHIẾN LƯỢC XÓA NHIỀU CHÍNH SÁCH ---
type PolicyBulkDeleteStrategy struct{}

func (s *PolicyBulkDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	var data struct {
		IDs []uint `json:"ids"`
	}
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &data); err != nil {
		return err
	}
	// Xóa đồng loạt theo danh sách ID đã lưu trong vé
	return tx.Where("id IN ? AND org_id = ?", data.IDs, ticket.OrgID).Delete(&models.Policy{}).Error
}

func (s *PolicyBulkDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return nil
}
