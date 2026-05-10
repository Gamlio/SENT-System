package strategies

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/utils"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// --- 1. CHIẾN LƯỢC TẠO MỚI NHÂN SỰ ---
type UserCreateStrategy struct{}

func (s *UserCreateStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// 1. Giải mã dữ liệu từ "Đơn" (Snapshot)
	var payload models.UserPayload
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &payload); err != nil {
		return err
	}

	// 2. Hash mật khẩu
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		return err
	}

	// 3. Bây giờ mới thực sự tạo User trong hệ thống (Hệ thống luôn sạch, không có rác PENDING)
	newUser := models.User{
		Username:           payload.Username,
		PasswordHash:       string(hashedPassword),
		FullName:           payload.FullName,
		Phone:              payload.Phone,
		Email:              payload.Email,
		GroupID:            payload.GroupID,
		OrgID:              &ticket.OrgID, // OrgID lấy từ Ticket
		PermAssetView:      payload.PermAssetView,
		PermAssetAction:    payload.PermAssetAction,
		PermAssetDelete:    payload.PermAssetDelete,
		PermAssetMove:      payload.PermAssetMove,
		PermPolicyView:     payload.PermPolicyView,
		PermPolicyManage:   payload.PermPolicyManage,
		PermIncidentView:   payload.PermIncidentView,
		PermIncidentAction: payload.PermIncidentAction,
		PermDocView:        payload.PermDocView,
		PermDocManage:      payload.PermDocManage,
		PermUserView:       payload.PermUserView,
		PermUserManage:     payload.PermUserManage,
		PermSystemConfig:   payload.PermSystemConfig,
		PermGroupManage:    payload.PermGroupManage,
		PermApprovalView:   payload.PermApprovalView,
		PermApprovalFinal:  payload.PermApprovalFinal,

		ApprovalStatus: "APPROVED", // Trạng thái là APPROVED vì đã được Sếp duyệt
	}

	if err := tx.Create(&newUser).Error; err != nil {
		return err
	}

	// 4. Lấy thông tin Công ty từ DB để gửi Email
	var org models.Organization
	if err := tx.First(&org, ticket.OrgID).Error; err == nil && newUser.Email != "" {
		// 5. Chạy ngầm tiến trình gửi Email ngay khi duyệt xong
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
					<p>Bạn có thể đăng nhập và bắt đầu làm việc tại:
                    <a href="%s/login/%s" style="color: #10b981; font-weight: bold;">Truy cập Dashboard</a></p>
                </div>
            `, u.FullName, companyCode, u.Username, companyCode)

			_ = utils.SendEmail([]string{u.Email}, subject, body)
		}(newUser, org.CompanyCode)
	}

	return nil
}

func (s *UserCreateStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// THEO TƯ DUY MỚI: Hệ thống SẠCH.
	// Lúc tạo đơn (Ticket), ta chưa hề INSERT vào bảng User (không tạo rác REJECTED).
	// Nên khi bị từ chối, ta KHÔNG CẦN LÀM GÌ trong bảng User cả.
	return nil
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
	PermAssetMove      bool   `json:"perm_asset_move"`
	PermPolicyView     bool   `json:"perm_policy_view"`
	PermPolicyManage   bool   `json:"perm_policy_manage"`
	PermIncidentView   bool   `json:"perm_incident_view"`
	PermIncidentAction bool   `json:"perm_incident_action"`
	PermDocView        bool   `json:"perm_doc_view"`
	PermDocManage      bool   `json:"perm_doc_manage"`
	PermUserView       bool   `json:"perm_user_view"`
	PermUserManage     bool   `json:"perm_user_manage"`
	PermSystemConfig   bool   `json:"perm_system_config"`
	PermGroupManage    bool   `json:"perm_group_manage"`
	PermApprovalView   bool   `json:"perm_approval_view"`
	PermApprovalFinal  bool   `json:"perm_approval_final"`
	GroupID            *uint  `json:"group_id"`
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
		"group_id":             payload.GroupID,
		"perm_asset_view":      payload.PermAssetView,
		"perm_asset_action":    payload.PermAssetAction,
		"perm_asset_delete":    payload.PermAssetDelete,
		"perm_asset_move":      payload.PermAssetMove,
		"perm_policy_view":     payload.PermPolicyView,
		"perm_policy_manage":   payload.PermPolicyManage,
		"perm_incident_view":   payload.PermIncidentView,
		"perm_incident_action": payload.PermIncidentAction,
		"perm_doc_view":        payload.PermDocView,
		"perm_doc_manage":      payload.PermDocManage,
		"perm_user_view":       payload.PermUserView,
		"perm_user_manage":     payload.PermUserManage,
		"perm_system_config":   payload.PermSystemConfig,
		"perm_group_manage":    payload.PermGroupManage,
		"perm_approval_view":   payload.PermApprovalView,
		"perm_approval_final":  payload.PermApprovalFinal,
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
