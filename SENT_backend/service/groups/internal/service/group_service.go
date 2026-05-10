package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"encoding/json"
	"fmt"
)

type GroupService struct{}

// ServiceCreateGroupRequest xử lý logic kiểm tra quyền và tạo Ticket phê duyệt (Hệ thống sạch)
func (s *GroupService) ServiceCreateGroupRequest(req models.PolicyGroupPayload, orgID uint, requester models.User) error {
	// 1. Kiểm tra quyền hạn của người gửi yêu cầu
	if !requester.PermGroupManage {
		return fmt.Errorf("bạn không có quyền quản lý nhóm (PermGroupManage)")
	}

	// 2. Kiểm tra trùng tên nhóm trong cùng một công ty (Org)
	var existing models.PolicyGroup
	err := database.DB.Where("name = ? AND org_id = ?", req.Name, orgID).First(&existing).Error
	if err == nil {
		return fmt.Errorf("tên nhóm '%s' đã tồn tại trong hệ thống", req.Name)
	}

	// 3. Đóng gói dữ liệu vào SnapshotData để chờ duyệt
	snapshot, _ := json.Marshal(req)

	// Tạo Ticket phê duyệt - Lúc này chưa INSERT vào bảng PolicyGroup
	ticket := models.ApprovalTicket{
		OrgID:        orgID,
		ModuleType:   "GROUP_CREATE", // Module này sẽ được xử lý bởi GroupCreateStrategy
		ActionType:   "CREATE",
		TargetName:   fmt.Sprintf("Nhóm mới: %s", req.Name),
		Status:       "PENDING",
		RequestedBy:  requester.Username,
		SnapshotData: string(snapshot),
	}

	// Lưu ticket vào DB để chờ Checker bấm "Duyệt"
	return database.DB.Create(&ticket).Error
}

// ServiceCreateGroupDirect thực hiện tạo phòng ban ngay lập tức cho Sếp
func (s *GroupService) ServiceCreateGroupDirect(req models.PolicyGroupPayload, orgID uint, requester models.User) error {
	// 1. Kiểm tra quyền Quản trị hệ thống. Chỉ sếp to nhất mới được tạo thẳng.
	if !requester.PermSystemConfig {
		return fmt.Errorf("chức năng tạo phòng ban trực tiếp chỉ dành cho cấp Quản trị hệ thống (PermSystemConfig)")
	}

	// 2. Kiểm tra trùng tên nhóm trong Tổ chức
	var count int64
	database.DB.Model(&models.PolicyGroup{}).Where("name = ? AND org_id = ?", req.Name, orgID).Count(&count)
	if count > 0 {
		return fmt.Errorf("phòng ban '%s' đã tồn tại trong hệ thống", req.Name)
	}

	// 3. Khởi tạo đối tượng mới với thông tin Audit
	newGroup := models.PolicyGroup{
		OrgID:       orgID,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   requester.Username, // Lưu tên người tạo
	}

	// Thực hiện tạo trực tiếp vào DB
	return database.DB.Create(&newGroup).Error
}

// ServiceUpdateGroupDirect thực hiện cập nhật phòng ban ngay lập tức
func (s *GroupService) ServiceUpdateGroupDirect(groupID uint, req models.PolicyGroupPayload, orgID uint, requester models.User) error {
	// 1. Kiểm tra quyền quản lý nhóm
	if !requester.PermGroupManage {
		return fmt.Errorf("bạn không có quyền quản lý nhóm (PermGroupManage)")
	}

	// 2. Kiểm tra xem nhóm có tồn tại không
	var group models.PolicyGroup
	if err := database.DB.Where("id = ? AND org_id = ?", groupID, orgID).First(&group).Error; err != nil {
		return fmt.Errorf("không tìm thấy phòng ban")
	}

	// 3. Kiểm tra trùng tên (nếu tên thay đổi)
	if group.Name != req.Name {
		var count int64
		database.DB.Model(&models.PolicyGroup{}).Where("name = ? AND org_id = ? AND id != ?", req.Name, orgID, groupID).Count(&count)
		if count > 0 {
			return fmt.Errorf("tên phòng ban '%s' đã tồn tại", req.Name)
		}
	}

	// 4. Cập nhật
	group.Name = req.Name
	group.Description = req.Description
	return database.DB.Save(&group).Error
}

// ServiceDeleteGroupDirect thực hiện xóa phòng ban ngay lập tức
func (s *GroupService) ServiceDeleteGroupRequest(groupID uint, reason string, orgID uint, requester models.User) error {
	// 1. Kiểm tra quyền quản lý nhóm
	if !requester.PermGroupManage {
		return fmt.Errorf("bạn không có quyền yêu cầu xóa nhóm (PermGroupManage)")
	}

	// 2. Kiểm tra xem nhóm có tồn tại không
	var group models.PolicyGroup
	if err := database.DB.Where("id = ? AND org_id = ?", groupID, orgID).First(&group).Error; err != nil {
		return fmt.Errorf("không tìm thấy phòng ban để xóa")
	}

	// 3. Tạo Ticket phê duyệt xóa
	ticket := models.ApprovalTicket{
		OrgID:         orgID,
		ModuleType:    "GROUP_DELETE", // Module này sẽ được đăng ký trong Strategy
		ActionType:    "DELETE",
		TargetID:      groupID,
		TargetName:    fmt.Sprintf("Xóa phòng ban: %s", group.Name),
		Status:        "PENDING",
		RequestedBy:   requester.Username,
		RequestReason: reason, // Lưu lý do xóa
	}

	return database.DB.Create(&ticket).Error
}
