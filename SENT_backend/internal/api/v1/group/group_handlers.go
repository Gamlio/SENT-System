package group

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	groupService "sent_backend/internal/service/groups"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HandleCreateGroup xử lý tạo phòng ban trực tiếp cho Sếp hoặc tạo ticket cho người khác
func HandleCreateGroup(c *gin.Context) {
	orgID, _ := c.Get("org_id")
	userID, _ := c.Get("user_id")

	// Sử dụng struct Payload để nhận dữ liệu từ Client
	var req models.PolicyGroupPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Lấy thông tin người thực hiện để check quyền
	var requester models.User
	if err := database.DB.First(&requester, userID.(uint)).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại"})
		return
	}

	svc := &groupService.GroupService{}

	// Nếu là Sếp (có PermSystemConfig), tạo trực tiếp
	if requester.PermSystemConfig {
		if err := svc.ServiceCreateGroupDirect(req, orgID.(uint), requester); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Phòng ban đã được tạo thành công!"})
	} else { // Nếu không phải sếp, tạo ticket phê duyệt
		if err := svc.ServiceCreateGroupRequest(req, orgID.(uint), requester); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Yêu cầu tạo nhóm đã được gửi, vui lòng chờ Quản trị viên duyệt!"})
	}
}

// HandleGetGroups lấy danh sách kèm Phân trang (Pagination)
func HandleGetGroups(c *gin.Context) {
	orgID, _ := c.Get("org_id")

	// Lấy tham số phân trang (mặc định page 1, limit 10)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var groups []models.PolicyGroup
	var total int64

	// Truy vấn có phân trang và sắp xếp mới nhất lên đầu
	query := database.DB.Model(&models.PolicyGroup{}).Where("org_id = ?", orgID)
	query.Count(&total)
	err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&groups).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi truy vấn danh sách"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  groups,
		"total": total,
		"page":  page,
	})
}

// HandleUpdateGroup cập nhật thông tin nhóm
func HandleUpdateGroup(c *gin.Context) {
	orgID, _ := c.Get("org_id")
	userID, _ := c.Get("user_id")
	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nhóm không hợp lệ"})
		return
	}

	var req models.PolicyGroupPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	var requester models.User
	if err := database.DB.First(&requester, userID.(uint)).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại"})
		return
	}

	svc := &groupService.GroupService{}
	if err := svc.ServiceUpdateGroupDirect(uint(groupID), req, orgID.(uint), requester); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Phòng ban đã được cập nhật thành công!"})
}

// HandleDeleteGroup xóa một nhóm
func HandleDeleteGroup(c *gin.Context) {
	orgID, _ := c.Get("org_id")
	userID, _ := c.Get("user_id")
	groupIDStr := c.Param("id")
	groupID, _ := strconv.ParseUint(groupIDStr, 10, 32)

	// Nhận lý do xóa từ Body
	var body struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp lý do xóa"})
		return
	}

	var requester models.User
	database.DB.First(&requester, userID.(uint))

	svc := &groupService.GroupService{}
	// Chuyển sang luồng tạo ticket thay vì xóa thẳng
	if err := svc.ServiceDeleteGroupRequest(uint(groupID), body.Reason, orgID.(uint), requester); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Yêu cầu xóa đã được gửi, vui lòng chờ duyệt."})
}
