package docs

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/repository"
	"sent_backend/internal/service/security" // Dùng service để lưu file

	"github.com/gin-gonic/gin"
)

// UploadDocument: Chuyên xử lý file PDF/Word quy định công ty
func UploadDocument(c *gin.Context) {
	title := c.PostForm("title")
	category := c.PostForm("category")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "Không tìm thấy file"})
		return
	}

	openedFile, _ := file.Open()
	defer openedFile.Close()
	buffer := make([]byte, file.Size)
	openedFile.Read(buffer)

	orgID := uint(1) // Tạm hardcode, sau này lấy từ JWT
	// Tái sử dụng hàm SavePolicyFile của Security Service (Hoặc đổi tên hàm này sau)
	err = security.SavePolicyFile(orgID, title, category, file.Filename, buffer)
	if err != nil {
		c.JSON(500, gin.H{"error": "Lỗi lưu file: " + err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Upload tài liệu thành công"})
}

// GetDocuments: Lấy danh sách tài liệu tri thức
func GetDocuments(c *gin.Context) {
	orgID := uint(1)
	var docs []models.PolicyDocument
	if err := database.DB.Where("org_id = ?", orgID).Find(&docs).Error; err != nil {
		c.JSON(500, gin.H{"error": "Lỗi truy vấn"})
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
	c.JSON(http.StatusOK, gin.H{"message": "Xóa tài liệu thành công"})
}

// UpdateDocumentStatus: Cập nhật tiêu đề/danh mục tài liệu
func UpdateDocument(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Title    string `json:"title"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu sai"})
		return
	}
	repository.UpdatePolicyStatus(id, req.Title, req.Category)
	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật thành công"})
}
