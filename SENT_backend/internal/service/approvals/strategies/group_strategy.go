// internal/service/approvals/strategies/group_strategy.go

package strategies

import (
	"sent_backend/internal/models"

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
