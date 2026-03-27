package incidents

import (
	"errors"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/scoring"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	DefaultCorrelationWindow = 24 * time.Hour // Cửa sổ thời gian gom nhóm sự cố
)

func (s *IncidentService) applyContextualMatrix(alertType, basePriority, departmentTag string) string {
	tag := strings.ToUpper(departmentTag)
	switch alertType {
	case "USB Violation":
		if tag == "DEV" {
			return "P4"
		}
		if tag == "FINANCE" {
			return "P1"
		}
		if tag == "PROD" {
			return "P2"
		}
	case "Software Violation", "Zero Trust Violation", "Defense Evasion":
		if tag == "DEV" {
			return "P3"
		}
		if tag == "FINANCE" || tag == "PROD" {
			return "P1"
		}
	case "Unauthorized Port", "Firewall Disabled":
		if tag == "DEV" {
			return "P2"
		}
		if tag == "FINANCE" || tag == "PROD" {
			return "P1"
		}
	}
	return basePriority
}

// TriggerSecurityEvent: Nhận tín hiệu từ Sensor và đưa vào quy trình Triage (Phân loại)
func (s *IncidentService) TriggerSecurityEvent(agent models.Agent, alertType, title, description, priority string) {
	// 1. Lọc qua Ma trận Ngữ cảnh để lấy Priority chuẩn
	finalPriority := s.applyContextualMatrix(alertType, priority, agent.DepartmentTag)

	alert := models.SecurityAlert{
		OrgID:       agent.OrgID,
		HWID:        agent.HWID,
		AlertType:   alertType,
		Title:       title,
		Description: description,
		Priority:    finalPriority,
		Severity:    s.getSeverityByPriority(finalPriority),
		IsResolved:  false,
	}
	database.DB.Create(&alert)

	var correlatedIncident models.Incident
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		timeWindow := time.Now().Add(-DefaultCorrelationWindow)
		err := tx.Where("agent_hw_id = ? AND type = ? AND status IN ('Open', 'Investigating') AND updated_at > ?",
			agent.HWID, alertType, timeWindow).First(&correlatedIncident).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			correlatedIncident = models.Incident{
				OrgID:       agent.OrgID,
				AgentHWID:   agent.HWID,
				Type:        alertType,
				Priority:    finalPriority,
				Severity:    s.getSeverityByPriority(finalPriority),
				Status:      "Open",
				Description: fmt.Sprintf("Hệ thống tự động phát hiện: %s", title),
			}
			if err := tx.Create(&correlatedIncident).Error; err != nil {
				return err
			}
		} else if err == nil {
			updates := map[string]interface{}{"updated_at": time.Now()}
			if s.shouldUpgradePriority(correlatedIncident.Priority, finalPriority) {
				updates["priority"] = finalPriority
				updates["severity"] = s.getSeverityByPriority(finalPriority)
			}
			tx.Model(&correlatedIncident).Updates(updates)
		}
		return tx.Model(&alert).Update("incident_id", correlatedIncident.ID).Error
	})

	// 4. Gọi Engine tính điểm v6.0
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
