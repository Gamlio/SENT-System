package incidents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/scoring" // Để tính lại điểm
	"time"

	"gorm.io/gorm"
)

type IncidentService struct{}

// AddActivity: Xử lý ghi chú, đổi trạng thái và lưu trữ bằng chứng ảnh
func (s *IncidentService) AddActivity(incidentID uint, userID uint, orgID uint, actionType, content string, files map[string][]byte) error {
	var incident models.Incident
	if err := database.DB.First(&incident, incidentID).Error; err != nil {
		return fmt.Errorf("không tìm thấy sự cố")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		oldStatus := incident.Status
		newStatus := oldStatus

		// 1. Xử lý chuyển trạng thái nghiệp vụ
		if actionType == "INVESTIGATE" && oldStatus == "Open" {
			newStatus = "Investigating"
			tx.Model(&incident).Updates(map[string]interface{}{"status": newStatus, "assignee_id": userID})
		} else if actionType == "RESOLVE" {
			newStatus = "Resolved"
			tx.Model(&incident).Update("status", newStatus)
			tx.Model(&models.SecurityAlert{}).Where("incident_id = ?", incident.ID).Update("is_resolved", true)
			defer scoring.RecalculateRiskScore(incident.AgentHWID) // Tính lại điểm sau khi commit
		}

		// 2. Lưu trữ tệp tin vật lý (Phân tách theo Org)
		var imageNames []string
		uploadDir := fmt.Sprintf("uploads/org_%d/incidents", orgID)
		_ = os.MkdirAll(uploadDir, os.ModePerm)

		for filename, data := range files {
			safeName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)
			if err := os.WriteFile(filepath.Join(uploadDir, safeName), data, 0644); err == nil {
				imageNames = append(imageNames, safeName)
			}
		}
		imagesJSON, _ := json.Marshal(imageNames)

		// 3. Tạo nhật ký hoạt động (Timeline)
		activity := models.IncidentActivity{
			IncidentID: incidentID,
			UserID:     &userID,
			ActionType: actionType,
			Content:    content,
			OldStatus:  oldStatus,
			NewStatus:  newStatus,
			Images:     string(imagesJSON),
		}
		return tx.Create(&activity).Error
	})
}

// UpdatePlaybook: Lưu tiến trình xử lý sự cố
func (s *IncidentService) UpdatePlaybook(incidentID uint, steps interface{}) error {
	jsonBytes, _ := json.Marshal(map[string]interface{}{"steps": steps})
	return database.DB.Model(&models.Incident{}).Where("id = ?", incidentID).Update("playbook_progress", string(jsonBytes)).Error
}

// GetIncidentDetail: Lấy toàn bộ thông tin sự cố kèm Timeline và Người dùng
func (s *IncidentService) GetIncidentDetail(id uint) (models.Incident, error) {
	var incident models.Incident
	err := database.DB.
		Preload("Agent").
		Preload("Alerts").
		Preload("Assignee").
		Preload("Activities.User").
		Preload("Activities", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at asc")
		}).
		First(&incident, id).Error
	return incident, err
}
