package service

import (
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/approvals/strategies"

	"gorm.io/gorm"
)

type ApprovalService struct{}

// ProcessReview: Xử lý phê duyệt đơn (Logic cốt lõi)
func (s *ApprovalService) ProcessReview(ticketID uint, status string, note string, reviewer string) error {
	var ticket models.ApprovalTicket
	if err := database.DB.First(&ticket, ticketID).Error; err != nil {
		return fmt.Errorf("không tìm thấy đơn yêu cầu")
	}

	if ticket.Status != "PENDING" {
		return fmt.Errorf("đơn này đã được xử lý trước đó")
	}

	// 1. Tìm chiến lược xử lý (Module Strategy)
	strategy := strategies.GetStrategy(ticket.ModuleType)
	if strategy == nil {
		return fmt.Errorf("hệ thống chưa hỗ trợ loại phê duyệt: %s", ticket.ModuleType)
	}

	// 2. Thực thi trong Transaction để đảm bảo an toàn dữ liệu
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Cập nhật thông tin Ticket
		ticket.Status = status
		ticket.ReviewNote = note
		ticket.ReviewedBy = reviewer

		if err := tx.Save(&ticket).Error; err != nil {
			return err
		}

		// Gọi Strategy tương ứng để thực hiện nghiệp vụ đích (VD: Đổi trạng thái Agent, tạo User...)
		var processErr error
		if status == "APPROVED" {
			processErr = strategy.OnApprove(tx, &ticket)
		} else if status == "REJECTED" {
			processErr = strategy.OnReject(tx, &ticket)
		}

		if processErr != nil {
			return fmt.Errorf("lỗi thực thi nghiệp vụ: %v", processErr)
		}

		return nil
	})
}
