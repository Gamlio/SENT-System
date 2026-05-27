package documents

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type DocumentService struct{}

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
		go s.convertToPDF(orgID, doc.ID, filePath, displayPdfPath)
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

	oldDoc := doc // Preserve old values for rollback
	return database.DB.Transaction(func(tx *gorm.DB) error {
		snapshot := map[string]interface{}{
			"action":            "update_content_or_file",
			"old_title":         oldDoc.Title,
			"old_category":      oldDoc.Category,
			"old_file_name":     oldDoc.FileName,
			"old_file_path":     oldDoc.FilePath,
			"old_original_name": oldDoc.OriginalName,
			"old_display_pdf":   oldDoc.DisplayPdfPath,
			"new_title":         title,
			"new_category":      category,
		}

		if file != nil {
			ext := strings.ToLower(filepath.Ext(fileName))
			if ext != ".doc" && ext != ".docx" {
				return fmt.Errorf("hệ thống chỉ chấp nhận định dạng Microsoft Word (.doc, .docx)")
			}

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
			if _, err := io.Copy(out, file); err != nil {
				return err
			}

			displayPdfName := strings.TrimSuffix(safeFileName, ext) + ".pdf"
			displayPdfPath := filepath.ToSlash(filepath.Join(pdfDir, displayPdfName))

			go s.convertToPDF(orgID, doc.ID, filePath, displayPdfPath)

			doc.FileName = safeFileName
			doc.FilePath = filepath.ToSlash(filePath)
			doc.OriginalName = fileName
			doc.DisplayPdfPath = displayPdfPath
			snapshot["new_file_name"] = safeFileName
			snapshot["new_file_path"] = filepath.ToSlash(filePath)
			snapshot["new_original_name"] = fileName
			snapshot["new_display_pdf"] = displayPdfPath
		}

		doc.Title = title
		doc.Category = category
		doc.ApprovalStatus = "PENDING" // Khóa lại chờ duyệt
		doc.UploadedBy = requester

		if err := tx.Save(&doc).Error; err != nil {
			return err
		}

		snapshot["doc_id"] = doc.ID
		snapshotBytes, _ := json.Marshal(snapshot)

		ticket := models.ApprovalTicket{
			OrgID:         orgID,
			ModuleType:    "DOCUMENT_UPLOAD",
			ActionType:    "UPDATE",
			TargetID:      doc.ID,
			TargetName:    fmt.Sprintf("[%s] %s (Cập nhật)", category, title),
			Status:        "PENDING",
			RequestedBy:   requester,
			RequestReason: reason,
			SnapshotData:  string(snapshotBytes),
		}
		return tx.Create(&ticket).Error
	})
}

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
func (s *DocumentService) convertToPDF(orgID uint, docID uint, wordPath, pdfPath string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	file, _ := os.Open(wordPath)
	defer file.Close()
	part, _ := writer.CreateFormFile("files", filepath.Base(wordPath))
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", "http://gotenberg:3000/forms/libreoffice/convert", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("🔥 Lỗi kết nối Gotenberg: %v\n", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("🔥 Gotenberg trả về lỗi: %d\n", resp.StatusCode)
		return fmt.Errorf("conversion failed")
	}
	out, _ := os.Create(pdfPath)
	defer out.Close()
	_, err = io.Copy(out, resp.Body)

	if err == nil {
		// go s.extractAndStoreText(orgID, docID, pdfPath)

		go s.extractAndStoreMarkdown(orgID, docID, wordPath)
	}

	return err
}
func (s *DocumentService) extractAndStoreText(orgID uint, docID uint, pdfPath string) error {
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var content strings.Builder
	totalPage := r.NumPage()

	for i := 1; i <= totalPage; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, _ := p.GetPlainText(nil)
		content.WriteString(text)
	}
	if database.DocumentContentCollection != nil {
		filter := bson.M{"doc_id": docID, "org_id": int64(orgID)}
		update := bson.M{"$set": bson.M{
			"doc_id":     docID,
			"org_id":     int64(orgID),
			"content":    content.String(),
			"updated_at": time.Now(),
		}}
		_, err = database.DocumentContentCollection.UpdateOne(
			context.TODO(), filter, update, options.Update().SetUpsert(true),
		)
	}
	return err
}
func (s *DocumentService) extractAndStoreMarkdown(orgID uint, docID uint, wordPath string) error {
	var err error

	mdPath := strings.TrimSuffix(wordPath, filepath.Ext(wordPath)) + ".md"

	cmd := exec.Command("pandoc", wordPath, "-f", "docx", "-t", "markdown", "-o", mdPath)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("🔥 Lỗi chạy Pandoc: %v\n", err)
		return err
	}

	var mdContent []byte
	mdContent, err = os.ReadFile(mdPath)
	if err != nil {
		return err
	}

	if database.DocumentContentCollection != nil {
		filter := bson.M{"doc_id": docID, "org_id": int64(orgID)}
		update := bson.M{"$set": bson.M{
			"doc_id":     docID,
			"org_id":     int64(orgID),
			"content":    string(mdContent),
			"updated_at": time.Now(),
		}}

		_, err = database.DocumentContentCollection.UpdateOne(
			context.TODO(), filter, update, options.Update().SetUpsert(true),
		)
	}

	return err
}
