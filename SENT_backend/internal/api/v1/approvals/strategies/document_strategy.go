package strategies

import (
	"sent_backend/internal/models"

	"gorm.io/gorm"
)

type DocumentUploadStrategy struct{}

func (s *DocumentUploadStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return tx.Model(&models.PolicyDocument{}).Where("id = ?", ticket.TargetID).Update("approval_status", "APPROVED").Error
}

func (s *DocumentUploadStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return tx.Model(&models.PolicyDocument{}).Where("id = ?", ticket.TargetID).Update("approval_status", "REJECTED").Error
}
