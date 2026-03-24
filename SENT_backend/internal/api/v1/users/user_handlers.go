package users

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	userService "sent_backend/internal/service/users"
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

// CreateUser (GATE)
func CreateUser(c *gin.Context) {
	orgID := getOrgIDFromContext(c)
	requesterID := getRequesterID(c)

	var req models.UserPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Lấy thông tin người thực hiện
	var requester models.User
	database.DB.First(&requester, requesterID)

	// Gọi Service Brain
	userSvc := &userService.UserService{}
	if err := userSvc.CreateUserRequest(req, orgID, requester); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi yêu cầu cấp tài khoản, vui lòng chờ duyệt!"})
}

// UpdateUser (GATE)
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	targetID, _ := strconv.ParseUint(idStr, 10, 32)
	orgID := getOrgIDFromContext(c)
	requesterID := getRequesterID(c)

	var req models.UserPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	var requester models.User
	database.DB.First(&requester, requesterID)

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
	targetID, _ := strconv.ParseUint(idStr, 10, 32)
	orgID := getOrgIDFromContext(c)
	requesterID := getRequesterID(c)

	var requester models.User
	database.DB.First(&requester, requesterID)

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
	var usersList []models.User
	database.DB.Where("org_id = ?", orgID).Order("created_at desc").Find(&usersList)
	c.JSON(http.StatusOK, usersList)
}
