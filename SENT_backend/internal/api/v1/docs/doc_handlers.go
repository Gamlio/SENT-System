package docs

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/security"

	"github.com/gin-gonic/gin"
)

// UploadDocument: Chuyên xử lý file PDF/Word quy định công ty (CÓ TẠO VÉ DUYỆT)
func UploadDocument(c *gin.Context) {
	title := c.PostForm("title")
	category := c.PostForm("category")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "Không tìm thấy file"})
		return
	}

	// 1. Xử lý lưu file vật lý vào thư mục
	uploadDir := filepath.Join("uploads", "policies")
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(500, gin.H{"error": "Không thể tạo thư mục lưu trữ"})
		return
	}

	safeFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	filePath := filepath.Join(uploadDir, safeFileName)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(500, gin.H{"error": "Lỗi lưu file vật lý: " + err.Error()})
		return
	}

	orgID := uint(1)               // Tạm hardcode, sau này lấy từ JWT
	uploaderName := "System_Admin" // Tạm hardcode

	// Bắt đầu Transaction Database
	tx := database.DB.Begin()

	// 2. Lưu vào DB với trạng thái PENDING
	doc := models.PolicyDocument{
		OrgID:          orgID,
		Title:          title,
		FileName:       safeFileName,
		FilePath:       filepath.ToSlash(filePath),
		Category:       category,
		IsProcessed:    false,
		ApprovalStatus: "PENDING", // <--- KHÓA LẠI CHỜ DUYỆT
		UploadedBy:     uploaderName,
	}

	if err := tx.Create(&doc).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Lỗi lưu Database"})
		return
	}

	// 3. ĐẺ RA VÉ CHỜ DUYỆT
	ticket := models.ApprovalTicket{
		OrgID:        orgID,
		ModuleType:   "DOCUMENT_UPLOAD",
		ActionType:   "CREATE",
		TargetID:     doc.ID,
		TargetName:   fmt.Sprintf("[%s] %s", category, title),
		Status:       "PENDING",
		RequestedBy:  uploaderName,
		SnapshotData: fmt.Sprintf(`{"File": "%s"}`, file.Filename),
	}

	if err := tx.Create(&ticket).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Lỗi tạo vé phê duyệt"})
		return
	}

	tx.Commit()
	c.JSON(200, gin.H{"message": "Tải tài liệu thành công. Đang chờ Admin phê duyệt."})
}

// GetDocuments: Lấy danh sách tài liệu (Hỗ trợ lọc theo trạng thái)
func GetDocuments(c *gin.Context) {
	orgID := uint(1)
	status := c.Query("status") // Hỗ trợ UI lấy riêng file PENDING hoặc APPROVED

	var docs []models.PolicyDocument
	query := database.DB.Where("org_id = ?", orgID)

	if status != "" {
		query = query.Where("approval_status = ?", status)
	}

	if err := database.DB.Where("approval_status = ?", "APPROVED").Find(&docs).Error; err != nil {
		c.JSON(500, gin.H{"error": "Lỗi truy xuất tài liệu"})
		return
	}

	c.JSON(200, docs)
}

// DeleteDocument: Xóa tài liệu
func DeleteDocument(c *gin.Context) {
	id := c.Param("id")
	orgID := uint(1)
	if err := security.DeletePolicyAndFile(id, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa file"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Xóa thành công"})
}
func UpdateDocument(c *gin.Context) {
	id := c.Param("id")
	title := c.PostForm("title")
	category := c.PostForm("category")

	var doc models.PolicyDocument
	if err := database.DB.First(&doc, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy tài liệu"})
		return
	}

	// Bắt đầu Transaction
	tx := database.DB.Begin()

	// 1. Xử lý nếu người dùng có đính kèm file MỚI để thay thế file CŨ
	file, err := c.FormFile("file")
	if err == nil && file != nil {
		// Xóa file cũ khỏi ổ cứng để tiết kiệm dung lượng
		os.Remove(doc.FilePath)

		// Lưu file mới
		uploadDir := filepath.Join("uploads", "policies")
		os.MkdirAll(uploadDir, os.ModePerm)
		safeFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
		filePath := filepath.Join(uploadDir, safeFileName)

		if err := c.SaveUploadedFile(file, filePath); err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Lỗi lưu file vật lý mới"})
			return
		}

		// Cập nhật đường dẫn mới
		doc.FileName = safeFileName
		doc.FilePath = filepath.ToSlash(filePath)
		doc.IsProcessed = false // Đánh dấu false để AI phải đọc (train) lại file mới này
	}

	// 2. Cập nhật thông tin và KHÓA TÀI LIỆU LẠI (PENDING)
	doc.Title = title
	doc.Category = category
	doc.ApprovalStatus = "PENDING"
	doc.UploadedBy = "System_Admin" // Tạm hardcode, sau lấy từ JWT

	if err := tx.Save(&doc).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Lỗi lưu cập nhật vào Database"})
		return
	}

	// 3. TẠO VÉ DUYỆT MỚI CHO SỰ KIỆN UPDATE NÀY
	ticket := models.ApprovalTicket{
		OrgID:        doc.OrgID,
		ModuleType:   "DOCUMENT_UPLOAD", // Giữ nguyên key này để Module Duyệt đơn xử lý tự động được
		ActionType:   "UPDATE",
		TargetID:     doc.ID,
		TargetName:   fmt.Sprintf("[%s] %s (Bản Cập Nhật)", doc.Category, doc.Title),
		Status:       "PENDING",
		RequestedBy:  doc.UploadedBy,
		SnapshotData: fmt.Sprintf(`{"File": "%s", "Action": "Thay đổi nội dung/tên"}`, doc.FileName),
	}

	if err := tx.Create(&ticket).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Lỗi tạo vé phê duyệt cho bản cập nhật"})
		return
	}

	tx.Commit()
	c.JSON(200, gin.H{"message": "Cập nhật tài liệu thành công. Vui lòng chờ Admin phê duyệt lại."})
}
