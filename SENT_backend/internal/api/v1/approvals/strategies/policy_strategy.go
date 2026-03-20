package strategies

import (
	"sent_backend/internal/models"

	"gorm.io/gorm"
)

type PolicyCreateStrategy struct{}

func (s *PolicyCreateStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Policy dùng TargetID (ID số nguyên)
	return tx.Model(&models.UniversalPolicy{}).Where("id = ?", ticket.TargetID).Update("approval_status", "APPROVED").Error
}

func (s *PolicyCreateStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return tx.Model(&models.UniversalPolicy{}).Where("id = ?", ticket.TargetID).Update("approval_status", "REJECTED").Error
}
