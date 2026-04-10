package strategies

import (
	"os"
	"sent_backend/internal/models"

	"gorm.io/gorm"
)

type DocumentUploadStrategy struct{}

func (s *DocumentUploadStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Trường hợp 1: Phê duyệt đơn XÓA tài liệu
	if ticket.ModuleType == "DOCUMENT_DELETE" {
		var doc models.Document
		if err := tx.First(&doc, ticket.TargetID).Error; err != nil {
			return err
		}
		// Xóa file vật lý trên server
		_ = os.Remove(doc.FilePath)
		// Xóa bản ghi trong DB
		return tx.Delete(&models.Document{}, ticket.TargetID).Error
	}

	// Trường hợp 2: Phê duyệt đơn TẢI LÊN hoặc CẬP NHẬT
	return tx.Model(&models.Document{}).
		Where("id = ? AND org_id = ?", ticket.TargetID, ticket.OrgID).
		Update("approval_status", "APPROVED").Error
}

func (s *DocumentUploadStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Nếu từ chối xóa, tài liệu vẫn giữ nguyên trạng thái cũ (APPROVED)
	if ticket.ModuleType == "DOCUMENT_DELETE" {
		return nil
	}

	// Nếu từ chối tải lên mới, đánh dấu REJECTED
	return tx.Model(&models.Document{}).
		Where("id = ? AND org_id = ?", ticket.TargetID, ticket.OrgID).
		Update("approval_status", "REJECTED").Error
}
