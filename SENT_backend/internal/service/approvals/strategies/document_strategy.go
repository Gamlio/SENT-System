package strategies

import (
	"encoding/json"
	"os"
	"sent_backend/internal/models"

	"gorm.io/gorm"
)

type DocumentUploadStrategy struct{}

type DocumentPayload struct {
	Title          string `json:"title"`
	FileName       string `json:"file_name"`
	FilePath       string `json:"file_path"`
	Category       string `json:"category"`
	OriginalPath   string `json:"original_path"`
	DisplayPdfPath string `json:"display_pdf_path"`
}

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
		if doc.OriginalPath != "" {
			_ = os.Remove(doc.OriginalPath)
		}
		if doc.DisplayPdfPath != "" {
			_ = os.Remove(doc.DisplayPdfPath)
		}
		// Xóa bản ghi trong DB
		return tx.Delete(&models.Document{}, ticket.TargetID).Error
	}

	// Trường hợp 2: Phê duyệt đơn TẢI LÊN mới
	// 1. Giải mã dữ liệu từ "Đơn" (Snapshot)
	var payload DocumentPayload
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &payload); err != nil {
		return err
	}

	// 2. Bây giờ mới thực sự tạo Document trong hệ thống
	newDoc := models.Document{
		Title:          payload.Title,
		FileName:       payload.FileName,
		FilePath:       payload.FilePath,
		Category:       payload.Category,
		OriginalPath:   payload.OriginalPath,
		DisplayPdfPath: payload.DisplayPdfPath,
		OrgID:          ticket.OrgID,
		IsProcessed:    true,
		ApprovalStatus: "APPROVED",
		UploadedBy:     ticket.RequestedBy,
		ApprovedBy:     ticket.ReviewedBy,
	}

	return tx.Create(&newDoc).Error
}

func (s *DocumentUploadStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Nếu từ chối xóa, tài liệu vẫn giữ nguyên
	if ticket.ActionType == "DELETE" || ticket.ModuleType == "DOCUMENT_DELETE" {
		return nil
	}

	// Nếu từ chối tải lên mới:
	// 1. Giải mã Snapshot để lấy đường dẫn file vật lý
	var payload DocumentPayload
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &payload); err == nil {
		// 2. Dọn dẹp rác (Xóa file vật lý đã upload tạm thời)
		if payload.FilePath != "" {
			_ = os.Remove(payload.FilePath)
		}
		if payload.OriginalPath != "" {
			_ = os.Remove(payload.OriginalPath)
		}
		if payload.DisplayPdfPath != "" {
			_ = os.Remove(payload.DisplayPdfPath)
		}
	}

	// Không cần cập nhật DB vì bản ghi Document chưa hề được tạo.
	return nil
}

