package strategies

import (
	"encoding/json"
	"errors"
	"sent_backend/internal/models"

	"gorm.io/gorm"
)

type PolicyCreateStrategy struct{}

func (s *PolicyCreateStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// 1. Kiểm tra tính toàn vẹn: Đảm bảo Ticket có thông tin người duyệt
	if ticket.ReviewedBy == "" {
		return errors.New("thiếu thông tin người phê duyệt chính sách")
	}

	// 2. Thực hiện cập nhật "Sâu":
	// - Đổi trạng thái sang APPROVED
	// - Ghi danh tính người duyệt vào bản ghi Policy (Audit)
	// - Kích hoạt luật (IsActive = true)
	// - Chốt chặn OrgID: Chỉ cập nhật nếu Policy đó thuộc về đúng Org của Ticket
	result := tx.Model(&models.Policy{}).
		Where("id = ? AND org_id = ?", ticket.TargetID, ticket.OrgID).
		Updates(map[string]interface{}{
			"approval_status": "APPROVED",
			"approved_by":     ticket.ReviewedBy, // Lưu lại danh tính người duyệt
			"is_active":       true,              // Tự động kích hoạt luật sau khi duyệt
		})

	if result.Error != nil {
		return result.Error
	}

	// 3. Kiểm tra xem có bản ghi nào được cập nhật không
	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy chính sách đích hoặc sai phạm phân quyền tổ chức")
	}

	return nil
}

func (s *PolicyCreateStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Khi từ chối, ta vẫn nên lưu lại người đã từ chối để làm bằng chứng (Audit)
	return tx.Model(&models.Policy{}).
		Where("id = ? AND org_id = ?", ticket.TargetID, ticket.OrgID).
		Updates(map[string]interface{}{
			"approval_status": "REJECTED",
			"approved_by":     ticket.ReviewedBy,
			"is_active":       false, // Đảm bảo luật không hoạt động nếu bị từ chối
		}).Error
}

type PolicyDeleteStrategy struct{}

func (s *PolicyDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Thực hiện xóa thật bản ghi trong DB khi được duyệt
	return tx.Delete(&models.Policy{}, ticket.TargetID).Error
}

func (s *PolicyDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return nil // Từ chối xóa thì không làm gì cả
}

// --- XÓA NHIỀU ---
type PolicyBulkDeleteStrategy struct{}

func (s *PolicyBulkDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	var data struct {
		IDs []uint `json:"ids"`
	}
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &data); err != nil {
		return err
	}
	// Xóa đồng loạt theo danh sách ID đã lưu trong vé
	return tx.Where("id IN ? AND org_id = ?", data.IDs, ticket.OrgID).Delete(&models.Policy{}).Error
}

func (s *PolicyBulkDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return nil
}
