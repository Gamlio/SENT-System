package policies

import (
	"fmt"
	"regexp"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"

	"github.com/gin-gonic/gin"
)

// --- 1. LẤY DANH SÁCH (Cập nhật để Frontend lấy theo trạng thái) ---
func GetPoliciesByCategory(c *gin.Context) {
	category := c.Query("category")
	status := c.Query("status") // Hỗ trợ lọc theo PENDING, APPROVED...
	var list []models.UniversalPolicy

	query := database.DB.Where("org_id = ?", 1)

	if category != "" {
		if category == "USB" {
			query = query.Where("category IN ?", []string{"USB", "DEVICE", "STORAGE", "OTHER"})
		} else {
			query = query.Where("category = ?", category)
		}
	}

	if status != "" {
		query = query.Where("approval_status = ?", status)
	}

	query.Order("created_at desc").Find(&list)
	c.JSON(200, list)
}

// --- 2. THÊM LUẬT MỚI LẺ (Sẽ bị đưa vào PENDING) ---
func AddUniversalPolicy(c *gin.Context) {
	var req models.UniversalPolicy
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu sai"})
		return
	}

	req.Category = normalizeCategory(req.Value, req.Category)
	if req.OrgID == 0 {
		req.OrgID = 1
	}

	req.ApprovalStatus = "PENDING"
	req.CreatedBy = "System_Admin"

	// 1. Tạo Policy trước (để lấy ID)
	if err := database.DB.Create(&req).Error; err != nil {
		c.JSON(500, gin.H{"error": "Lỗi Database"})
		return
	}

	// 2. [MỚI] TẠO TICKET VÀO TRUNG TÂM PHÊ DUYỆT
	ticket := models.ApprovalTicket{
		OrgID:        req.OrgID,
		ModuleType:   "POLICY_CREATE",
		ActionType:   "CREATE",
		TargetID:     req.ID, // ID của Policy vừa tạo
		TargetName:   fmt.Sprintf("[%s] %s", req.Category, req.Value),
		Status:       "PENDING",
		RequestedBy:  "Nhân viên SOC", // Tạm hardcode
		SnapshotData: fmt.Sprintf(`{"title": "%s", "value": "%s"}`, req.Title, req.Value),
	}
	database.DB.Create(&ticket)

	c.JSON(200, gin.H{"message": "Đã gửi yêu cầu. Chờ phê duyệt.", "data": req})
}

// --- 3. THÊM NHIỀU TỪ EXCEL/WORD (Sẽ bị đưa vào PENDING) ---
func AddBulkPolicies(c *gin.Context) {
	type ExcelRow struct {
		Title      string `json:"title"`
		Category   string `json:"category"`
		Value      string `json:"value"`
		PolicyType string `json:"policy_type"`
		TargetType string `json:"target_type"`
	}

	var req struct {
		Policies []ExcelRow `json:"policies"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu form Excel không hợp lệ"})
		return
	}

	if len(req.Policies) == 0 {
		c.JSON(400, gin.H{"error": "File Excel không có dữ liệu"})
		return
	}

	tx := database.DB.Begin()
	count := 0

	for _, row := range req.Policies {
		val := strings.TrimSpace(row.Value)
		if val == "" {
			continue
		}

		pType := strings.ToUpper(strings.TrimSpace(row.PolicyType))
		if pType == "" {
			pType = "BLACKLIST"
		}

		tType := strings.ToUpper(strings.TrimSpace(row.TargetType))
		if tType == "" {
			tType = "GLOBAL"
		}

		finalCategory := normalizeCategory(val, row.Category)

		policy := models.UniversalPolicy{
			OrgID:          1,
			Title:          strings.TrimSpace(row.Title),
			Category:       finalCategory,
			Value:          val,
			PolicyType:     pType,
			TargetType:     tType,
			IsActive:       true,
			ApprovalStatus: "PENDING",
			CreatedBy:      "System_Import",
		}

		if policy.Title == "" {
			policy.Title = "Imported: " + val
		}

		// 1. Lưu Policy
		if err := tx.Create(&policy).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Lỗi lưu tại dòng: " + val})
			return
		}

		// 2. [MỚI] TẠO TICKET CHO TỪNG DÒNG EXCEL
		ticket := models.ApprovalTicket{
			OrgID:        1,
			ModuleType:   "POLICY_CREATE",
			ActionType:   "CREATE",
			TargetID:     policy.ID,
			TargetName:   fmt.Sprintf("[%s] %s", policy.Category, policy.Value),
			Status:       "PENDING",
			RequestedBy:  "System_Import",
			SnapshotData: fmt.Sprintf(`{"title": "%s", "value": "%s"}`, policy.Title, policy.Value),
		}

		if err := tx.Create(&ticket).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Lỗi tạo vé phê duyệt tại dòng: " + val})
			return
		}

		count++
	}

	tx.Commit()
	c.JSON(200, gin.H{
		"message": fmt.Sprintf("Đã nạp thành công %d luật từ Excel. Đang chờ phê duyệt.", count),
		"count":   count,
	})
}

// --- 4. [MỚI] API PHÊ DUYỆT HOẶC TỪ CHỐI (Dành riêng cho SOC Manager) ---
func ReviewPolicy(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"` // Truyền lên "APPROVED" hoặc "REJECTED"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	// Xác thực: Chỗ này sau này bạn sẽ check xem JWT Role có phải là "MANAGER" hay không
	// giả sử tạm thời ai cũng duyệt được để test:
	reviewerName := "SOC_Manager" // Sẽ lấy từ JWT Token

	var policy models.UniversalPolicy
	if err := database.DB.First(&policy, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy chính sách"})
		return
	}

	// Cập nhật trạng thái
	policy.ApprovalStatus = req.Status
	policy.ApprovedBy = reviewerName

	if err := database.DB.Save(&policy).Error; err != nil {
		c.JSON(500, gin.H{"error": "Không thể cập nhật trạng thái"})
		return
	}

	c.JSON(200, gin.H{"message": "Đã " + req.Status + " chính sách thành công"})
}

// DeletePolicy: (Giữ nguyên)
func DeletePolicy(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.UniversalPolicy{}, id).Error; err != nil {
		c.JSON(500, gin.H{"error": "Không thể xóa"})
		return
	}
	c.JSON(200, gin.H{"message": "Đã xóa"})
}

