package documents

import (
	"fmt"
	"os"
	"path/filepath"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"gorm.io/gorm"
)

type DocumentService struct{}

// CreateUploadRequest: Xử lý lưu file và tạo đơn phê duyệt mới
func (s *DocumentService) CreateUploadRequest(orgID uint, title, category, fileName string, fileContent []byte, uploader string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Lưu file vật lý
		uploadDir := filepath.Join("uploads", "policies")
		os.MkdirAll(uploadDir, os.ModePerm)

		safeFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), fileName)
		filePath := filepath.Join(uploadDir, safeFileName)
		if err := os.WriteFile(filePath, fileContent, 0644); err != nil {
			return err
		}

		// 2. Tạo Record nháp (PENDING)
		doc := models.PolicyDocument{
			OrgID:          orgID,
			Title:          title,
			FileName:       safeFileName,
			FilePath:       filepath.ToSlash(filePath),
			Category:       category,
			ApprovalStatus: "PENDING",
			UploadedBy:     uploader,
		}
		if err := tx.Create(&doc).Error; err != nil {
			return err
		}

		// 3. Đẻ vé phê duyệt
		ticket := models.ApprovalTicket{
			OrgID:        orgID,
			ModuleType:   "DOCUMENT_UPLOAD",
			ActionType:   "CREATE",
			TargetID:     doc.ID,
			TargetName:   fmt.Sprintf("[%s] %s", category, title),
			Status:       "PENDING",
			RequestedBy:  uploader,
			SnapshotData: fmt.Sprintf(`{"file": "%s", "size": %d}`, fileName, len(fileContent)),
		}
		return tx.Create(&ticket).Error
	})
}

// CreateUpdateRequest: Tạo đơn yêu cầu cập nhật nội dung/file
func (s *DocumentService) CreateUpdateRequest(docID uint, orgID uint, title, category string, fileName string, fileContent []byte, requester string) error {
	var doc models.PolicyDocument
	if err := database.DB.Where("id = ? AND org_id = ?", docID, orgID).First(&doc).Error; err != nil {
		return fmt.Errorf("không tìm thấy tài liệu")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Nếu có file mới đi kèm, thực hiện ghi đè và dọn dẹp file cũ
		if len(fileContent) > 0 {
			os.Remove(doc.FilePath) // Xóa file cũ

			uploadDir := filepath.Join("uploads", "policies")
			safeFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), fileName)
			filePath := filepath.Join(uploadDir, safeFileName)
			os.WriteFile(filePath, fileContent, 0644)

			doc.FileName = safeFileName
			doc.FilePath = filepath.ToSlash(filePath)
			doc.IsProcessed = false
		}

		doc.Title = title
		doc.Category = category
		doc.ApprovalStatus = "PENDING" // Khóa lại chờ duyệt
		doc.UploadedBy = requester

		if err := tx.Save(&doc).Error; err != nil {
			return err
		}

		// Tạo vé duyệt cho bản cập nhật
		ticket := models.ApprovalTicket{
			OrgID:        orgID,
			ModuleType:   "DOCUMENT_UPLOAD",
			ActionType:   "UPDATE",
			TargetID:     doc.ID,
			TargetName:   fmt.Sprintf("[%s] %s (Cập nhật)", category, title),
			Status:       "PENDING",
			RequestedBy:  requester,
			SnapshotData: `{"action": "update_content_or_file"}`,
		}
		return tx.Create(&ticket).Error
	})
}

// CreateDeleteRequest: Tạo đơn yêu cầu xóa tài liệu (Zero Trust)
func (s *DocumentService) CreateDeleteRequest(docID uint, orgID uint, requester string) error {
	var doc models.PolicyDocument
	if err := database.DB.Where("id = ? AND org_id = ?", docID, orgID).First(&doc).Error; err != nil {
		return fmt.Errorf("không tìm thấy tài liệu")
	}

	ticket := models.ApprovalTicket{
		OrgID:        orgID,
		ModuleType:   "DOCUMENT_DELETE",
		ActionType:   "DELETE",
		TargetID:     doc.ID,
		TargetName:   fmt.Sprintf("Xóa tài liệu: %s", doc.Title),
		Status:       "PENDING",
		RequestedBy:  requester,
		SnapshotData: fmt.Sprintf(`{"id": %d, "title": "%s"}`, doc.ID, doc.Title),
	}
	return database.DB.Create(&ticket).Error
}

// GetDocuments: Lấy danh sách tài liệu theo Org và trạng thái
func (s *DocumentService) GetDocuments(orgID uint, status string) ([]models.PolicyDocument, error) {
	var docs []models.PolicyDocument
	query := database.DB.Where("org_id = ?", orgID)

	if status != "" {
		query = query.Where("approval_status = ?", status)
	} else {
		// Mặc định nếu không truyền status thì chỉ lấy những cái đã được duyệt
		query = query.Where("approval_status = ?", "APPROVED")
	}

	err := query.Order("created_at desc").Find(&docs).Error
	return docs, err
}
