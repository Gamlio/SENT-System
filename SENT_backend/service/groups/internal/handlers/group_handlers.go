package handlers

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	groupService "SENT_backend/service/groups/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func HandleCreateGroup(c *gin.Context) {
	orgID := c.GetUint("org_id")
	userID := c.GetUint("user_id")

	// Validate context values
	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được để trống"})
		return
	}
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UserID không được để trống"})
		return
	}

	var req models.PolicyGroupPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ. Vui lòng kiểm tra: name (3-100 ký tự), description (tùy chọn, max 500 ký tự). Chi tiết: " + err.Error()})
		return
	}

	var requester models.User
	if err := database.DB.First(&requester, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại"})
		return
	}

	svc := &groupService.GroupService{}

	if requester.PermSystemConfig {
		if err := svc.ServiceCreateGroupDirect(req, orgID, requester); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Phòng ban đã được tạo thành công!"})
	} else {
		if err := svc.ServiceCreateGroupRequest(req, orgID, requester); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Yêu cầu tạo nhóm đã được gửi, vui lòng chờ Quản trị viên duyệt!"})
	}
}

func HandleGetGroups(c *gin.Context) {
	orgIDVal, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được tìm thấy"})
		return
	}

	orgID, ok := orgIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không hợp lệ"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var groups []models.PolicyGroup
	var total int64

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

func HandleUpdateGroup(c *gin.Context) {
	orgIDVal, _ := c.Get("org_id")
	userIDVal, _ := c.Get("user_id")

	orgID, ok := orgIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không hợp lệ"})
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UserID không hợp lệ"})
		return
	}

	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nhóm không hợp lệ"})
		return
	}

	var req models.PolicyGroupPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ. Vui lòng kiểm tra: name (3-100 ký tự), description (tùy chọn, max 500 ký tự). Chi tiết: " + err.Error()})
		return
	}

	var requester models.User
	if err := database.DB.First(&requester, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại"})
		return
	}

	svc := &groupService.GroupService{}
	if err := svc.ServiceUpdateGroupDirect(uint(groupID), req, orgID, requester); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Phòng ban đã được cập nhật thành công!"})
}

func HandleDeleteGroup(c *gin.Context) {
	orgIDVal, _ := c.Get("org_id")
	userIDVal, _ := c.Get("user_id")
	
	orgID, ok := orgIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không hợp lệ"})
		return
	}
	
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UserID không hợp lệ"})
		return
	}
	
	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nhóm không hợp lệ"})
		return
	}

	var body struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng cung cấp lý do xóa"})
		return
	}

	var requester models.User
	if err := database.DB.First(&requester, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại"})
		return
	}

	svc := &groupService.GroupService{}

	if err := svc.ServiceDeleteGroupRequest(uint(groupID), body.Reason, orgID, requester); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Yêu cầu xóa đã được gửi, vui lòng chờ duyệt."})
}

func HandleGetGroupDetail(c *gin.Context) {
	orgID := c.GetUint("org_id")
	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OrgID không được tìm thấy"})
		return
	}
	
	groupIDStr := c.Param("id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID nhóm không hợp lệ"})
		return
	}

	var group models.PolicyGroup
	if err := database.DB.Where("id = ? AND org_id = ?", uint(groupID), orgID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy phòng ban"})
		return
	}

	var assets []models.Asset
	database.DB.Where("group_id = ? AND org_id = ?", uint(groupID), orgID).Find(&assets)

	c.JSON(http.StatusOK, gin.H{
		"group":  group,
		"assets": assets,
	})
}
