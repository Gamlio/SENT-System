package policies

import (
	"fmt"
	"net/http"
	"regexp"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// --- 1. LẤY DANH SÁCH (Cập nhật để Frontend lấy theo trạng thái) ---
func GetPoliciesByCategory(c *gin.Context) {
	category := c.Query("category")
	status := c.Query("status") // Hỗ trợ lọc theo PENDING, APPROVED...
	orgID := c.GetUint("org_id")
	var list []models.UniversalPolicy
	query := database.DB.Where("org_id = ?", orgID)

	if category != "" {
		if category == "USB" {
			query = query.Where("category IN ?", []string{"USB", "DEVICE", "STORAGE"})
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

	orgID := c.GetUint("org_id")
	username, _ := c.Get("username")

	req.OrgID = orgID
	req.Category = normalizeCategory(req.Value, req.Category)
	req.ApprovalStatus = "PENDING"
	req.CreatedBy = username.(string)

	if err := database.DB.Create(&req).Error; err != nil {
		c.JSON(500, gin.H{"error": "Lỗi Database"})
		return
	}

	// Tạo vé phê duyệt tập trung[cite: 45]
	ticket := models.ApprovalTicket{
		OrgID:        orgID,
		ModuleType:   "POLICY_CREATE",
		ActionType:   "CREATE",
		TargetID:     req.ID,
		TargetName:   fmt.Sprintf("[%s] %s", req.Category, req.Value),
		Status:       "PENDING",
		RequestedBy:  username.(string),
		SnapshotData: fmt.Sprintf(`{"title": "%s", "value": "%s"}`, req.Title, req.Value),
	}
	database.DB.Create(&ticket)

	c.JSON(200, gin.H{"message": "Đã gửi yêu cầu phê duyệt chính sách."})
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
	orgIDFromToken := c.GetUint("org_id")

	// 2. HWID mà máy trạm khai báo
	hwid := c.Query("hwid")
	if hwid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thiếu thông tin HWID"})
		return
	}

	// 3. CHỐT CHẶN BẢO MẬT: Kiểm tra máy trạm có thuộc về tổ chức này không
	var agent models.Agent
	if err := database.DB.Where("hw_id = ? AND org_id = ?", hwid, orgIDFromToken).First(&agent).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Thiết bị không thuộc phạm vi quản lý của tổ chức hoặc chưa đăng ký!"})
		return
	}

	// 4. Nếu hợp lệ, tính toán bộ luật hiệu dụng[cite: 48]
	policies := CalculateEffectivePolicies(orgIDFromToken, hwid)

	c.JSON(http.StatusOK, gin.H{
		"status":   "synchronized",
		"sync_at":  time.Now(),
		"policies": policies,
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
