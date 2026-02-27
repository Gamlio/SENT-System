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

// TẠO TÀI KHOẢN MỚI (CHỈ TRONG CÔNG TY)
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
		Role              string `json:"role"`
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
	isAdmin := req.Role == "ADMIN"

	newUser := models.User{
		Username:          req.Username,
		PasswordHash:      string(hashedPassword),
		FullName:          req.FullName,
		Phone:             req.Phone,
		Email:             req.Email,
		Role:              req.Role,
		OrgID:             &orgID,
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

// LẤY DANH SÁCH TÀI KHOẢN (LỌC THEO ORG_ID)
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

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	orgID := getOrgIDFromContext(c)

	// Lấy Role của người đang thực hiện thao tác (Requester)
	requesterRole := c.GetString("role") // Middleware đã set cái này

	// 1. Tìm User cần sửa
	var targetUser models.User
	if err := database.DB.Where("id = ? AND org_id = ?", id, orgID).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng"})
		return
	}

	// --- BẢO MẬT: CHẶN QUYỀN NHÂN VIÊN ---
	if requesterRole != "ADMIN" {
		// Rule 1: Nhân viên không được phép sửa thông tin của Sếp (Admin)
		if targetUser.Role == "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền chỉnh sửa tài khoản Quản trị viên (Admin)"})
			return
		}
	}
	// --------------------------------------

	// 2. Hứng dữ liệu update
	var req struct {
		FullName           string `json:"full_name"`
		Phone              string `json:"phone"`
		Email              string `json:"email"`
		Role               string `json:"role"`
		CanViewAgents      bool   `json:"can_view_agents"`
		CanManageAgents    bool   `json:"can_manage_agents"`
		CanViewDocs        bool   `json:"can_view_docs"`
		CanManageDocs      bool   `json:"can_manage_docs"`
		CanManagePolicies  bool   `json:"can_manage_policies"`
		CanManageIncidents bool   `json:"can_manage_incidents"`
		CanManageUsers     bool   `json:"can_manage_users"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// --- BẢO MẬT: CHẶN LEO THANG ĐẶC QUYỀN ---
	if requesterRole != "ADMIN" {
		// Rule 2: Nhân viên không được phép tự set mình hoặc người khác lên ADMIN
		if req.Role == "ADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền chỉ định vai trò Admin"})
			return
		}
	}
	// ------------------------------------------

	// 3. Logic cập nhật
	// Nếu người thực hiện là Admin -> Update theo ý họ
	// Nếu là User -> Chỉ update các quyền hạn cho phép (đã lọc ở trên)

	isAdmin := req.Role == "ADMIN" // Logic cũ của bạn
	updates := map[string]interface{}{
		"full_name":            req.FullName,
		"phone":                req.Phone,
		"email":                req.Email,
		"role":                 req.Role,
		"can_view_agents":      isAdmin || req.CanViewAgents,
		"can_manage_agents":    isAdmin || req.CanManageAgents,
		"can_view_docs":        isAdmin || req.CanViewDocs,
		"can_manage_docs":      isAdmin || req.CanManageDocs,
		"can_manage_policies":  isAdmin || req.CanManagePolicies,
		"can_manage_incidents": isAdmin || req.CanManageIncidents,
		"can_manage_users":     isAdmin || req.CanManageUsers,
	}

	if err := database.DB.Model(&targetUser).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi cập nhật người dùng"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật thành công!"})
}

// XÓA TÀI KHOẢN (CŨNG CẦN BẢO VỆ)
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	orgID := getOrgIDFromContext(c)
	requesterRole := c.GetString("role")

	// Tìm user định xóa xem nó là ai
	var targetUser models.User
	if err := database.DB.Where("id = ? AND org_id = ?", id, orgID).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy người dùng"})
		return
	}

	// Rule: Chỉ Admin mới được xóa Admin
	if requesterRole != "ADMIN" && targetUser.Role == "ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không được phép xóa tài khoản Admin"})
		return
	}

	// Thực hiện xóa
	if err := database.DB.Delete(&targetUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi xóa tài khoản"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đã xóa tài khoản thành công"})
}
