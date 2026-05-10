package approvals

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	service "SENT_backend/service/approvals/internal/service"
	"net/http" // Import thư mục strategies vừa tạo
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /api/v1/approvals
func GetTickets(c *gin.Context) {
	status := c.Query("status")
	module := c.Query("module_type")

	var tickets []models.ApprovalTicket
	query := database.DB.Model(&models.ApprovalTicket{})

	if status != "" && status != "ALL" {
		query = query.Where("status = ?", status)
	} else if status == "" {
		// Mặc định nếu không truyền gì (lần đầu load trang) thì chỉ hiện PENDING
		query = query.Where("status = ?", "PENDING")
	}

	if module != "" {
		query = query.Where("module_type = ?", module)
	}

	query.Order("created_at desc").Find(&tickets)
	c.JSON(http.StatusOK, tickets)
}

// PUT /api/v1/approvals/:id/review
func ReviewTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var req struct {
		Status     string `json:"status"`
		ReviewNote string `json:"review_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// 2. Xác định danh tính người duyệt từ JWT
	reviewer := "System_Admin"
	if username, exists := c.Get("username"); exists {
		reviewer = username.(string)
	}

	// 3. Gọi Service để xử lý (Brain)
	approvalSvc := &service.ApprovalService{}
	err := approvalSvc.ProcessReview(uint(id), req.Status, req.ReviewNote, reviewer)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xử lý đơn thành công"})
}
