package security

import (
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"
	"time"
)

// CalculateEffectivePolicies: (Giữ nguyên như cũ)
func CalculateEffectivePolicies(orgID uint, hwid string) []models.UniversalPolicy {
	var allPolicies []models.UniversalPolicy
	database.DB.Where("org_id = ? AND is_active = ?", orgID, true).Find(&allPolicies)
	
	// ... (Giữ nguyên logic lọc policy cũ của bạn) ...
	effectiveMap := make(map[string]models.UniversalPolicy)
	for _, p := range allPolicies {
		key := p.Category + "|" + p.Value
		if p.TargetType == "GLOBAL" {
			if _, exists := effectiveMap[key]; !exists { effectiveMap[key] = p }
		} else if p.TargetType == "SPECIFIC" {
			for _, targetID := range p.TargetHWIDs {
				if targetID == hwid {
					effectiveMap[key] = p
					break
				}
			}
		}
	}
	var finalPolicies []models.UniversalPolicy
	for _, p := range effectiveMap { finalPolicies = append(finalPolicies, p) }
	return finalPolicies
}

// CheckSoftwareCompliance: (Sửa đoạn gọi hàm create)
func CheckSoftwareCompliance(agent models.Agent, data interface{}) {
	policies := CalculateEffectivePolicies(agent.OrgID, agent.HWID)
	softwareList, ok := data.([]interface{})
	if !ok { return }

	for _, p := range policies {
		if p.Category != "SOFTWARE" || p.PolicyType != "BLACKLIST" { continue }
		for _, item := range softwareList {
			swMap, ok := item.(map[string]interface{})
			if !ok { continue }
			swName := fmt.Sprintf("%v", swMap["software_name"])

			if strings.Contains(strings.ToLower(swName), strings.ToLower(p.Value)) {
				// Tìm thấy vi phạm -> Gọi hàm xử lý thông minh
				desc := fmt.Sprintf("Phát hiện phần mềm cấm: %s (Luật: %s)", swName, p.Title)
				handleSecurityViolation(agent, p, desc, "High")
				break // Đã bắt được lỗi này thì break loop software, check luật khác
			}
		}
	}
}

// CheckUSBCompliance: (Sửa đoạn gọi hàm create)
func CheckUSBCompliance(agent models.Agent, data interface{}) {
	policies := CalculateEffectivePolicies(agent.OrgID, agent.HWID)
	usbList, ok := data.([]interface{})
	if !ok { return }

	for _, p := range policies {
		if p.Category != "USB" { continue }
		for _, item := range usbList {
			u, _ := item.(map[string]interface{})
			deviceID := fmt.Sprintf("%v", u["device_id"])
			deviceName := fmt.Sprintf("%v", u["device_name"])

			if p.PolicyType == "BLACKLIST" && strings.Contains(strings.ToLower(deviceID), strings.ToLower(p.Value)) {
				desc := fmt.Sprintf("Thiết bị USB bị cấm: %s (%s)", deviceName, deviceID)
				handleSecurityViolation(agent, p, desc, "High")
			}
		}
	}
}

// --- LOGIC MỚI: GOM NHÓM SỰ CỐ (CASE MANAGEMENT) ---
func handleSecurityViolation(agent models.Agent, policy models.UniversalPolicy, desc string, severity string) {
	// 1. Tìm xem máy này có Incident nào đang MỞ (Open hoặc InProgress) không?
	var activeIncident models.Incident
	err := database.DB.Where("agent_hw_id = ? AND status IN ?", agent.HWID, []string{"Open", "InProgress"}).First(&activeIncident).Error

	var incidentID uint

	if err == nil {
		// TRƯỜNG HỢP 1: ĐÃ CÓ SỰ CỐ -> CẬP NHẬT
		incidentID = activeIncident.ID
		fmt.Printf(">> [GOM NHÓM] Phát hiện cảnh báo mới cho Case #%d (Máy: %s)\n", incidentID, agent.Hostname)
		
		// Cập nhật thời gian update để nó nổi lên đầu danh sách
		database.DB.Model(&activeIncident).Updates(map[string]interface{}{
			"updated_at": time.Now(),
			"description": activeIncident.Description + " | " + desc, // Nối thêm mô tả (hoặc giữ nguyên tùy bạn)
		})

	} else {
		// TRƯỜNG HỢP 2: CHƯA CÓ -> TẠO MỚI (CASE MỚI)
		newIncident := models.Incident{
			OrgID:       agent.OrgID,
			AgentHWID:   agent.HWID,
			Type:        "Policy Violation (" + policy.Category + ")",
			Severity:    severity,
			Status:      "Open",
			Description: desc, // Mô tả ban đầu
			AIAnalysis:  "Hệ thống tự động khởi tạo hồ sơ sự cố mới.",
		}
		database.DB.Create(&newIncident)
		incidentID = newIncident.ID
		fmt.Printf(">> [TẠO MỚI] Khởi tạo Case #%d cho máy %s\n", incidentID, agent.Hostname)
	}

	// 2. Luôn luôn tạo Alert (Bằng chứng) và gắn vào IncidentID tìm được
	// (Nhưng kiểm tra trùng lặp alert để tránh spam 1000 dòng giống hệt nhau trong 1 phút)
	var duplicateAlert int64
	database.DB.Model(&models.SecurityAlert{}).
		Where("incident_id = ? AND description = ? AND created_at > ?", incidentID, desc, time.Now().Add(-5*time.Minute)).
		Count(&duplicateAlert)

	if duplicateAlert == 0 {
		alert := models.SecurityAlert{
			OrgID:       agent.OrgID,
			HWID:        agent.HWID,
			IncidentID:  &incidentID, // Gắn vào Case
			AlertType:   policy.Category + "_VIOLATION",
			Title:       policy.Title,
			Description: desc,
			Severity:    severity,
		}
		database.DB.Create(&alert)
	}
}