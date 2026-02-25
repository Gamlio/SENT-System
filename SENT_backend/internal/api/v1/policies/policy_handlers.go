package policies

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/repository"
	"sent_backend/internal/service"

	"github.com/gin-gonic/gin"
)

func UploadPolicyHandler(c *gin.Context) {
	title := c.PostForm("title")
	category := c.PostForm("category")
	file, err := c.FormFile("file") // Lấy file từ key 'file'
	if err != nil {
		c.JSON(400, gin.H{"error": "Không tìm thấy file gửi kèm"})
		return
	}

	openedFile, _ := file.Open()
	defer openedFile.Close()
	buffer := make([]byte, file.Size)
	openedFile.Read(buffer)

	orgID := uint(1) // Tạm thời
	err = service.SavePolicyFile(orgID, title, category, file.Filename, buffer)
	if err != nil {
		c.JSON(500, gin.H{"error": "Lỗi khi lưu file"})
		return
	}
	c.JSON(200, gin.H{"message": "Upload thành công"})
}
func GetPoliciesHandler(c *gin.Context) {
	orgID := uint(1)
	var docs []models.PolicyDocument // Chắc chắn dùng đúng struct này
	if err := database.DB.Where("org_id = ?", orgID).Find(&docs).Error; err != nil {
		c.JSON(500, gin.H{"error": "Lỗi truy vấn"})
		return
	}
	c.JSON(200, docs)
}

// DeletePolicyHandler: Xóa tài liệu
func DeletePolicyHandler(c *gin.Context) {
	id := c.Param("id")
	orgID := uint(1)

	if err := service.DeletePolicyAndFile(id, orgID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể xóa tài liệu"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Xóa thành công"})
}

// UpdatePolicyHandler: Sửa thông tin tài liệu
func UpdatePolicyHandler(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Title    string `json:"title"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	repository.UpdatePolicyStatus(id, req.Title, req.Category)
	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật thành công"})
}

// Lấy danh sách chính sách theo loại (Software, USB,...)
func GetPoliciesByCategory(c *gin.Context) {
	category := c.Query("category") // Lấy từ URL: ?category=SOFTWARE
	var list []models.UniversalPolicy

	query := database.DB.Where("org_id = ?", 1) // Tạm hardcode OrgID
	if category != "" {
		query = query.Where("category = ?", category)
	}

	query.Find(&list)
	c.JSON(200, list)
}

func AddUniversalPolicy(c *gin.Context) {
	var req models.UniversalPolicy
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}
	// Lưu vào bảng chính sách kỹ thuật
	database.DB.Create(&req)
	c.JSON(200, gin.H{"message": "Đã lưu chính sách kỹ thuật"})
}

// Xóa chính sách kỹ thuật (Software/USB) khỏi bảng UniversalPolicy
func DeletePolicy(c *gin.Context) {
	id := c.Param("id")

	// Xóa trực tiếp bằng GORM
	if err := database.DB.Delete(&models.UniversalPolicy{}, id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Lỗi khi xóa chính sách"})
		return
	}

	c.JSON(200, gin.H{"message": "Đã xóa chính sách kỹ thuật"})
}
