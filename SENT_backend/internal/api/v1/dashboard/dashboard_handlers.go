package dashboard

import (
	"net/http"

	"sent_backend/internal/service/controllers"

	"github.com/gin-gonic/gin"
)

// Khởi tạo một Controller trung tâm
var dashController = &controllers.DashboardController{}

// GetDashboardStats: Handler siêu mỏng (Thin Handler)
func GetDashboardStats(c *gin.Context) {
	// 1. Nhận Context từ Request (VD: OrgID của User)
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không xác định được tổ chức"})
		return
	}
	orgID := orgIDVal.(uint)

	// 2. Chuyển tiếp công việc nặng nhọc cho Controller/Service
	summary, err := dashController.FetchDetailedSummary(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy xuất dữ liệu tổng quan"})
		return
	}

	// 3. Trả về kết quả
	c.JSON(http.StatusOK, summary)
}
