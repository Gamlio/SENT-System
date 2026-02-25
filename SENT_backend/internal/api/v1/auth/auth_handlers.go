package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sent_backend/internal/auth"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Hàm sinh mã công ty ngẫu nhiên (VD: SME-A1B2C3)
func generateCompanyCode() string {
	b := make([]byte, 3) // 3 bytes = 6 ký tự Hex
	rand.Read(b)
	return "SME-" + strings.ToUpper(hex.EncodeToString(b))
}

func RegisterSMEHandler(c *gin.Context) {
	var req struct {
		CompanyName string `json:"company_name"`
		// ĐÃ XÓA CompanyCode Ở ĐÂY, KHÔNG NHẬN TỪ FRONTEND NỮA
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu sai định dạng"})
		return
	}

	// 1. Tự động sinh Mã Công Ty
	newCompanyCode := generateCompanyCode()

	// 2. Đảm bảo mã sinh ra không bị trùng (dù xác suất cực thấp)
	var existingOrg models.Organization
	for {
		if err := database.DB.Where("company_code = ?", newCompanyCode).First(&existingOrg).Error; err != nil {
			break // Nếu không tìm thấy (lỗi record not found) -> Mã này an toàn để dùng
		}
		newCompanyCode = generateCompanyCode() // Nếu trùng thì sinh lại
	}

	org := models.Organization{
		Name:              req.CompanyName,
		CompanyCode:       newCompanyCode, // Lưu mã tự sinh
		EnrollTokenPrefix: "SENT-" + newCompanyCode,
	}
	database.DB.Create(&org)

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	user := models.User{
		Username:          req.Username,
		PasswordHash:      string(hashed),
		RoleLevel:         3,
		OrgID:             &org.ID,
		CanManageAgents:   true,
		CanManagePolicies: true,
		CanManageDocs:     true,
		CanManageUsers:    true,
	}
	database.DB.Create(&user)

	// TRẢ MÃ CÔNG TY VỀ CHO FRONTEND HIỂN THỊ
	c.JSON(http.StatusOK, gin.H{
		"message":      "Đăng ký công ty và tài khoản thành công",
		"company_code": newCompanyCode,
	})
}

// ... (Hàm LoginHandler giữ nguyên như cũ vì vẫn cần nhận company_code để đăng nhập) ...
func LoginHandler(c *gin.Context) {
	var req struct {
		CompanyCode string `json:"company_code"` // Giờ đã trở thành Tùy chọn (Optional)
		Username    string `json:"username"`
		Password    string `json:"password"`
	}
	c.ShouldBindJSON(&req)

	var user models.User

	// LUỒNG 1: NẾU NGƯỜI DÙNG CÓ NHẬP MÃ CÔNG TY
	if req.CompanyCode != "" {
		var org models.Organization
		if err := database.DB.Where("company_code = ?", req.CompanyCode).First(&org).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Mã công ty không hợp lệ"})
			return
		}

		if err := database.DB.Where("username = ? AND org_id = ?", req.Username, org.ID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại trong công ty này"})
			return
		}
	} else {
		// LUỒNG 2: NẾU BỎ TRỐNG MÃ CÔNG TY -> Tìm kiếm toàn cầu
		var users []models.User
		database.DB.Where("username = ?", req.Username).Find(&users)

		if len(users) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại trên hệ thống"})
			return
		} else if len(users) > 1 {
			// Bắt trúng trường hợp trùng tên ở 2 công ty khác nhau
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tên đăng nhập này thuộc nhiều công ty. Vui lòng nhập Mã công ty để xác định!"})
			return
		}

		// Nếu tên này là duy nhất toàn cầu -> Lấy luôn user đó
		user = users[0]
	}

	// Kiểm tra mật khẩu (Dùng chung cho cả 2 luồng)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mật khẩu không chính xác"})
		return
	}

	token, _ := auth.GenerateToken(user.Username)

	// TÌM LẤY MÃ CÔNG TY ĐỂ TRẢ VỀ FRONTEND
	var currentOrg models.Organization
	database.DB.Where("id = ?", user.OrgID).First(&currentOrg)

	c.JSON(http.StatusOK, gin.H{
		"token":        token,
		"username":     user.Username,
		"level":        user.RoleLevel,
		"org_id":       user.OrgID,
		"company_code": currentOrg.CompanyCode, // BỔ SUNG DÒNG NÀY
		"permissions": gin.H{
			"agents":   user.CanManageAgents,
			"policies": user.CanManagePolicies,
			"docs":     user.CanManageDocs,
			"users":    user.CanManageUsers,
		},
	})
}
