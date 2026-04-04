package strategies

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/models"
	"sent_backend/internal/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// --- 1. CHIẾN LƯỢC TẠO MỚI NHÂN SỰ ---
type UserCreateStrategy struct{}

func (s *UserCreateStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// 1. Kích hoạt tài khoản (Đổi PENDING -> APPROVED)
	if err := tx.Model(&models.User{}).Where("id = ?", ticket.TargetID).Update("approval_status", "APPROVED").Error; err != nil {
		return err
	}

	// 2. Lấy thông tin User và Công ty từ DB để gửi Email
	var user models.User
	if err := tx.First(&user, ticket.TargetID).Error; err == nil && user.Email != "" {

		var org models.Organization
		tx.First(&org, user.OrgID)

		// 3. Chạy ngầm tiến trình gửi Email ngay khi duyệt xong
		go func(u models.User, companyCode string) {
			subject := "Tài khoản nhân sự SENT SOC của bạn đã được phê duyệt"
			body := fmt.Sprintf(`
                <div style="font-family: sans-serif; padding: 20px; color: #333;">
                    <h2 style="color: #10b981;">Chào mừng %s!</h2>
                    <p>Tài khoản nhân sự SOC của bạn đã được Quản trị viên <b>phê duyệt thành công</b>.</p>
                    <div style="background: #f8fafc; padding: 15px; border-left: 4px solid #10b981; margin: 20px 0;">
                        <p><b>Mã Workspace:</b> %s</p>
                        <p><b>Tên đăng nhập:</b> %s</p>
                        <p><b>Mật khẩu:</b> (Vui lòng liên hệ người cấp tài khoản để nhận)</p>
                    </div>
                    <p>Bạn có thể đăng nhập và bắt đầu làm việc tại: <a href="http://localhost:3000/login/%s" style="color: #10b981; font-weight: bold;">Truy cập Dashboard</a></p>
                </div>
            `, u.FullName, companyCode, u.Username, companyCode)

			_ = utils.SendEmail([]string{u.Email}, subject, body)
		}(user, org.CompanyCode)
	}

	return nil
}

func (s *UserCreateStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Nếu Sếp từ chối, đánh dấu bản ghi nháp thành REJECTED (hoặc có thể dùng lệnh Xóa luôn cũng được)
	return tx.Model(&models.User{}).Where("id = ?", ticket.TargetID).Update("approval_status", "REJECTED").Error
}

// --- 2. CHIẾN LƯỢC CẬP NHẬT QUYỀN NHÂN SỰ ---
type UserUpdateStrategy struct{}

type userUpdatePayload struct {
	Password           string `json:"password"`
	FullName           string `json:"full_name"`
	Phone              string `json:"phone"`
	Email              string `json:"email"`
	PermAssetView      bool   `json:"perm_asset_view"`
	PermAssetAction    bool   `json:"perm_asset_action"`
	PermAssetDelete    bool   `json:"perm_asset_delete"`
	PermPolicyView     bool   `json:"perm_policy_view"`
	PermPolicyAction   bool   `json:"perm_policy_action"`
	PermIncidentView   bool   `json:"perm_incident_view"`
	PermIncidentAction bool   `json:"perm_incident_action"`
	PermDocView        bool   `json:"perm_doc_view"`
	PermDocManage      bool   `json:"perm_doc_manage"`
	PermUserManage     bool   `json:"perm_user_manage"`
	PermApprovalManage bool   `json:"perm_approval_manage"`
}

func (s *UserUpdateStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// 1. Bung nén JSON chứa các quyền mới ra
	var payload userUpdatePayload
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &payload); err != nil {
		return err
	}

	// 2. Chuẩn bị dữ liệu cập nhật
	updates := map[string]interface{}{
		"full_name":            payload.FullName,
		"phone":                payload.Phone,
		"email":                payload.Email,
		"perm_asset_view":      payload.PermAssetView,
		"perm_asset_action":    payload.PermAssetAction,
		"perm_asset_delete":    payload.PermAssetDelete,
		"perm_policy_view":     payload.PermPolicyView,
		"perm_policy_action":   payload.PermPolicyAction,
		"perm_incident_view":   payload.PermIncidentView,
		"perm_incident_action": payload.PermIncidentAction,
		"perm_doc_view":        payload.PermDocView,
		"perm_doc_manage":      payload.PermDocManage,
		"perm_user_manage":     payload.PermUserManage,
		"perm_approval_manage": payload.PermApprovalManage,
	}

	if payload.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
		updates["password_hash"] = string(hashed)
	}

	// 3. Ghi đè vào Database
	return tx.Model(&models.User{}).Where("id = ?", ticket.TargetID).Updates(updates).Error
}

func (s *UserUpdateStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Không làm gì cả, user vẫn giữ nguyên quyền cũ
	return nil
}

// --- 3. CHIẾN LƯỢC XÓA NHÂN SỰ ---
type UserDeleteStrategy struct{}

func (s *UserDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// Khi sếp bấm Duyệt xóa -> Lệnh Delete mới thực sự được kích hoạt
	return tx.Delete(&models.User{}, ticket.TargetID).Error
}

func (s *UserDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return nil
}
