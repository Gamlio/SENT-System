package policies

import (
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"
)

// CalculateEffectivePolicies: Tính toán bộ luật cuối cùng cho 1 máy cụ thể
func CalculateEffectivePolicies(orgID uint, agentHWID string) []models.UniversalPolicy {
	var allPolicies []models.UniversalPolicy

	// 1. Lấy TOÀN BỘ chính sách của công ty (Active = true)
	// (Lấy hết 1 lần rồi lọc trên RAM để giảm số lần query DB)
	database.DB.Where("org_id = ? AND is_active = ? AND approval_status = ?", orgID, true, "APPROVED").Find(&allPolicies)
	// Map để khử trùng lặp (Key = Category + Value), Luật SPECIFIC sẽ ghi đè GLOBAL
	effectiveMap := make(map[string]models.UniversalPolicy)

	// 2. DUYỆT QUA CÁC LUẬT GLOBAL TRƯỚC (Ưu tiên thấp)
	for _, p := range allPolicies {
		if p.TargetType == "GLOBAL" {
			key := p.Category + "|" + p.Value
			effectiveMap[key] = p
		}
	}

	// 3. DUYỆT QUA CÁC LUẬT SPECIFIC SAU (Ưu tiên cao - Ghi đè)
	for _, p := range allPolicies {
		if p.TargetType == "SPECIFIC" {
			// Kiểm tra xem máy này có nằm trong danh sách áp dụng không
			isTargeted := false
			for _, targetID := range p.TargetHWIDs {
				if targetID == agentHWID {
					isTargeted = true
					break
				}
			}

			if isTargeted {
				key := p.Category + "|" + p.Value
				// GHI ĐÈ: Nếu đã có Global, luật này sẽ thay thế nó
				effectiveMap[key] = p
			}
		}
	}

	// 4. Chuyển Map thành Slice kết quả
	var finalPolicies []models.UniversalPolicy
	for _, p := range effectiveMap {
		finalPolicies = append(finalPolicies, p)
	}

	return finalPolicies
}

// 1. Lấy danh sách luật áp dụng cho một máy cụ thể (Logic Gộp Global + Specific)
func GetEffectivePolicies(agentHWID string) []models.UniversalPolicy {
	var allPolicies []models.UniversalPolicy
	// Lấy tất cả policy active
	database.DB.Where("is_active = ?", true).Find(&allPolicies)

	effectiveMap := make(map[string]models.UniversalPolicy)

	// Duyệt Global trước
	for _, p := range allPolicies {
		if p.TargetType == "GLOBAL" {
			key := p.Category + "|" + p.Value
			effectiveMap[key] = p
		}
	}

	// Duyệt Specific (Ghi đè)
	for _, p := range allPolicies {
		if p.TargetType == "SPECIFIC" {
			// Check xem HWID này có trong danh sách TargetHWIDs không
			isTarget := false
			for _, targetID := range p.TargetHWIDs {
				if targetID == agentHWID {
					isTarget = true
					break
				}
			}
			if isTarget {
				key := p.Category + "|" + p.Value
				effectiveMap[key] = p
			}
		}
	}

	var results []models.UniversalPolicy
	for _, p := range effectiveMap {
		results = append(results, p)
	}
	return results
}

// 2. Hàm kiểm tra tuân thủ (Core Logic)
func CheckCompliance(agentHWID string, softwareList []models.SoftwareItem, usbList []models.USBLog) {
	policies := GetEffectivePolicies(agentHWID)

	// --- A. KIỂM TRA PHẦN MỀM ---
	for _, sw := range softwareList {
		for _, rule := range policies {
			if rule.Category == "SOFTWARE" && rule.PolicyType == "BLACKLIST" {
				// Nếu tên phần mềm chứa từ khóa cấm (VD: "game" trong "MyGame.exe")
				if strings.Contains(strings.ToLower(sw.SoftwareName), strings.ToLower(rule.Value)) {
					createAlert(agentHWID, "SOFTWARE_BLACKLIST", "Phát hiện phần mềm cấm: "+sw.SoftwareName, "HIGH")
				}
			}
		}
	}

	// --- B. KIỂM TRA USB ---
	for _, usb := range usbList {
		// Chỉ check sự kiện cắm vào (connected) hoặc hiện hữu
		if usb.EventType == "connected" || usb.EventType == "current" {
			isViolation := false

			// Logic 1: Check Blacklist cụ thể
			for _, rule := range policies {
				if rule.Category == "USB" && rule.PolicyType == "BLACKLIST" {
					if strings.Contains(usb.DeviceID, rule.Value) {
						isViolation = true
						createAlert(agentHWID, "USB_BLACKLIST", "Phát hiện USB cấm: "+usb.DeviceName, "CRITICAL")
					}
				}
			}

			// Logic 2: Nếu hệ thống bật chế độ Whitelist (Chỉ cho phép USB đăng ký)
			// (Đây là logic nâng cao, tạm thời ta chỉ làm Blacklist trước cho đơn giản theo yêu cầu của bạn)
			if isViolation {
				// Đã xử lý ở trên
			}
		}
	}
}

// Helper: Tạo cảnh báo vào DB
func createAlert(hwid, alertType, desc, severity string) {
	// Kiểm tra xem cảnh báo này đã tồn tại và chưa xử lý chưa (để tránh spam DB)
	var count int64
	database.DB.Model(&models.SecurityAlert{}).Where("hw_id = ? AND description = ? AND is_resolved = ?", hwid, desc, false).Count(&count)

	if count == 0 {
		alert := models.SecurityAlert{
			OrgID:       1, // Tạm hardcode
			HWID:        hwid,
			AlertType:   alertType,
			Title:       "Vi phạm chính sách an ninh",
			Description: desc,
			Severity:    severity,
			IsResolved:  false,
		}
		database.DB.Create(&alert)
	}
}
