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
	database.DB.Where("org_id = ? AND is_active = ? AND approval_status = ?", orgID, true, "APPROVED").Find(&allPolicies)
	// Map để khử trùng lặp (Key = Category + Value), Luật SPECIFIC sẽ ghi đè GLOBAL
	effectiveMap := make(map[string]models.UniversalPolicy)
	for _, p := range allPolicies {
		key := p.Category + "|" + p.Value
		if p.TargetType == "GLOBAL" {
			if _, exists := effectiveMap[key]; !exists {
				effectiveMap[key] = p
			}
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
	for _, p := range effectiveMap {
		finalPolicies = append(finalPolicies, p)
	}
	return finalPolicies
}

// CheckSoftwareCompliance: (Sửa đoạn gọi hàm create)
func CheckSoftwareCompliance(agent models.Agent, data interface{}) {
	policies := CalculateEffectivePolicies(agent.OrgID, agent.HWID)

	softwareList, ok := data.([]interface{})
	if !ok {
		return
	}

	violationFound := false

	for _, item := range softwareList {
		swMap, _ := item.(map[string]interface{})
		swName := fmt.Sprintf("%v", swMap["software_name"])

		for _, p := range policies {
			if p.Category == "SOFTWARE_BLACKLIST" && p.Value == swName {
				violationFound = true
				// Gọi CreateAlert (Hàm này giờ đã tự tạo Incident + Playbook)
				CreateAlert(agent, "Software Violation",
					fmt.Sprintf("Phát hiện phần mềm cấm: %s", swName),
					"Máy trạm đã cài đặt phần mềm nằm trong danh sách đen.", "Medium")
			}
		}
	}

	// [LOGIC MỚI] NẾU KHÔNG CÒN VI PHẠM -> TỰ ĐỘNG ĐÓNG SỰ CỐ
	if !violationFound {
		AutoResolveIncident(agent.HWID, "Software Violation")
	}
}

// CheckUSBCompliance: (Sửa đoạn gọi hàm create)
func CheckUSBCompliance(agent models.Agent, data interface{}) {
	policies := CalculateEffectivePolicies(agent.OrgID, agent.HWID)
	usbList, ok := data.([]interface{})
	if !ok {
		return
	}

	for _, p := range policies {
		if p.Category != "USB" {
			continue
		}
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
			"updated_at":  time.Now(),
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

// AutoResolveIncident: Tự động đóng Case nếu Agent báo cáo đã sạch
func AutoResolveIncident(hwid string, incidentType string) {
	var incident models.Incident

	// Tìm sự cố đang mở của máy này
	err := database.DB.Where("agent_hw_id = ? AND type = ? AND status != ?", hwid, incidentType, "Resolved").
		First(&incident).Error

	if err == nil {
		// 1. Cập nhật trạng thái
		database.DB.Model(&incident).Updates(map[string]interface{}{
			"status":      "Resolved",
			"description": incident.Description + " [AUTO: Đã khắc phục]",
		})

		// 2. Thêm Activity Log (Để hiện lên Timeline bên phải)
		activity := models.IncidentActivity{
			IncidentID: incident.ID,
			UserID:     nil, // nil đại diện cho System/AI
			ActionType: "RESOLVE",
			Content:    "Hệ thống giám sát xác nhận máy trạm không còn vi phạm. Tự động đóng hồ sơ.",
			OldStatus:  incident.Status,
			NewStatus:  "Resolved",
			CreatedAt:  time.Now(),
		}
		database.DB.Create(&activity)

		// 3. Đánh dấu các Alert con là đã giải quyết
		database.DB.Model(&models.SecurityAlert{}).
			Where("incident_id = ?", incident.ID).
			Update("is_resolved", true)

		fmt.Printf("✅ AUTO-RESOLVED Incident #%d for %s\n", incident.ID, hwid)
	}
}
