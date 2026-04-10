package incidents

import (
	"context"
	"errors"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type IncidentService struct{}

// GetIncidentList: Lấy danh sách sự cố từ PostgreSQL
func (s *IncidentService) GetIncidentList(orgID uint) ([]models.Incident, error) {
	var incidents []models.Incident
	// Preload Asset để hiển thị thông tin máy trạm trong danh sách
	err := database.DB.Preload("Asset").Where("org_id = ?", orgID).Order("created_at DESC").Find(&incidents).Error
	return incidents, err
}

// GetIncidentDetail: Lấy chi tiết sự cố và toàn bộ Audit Logs
func (s *IncidentService) GetIncidentDetail(ctx context.Context, incidentID uint) (*models.Incident, []models.IncidentAudit, error) {
	// 1. Lấy thông tin chính từ Postgres
	var incident models.Incident
	if err := database.DB.Preload("Asset").Preload("Assignee").First(&incident, incidentID).Error; err != nil {
		return nil, nil, err
	}

	// 2. Lấy toàn bộ Audit Logs từ MongoDB
	var audits []models.IncidentAudit
	// FIX: Phải cast incidentID sang int64 để khớp với dữ liệu trong Mongo
	filter := bson.M{"incident_id": int64(incidentID)}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	cursor, err := database.IncidentAuditCollection.Find(ctx, filter, opts)
	if err != nil {
		return &incident, nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &audits); err != nil {
		return &incident, nil, err
	}

	return &incident, audits, nil
}

// CloseIncidentWorkflow: Luồng đóng sự cố bắt buộc có Audit
func (s *IncidentService) CloseIncidentWorkflow(ctx context.Context, audit *models.IncidentAudit) error {
	// Kiểm tra bằng chứng Baseline Scan trước khi cho phép đóng
	if audit.EvidenceData == "" || audit.EvidenceData == "{}" {
		return errors.New("không thể đóng: Thiếu bằng chứng quét Baseline Scan để xác minh an toàn")
	}

	// 1. Niêm phong bản ghi Audit
	audit.CreatedAt = time.Now()
	audit.ActionType = "CLOSE"
	audit.NewStatus = "Resolved"
	audit.GenerateAuditHash()

	// 2. Lưu log vào MongoDB
	if _, err := database.IncidentAuditCollection.InsertOne(ctx, audit); err != nil {
		return err
	}

	// 3. Cập nhật trạng thái sự cố và tóm tắt xử lý vào Postgres
	return database.DB.Model(&models.Incident{}).Where("id = ?", audit.IncidentID).Updates(map[string]interface{}{
		"status":             "Resolved",
		"resolution_summary": audit.Content,
	}).Error
}

// internal/api/v1/incidents/incident_service.go

func (s *IncidentService) AssignIncident(ctx context.Context, incidentID, assigneeID, userID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var incident models.Incident
		if err := tx.First(&incident, incidentID).Error; err != nil {
			return err
		}

		oldStatus := incident.Status
		// Cập nhật người xử lý trong Postgres
		if err := tx.Model(&incident).Updates(map[string]interface{}{
			"assignee_id": assigneeID,
			"status":      "Investigating",
		}).Error; err != nil {
			return err
		}

		audit := models.IncidentAudit{
			IncidentID: int64(incidentID), // FIX: uint -> int64
			UserID:     nil,
			ActionType: "ASSIGN",
			Content:    fmt.Sprintf("Chỉ định nhân viên (ID: %d) xử lý sự cố.", assigneeID),
			OldStatus:  oldStatus,
			NewStatus:  "Investigating",
			CreatedAt:  time.Now(),
		}
		audit.GenerateAuditHash()
		_, err := database.IncidentAuditCollection.InsertOne(ctx, audit)
		return err
	})
}
