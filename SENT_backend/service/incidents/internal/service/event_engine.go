package incidents

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

const (
	DefaultCorrelationWindow = 24 * time.Hour
)

func (s *IncidentService) applyContextualMatrix(alertType, basePriority, departmentTag string) string {
	tag := strings.ToUpper(departmentTag)
	switch alertType {
	case "USB Violation":
		if tag == "DEV" {
			return "P3"
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

func (s *IncidentService) TriggerSecurityEvent(ctx context.Context, asset models.Asset, alertType, title, description, priority string) {
	finalPriority := s.applyContextualMatrix(alertType, priority, asset.DepartmentTag)

	alert := models.SecurityAlert{
		OrgID:       asset.OrgID,
		AssetHWID:   asset.AssetHWID,
		AlertType:   alertType,
		Title:       title,
		Description: description,
		Priority:    finalPriority,
		Severity:    s.getSeverityByPriority(finalPriority),
		IsResolved:  false,
		CreatedAt:   time.Now(),
	}

	if database.SecurityAlertCollection != nil {
		res, err := database.SecurityAlertCollection.InsertOne(ctx, alert)
		if err == nil {
			if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
				alert.ID = oid
			}
		}
	}

	// Kích hoạt tính lại điểm rủi ro bất đồng bộ
	go func(hwid string) {
		// Cổng 8000 là cổng của Scoring Service trong docker-compose
		url := fmt.Sprintf("http://scoring-service:8000/api/v1/scoring/recalculate/%s", hwid)

		resp, err := http.Post(url, "application/json", nil)
		if err != nil {
			fmt.Printf("⚠️ Lỗi gọi Scoring API cho máy %s: %v\n", hwid, err)
			return
		}
		defer resp.Body.Close()
	}(asset.AssetHWID)
}

func (s *IncidentService) AutoResolveIncident(hwid string, alertType string) {
	var incident models.Incident
	err := database.DB.Where("asset_hwid = ? AND type = ? AND status != ?", hwid, alertType, "Resolved").First(&incident).Error

	if err == nil {
		database.DB.Transaction(func(tx *gorm.DB) error {
			tx.Model(&incident).Updates(map[string]interface{}{
				"status":      "Resolved",
				"description": incident.Description + " [Auto-Resolved]",
			})

			audit := models.IncidentAudit{
				IncidentID:   uint(incident.ID),
				UserID:       nil,
				ActionType:   "RESOLVE",
				Content:      "AI SOC xác nhận thiết bị đã sạch vi phạm. Tự động đóng hồ sơ.",
				OldStatus:    incident.Status,
				NewStatus:    "Resolved",
				EvidenceData: "{\"system_check\": \"verified_clean\"}",
				CreatedAt:    time.Now(),
			}

			audit.GenerateAuditHash()
			database.IncidentAuditCollection.InsertOne(context.TODO(), audit)
			return nil
		})
	}
}

func (s *IncidentService) getSeverityByPriority(p string) string {
	mapping := map[string]string{"P1": "Critical", "P2": "High", "P3": "Medium"}
	if val, ok := mapping[p]; ok {
		return val
	}
	return "Medium"
}

func (s *IncidentService) shouldUpgradePriority(current, new string) bool {
	levels := map[string]int{"P1": 3, "P2": 2, "P3": 1}
	return levels[new] > levels[current]
}
