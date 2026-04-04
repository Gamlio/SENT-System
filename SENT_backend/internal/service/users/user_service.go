package users

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct{}

// CreateUserRequest: Xử lý yêu cầu tạo nhân sự mới
func (s *UserService) CreateUserRequest(req models.UserPayload, orgID uint, requester models.User) error {
	// 1. Kiểm tra quyền hạn cấp phát
	if (req.PermApprovalManage && !requester.PermApprovalManage) ||
		(req.PermUserManage && !requester.PermUserManage) {
		return fmt.Errorf("bạn không có quyền cấp phát các đặc quyền quản trị cao cấp")
	}

	// 2. Kiểm tra trùng lặp username
	var existingUser models.User
	if err := database.DB.Where("username = ? AND org_id = ?", req.Username, orgID).First(&existingUser).Error; err == nil {
		return fmt.Errorf("tên đăng nhập này đã tồn tại")
	}

	// 3. Băm mật khẩu
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Tạo bản ghi User ở trạng thái PENDING
		newUser := models.User{
			Username:       req.Username,
			PasswordHash:   string(hashedPassword),
			FullName:       req.FullName,
			Phone:          req.Phone,
			Email:          req.Email,
			OrgID:          &orgID,
			ApprovalStatus: "PENDING",
			// Gán các quyền từ request
			PermAssetView:      req.PermAssetView,
			PermAssetAction:    req.PermAssetAction,
			PermAssetDelete:    req.PermAssetDelete,
			PermPolicyView:     req.PermPolicyView,
			PermPolicyAction:   req.PermPolicyAction,
			PermIncidentView:   req.PermIncidentView,
			PermIncidentAction: req.PermIncidentAction,
			PermDocView:        req.PermDocView,
			PermDocManage:      req.PermDocManage,
			PermUserManage:     req.PermUserManage,
			PermApprovalManage: req.PermApprovalManage,
		}

		if err := tx.Create(&newUser).Error; err != nil {
			return err
		}

		// Tạo Ticket phê duyệt
		snapshot, _ := json.Marshal(req)
		ticket := models.ApprovalTicket{
			OrgID:        orgID,
			ModuleType:   "USER_CREATE",
			ActionType:   "CREATE",
			TargetID:     newUser.ID,
			TargetName:   fmt.Sprintf("Tài khoản: %s", newUser.Username),
			Status:       "PENDING",
			RequestedBy:  requester.Username,
			SnapshotData: string(snapshot),
		}

		return tx.Create(&ticket).Error
	})
}

// UpdateUserRequest: Tạo đơn yêu cầu thay đổi quyền hạn/thông tin
func (s *UserService) UpdateUserRequest(targetID uint, req models.UserPayload, orgID uint, requester models.User) error {
	var targetUser models.User
	if err := database.DB.Where("id = ? AND org_id = ?", targetID, orgID).First(&targetUser).Error; err != nil {
		return fmt.Errorf("không tìm thấy người dùng")
	}

	// Chốt chặn: Không thể tự tước quyền quản lý của chính mình
	if targetUser.ID == requester.ID && !req.PermUserManage {
		return fmt.Errorf("bạn không thể tự tước quyền Quản lý nhân sự của chính mình")
	}

	if (req.PermApprovalManage && !requester.PermApprovalManage) ||
		(req.PermUserManage && !requester.PermUserManage) {
		return fmt.Errorf("bạn không có quyền cấp phát đặc quyền quản trị")
	}

	snapshot, _ := json.Marshal(req)
	ticket := models.ApprovalTicket{
		OrgID:        orgID,
		ModuleType:   "USER_UPDATE",
		ActionType:   "UPDATE",
		TargetID:     targetUser.ID,
		TargetName:   fmt.Sprintf("Sửa quyền: %s", targetUser.Username),
		Status:       "PENDING",
		RequestedBy:  requester.Username,
		SnapshotData: string(snapshot),
	}

	return database.DB.Create(&ticket).Error
}

// DeleteUserRequest: Tạo đơn yêu cầu xóa nhân sự
func (s *UserService) DeleteUserRequest(targetID uint, orgID uint, requester models.User) error {
	var targetUser models.User
	if err := database.DB.Where("id = ? AND org_id = ?", targetID, orgID).First(&targetUser).Error; err != nil {
		return fmt.Errorf("không tìm thấy tài khoản")
	}

	if targetUser.ID == requester.ID {
		return fmt.Errorf("bạn không thể tự xóa chính mình")
	}

	ticket := models.ApprovalTicket{
		OrgID:        orgID,
		ModuleType:   "USER_DELETE",
		ActionType:   "DELETE",
		TargetID:     targetUser.ID,
		TargetName:   fmt.Sprintf("Xóa: %s", targetUser.Username),
		Status:       "PENDING",
		RequestedBy:  requester.Username,
		SnapshotData: `{"reason": "Yêu cầu gỡ bỏ nhân sự từ Admin"}`,
	}

	return database.DB.Create(&ticket).Error
}
