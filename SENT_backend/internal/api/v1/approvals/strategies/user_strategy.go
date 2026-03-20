package strategies

import (
	"sent_backend/internal/models"

	"gorm.io/gorm"
)

type UserCreateStrategy struct{}

func (s *UserCreateStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return tx.Model(&models.User{}).Where("id = ?", ticket.TargetID).Update("approval_status", "APPROVED").Error
}

func (s *UserCreateStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return tx.Model(&models.User{}).Where("id = ?", ticket.TargetID).Update("approval_status", "REJECTED").Error
}
