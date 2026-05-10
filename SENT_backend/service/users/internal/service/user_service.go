package users

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"encoding/json"
	"fmt"
)

type UserService struct{}

// CreateUserRequest: Xử lý yêu cầu tạo nhân sự mới
func (s *UserService) CreateUserRequest(req models.UserPayload, orgID uint, requester models.User) error {
	// 1. Kiểm tra quyền hạn cấp phát: "Không được cho đi thứ mình không có"
	if (req.PermSystemConfig && !requester.PermSystemConfig) ||
		(req.PermApprovalFinal && !requester.PermApprovalFinal) ||
		(req.PermUserManage && !requester.PermUserManage) ||
		(req.PermUserView && !requester.PermUserView) ||
		(req.PermAssetMove && !requester.PermAssetMove) ||
		(req.PermAssetView && !requester.PermAssetView) ||
		(req.PermAssetAction && !requester.PermAssetAction) ||
		(req.PermAssetDelete && !requester.PermAssetDelete) ||
		(req.PermPolicyManage && !requester.PermPolicyManage) ||
		(req.PermApprovalView && !requester.PermApprovalView) ||
		(req.PermGroupManage && !requester.PermGroupManage) {
		return fmt.Errorf("bạn không có quyền cấp phát các đặc quyền cao hơn quyền của mình")
	}

	// 2. Kiểm tra trùng lặp username
	var existingUser models.User
	if err := database.DB.Where("username = ? AND org_id = ?", req.Username, orgID).First(&existingUser).Error; err == nil {
		return fmt.Errorf("tên đăng nhập này đã tồn tại")
	}

	// 3. Tạo Ticket phê duyệt (Hệ thống SẠCH - Không tạo rác PENDING)
	snapshot, _ := json.Marshal(req)
	ticket := models.ApprovalTicket{
		OrgID:        orgID,
		ModuleType:   "USER_CREATE",
		ActionType:   "CREATE",
		TargetID:     0, // Chưa có ID thực vì đợi Sếp duyệt mới INSERT
		TargetName:   fmt.Sprintf("Tài khoản: %s", req.Username),
		Status:       "PENDING",
		RequestedBy:  requester.Username,
		SnapshotData: string(snapshot),
	}

	return database.DB.Create(&ticket).Error
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

	if (req.PermSystemConfig && !requester.PermSystemConfig) ||
		(req.PermApprovalFinal && !requester.PermApprovalFinal) ||
		(req.PermUserManage && !requester.PermUserManage) ||
		(req.PermUserView && !requester.PermUserView) ||
		(req.PermAssetMove && !requester.PermAssetMove) ||
		(req.PermAssetView && !requester.PermAssetView) ||
		(req.PermAssetAction && !requester.PermAssetAction) ||
		(req.PermAssetDelete && !requester.PermAssetDelete) ||
		(req.PermPolicyManage && !requester.PermPolicyManage) ||
		(req.PermGroupManage && !requester.PermGroupManage) ||
		(req.PermApprovalView && !requester.PermApprovalView) {
		return fmt.Errorf("bạn không có quyền cấp phát các đặc quyền cao hơn quyền của mình")
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
