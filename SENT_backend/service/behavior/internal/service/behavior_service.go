package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BehaviorService struct{}

func (s *BehaviorService) LogBehavior(ctx context.Context, asset models.Asset, category, value, title, desc string) (*models.SecurityAlert, error) {
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

	priority := s.applyContextualMatrix(category, "P3", asset.DepartmentTag)

	alert := models.SecurityAlert{
		OrgID:       uint(asset.OrgID),
		AssetHWID:   asset.AssetHWID,
		AlertType:   category + " Violation",
		Title:       title,
		Description: desc + " | Lý do: " + result.Reason,
		Priority:    priority,
		Severity:    s.getSeverityByPriority(priority),
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

	return basePriority
}

func (s *BehaviorService) getSeverityByPriority(p string) string {
	mapping := map[string]string{"P1": "Critical", "P2": "High", "P3": "Medium", "P4": "Low"}
	return mapping[p]
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
