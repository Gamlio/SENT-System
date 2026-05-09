package behavior

import (
	"context"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/policies"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BehaviorService struct{}

// LogBehavior: Tiếp nhận mọi hành vi, kiểm tra vi phạm và lưu vào MongoDB
func (s *BehaviorService) LogBehavior(ctx context.Context, asset models.Asset, category, value, title, desc string) (*models.SecurityAlert, error) {
	// 1. Kiểm tra chính sách (Dịch chuyển từ Policy sang Behavior)
	policySvc := &policies.PolicyService{}
	isViolation, reason, _ := policySvc.CheckPolicyViolation(asset.OrgID, asset.AssetHWID, category, value)

	if !isViolation {
		return nil, nil // Không có vi phạm, không cần tạo Alert
	}

	// 2. Phân tích mức độ ưu tiên dựa trên ma trận ngữ cảnh (Contextual Matrix)
	// Logic này được chuyển từ event_engine.go sang đây
	priority := s.applyContextualMatrix(category, "P3", asset.DepartmentTag)

	// 3. Tạo bản ghi Hành vi (SecurityAlert) trong MongoDB
	alert := models.SecurityAlert{
		OrgID:       uint(asset.OrgID),
		AssetHWID:   asset.AssetHWID,
		AlertType:   category + " Violation",
		Title:       title,
		Description: desc + " | Lý do: " + reason,
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

// applyContextualMatrix: Phân tích sâu mức độ nguy hiểm dựa trên phòng ban
func (s *BehaviorService) applyContextualMatrix(category, basePriority, tag string) string {
	// Giữ nguyên logic switch-case từ event_engine.go của bạn
	return basePriority
}

func (s *BehaviorService) getSeverityByPriority(p string) string {
	mapping := map[string]string{"P1": "Critical", "P2": "High", "P3": "Medium", "P4": "Low"}
	return mapping[p]
}

// GetBehaviors: Truy vấn danh sách hành vi từ MongoDB theo OrgID
func (s *BehaviorService) GetBehaviors(ctx context.Context, orgID uint, page, limit int64) ([]models.SecurityAlert, int64, error) {
	alerts := []models.SecurityAlert{}
	if database.SecurityAlertCollection == nil {
		return alerts, 0, nil
	}

	filter := bson.M{"org_id": orgID} // [BẢO MẬT] Luôn lọc theo org_id để tránh IDOR

	// [HIỆU SUẤT] Đếm tổng số bản ghi trực tiếp từ MongoDB
	total, err := database.SecurityAlertCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// [PHÂN TRANG] Chỉ lấy đúng số lượng cần thiết
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
	// Kiểm tra quyền sở hữu bản ghi
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

	// Gắn ID sự cố vào log NoSQL
	incidentID := incident.ID
	_, err = database.SecurityAlertCollection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"incident_id": &incidentID, "is_resolved": true}},
	)

	return &incident, err
}
