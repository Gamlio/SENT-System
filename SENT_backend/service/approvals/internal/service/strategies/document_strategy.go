package strategies

import (
	"SENT_backend/pkg/models"
	"encoding/json"
	"os"

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
	// Cập nhật bản ghi nháp thành APPROVED và dọn dẹp tệp cũ nếu cần
	if ticket.ActionType == "UPDATE" {
		var snapshot struct {
			OldFilePath string `json:"old_file_path"`
			NewFilePath string `json:"new_file_path"`
		}
		if err := json.Unmarshal([]byte(ticket.SnapshotData), &snapshot); err == nil {
			if snapshot.OldFilePath != "" && snapshot.NewFilePath != "" && snapshot.OldFilePath != snapshot.NewFilePath {
				_ = os.Remove(snapshot.OldFilePath)
			}
		}
	}

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

	// Nếu từ chối cập nhật: Khôi phục lại bản ghi cũ và trạng thái APPROVED
	if ticket.ActionType == "UPDATE" {
		var snapshot struct {
			OldTitle        string `json:"old_title"`
			OldCategory     string `json:"old_category"`
			OldFileName     string `json:"old_file_name"`
			OldFilePath     string `json:"old_file_path"`
			OldOriginalName string `json:"old_original_name"`
			OldDisplayPdf   string `json:"old_display_pdf"`
			NewFilePath     string `json:"new_file_path"`
		}
		if err := json.Unmarshal([]byte(ticket.SnapshotData), &snapshot); err != nil {
			return tx.Model(&models.Document{}).Where("id = ?", ticket.TargetID).Updates(map[string]interface{}{
				"approval_status": "APPROVED",
			}).Error
		}

		updates := map[string]interface{}{
			"title":           snapshot.OldTitle,
			"category":        snapshot.OldCategory,
			"approval_status": "APPROVED",
		}
		if snapshot.OldFileName != "" {
			updates["file_name"] = snapshot.OldFileName
		}
		if snapshot.OldFilePath != "" {
			updates["file_path"] = snapshot.OldFilePath
		}
		if snapshot.OldOriginalName != "" {
			updates["original_name"] = snapshot.OldOriginalName
		}
		if snapshot.OldDisplayPdf != "" {
			updates["display_pdf_path"] = snapshot.OldDisplayPdf
		}

		if err := tx.Model(&models.Document{}).Where("id = ?", ticket.TargetID).Updates(updates).Error; err != nil {
			return err
		}

		if snapshot.NewFilePath != "" && snapshot.NewFilePath != snapshot.OldFilePath {
			_ = os.Remove(snapshot.NewFilePath)
		}
		return nil
	}

	return nil
}
