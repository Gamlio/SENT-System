package incidents

import (
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/scoring"
	"time"

	"gorm.io/gorm"
)

// TriggerSecurityEvent: Nhận tín hiệu từ Sensor và đưa vào quy trình Triage (Phân loại)
func (s *IncidentService) TriggerSecurityEvent(agent models.Agent, alertType, title, description, priority string) {
	// 1. Tạo bản ghi Alert (Bằng chứng thô)
	alert := models.SecurityAlert{
		OrgID:       agent.OrgID,
		HWID:        agent.HWID,
		AlertType:   alertType,
		Title:       title,
		Description: description,
		Severity:    s.getSeverityByPriority(priority), // Map P1->Critical...
		IsResolved:  false,
	}
	database.DB.Create(&alert)

	// 2. THUẬT TOÁN CORRELATION: Kiểm tra xem có Case nào tương tự đang mở không?
	var activeIncident models.Incident
	timeWindow := time.Now().Add(-24 * time.Hour)

	// Tìm Case: Cùng Máy + Cùng Loại Lỗi + Trạng thái chưa đóng + Trong 24h
	err := database.DB.Where(
		"agent_hw_id = ? AND type = ? AND status IN ('Open', 'Investigating') AND updated_at > ?",
		agent.HWID, alertType, timeWindow,
	).First(&activeIncident).Error

	if err != nil {
		// TRƯỜNG HỢP A: Chưa có Case phù hợp -> Khởi tạo Case mới tinh
		activeIncident = models.Incident{
			OrgID:       agent.OrgID,
			AgentHWID:   agent.HWID,
			Type:        alertType,
			Priority:    priority,
			Severity:    s.getSeverityByPriority(priority),
			Status:      "Open",
			Description: fmt.Sprintf("Hệ thống tự động phát hiện chuỗi sự kiện: %s", title),
		}
		database.DB.Create(&activeIncident)
	} else {
		// TRƯỜNG HỢP B: Đã có Case -> Nối Alert vào và nâng cấp độ nghiêm trọng nếu cần
		updates := map[string]interface{}{"updated_at": time.Now()}

		if s.shouldUpgradePriority(activeIncident.Priority, priority) {
			updates["priority"] = priority
			updates["severity"] = s.getSeverityByPriority(priority)
		}
		database.DB.Model(&activeIncident).Updates(updates)
	}

	// 3. Liên kết Alert vào Case và tính lại điểm rủi ro
	database.DB.Model(&alert).Update("incident_id", activeIncident.ID)
	scoring.RecalculateRiskScore(agent.HWID)
}

// AutoResolveIncident: Tự động đóng Case nếu Agent báo cáo trạng thái đã an toàn
func (s *IncidentService) AutoResolveIncident(hwid string, alertType string) {
	var incident models.Incident
	err := database.DB.Where("agent_hw_id = ? AND type = ? AND status != ?", hwid, alertType, "Resolved").First(&incident).Error

	if err == nil {
		database.DB.Transaction(func(tx *gorm.DB) error {
			tx.Model(&incident).Updates(map[string]interface{}{
				"status":      "Resolved",
				"description": incident.Description + " [Hệ thống tự động đóng do vi phạm đã được khắc phục]",
			})

			// Đóng toàn bộ alert liên quan
			tx.Model(&models.SecurityAlert{}).Where("incident_id = ?", incident.ID).Update("is_resolved", true)

			// Ghi log vào Timeline
			tx.Create(&models.IncidentActivity{
				IncidentID: incident.ID,
				ActionType: "RESOLVE",
				Content:    "AI SOC xác nhận thiết bị đã sạch vi phạm. Tự động kết thúc hồ sơ.",
				OldStatus:  incident.Status,
				NewStatus:  "Resolved",
			})
			return nil
		})
		scoring.RecalculateRiskScore(hwid) // Cập nhật lại Risk Score về 0
	}
}

// Helpers nội bộ
func (s *IncidentService) getSeverityByPriority(p string) string {
	mapping := map[string]string{"P1": "Critical", "P2": "High", "P3": "Medium", "P4": "Low"}
	if val, ok := mapping[p]; ok {
		return val
	}
	return "Low"
}

func (s *IncidentService) shouldUpgradePriority(current, new string) bool {
	levels := map[string]int{"P1": 4, "P2": 3, "P3": 2, "P4": 1}
	return levels[new] > levels[current]
}
