package incidents

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	// Route behavior/logging through Behavior Service so only Behavior triggers scoring
	finalPriority := s.applyContextualMatrix(alertType, priority, asset.DepartmentTag)

	payload, _ := json.Marshal(map[string]interface{}{
		"org_id":        asset.OrgID,
		"asset_hwid":    asset.AssetHWID,
		"category":      alertType,
		"value":         "",
		"title":         title,
		"desc":          description,
		"base_priority": finalPriority,
	})

	// Send to Behavior Service which will persist the alert and trigger scoring
	url := "http://behavior-service:8000/api/v1/behaviors/log"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("⚠️ Lỗi gọi Behavior Service cho máy %s: %v\n", asset.AssetHWID, err)
		return
	}
	defer resp.Body.Close()
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
