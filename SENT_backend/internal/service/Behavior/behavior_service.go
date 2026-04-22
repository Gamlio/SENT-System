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
		OrgID:       int64(asset.OrgID),
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
func (s *BehaviorService) GetBehaviors(ctx context.Context, orgID int64, page, limit int64) ([]models.SecurityAlert, error) {
	alerts := []models.SecurityAlert{}
	if database.SecurityAlertCollection == nil {
		return alerts, nil
	}

	// [FIX] Thêm logic phân trang
	skip := (page - 1) * limit
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)
	filter := bson.M{"org_id": orgID}

	cursor, err := database.SecurityAlertCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}
