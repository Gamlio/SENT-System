package service

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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BehaviorService struct{}

// LogBehavior now accepts a basePriority from the calling service (e.g., asset-service)
func (s *BehaviorService) LogBehavior(ctx context.Context, asset models.Asset, category, value, title, desc, basePriority string) (*models.SecurityAlert, error) {
	checkURL := "http://policy-service:8000/api/v1/policies/check-violation"
	payload, _ := json.Marshal(map[string]interface{}{
		"org_id":     asset.OrgID,
		"asset_hwid": asset.AssetHWID,
		"category":   category,
		"value":      value,
	})

	resp, err := http.Post(checkURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối tới Policy Service: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		IsViolation bool   `json:"is_violation"`
		Reason      string `json:"reason"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	// if !result.IsViolation {
	// 	return nil, nil
	// }

	// Use the priority from the upstream service as the base, instead of hardcoding "P3"
	finalPriority := s.applyContextualMatrix(category, basePriority, asset.DepartmentTag)

	// Only append the reason if it's not empty
	finalDesc := desc
	if result.Reason != "" {
		finalDesc = fmt.Sprintf("%s | Lý do: %s", desc, result.Reason)
	}

	alert := models.SecurityAlert{
		OrgID:       uint(asset.OrgID),
		AssetHWID:   asset.AssetHWID,
		AlertType:   category + " Violation",
		Title:       title,
		Description: finalDesc,
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

	return &alert, nil
}

func (s *BehaviorService) applyContextualMatrix(category, basePriority, tag string) string {
	// This logic is synchronized with incident_service to ensure consistency
	upperTag := strings.ToUpper(tag)

	switch category {
	case "USB Violation":
		if upperTag == "DEV" {
			return "P3"
		}
		if upperTag == "FINANCE" {
			return "P1"
		}
		if upperTag == "PROD" {
			return "P2"
		}
	// software.go can send "Software Violation", "Zero Trust Violation", "Defense Evasion"
	// antivirus.go can send "Malware"
	case "Software Violation", "Zero Trust Violation", "Defense Evasion", "Malware":
		if upperTag == "DEV" {
			return "P3"
		}
		if upperTag == "FINANCE" || upperTag == "PROD" {
			return "P1"
		}
	case "Unauthorized Port", "Firewall Disabled":
		if upperTag == "DEV" {
			return "P2"
		}
		if upperTag == "FINANCE" || upperTag == "PROD" {
			return "P1"
		}
	}
	return basePriority // Return the original priority if no specific rule matches
}

func (s *BehaviorService) getSeverityByPriority(p string) string {
	mapping := map[string]string{"P1": "Critical", "P2": "High", "P3": "Medium"}
	if val, ok := mapping[p]; ok {
		return val
	}
	return "Medium"
}

func (s *BehaviorService) GetBehaviors(ctx context.Context, orgID uint, page, limit int64) ([]models.SecurityAlert, int64, error) {
	alerts := []models.SecurityAlert{}
	if database.SecurityAlertCollection == nil {
		return alerts, 0, nil
	}

	filter := bson.M{"org_id": orgID}
	total, err := database.SecurityAlertCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := (page - 1) * limit
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := database.SecurityAlertCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &alerts); err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}
func (s *BehaviorService) GetBehaviorDetail(ctx context.Context, id string) (*models.SecurityAlert, error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}

	var alert models.SecurityAlert
	err := database.SecurityAlertCollection.FindOne(ctx, filter).Decode(&alert)
	return &alert, err
}
func (s *BehaviorService) CreateIncident(ctx context.Context, alertID string, orgID uint, manualData models.Incident) (*models.Incident, error) {
	objID, _ := primitive.ObjectIDFromHex(alertID)
	var alert models.SecurityAlert
	err := database.SecurityAlertCollection.FindOne(ctx, bson.M{"_id": objID, "org_id": orgID}).Decode(&alert)
	if err != nil {
		return nil, err
	}

	incident := models.Incident{
		OrgID:       orgID,
		AssetHWID:   alert.AssetHWID,
		Type:        alert.AlertType,
		Severity:    manualData.Severity,
		Priority:    alert.Priority,
		Status:      "OPEN",
		Description: manualData.Description,
	}

	if err := database.DB.Create(&incident).Error; err != nil {
		return nil, err
	}

	incidentID := incident.ID
	_, err = database.SecurityAlertCollection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"incident_id": &incidentID, "is_resolved": true}},
	)

	return &incident, err
}
