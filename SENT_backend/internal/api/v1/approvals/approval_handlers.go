package approvals

import (
	"fmt"
	"net/http"
	"sent_backend/internal/api/v1/approvals/strategies" // Import thư mục strategies vừa tạo
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GET /api/v1/approvals
func GetTickets(c *gin.Context) {
	status := c.Query("status")
	module := c.Query("module_type")

	var tickets []models.ApprovalTicket
	query := database.DB.Model(&models.ApprovalTicket{})

	if status != "" {
		query = query.Where("status = ?", status)
	} else {
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
	id := c.Param("id")
	var req struct {
		Status     string `json:"status"`
		ReviewNote string `json:"review_note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	var ticket models.ApprovalTicket
	if err := database.DB.First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy đơn yêu cầu"})
		return
	}

	if ticket.Status != "PENDING" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Đơn này đã được xử lý"})
		return
	}

	// 1. TÌM CHIẾN LƯỢC XỬ LÝ DỰA VÀO MODULE TYPE
	strategy := strategies.GetStrategy(ticket.ModuleType)
	if strategy == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hệ thống chưa hỗ trợ loại phê duyệt này!"})
		return
	}

	tx := database.DB.Begin()

	// 2. Cập nhật trạng thái Ticket
	// Lấy tên người duyệt từ Context (đã gán bởi JWT Middleware)
	usernameVal, exists := c.Get("username")
	if exists {
		ticket.ReviewedBy = usernameVal.(string)
	} else {
		ticket.ReviewedBy = "System_Admin"
	}

	ticket.Status = req.Status
	ticket.ReviewNote = req.ReviewNote

	if err := tx.Save(&ticket).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lưu trạng thái đơn"})
		return
	}

	// 3. THỰC THI NGHIỆP VỤ BẰNG STRATEGY
	var processErr error
	if req.Status == "APPROVED" {
		processErr = strategy.OnApprove(tx, &ticket)
	} else if req.Status == "REJECTED" {
		processErr = strategy.OnReject(tx, &ticket)
	}

	// Nếu xử lý đích bị lỗi -> Rollback toàn bộ
	if processErr != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật bảng dữ liệu đích"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Đã %s thành công", req.Status)})
}
