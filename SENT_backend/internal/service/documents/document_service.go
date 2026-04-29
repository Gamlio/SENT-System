package documents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DocumentService struct{}

// CreateUploadRequest: Xử lý lưu file và tạo đơn phê duyệt mới
func (s *DocumentService) CreateUploadRequest(orgID uint, title, category, fileName string, file io.Reader, fileSize int64, uploader string) error {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".doc" && ext != ".docx" {
		return fmt.Errorf("hệ thống chỉ chấp nhận định dạng Microsoft Word (.doc, .docx)")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		uploadDir := filepath.Join("uploads", "documents")
		pdfDir := filepath.Join("uploads", "pdf_previews")
		os.MkdirAll(uploadDir, os.ModePerm)
		os.MkdirAll(pdfDir, os.ModePerm)

		// 2. Lưu file vật lý với tên an toàn (để tránh trùng)
		safeFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		filePath := filepath.Join(uploadDir, safeFileName)

		out, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer out.Close()
		if _, err := io.Copy(out, file); err != nil {
			return err
		}

		// 3. Xử lý chuyển đổi PDF (Placeholder cho Worker)
		displayPdfName := strings.TrimSuffix(safeFileName, ext) + ".pdf"
		displayPdfPath := filepath.ToSlash(filepath.Join(pdfDir, displayPdfName))

		// 4. Tạo Record nháp (PENDING)
		doc := models.Document{
			OrgID:          orgID,
			Title:          title,
			FileName:       safeFileName,
			FilePath:       filepath.ToSlash(filePath),
			Category:       category,
			OriginalName:   fileName,
			DisplayPdfPath: displayPdfPath,
			ApprovalStatus: "PENDING",
			UploadedBy:     uploader,
		}
		if err := tx.Create(&doc).Error; err != nil {
			return err
		}
		go s.convertToPDF(filePath, displayPdfPath)
		// 5. Đẻ vé phê duyệt
		ticket := models.ApprovalTicket{
			OrgID:       orgID,
			ModuleType:  "DOCUMENT_UPLOAD",
			ActionType:  "CREATE",
			TargetID:    doc.ID,
			TargetName:  fmt.Sprintf("[%s] %s", category, title),
			Status:      "PENDING",
			RequestedBy: uploader,
		}

		payload := map[string]interface{}{
			"title":            title,
			"category":         category,
			"file_name":        safeFileName,
			"file_path":        filepath.ToSlash(filePath),
			"original_name":    fileName,
			"display_pdf_path": displayPdfPath,
			"size":             fileSize,
		}
		snapshotBytes, _ := json.Marshal(payload)
		ticket.SnapshotData = string(snapshotBytes)

		return tx.Create(&ticket).Error
	})
}

// CreateUpdateRequest: Tạo đơn yêu cầu cập nhật nội dung/file
func (s *DocumentService) CreateUpdateRequest(docID uint, orgID uint, title, category string, fileName string, file io.Reader, reason string, requester string) error {
	var doc models.Document
	if err := database.DB.Where("id = ? AND org_id = ?", docID, orgID).First(&doc).Error; err != nil {
		return fmt.Errorf("không tìm thấy tài liệu")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Nếu có file mới đi kèm, thực hiện ghi đè và dọn dẹp file cũ
		if file != nil {
			ext := strings.ToLower(filepath.Ext(fileName))
			if ext != ".doc" && ext != ".docx" {
				return fmt.Errorf("hệ thống chỉ chấp nhận định dạng Microsoft Word (.doc, .docx)")
			}

			os.Remove(doc.FilePath) // Xóa file cũ

			uploadDir := filepath.Join("uploads", "documents")
			pdfDir := filepath.Join("uploads", "pdf_previews")
			os.MkdirAll(uploadDir, os.ModePerm)
			os.MkdirAll(pdfDir, os.ModePerm)

			safeFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
			filePath := filepath.Join(uploadDir, safeFileName)

			out, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer out.Close()
			io.Copy(out, file)

			displayPdfName := strings.TrimSuffix(safeFileName, ext) + ".pdf"
			displayPdfPath := filepath.ToSlash(filepath.Join(pdfDir, displayPdfName))

			doc.FileName = safeFileName
			doc.FilePath = filepath.ToSlash(filePath)
			doc.OriginalName = fileName
			doc.DisplayPdfPath = displayPdfPath
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
			OrgID:         orgID,
			ModuleType:    "DOCUMENT_UPLOAD",
			ActionType:    "UPDATE",
			TargetID:      doc.ID,
			TargetName:    fmt.Sprintf("[%s] %s (Cập nhật)", category, title),
			Status:        "PENDING",
			RequestedBy:   requester,
			RequestReason: reason,
			SnapshotData:  `{"action": "update_content_or_file"}`,
		}
		return tx.Create(&ticket).Error
	})
}

// CreateDeleteRequest: Tạo đơn yêu cầu xóa tài liệu kèm lý do (Zero Trust)
func (s *DocumentService) CreateDeleteRequest(docID uint, orgID uint, requester string, reason string) error {
	var doc models.Document
	if err := database.DB.Where("id = ? AND org_id = ?", docID, orgID).First(&doc).Error; err != nil {
		return fmt.Errorf("không tìm thấy tài liệu")
	}

	ticket := models.ApprovalTicket{
		OrgID:         orgID,
		ModuleType:    "DOCUMENT_DELETE",
		ActionType:    "DELETE",
		TargetID:      doc.ID,
		TargetName:    fmt.Sprintf("Xóa tài liệu: %s", doc.Title),
		Status:        "PENDING",
		RequestedBy:   requester,
		RequestReason: reason, // [BỔ SUNG] Lưu lý do xóa vào đơn
		SnapshotData:  fmt.Sprintf(`{"id": %d, "title": "%s"}`, doc.ID, doc.Title),
	}
	return database.DB.Create(&ticket).Error
}

// GetDocumentByID: Lấy chi tiết tài liệu theo ID và OrgID
func (s *DocumentService) GetDocumentByID(id uint, orgID uint) (models.Document, error) {
	var doc models.Document
	err := database.DB.Where("id = ? AND org_id = ?", id, orgID).First(&doc).Error
	return doc, err
}

// GetDocuments: Lấy danh sách tài liệu theo Org và trạng thái
func (s *DocumentService) GetDocuments(orgID uint, status string) ([]models.Document, error) {
	var docs []models.Document
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
func (s *DocumentService) convertToPDF(wordPath, pdfPath string) error {
	// 1. Chuẩn bị request gửi sang Gotenberg
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Mở file Word gốc
	file, _ := os.Open(wordPath)
	defer file.Close()

	part, _ := writer.CreateFormFile("files", filepath.Base(wordPath))
	io.Copy(part, file)
	writer.Close()

	// 2. Gọi API Gotenberg (sử dụng tên service trong Docker là 'gotenberg')
	req, _ := http.NewRequest("POST", "http://gotenberg:3000/forms/libreoffice/convert", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 3. Lưu file PDF nhận được vào thư mục pdf_previews
	out, _ := os.Create(pdfPath)
	defer out.Close()
	_, err = io.Copy(out, resp.Body)

	return err
}
