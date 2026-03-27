package incidents

import (
	"fmt"
	"sent_backend/internal/models"
	"sent_backend/internal/service/scoring"
	"time"

	"gorm.io/gorm"
)

// IncidentService handles business logic related to incidents
type IncidentService struct {
	db *gorm.DB
}

// NewIncidentService creates a new IncidentService instance
func NewIncidentService(db *gorm.DB) *IncidentService {
	return &IncidentService{
		db: db,
	}
}

// CreateIncidentFromAlert processes a security alert and creates an incident, updating agent scores.
func (s *IncidentService) CreateIncidentFromAlert(alert *models.SecurityAlert, agent *models.Agent) (*models.Incident, error) {
	// Dùng logic phân loại trực tiếp nếu không có Matrix Service riêng
	priority := "P4" // Mặc định
	if alert.Priority != "" {
		priority = alert.Priority
	}

	// 1. Create the Incident
	incident := models.Incident{
		OrgID:       alert.OrgID,
		AgentHWID:   alert.HWID, // SỬA: Dùng HWID thay vì ID
		Type:        alert.AlertType,
		Severity:    alert.Severity,
		Priority:    priority,
		Status:      "Open",
		Description: alert.Description,
	}

	if err := s.db.Create(&incident).Error; err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}

	// Link the alert to the new incident
	alert.IncidentID = &incident.ID
	if err := s.db.Save(alert).Error; err != nil {
		fmt.Printf("Warning: Failed to link alert %d to incident %d: %v\n", alert.ID, incident.ID, err)
	}

	// 2. SỬA: Cập nhật LastIncidentAt trực tiếp bằng GORM thông qua HWID
	now := time.Now()
	if err := s.db.Model(&models.Agent{}).Where("hw_id = ?", agent.HWID).Update("last_incident_at", &now).Error; err != nil {
		fmt.Printf("Warning: Failed to update LastIncidentAt for agent %s: %v\n", agent.HWID, err)
	}

	// 3. SỬA: Gọi trực tiếp hàm tính điểm toàn cục trong score_service.go
	scoring.RecalculateRiskScore(agent.HWID)

	return &incident, nil
}

// ResolveIncident handles resolving an incident and updating agent scores.
func (s *IncidentService) ResolveIncident(incidentID uint, agentHWID string, resolutionSummary string) error {
	var incident models.Incident
	if err := s.db.First(&incident, incidentID).Error; err != nil {
		return fmt.Errorf("incident not found: %w", err)
	}

	// SỬA: Bỏ trường ResolvedAt vì không có trong models. Cập nhật Status và Summary
	incident.Status = "Resolved"
	incident.ResolutionSummary = resolutionSummary

	if err := s.db.Save(&incident).Error; err != nil {
		return fmt.Errorf("failed to update incident status: %w", err)
	}

	// Close all associated alerts
	if err := s.db.Model(&models.SecurityAlert{}).Where("incident_id = ?", incidentID).Update("is_resolved", true).Error; err != nil {
		fmt.Printf("Warning: Failed to resolve associated alerts for incident %d: %v\n", incidentID, err)
	}

	// SỬA: Gọi tính lại điểm rủi ro bằng hàm toàn cục
	scoring.RecalculateRiskScore(agentHWID)

	return nil
}

// GetAllIncidents fetches all incidents with preloaded agent information.
func (s *IncidentService) GetAllIncidents(orgID uint) ([]models.Incident, error) {
	var incidents []models.Incident
	err := s.db.Where("org_id = ?", orgID).Preload("Agent").Order("created_at desc").Find(&incidents).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch incidents: %w", err)
	}
	return incidents, nil
}

// GetIncidentByID fetches a single incident with all related preloads.
func (s *IncidentService) GetIncidentByID(incidentID uint) (*models.Incident, error) {
	var incident models.Incident
	err := s.db.
		Preload("Agent").
		Preload("Alerts").
		Preload("Assignee").
		Preload("Activities.User").
		Preload("Activities", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at asc")
		}).
		First(&incident, incidentID).Error
	if err != nil {
		return nil, fmt.Errorf("incident not found: %w", err)
	}
	return &incident, nil
}

// UpdateIncidentStatus updates the status of an incident.
func (s *IncidentService) UpdateIncidentStatus(incidentID uint, newStatus string, assigneeID *uint) error {
	updates := map[string]interface{}{
		"status": newStatus,
	}
	if assigneeID != nil {
		updates["assignee_id"] = assigneeID
	}
	if err := s.db.Model(&models.Incident{}).Where("id = ?", incidentID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update incident status: %w", err)
	}
	return nil
}

// UpdatePlaybookProgress updates the playbook progress for an incident.
func (s *IncidentService) UpdatePlaybookProgress(incidentID uint, progressJSON string) error {
	if err := s.db.Model(&models.Incident{}).Where("id = ?", incidentID).Update("playbook_progress", progressJSON).Error; err != nil {
		return fmt.Errorf("failed to update playbook progress: %w", err)
	}
	return nil
}

// AddIncidentActivity adds a new activity to an incident.
func (s *IncidentService) AddIncidentActivity(activity *models.IncidentActivity) error {
	if err := s.db.Create(activity).Error; err != nil {
		return fmt.Errorf("failed to add incident activity: %w", err)
	}
	return nil
}
