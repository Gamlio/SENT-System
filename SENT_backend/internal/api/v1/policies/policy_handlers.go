package policies

import (
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetPoliciesByCategory: Lấy luật Software/USB/Network (JSON)
func GetPoliciesByCategory(c *gin.Context) {
	category := c.Query("category")
	var list []models.UniversalPolicy
	query := database.DB.Where("org_id = ?", 1)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	query.Find(&list)
	c.JSON(200, list)
}

// AddUniversalPolicy: Thêm luật kỹ thuật mới
func AddUniversalPolicy(c *gin.Context) {
	var req models.UniversalPolicy
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu sai"})
		return
	}
	// Mặc định OrgID = 1 nếu chưa có Auth
	if req.OrgID == 0 {
		req.OrgID = 1
	}
	database.DB.Create(&req)
	c.JSON(200, gin.H{"message": "Đã lưu chính sách kỹ thuật"})
}

// DeletePolicy: Xóa luật kỹ thuật
func DeletePolicy(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.UniversalPolicy{}, id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Lỗi xóa"})
		return
	}
	c.JSON(200, gin.H{"message": "Đã xóa chính sách"})
}

// SyncPoliciesForAgent: Agent gọi API này để lấy bộ luật "Effective"
func SyncPoliciesForAgent(c *gin.Context) {
	// Giả sử Agent gửi HWID qua Query hoặc Header (Thực tế nên lấy từ Token Claims)
	hwid := c.Query("hwid")
	if hwid == "" {
		c.JSON(400, gin.H{"error": "Thiếu HWID"})
		return
	}

	orgID := uint(1) // Tạm hardcode, sau này lấy từ Auth Middleware của Agent

	// Gọi Engine tính toán
	policies := CalculateEffectivePolicies(orgID, hwid)

	c.JSON(200, gin.H{
		"sync_time": "now",
		"count":     len(policies),
		"policies":  policies,
	})
}

// AddBulkPolicies: API chuyên dụng để nạp danh sách luật (Bulk Insert)
func AddBulkPolicies(c *gin.Context) {
	// Định nghĩa struct nhận dữ liệu riêng cho API này
	var req struct {
		Title       string   `json:"title"`
		Category    string   `json:"category"`
		PolicyType  string   `json:"policy_type"`
		TargetType  string   `json:"target_type"`
		TargetHWIDs []string `json:"target_hwids"`
		Values      []string `json:"values"` // <--- Quan trọng: Nhận mảng giá trị
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	if len(req.Values) == 0 {
		c.JSON(400, gin.H{"error": "Danh sách giá trị không được để trống"})
		return
	}

	orgID := uint(1) // Tạm hardcode

	// Bắt đầu Transaction (Lưu tất cả hoặc không lưu gì cả)
	tx := database.DB.Begin()

	for _, val := range req.Values {
		policy := models.UniversalPolicy{
			OrgID:       orgID,
			Title:       req.Title,
			Category:    req.Category,
			PolicyType:  req.PolicyType,
			TargetType:  req.TargetType,
			TargetHWIDs: req.TargetHWIDs,
			Value:       val,
			IsActive:    true,
		}

		if err := tx.Create(&policy).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Lỗi lưu DB: " + err.Error()})
			return
		}
	}

	tx.Commit()
	c.JSON(200, gin.H{"message": "Đã nạp thành công danh sách chính sách"})
}
