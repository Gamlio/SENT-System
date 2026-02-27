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
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		// Nếu err == nil nghĩa là TÌM THẤY user này trong DB -> Chặn luôn
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tên đăng nhập này đã có người sử dụng. Vui lòng chọn tên khác!"})
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
		Username:           req.Username,
		PasswordHash:       string(hashed),
		Role:               "ADMIN", // Người tạo công ty sẽ là ADMIN
		OrgID:              &org.ID,
		CanViewAgents:      true,
		CanViewDocs:        true,
		CanManageAgents:    true,
		CanManagePolicies:  true,
		CanManageDocs:      true,
		CanManageUsers:     true,
		CanManageIncidents: true,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		// Nếu lỗi do trùng tên đăng nhập
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "uni_users_username") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tên đăng nhập đã tồn tại! Vui lòng chọn tên khác."})
			return // <--- BẮT BUỘC PHẢI CÓ RETURN ĐỂ DỪNG LẠI, KHÔNG CHẠY XUỐNG DƯỚI NỮA
		}

		// Nếu là lỗi DB khác
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi tạo tài khoản"})
		return
	}

	// 3. Đoạn này chỉ chạy khi Create() thành công mỹ mãn
	c.JSON(http.StatusOK, gin.H{
		"message":      "Đăng ký thành công",
		"company_code": newCompanyCode,
	})
}

// ... (Hàm LoginHandler giữ nguyên như cũ vì vẫn cần nhận company_code để đăng nhập) ...
func LoginHandler(c *gin.Context) {
	var req struct {
		CompanyCode string `json:"company_code"` // Không bắt buộc
		Username    string `json:"username"`
		Password    string `json:"password"`
	}
	c.ShouldBindJSON(&req)

	var user models.User

	// LUỒNG 1: USER CÓ NHẬP MÃ CÔNG TY
	if req.CompanyCode != "" {
		var org models.Organization
		if err := database.DB.Where("company_code = ?", req.CompanyCode).First(&org).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Mã công ty không tồn tại"})
			return
		}
		// Tìm User theo tên VÀ theo mã công ty
		if err := database.DB.Where("username = ? AND org_id = ?", req.Username, org.ID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại trong công ty này"})
			return
		}
	} else {
		// LUỒNG 2: USER KHÔNG NHẬP MÃ CÔNG TY -> Quét toàn hệ thống
		var users []models.User
		database.DB.Where("username = ?", req.Username).Find(&users)

		if len(users) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tài khoản không tồn tại trên hệ thống"})
			return
		} else if len(users) > 1 {
			// Đây là mấu chốt: Trùng tên ở 2 công ty khác nhau
			c.JSON(http.StatusBadRequest, gin.H{
				"error":                "Tên đăng nhập này thuộc nhiều công ty. Vui lòng nhập Mã công ty (Company Code) để đăng nhập!",
				"require_company_code": true, // Báo cho React biết để hiện ô nhập Mã CTY lên
			})
			return
		}
		// Nếu chỉ có 1 người duy nhất trên hệ thống -> Tự động lấy người đó
		user = users[0]
	}

	// Kiểm tra mật khẩu
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Mật khẩu không chính xác"})
		return
	}

	// Cấp Token
	token, _ := auth.GenerateToken(user.Username, *user.OrgID)
	var currentOrg models.Organization
	database.DB.Where("id = ?", user.OrgID).First(&currentOrg)

	// Trả về dữ liệu kèm quyền hạn chi tiết
	c.JSON(http.StatusOK, gin.H{
		"token":        token,
		"username":     user.Username,
		"role":         user.Role, // "ADMIN" hoặc "USER"
		"company_code": currentOrg.CompanyCode,
		"permissions": map[string]bool{
			"view_agents":      user.CanViewAgents,
			"manage_agents":    user.CanManageAgents,
			"view_docs":        user.CanViewDocs,
			"manage_docs":      user.CanManageDocs,
			"manage_policies":  user.CanManagePolicies,
			"manage_incidents": user.CanManageIncidents,
			"manage_users":     user.CanManageUsers,
		},
	})
}
