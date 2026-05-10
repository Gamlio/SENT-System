// pkg/service/approvals/strategies/group_strategy.go

package strategies

import (
	"SENT_backend/pkg/models"
	"encoding/json"

	"gorm.io/gorm"
)

type GroupDeleteStrategy struct{}

func (s *GroupDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// 1. Cập nhật người xóa vào bản ghi trước khi soft-delete
	tx.Model(&models.PolicyGroup{}).Where("id = ?", ticket.TargetID).
		Update("deleted_by", ticket.ReviewedBy) // Người duyệt là người thực thi xóa cuối cùng

	// 2. Thực hiện xóa (Soft delete)
	return tx.Delete(&models.PolicyGroup{}, ticket.TargetID).Error
}

func (s *GroupDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Không làm gì cả nếu bị từ chối xóa
	return nil
}

type GroupCreateStrategy struct{}

func (s *GroupCreateStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Giải mã dữ liệu từ Snapshot đã lưu lúc làm đơn
	var payload models.PolicyGroupPayload
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &payload); err != nil {
		return err
	}

	// Thực hiện tạo mới bản ghi vào Database sau khi được duyệt
	newGroup := models.PolicyGroup{
		OrgID:       ticket.OrgID,
		Name:        payload.Name,
		Description: payload.Description,
	}
	return tx.Create(&newGroup).Error
}

func (s *GroupCreateStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return nil // Không cần làm gì nếu từ chối tạo mới
}

// --- 2. CHIẾN LƯỢC CẬP NHẬT NHÓM ---
type GroupUpdateStrategy struct{}

func (s *GroupUpdateStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	var payload models.PolicyGroupPayload
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &payload); err != nil {
		return err
	}

	// Cập nhật thông tin dựa trên TargetID của đơn
	return tx.Model(&models.PolicyGroup{}).Where("id = ? AND org_id = ?", ticket.TargetID, ticket.OrgID).
		Updates(map[string]interface{}{
			"name":        payload.Name,
			"description": payload.Description,
		}).Error
}

func (s *GroupUpdateStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return nil // Giữ nguyên thông tin cũ nếu từ chối cập nhật
}
