package approvals

import (
	"fmt"
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GET /api/v1/approvals
// Lấy danh sách các đơn cần duyệt
func GetTickets(c *gin.Context) {
	status := c.Query("status")      // PENDING, APPROVED, REJECTED
	module := c.Query("module_type") // AGENT, POLICY, DOCUMENT

	var tickets []models.ApprovalTicket
	query := database.DB.Model(&models.ApprovalTicket{})

	// Mặc định chỉ lấy đơn đang chờ
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
// Trưởng phòng SOC thực hiện duyệt hoặc từ chối
func ReviewTicket(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status     string `json:"status"`      // Bắt buộc: "APPROVED" hoặc "REJECTED"
		ReviewNote string `json:"review_note"` // Lý do từ chối (nếu có)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// 1. Tìm đơn trong DB
	var ticket models.ApprovalTicket
	if err := database.DB.First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy đơn yêu cầu"})
		return
	}

	if ticket.Status != "PENDING" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Đơn này đã được xử lý"})
		return
	}

	// Bắt đầu Transaction (Đảm bảo an toàn dữ liệu)
	tx := database.DB.Begin()

	// 2. Cập nhật trạng thái Đơn (Ticket)
	ticket.Status = req.Status
	ticket.ReviewNote = req.ReviewNote
	ticket.ReviewedBy = "Admin_SOC" // TODO: Sau này lấy từ JWT Token của người đăng nhập

	if err := tx.Save(&ticket).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lưu trạng thái đơn"})
		return
	}

	// 3. [QUAN TRỌNG] Tự động cập nhật bảng đích dựa theo Module
	if req.Status == "APPROVED" {
		switch ticket.ModuleType {
		case "AGENT_ENROLL":
			// Kích hoạt máy trạm
			if err := tx.Model(&models.Agent{}).Where("id = ?", ticket.TargetID).Update("status", "ACTIVE").Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi kích hoạt máy trạm"})
				return
			}
		case "POLICY_CREATE":
			// Kích hoạt luật bảo mật
			if err := tx.Model(&models.UniversalPolicy{}).Where("id = ?", ticket.TargetID).Update("approval_status", "APPROVED").Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi kích hoạt chính sách"})
				return
			}
		case "DOCUMENT_UPLOAD":
			if err := tx.Model(&models.PolicyDocument{}).Where("id = ?", ticket.TargetID).Update("approval_status", "APPROVED").Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi kích hoạt tài liệu"})
				return
			}
		}
	} else if req.Status == "REJECTED" {
		switch ticket.ModuleType {
		case "AGENT_ENROLL":
			tx.Model(&models.Agent{}).Where("id = ?", ticket.TargetID).Update("status", "REJECTED")
		case "POLICY_CREATE":
			tx.Model(&models.UniversalPolicy{}).Where("id = ?", ticket.TargetID).Update("approval_status", "REJECTED")
		case "DOCUMENT_UPLOAD":
			tx.Model(&models.PolicyDocument{}).Where("id = ?", ticket.TargetID).Update("approval_status", "REJECTED")
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Đã %s thành công", req.Status)})
}
