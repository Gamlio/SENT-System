package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	docService "SENT_backend/service/documents/internal/service" // Alias
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func UploadDocument(c *gin.Context) {
	title := c.PostForm("title")
	category := c.PostForm("category")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "Không tìm thấy file"})
		return
	}

	// Mở stream file để chuyển xuống Service (Không đọc toàn bộ vào RAM)
	f, _ := file.Open()
	defer f.Close()

	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	svc := &docService.DocumentService{}
	if err := svc.CreateUploadRequest(orgID, title, category, file.Filename, f, file.Size, username.(string)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Đã gửi yêu cầu tải tài liệu, vui lòng chờ duyệt."})
}

func UpdateDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	title := c.PostForm("title")
	category := c.PostForm("category")
	reason := c.PostForm("reason")
	var fileReader io.Reader
	var fileName string
	file, err := c.FormFile("file")
	if err == nil {
		f, _ := file.Open()
		fileReader = f
		fileName = file.Filename
		defer f.Close()
	}

	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	svc := &docService.DocumentService{}
	err = svc.CreateUpdateRequest(uint(id), orgID, title, category, fileName, fileReader, reason, username.(string))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Đã gửi yêu cầu cập nhật tài liệu."})
}
func GetDocuments(c *gin.Context) {
	orgID := c.GetUint("org_id")
	status := c.Query("status")

	svc := &docService.DocumentService{}
	list, err := svc.GetDocuments(orgID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách tài liệu"})
		return
	}

	c.JSON(http.StatusOK, list)
}
func DeleteDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Vui lòng cung cấp lý do xóa"})
		return
	}

	svc := &docService.DocumentService{}
	if err := svc.CreateDeleteRequest(uint(id), orgID, username.(string), req.Reason); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Đã gửi yêu cầu xóa tài liệu kèm lý do."})
}

func DownloadDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	orgID := c.GetUint("org_id")

	svc := &docService.DocumentService{}
	doc, err := svc.GetDocumentByID(uint(id), orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tài liệu đã bốc hơi hoặc bạn không có quyền"})
		return
	}

	c.FileAttachment(doc.FilePath, doc.OriginalName)
}
func ApproveDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	// Cập nhật trạng thái phê duyệt trong PostgreSQL
	err := database.DB.Model(&models.Document{}).
		Where("id = ? AND org_id = ?", uint(id), orgID).
		Updates(map[string]interface{}{
			"approval_status": "APPROVED",
			"approved_by":     username,
		}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi phê duyệt tài liệu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tài liệu đã được phê duyệt và sẵn sàng sử dụng"})
}
