package security

import (
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/scoring"
	"time"
)

// TriggerSecurityEvent: Nhận lỗi và tự động phân bổ vào Hồ sơ Sự cố (Incident)
func TriggerSecurityEvent(agent models.Agent, alertType, title, description, priority string) {
	// 1. Lưu Cảnh báo (Alert) thô vào DB
	alert := models.SecurityAlert{
		OrgID:       agent.OrgID,
		HWID:        agent.HWID,
		AlertType:   alertType, // VD: "Malware Detected", "USB Violation", "Firewall Disabled"
		Title:       title,
		Description: description,
		IsResolved:  false,
	}
	database.DB.Create(&alert)

	// 2. THUẬT TOÁN CORRELATION: EXACT MATCHING (1 Loại Lỗi = 1 Case Riêng)
	var activeIncident models.Incident

	// Khung thời gian: Chỉ gom vào Case nếu Case đó có hoạt động trong 24h qua
	timeWindow := time.Now().Add(-24 * time.Hour)

	// QUY TẮC VÀNG: Cùng Máy + Khớp 100% Loại Lỗi + Đang Mở + Trong 24h
	err := database.DB.Where(
		"agent_hw_id = ? AND status IN ('Open', 'Investigating') AND type = ? AND updated_at > ?",
		agent.HWID, alertType, timeWindow,
	).First(&activeIncident).Error

	if err != nil {
		// CHƯA CÓ CASE PHÙ HỢP: Lập Case mới tinh
		activeIncident = models.Incident{
			AgentHWID: agent.HWID,
			OrgID:     agent.OrgID,
			Type:      alertType, // Tên Case = Tên Loại Lỗi
			Priority:  priority,
			Severity:  getSeverityByPriority(priority),
			Status:    "Open",
		}
		database.DB.Create(&activeIncident)
		fmt.Printf("🚨 [NEW CASE] Đã lập hồ sơ sự cố mới: %s cho máy %s\n", alertType, agent.Hostname)
	} else {
		// ĐÃ CÓ CASE: Nối Alert vào và làm mới thời gian
		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}

		// Nâng cấp độ nghiêm trọng nếu Alert mới nguy hiểm hơn
		if shouldUpgradePriority(activeIncident.Priority, priority) {
			updates["priority"] = priority
			updates["severity"] = getSeverityByPriority(priority)
		}
		database.DB.Model(&activeIncident).Updates(updates)
		fmt.Printf("🔄 [UPDATE CASE] Đã gom cảnh báo vào Case: %s của máy %s\n", alertType, agent.Hostname)
	}

	// 3. Liên kết Cảnh báo vào Hồ sơ Sự cố
	database.DB.Model(&alert).Update("incident_id", activeIncident.ID)

	// 4. Kích hoạt tính lại Điểm Rủi Ro (Risk Score) ngay lập tức
	scoring.RecalculateRiskScore(agent.HWID)
}

func getSeverityByPriority(p string) string {
	switch p {
	case "P1":
		return "Critical"
	case "P2":
		return "High"
	case "P3":
		return "Medium"
	default:
		return "Low"
	}
}

func shouldUpgradePriority(current, new string) bool {
	levels := map[string]int{"P1": 4, "P2": 3, "P3": 2, "P4": 1}
	return levels[new] > levels[current]
}
