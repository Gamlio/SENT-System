package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// --- HELPERS --- (Giữ nguyên getOrgIDFromContext và getRequesterID)
func getOrgIDFromContext(c *gin.Context) uint {
	rawOrgID, exists := c.Get("org_id")
	if !exists {
		return 0
	}
	if floatVal, ok := rawOrgID.(float64); ok {
		return uint(floatVal)
	}
	if uintVal, ok := rawOrgID.(uint); ok {
		return uintVal
	}
	return 0
}

func getRequesterID(c *gin.Context) uint {
	rawUserID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	if floatVal, ok := rawUserID.(float64); ok {
		return uint(floatVal)
	}
	if uintVal, ok := rawUserID.(uint); ok {
		return uintVal
	}
	return 0
}

type UserPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`

	PermAgentView      bool `json:"perm_agent_view"`
	PermAgentAction    bool `json:"perm_agent_action"`
	PermAgentDelete    bool `json:"perm_agent_delete"`
	PermPolicyView     bool `json:"perm_policy_view"`
	PermPolicyAction   bool `json:"perm_policy_action"`
	PermIncidentView   bool `json:"perm_incident_view"`
	PermIncidentAction bool `json:"perm_incident_action"`
	PermDocView        bool `json:"perm_doc_view"`
	PermDocManage      bool `json:"perm_doc_manage"`
	PermUserManage     bool `json:"perm_user_manage"`
	PermApprovalManage bool `json:"perm_approval_manage"`
}

// 1. CREATE USER (Đã tích hợp Approval)
func CreateUser(c *gin.Context) {
	orgID := getOrgIDFromContext(c)
	requesterID := getRequesterID(c)

	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không xác định được tổ chức"})
		return
	}

	var req UserPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu gửi lên không hợp lệ"})
		return
	}

	var requester models.User
	if err := database.DB.First(&requester, requesterID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Lỗi xác thực người thực hiện"})
		return
	}

	if (req.PermApprovalManage && !requester.PermApprovalManage) ||
		(req.PermUserManage && !requester.PermUserManage) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền cấp phát các đặc quyền quản trị cao cấp"})
		return
	}

	var existingUser models.User
	if err := database.DB.Where("username = ? AND org_id = ?", req.Username, orgID).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Tên đăng nhập này đã tồn tại trong hệ thống"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)

	// CHÚ Ý: Chuyển Status thành PENDING thay vì APPROVED
	newUser := models.User{
		Username:           req.Username,
		PasswordHash:       string(hashedPassword),
		FullName:           req.FullName,
		Phone:              req.Phone,
		Email:              req.Email,
		OrgID:              &orgID,
		ApprovalStatus:     "PENDING", // <--- Bị khóa cho đến khi duyệt
		PermAgentView:      req.PermAgentView,
		PermAgentAction:    req.PermAgentAction,
		PermAgentDelete:    req.PermAgentDelete,
		PermPolicyView:     req.PermPolicyView,
		PermPolicyAction:   req.PermPolicyAction,
		PermIncidentView:   req.PermIncidentView,
		PermIncidentAction: req.PermIncidentAction,
		PermDocView:        req.PermDocView,
		PermDocManage:      req.PermDocManage,
		PermUserManage:     req.PermUserManage,
		PermApprovalManage: req.PermApprovalManage,
	}

	// Bắt đầu Transaction
	tx := database.DB.Begin()

	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi tạo tài khoản"})
		return
	}

	// TẠO TICKET PHÊ DUYỆT
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

	if err := tx.Create(&ticket).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi tạo đơn phê duyệt"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu cấp tài khoản, vui lòng chờ duyệt!"})
}

// 2. GET USERS (Giữ nguyên)
func GetUsers(c *gin.Context) {
	orgID := getOrgIDFromContext(c)
	var usersList []models.User

	if err := database.DB.Where("org_id = ?", orgID).Order("created_at desc").Find(&usersList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn dữ liệu"})
		return
	}
	c.JSON(http.StatusOK, usersList)
}

// 3. UPDATE USER (Đã tích hợp Approval)
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	orgID := getOrgIDFromContext(c)
	requesterID := getRequesterID(c)

	var targetUser models.User
	if err := database.DB.Where("id = ? AND org_id = ?", id, orgID).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng"})
		return
	}

	var req UserPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	if targetUser.ID == requesterID && !req.PermUserManage {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bạn không thể tự tước quyền Quản lý nhân sự của chính mình"})
		return
	}

	var requester models.User
	database.DB.First(&requester, requesterID)
	if (req.PermApprovalManage && !requester.PermApprovalManage) ||
		(req.PermUserManage && !requester.PermUserManage) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền cấp phát đặc quyền này"})
		return
	}

	// TẠO TICKET SỬA THAY VÌ UPDATE TRỰC TIẾP
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

	if err := database.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi tạo đơn phê duyệt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu thay đổi quyền, vui lòng chờ duyệt!"})
}

// 4. DELETE USER (Đã tích hợp Approval)
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	orgID := getOrgIDFromContext(c)
	requesterID := getRequesterID(c)

	var targetUser models.User
	if err := database.DB.Where("id = ? AND org_id = ?", id, orgID).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy tài khoản"})
		return
	}

	if targetUser.ID == requesterID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không thể xóa chính tài khoản đang đăng nhập"})
		return
	}

	var requester models.User
	database.DB.First(&requester, requesterID)

	// TẠO TICKET XÓA
	ticket := models.ApprovalTicket{
		OrgID:        orgID,
		ModuleType:   "USER_DELETE",
		ActionType:   "DELETE",
		TargetID:     targetUser.ID,
		TargetName:   fmt.Sprintf("Xóa: %s", targetUser.Username),
		Status:       "PENDING",
		RequestedBy:  requester.Username,
		SnapshotData: `{"Lý do": "Yêu cầu gỡ bỏ nhân sự"}`,
	}

	if err := database.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi tạo đơn phê duyệt"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu xóa tài khoản, vui lòng chờ duyệt!"})
}
