package docs

import (
	"io"
	"net/http"
	docService "sent_backend/internal/service/documents" // Alias
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

	// Đọc dữ liệu file vào byte slice để chuyển xuống Service
	f, _ := file.Open()
	fileBytes, _ := io.ReadAll(f)
	defer f.Close()

	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	svc := &docService.DocumentService{}
	if err := svc.CreateUploadRequest(orgID, title, category, file.Filename, fileBytes, username.(string)); err != nil {
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

	var fileBytes []byte
	var fileName string
	file, err := c.FormFile("file")
	if err == nil {
		f, _ := file.Open()
		fileBytes, _ = io.ReadAll(f)
		fileName = file.Filename
		defer f.Close()
	}

	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	svc := &docService.DocumentService{}
	err = svc.CreateUpdateRequest(uint(id), orgID, title, category, fileName, fileBytes, username.(string))
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

	svc := &docService.DocumentService{}
	if err := svc.CreateDeleteRequest(uint(id), orgID, username.(string)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Đã gửi yêu cầu xóa tài liệu."})
}
