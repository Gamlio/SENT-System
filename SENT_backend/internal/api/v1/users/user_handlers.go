package users

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Hàm hỗ trợ: Rút OrgID từ Context một cách an toàn (tránh lỗi Panic)
func getOrgIDFromContext(c *gin.Context) uint {
	rawOrgID, exists := c.Get("org_id")
	if !exists {
		return 0
	}
	if floatVal, ok := rawOrgID.(float64); ok {
		return uint(floatVal)
	} else if uintVal, ok := rawOrgID.(uint); ok {
		return uintVal
	}
	return 0
}

// 1. TẠO TÀI KHOẢN MỚI (CHỈ TRONG CÔNG TY)
func CreateUser(c *gin.Context) {
	orgID := getOrgIDFromContext(c)
	if orgID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không xác định được tổ chức của bạn"})
		return
	}

	var req struct {
		Username          string `json:"username"`
		Password          string `json:"password"`
		FullName          string `json:"full_name"`
		Phone             string `json:"phone"`
		Email             string `json:"email"`
		RoleLevel         int    `json:"role_level"`
		CanManageAgents   bool   `json:"can_manage_agents"`
		CanManagePolicies bool   `json:"can_manage_policies"`
		CanManageDocs     bool   `json:"can_manage_docs"`
		CanManageUsers    bool   `json:"can_manage_users"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	var existingUser models.User
	if err := database.DB.Where("username = ? AND org_id = ?", req.Username, orgID).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Tên đăng nhập này đã có người sử dụng trong công ty của bạn"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi mã hóa mật khẩu"})
		return
	}

	// Nếu tạo user Level 2 (Admin) thì tự động full quyền
	isAdmin := req.RoleLevel >= 2

	newUser := models.User{
		Username:          req.Username,
		PasswordHash:      string(hashedPassword),
		FullName:          req.FullName,
		Phone:             req.Phone,
		Email:             req.Email,
		RoleLevel:         req.RoleLevel,
		OrgID:             &orgID, // GẮN CHẶT TÀI KHOẢN VÀO CÔNG TY CỦA NGƯỜI TẠO
		CanManageAgents:   isAdmin || req.CanManageAgents,
		CanManagePolicies: isAdmin || req.CanManagePolicies,
		CanManageDocs:     isAdmin || req.CanManageDocs,
		CanManageUsers:    isAdmin || req.CanManageUsers,
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo tài khoản"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tạo tài khoản thành công!"})
}

// 2. LẤY DANH SÁCH TÀI KHOẢN (LỌC THEO ORG_ID)
func GetUsers(c *gin.Context) {
	orgID := getOrgIDFromContext(c)
	var usersList []models.User

	// CHỈ TÌM CÁC USER THUỘC VỀ CÔNG TY CỦA MÌNH
	if err := database.DB.Where("org_id = ?", orgID).Find(&usersList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi truy vấn danh sách người dùng"})
		return
	}

	c.JSON(http.StatusOK, usersList)
}

// 3. XÓA TÀI KHOẢN (KIỂM TRA CHÉO ORG_ID)
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	orgID := getOrgIDFromContext(c)

	// CHỈ CHO PHÉP XÓA NẾU USER ĐÓ CÓ CÙNG ORG_ID VỚI ADMIN
	result := database.DB.Where("org_id = ?", orgID).Delete(&models.User{}, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi xóa tài khoản"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng hoặc bạn không có quyền xóa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa tài khoản thành công"})
}
