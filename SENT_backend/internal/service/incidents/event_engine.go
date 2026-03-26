package incidents

import (
	"errors"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/scoring"
	"time"

	"gorm.io/gorm"
)

const (
	DefaultCorrelationWindow = 24 * time.Hour // Cửa sổ thời gian gom nhóm sự cố
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

	// 2. THUẬT TOÁN CORRELATION: Tìm hoặc Tạo Incident trong một transaction để tránh race condition
	var correlatedIncident models.Incident
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		timeWindow := time.Now().Add(-DefaultCorrelationWindow)

		// Tìm Case: Cùng Máy + Cùng Loại Lỗi + Trạng thái chưa đóng + Trong cửa sổ thời gian
		err := tx.Where(
			"agent_hw_id = ? AND type = ? AND status IN ('Open', 'Investigating') AND updated_at > ?",
			agent.HWID, alertType, timeWindow,
		).First(&correlatedIncident).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// TRƯỜNG HỢP A: Chưa có Case phù hợp -> Khởi tạo Case mới tinh
			correlatedIncident = models.Incident{
				OrgID:       agent.OrgID,
				AgentHWID:   agent.HWID,
				Type:        alertType,
				Priority:    priority,
				Severity:    s.getSeverityByPriority(priority),
				Status:      "Open",
				Description: fmt.Sprintf("Hệ thống tự động phát hiện chuỗi sự kiện: %s", title),
			}
			if err := tx.Create(&correlatedIncident).Error; err != nil {
				return err
			}
		} else if err == nil {
			// TRƯỜNG HỢP B: Đã có Case -> Nâng cấp độ nghiêm trọng nếu cần
			updates := map[string]interface{}{"updated_at": time.Now()}

			if s.shouldUpgradePriority(correlatedIncident.Priority, priority) {
				updates["priority"] = priority
				updates["severity"] = s.getSeverityByPriority(priority)
			}
			if err := tx.Model(&correlatedIncident).Updates(updates).Error; err != nil {
				return err
			}
		} else {
			return err // Lỗi khác
		}

		// 3. Liên kết Alert vào Case
		return tx.Model(&alert).Update("incident_id", correlatedIncident.ID).Error
	})

	// 4. Tính lại điểm rủi ro sau khi transaction thành công
	if err == nil {
		scoring.RecalculateRiskScore(agent.HWID)
	}
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
