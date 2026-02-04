package v1

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
)

// PushDataFromAgent: API dành riêng cho Go Agent đẩy thông tin
func PushDataFromAgent(c *gin.Context) {
	var req models.Agent
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu Agent sai định dạng"})
		return
	}

	// Cập nhật hoặc tạo mới dựa trên HWID
	var agent models.Agent
	result := database.DB.Where("hwid = ?", req.HWID).First(&agent)

	if result.Error != nil {
		database.DB.Create(&req) // Tạo mới nếu chưa có
	} else {
		database.DB.Model(&agent).Updates(req) // Cập nhật thông số mới nhất
	}

	c.JSON(http.StatusOK, gin.H{"status": "Dữ liệu đã được tiếp nhận"})
}

// GetAssetsHandler: Lấy danh sách máy trạm (Có lọc theo OrgID)
func GetAssetsHandler(c *gin.Context) {
	// Lấy OrgID từ Middleware (người dùng đang đăng nhập)
	userOrgID := c.GetUint("org_id")

	var assets []models.Agent
	database.DB.Where("org_id = ?", userOrgID).Find(&assets) // Cô lập dữ liệu SME

	c.JSON(http.StatusOK, assets)
}