// DeleteBulkPolicies: Xóa nhiều chính sách cùng lúc
func DeleteBulkPolicies(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"error": "Không có chính sách nào để xóa"})
		return
	}

	if err := database.DB.Delete(&models.UniversalPolicy{}, req.IDs).Error; err != nil {
		c.JSON(500, gin.H{"error": "Không thể xóa"})
		return
	}
	c.JSON(200, gin.H{"message": "Đã xóa thành công"})
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

// --- HÀM KIỂM TRA SỐ (Dùng cho kiểm tra Port) ---
func isNumeric(s string) bool {
	// Dọn dẹp khoảng trắng trước khi check
	s = strings.TrimSpace(s)
	match, _ := regexp.MatchString("^[0-9]+$", s)
	return match
}

// --- HÀM CHUẨN HÓA DANH MỤC THÔNG MINH ---
func normalizeCategory(value string, inputCategory string) string {
	// 1. Dọn dẹp dữ liệu rác từ Excel (khoảng trắng thừa, chữ hoa chữ thường)
	val := strings.ToUpper(strings.TrimSpace(value))
	cat := strings.ToUpper(strings.TrimSpace(inputCategory))

	// 2. ƯU TIÊN 1: TÔN TRỌNG FORM MẪU EXCEL
	// Nếu người dùng đã chọn từ Dropdown trong Excel, lấy luôn kết quả đó.
	// Bắt luôn cả trường hợp user gõ nhầm (ví dụ: "NETWOR" trong file mẫu của bạn)
	if cat == "SOFTWARE" || cat == "APP" {
		return "SOFTWARE"
	}
	if cat == "USB" || cat == "DEVICE" {
		return "USB"
	}
	if cat == "NETWORK" || cat == "NETWOR" || cat == "PORT" {
		return "NETWORK"
	}

	// 3. ƯU TIÊN 2: FALLBACK (TỰ ĐOÁN) NẾU FORM BỊ BỎ TRỐNG HOẶC LỖI
	if cat == "" || cat == "OTHER" {
		// Đoán là Phần mềm
		if strings.HasSuffix(val, ".EXE") || strings.HasSuffix(val, ".MSI") || strings.HasSuffix(val, ".BAT") {
			return "SOFTWARE"
		}
		// Đoán là Mạng / Port
		if isNumeric(val) || (strings.Contains(val, ".") && !strings.Contains(val, "EXE")) {
			return "NETWORK"
		}
		// Nếu không phải phần mềm, không phải Port -> Gom hết vào USB (ví dụ: Generic Flash Disk)
		return "USB"
	}

	// Nếu user nhập một danh mục lạ hoắc, vẫn giữ nguyên để lưu vào DB (dễ debug)
	if cat != "" {
		return cat
	}
	return "OTHER"
}
