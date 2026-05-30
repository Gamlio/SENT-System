package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	userService "SENT_backend/service/users/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getOrgIDFromContext(c *gin.Context) uint {
	if val, exists := c.Get("org_id"); exists {
		if uintVal, ok := val.(uint); ok {
			return uintVal
		}
	}
	return 0
}

func getRequesterID(c *gin.Context) uint {
	if val, exists := c.Get("user_id"); exists {
		if uintVal, ok := val.(uint); ok {
			return uintVal
		}
	}
	return 0
}

func CreateUser(c *gin.Context) {
	orgID := getOrgIDFromContext(c)
	requesterID := getRequesterID(c)

	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được tìm thấy"})
		return
	}

	if requesterID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UserID không được tìm thấy"})
		return
	}

	var req models.UserPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ. Chi tiết: " + err.Error()})
		return
	}

	var requester models.User
	if err := database.DB.Where("id = ? AND org_id = ?", requesterID, orgID).First(&requester).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại hoặc không thuộc tổ chức của bạn"})
		return
	}

	userSvc := &userService.UserService{}
	if err := userSvc.CreateUserRequest(req, orgID, requester); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu cấp tài khoản, vui lòng chờ duyệt!"})
}
func GetUserDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	orgID := getOrgIDFromContext(c)
	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được tìm thấy"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ? AND org_id = ?", uint(id), orgID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUserPermissions - Lấy danh sách quyền của một người dùng
func GetUserPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	orgID := getOrgIDFromContext(c)
	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được tìm thấy"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ? AND org_id = ?", uint(id), orgID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng"})
		return
	}

	// Trả về toàn bộ quyền dưới dạng map
	c.JSON(http.StatusOK, gin.H{
		"id":                   user.ID,
		"username":             user.Username,
		"full_name":            user.FullName,
		"perm_asset_view":      user.PermAssetView,
		"perm_asset_action":    user.PermAssetAction,
		"perm_asset_delete":    user.PermAssetDelete,
		"perm_asset_move":      user.PermAssetMove,
		"perm_policy_view":     user.PermPolicyView,
		"perm_policy_manage":   user.PermPolicyManage,
		"perm_incident_view":   user.PermIncidentView,
		"perm_incident_action": user.PermIncidentAction,
		"perm_doc_view":        user.PermDocView,
		"perm_doc_manage":      user.PermDocManage,
		"perm_user_view":       user.PermUserView,
		"perm_user_manage":     user.PermUserManage,
		"perm_system_config":   user.PermSystemConfig,
		"perm_group_manage":    user.PermGroupManage,
		"perm_approval_view":   user.PermApprovalView,
		"perm_approval_final":  user.PermApprovalFinal,
	})
}

// GetCurrentUserPermissions - Lấy quyền của người dùng hiện tại
func GetCurrentUserPermissions(c *gin.Context) {
	userID := getRequesterID(c)
	orgID := getOrgIDFromContext(c)

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UserID không được tìm thấy"})
		return
	}

	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được tìm thấy"})
		return
	}

	var user models.User
	if err := database.DB.Where("id = ? AND org_id = ?", userID, orgID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy thông tin người dùng"})
		return
	}

	// Trả về toàn bộ quyền dưới dạng map
	c.JSON(http.StatusOK, gin.H{
		"perm_asset_view":      user.PermAssetView,
		"perm_asset_action":    user.PermAssetAction,
		"perm_asset_delete":    user.PermAssetDelete,
		"perm_asset_move":      user.PermAssetMove,
		"perm_policy_view":     user.PermPolicyView,
		"perm_policy_manage":   user.PermPolicyManage,
		"perm_incident_view":   user.PermIncidentView,
		"perm_incident_action": user.PermIncidentAction,
		"perm_doc_view":        user.PermDocView,
		"perm_doc_manage":      user.PermDocManage,
		"perm_user_view":       user.PermUserView,
		"perm_user_manage":     user.PermUserManage,
		"perm_system_config":   user.PermSystemConfig,
		"perm_group_manage":    user.PermGroupManage,
		"perm_approval_view":   user.PermApprovalView,
		"perm_approval_final":  user.PermApprovalFinal,
	})
}

func UpdatePermissions(c *gin.Context) {
	UpdateUser(c)
}
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	targetID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	orgID := getOrgIDFromContext(c)
	requesterID := getRequesterID(c)

	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được tìm thấy"})
		return
	}

	if requesterID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UserID không được tìm thấy"})
		return
	}

	var req models.UserPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ. Chi tiết: " + err.Error()})
		return
	}

	var requester models.User
	if err := database.DB.Where("id = ? AND org_id = ?", requesterID, orgID).First(&requester).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại hoặc không thuộc tổ chức của bạn"})
		return
	}

	// Gọi Service Brain
	userSvc := &userService.UserService{}
	if err := userSvc.UpdateUserRequest(uint(targetID), req, orgID, requester); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu thay đổi, vui lòng chờ duyệt!"})
}

// DeleteUser (GATE)
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	targetID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	orgID := getOrgIDFromContext(c)
	requesterID := getRequesterID(c)

	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được tìm thấy"})
		return
	}

	if requesterID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UserID không được tìm thấy"})
		return
	}

	var requester models.User
	if err := database.DB.Where("id = ? AND org_id = ?", requesterID, orgID).First(&requester).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại hoặc không thuộc tổ chức của bạn"})
		return
	}

	// Gọi Service Brain
	userSvc := &userService.UserService{}
	if err := userSvc.DeleteUserRequest(uint(targetID), orgID, requester); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu xóa, vui lòng chờ duyệt!"})
}

// GetUsers vẫn giữ ở Gate vì chỉ là truy vấn
func GetUsers(c *gin.Context) {
	orgID := getOrgIDFromContext(c)
	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được tìm thấy"})
		return
	}

	var usersList []models.User
	if err := database.DB.Where("org_id = ?", orgID).Order("created_at desc").Find(&usersList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn danh sách người dùng"})
		return
	}

	c.JSON(http.StatusOK, usersList)
}
