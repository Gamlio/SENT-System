package strategies

import (
	"os"
	"sent_backend/internal/models"

	"gorm.io/gorm"
)

type DocumentUploadStrategy struct{}

func (s *DocumentUploadStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Trường hợp 1: Phê duyệt đơn XÓA tài liệu
	// (Vẫn giữ logic xóa cũ vì tài liệu đang tồn tại)
	if ticket.ActionType == "DELETE" || ticket.ModuleType == "DOCUMENT_DELETE" {
		var doc models.Document
		if err := tx.First(&doc, ticket.TargetID).Error; err != nil {
			return err
		}
		// Xóa file vật lý trên server
		_ = os.Remove(doc.FilePath)
		if doc.DisplayPdfPath != "" {
			_ = os.Remove(doc.DisplayPdfPath)
		}
		// Xóa bản ghi trong DB
		return tx.Delete(&models.Document{}, ticket.TargetID).Error
	}

	// Trường hợp 2: Phê duyệt đơn TẢI LÊN/CẬP NHẬT
	// Cập nhật bản ghi nháp thành APPROVED thay vì tạo mới để tránh nhân đôi dữ liệu
	return tx.Model(&models.Document{}).Where("id = ?", ticket.TargetID).Updates(map[string]interface{}{
		"approval_status": "APPROVED",
		"approved_by":     ticket.ReviewedBy,
	}).Error
}

func (s *DocumentUploadStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Nếu từ chối xóa, tài liệu vẫn giữ nguyên
	if ticket.ActionType == "DELETE" || ticket.ModuleType == "DOCUMENT_DELETE" {
		return nil
	}

	// Nếu từ chối tải lên mới: Xóa bản ghi nháp và dọn dẹp file vật lý
	if ticket.ActionType == "CREATE" {
		var doc models.Document
		if err := tx.First(&doc, ticket.TargetID).Error; err == nil {
			if doc.FilePath != "" {
				_ = os.Remove(doc.FilePath)
			}
			if doc.DisplayPdfPath != "" {
				_ = os.Remove(doc.DisplayPdfPath)
			}
			return tx.Delete(&models.Document{}, ticket.TargetID).Error
		}
	}

	// Nếu từ chối cập nhật: Khôi phục lại trạng thái APPROVED
	if ticket.ActionType == "UPDATE" {
		return tx.Model(&models.Document{}).Where("id = ?", ticket.TargetID).Updates(map[string]interface{}{
			"approval_status": "APPROVED",
		}).Error
	}

	return nil
}
