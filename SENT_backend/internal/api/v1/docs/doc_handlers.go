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
	err = svc.CreateUpdateRequest(uint(id), orgID, title, category, fileName, fileReader, username.(string), reason)
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

	// [MỚI] Khai báo struct để nhận lý do từ Frontend
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Vui lòng cung cấp lý do xóa"})
		return
	}

	svc := &docService.DocumentService{}
	// Truyền req.Reason vào service
	if err := svc.CreateDeleteRequest(uint(id), orgID, username.(string), req.Reason); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Đã gửi yêu cầu xóa tài liệu kèm lý do."})
}

// DownloadDocument: Trả về file Word gốc với tên thuở sơ khai
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

	// [QUAN TRỌNG] FileAttachment sẽ ép trình duyệt tải xuống với tên gốc (OriginalName)
	// thay vì cái tên dán nhãn timestamp trên server.
	c.FileAttachment(doc.FilePath, doc.OriginalName)
}
