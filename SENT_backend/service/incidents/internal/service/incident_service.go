package incidents

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type IncidentService struct{}

// 1. TRUY VẤN DỮ LIỆU

// GetIncidentList: Lấy danh sách sự cố có phân trang
func (s *IncidentService) GetIncidentList(orgID uint, page, limit int) ([]models.Incident, int64, error) {
	var incidents []models.Incident
	var total int64

	// Đếm tổng số bản ghi trước để tính số trang[cite: 58]
	database.DB.Model(&models.Incident{}).Where("org_id = ?", orgID).Count(&total)

	offset := (page - 1) * limit

	err := database.DB.Preload("Asset").Preload("Assignee").
		Where("org_id = ?", orgID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&incidents).Error

	return incidents, total, err
}

// GetIncidentDetail: Lấy chi tiết sự cố và toàn bộ Audit Logs[cite: 58]
func (s *IncidentService) GetIncidentDetail(ctx context.Context, incidentID uint, orgID uint) (*models.Incident, []models.IncidentAudit, error) {
	var incident models.Incident
	// Chống IDOR bằng cách lọc theo org_id[cite: 58]
	err := database.DB.Preload("Asset").Preload("Assignee").
		Where("id = ? AND org_id = ?", incidentID, orgID).First(&incident).Error
	if err != nil {
		return nil, nil, err
	}

	// 2. [QUAN TRỌNG] Lấy bằng chứng USB/Hành vi từ MongoDB để hiện ở cột trái
	alerts := []models.SecurityAlert{}
	alertFilter := bson.M{"incident_id": incidentID}
	if database.SecurityAlertCollection != nil {
		alertCursor, _ := database.SecurityAlertCollection.Find(ctx, alertFilter)
		if alertCursor != nil {
			alertCursor.All(ctx, &alerts)
			incident.Alerts = alerts // Gán vào struct Incident để trả về cho Frontend
		}
	}

	// Luôn khởi tạo mảng rỗng để tránh trả về 'null' cho Frontend gây crash[cite: 58]
	audits := []models.IncidentAudit{}
	filter := bson.M{"incident_id": incidentID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(500)

	cursor, err := database.IncidentAuditCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &audits); err != nil {
		return &incident, audits, err
	}

	return &incident, audits, nil
}

// 2. LOGIC NGHIỆP VỤ CHÍNH (VÒNG ĐỜI SỰ CỐ)

// AssignIncident: Admin phân công và ghi log[cite: 58]
func (s *IncidentService) AssignIncident(ctx context.Context, incidentID, assigneeID, userID uint, orgID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var incident models.Incident
		if err := tx.Where("id = ? AND org_id = ?", incidentID, orgID).First(&incident).Error; err != nil {
			return err
		}

		oldStatus := incident.Status
		if err := tx.Model(&incident).Updates(map[string]interface{}{
			"assignee_id": assigneeID,
			"status":      "Investigating",
		}).Error; err != nil {
			return err
		}

		audit := models.IncidentAudit{
			IncidentID: uint(incidentID),
			UserID:     &userID,
			ActionType: "ASSIGN",
			Content:    fmt.Sprintf("Chỉ định nhân viên xử lý sự cố."),
			OldStatus:  oldStatus,
			NewStatus:  "Investigating",
			CreatedAt:  time.Now(),
		}
		return s.createAndChainAudit(ctx, tx, &audit)
	})
}

// CloseIncidentWorkflow: Kiểm tra điều kiện và đóng sự cố[cite: 58]
func (s *IncidentService) CloseIncidentWorkflow(ctx context.Context, audit *models.IncidentAudit, orgID uint) error {

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var incident models.Incident
		if err := tx.Where("id = ? AND org_id = ?", audit.IncidentID, orgID).First(&incident).Error; err != nil {
			return err
		}

		audit.CreatedAt = time.Now()
		audit.ActionType = "CLOSE"
		audit.NewStatus = "Resolved"

		if err := s.createAndChainAudit(ctx, tx, audit); err != nil {
			return err
		}

		return tx.Model(&incident).Updates(map[string]interface{}{
			"status":             "Resolved",
			"resolution_summary": audit.Content,
		}).Error
	})
}

// EscalateToIncident: Nâng cấp Alert (MongoDB) lên Incident (Postgres)[cite: 58]
func (s *IncidentService) EscalateToIncident(ctx context.Context, alertIDStr string, adminID uint, orgID uint) (*models.Incident, error) {
	alertID, err := primitive.ObjectIDFromHex(alertIDStr)
	if err != nil {
		return nil, errors.New("ID hành vi không hợp lệ")
	}

	var alert models.SecurityAlert
	err = database.SecurityAlertCollection.FindOne(ctx, bson.M{"_id": alertID, "org_id": orgID}).Decode(&alert)
	if err != nil {
		return nil, errors.New("không tìm thấy hành vi vi phạm")
	}

	if alert.IsResolved {
		return nil, errors.New("hành vi này đã được xử lý")
	}

	incident := models.Incident{
		OrgID:       uint(alert.OrgID),
		AssetHWID:   alert.AssetHWID,
		Type:        alert.AlertType,
		Priority:    alert.Priority,
		Severity:    alert.Severity,
		Status:      "Open",
		Description: fmt.Sprintf("Nâng cấp từ hành vi: %s\nChi tiết: %s", alert.Title, alert.Description),
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&incident).Error; err != nil {
			return err
		}

		// Đánh dấu alert đã xử lý trong Mongo
		update := bson.M{"$set": bson.M{"incident_id": incident.ID, "is_resolved": true}}
		if _, err := database.SecurityAlertCollection.UpdateOne(ctx, bson.M{"_id": alertID}, update); err != nil {
			return err
		}

		audit := models.IncidentAudit{
			IncidentID: uint(incident.ID),
			UserID:     &adminID,
			ActionType: "ESCALATE",
			Content:    fmt.Sprintf("Hệ thống nâng cấp hành vi thành sự cố chính thức."),
			NewStatus:  "Open",
			CreatedAt:  time.Now(),
		}
		return s.createAndChainAudit(ctx, tx, &audit)
	})

	return &incident, err
}

// 3. HỆ THỐNG CHAINING HASH (BLOCKCHAIN CORE)

// CreateChainedAudit: Public API để tạo Log xâu chuỗi (Dùng cho AuditChat)
func (s *IncidentService) CreateChainedAudit(ctx context.Context, audit *models.IncidentAudit, orgID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		return s.createAndChainAudit(ctx, tx, audit)
	})
}

// createAndChainAudit: (Private) Logic lõi để nối chuỗi Hash[cite: 58]
func (s *IncidentService) createAndChainAudit(ctx context.Context, tx *gorm.DB, audit *models.IncidentAudit) error {
	// Lấy mã băm của bản ghi cuối cùng để nối chuỗi[cite: 58]
	previousHash, err := s.getLastAuditHash(ctx, audit.IncidentID)
	if err != nil {
		return err
	}
	audit.PreviousHash = previousHash

	// Niêm phong bản ghi hiện tại[cite: 57, 58]
	audit.GenerateAuditHash()

	// Lưu vào MongoDB
	if _, err := database.IncidentAuditCollection.InsertOne(ctx, audit); err != nil {
		return err
	}
	fmt.Println("Audit saved with ID:", audit.IncidentID)

	// Cập nhật Golden Hash vào Postgres để đối soát[cite: 58]
	return tx.Model(&models.Incident{}).Where("id = ?", audit.IncidentID).
		Update("last_audit_hash", audit.AuditHash).Error
}

// getLastAuditHash: Tìm hash của bản ghi gần nhất[cite: 58]
func (s *IncidentService) getLastAuditHash(ctx context.Context, incidentID uint) (string, error) {
	var lastAudit models.IncidentAudit
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	filter := bson.M{"incident_id": incidentID}

	err := database.IncidentAuditCollection.FindOne(ctx, filter, opts).Decode(&lastAudit)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", nil // Bản ghi đầu tiên
		}
		return "", err
	}
	return lastAudit.AuditHash, nil
}

// VerifyIncidentAuditChain: Xác minh toàn bộ chuỗi log[cite: 58]
func (s *IncidentService) VerifyIncidentAuditChain(ctx context.Context, incidentID uint, orgID uint) (bool, string, error) {
	var incident models.Incident
	if err := database.DB.Where("id = ? AND org_id = ?", incidentID, orgID).First(&incident).Error; err != nil {
		return false, "Sự cố không tồn tại", err
	}

	filter := bson.M{"incident_id": incidentID}
	cursor, err := database.IncidentAuditCollection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}))
	if err != nil {
		return false, "Lỗi truy vấn Mongo", err
	}
	defer cursor.Close(ctx)

	var audits []models.IncidentAudit
	if err = cursor.All(ctx, &audits); err != nil {
		return false, "Lỗi đọc dữ liệu", err
	}

	currentPrevHash := ""
	for i, record := range audits {
		// 1. Kiểm tra tính liên kết[cite: 58]
		if record.PreviousHash != currentPrevHash {
			return false, fmt.Sprintf("Đứt gãy tại bản ghi #%d", i+1), nil
		}

		// 2. Tính toán lại hash để phát hiện sửa đổi nội dung[cite: 58]
		storedHash := record.AuditHash
		record.AuditHash = ""
		record.GenerateAuditHash()
		if storedHash != record.AuditHash {
			return false, fmt.Sprintf("Nội dung bản ghi #%d bị thay đổi", i+1), nil
		}
		currentPrevHash = storedHash
	}

	// 3. Đối soát với Golden Hash trong SQL[cite: 58]
	if len(audits) > 0 && audits[len(audits)-1].AuditHash != incident.LastAuditHash {
		return false, "Golden Hash không khớp", nil
	}

	return true, "Dữ liệu nguyên bản", nil
}

// 4. UTILS

func (s *IncidentService) isNoteComplexEnough(note string) bool {
	cleanNote := strings.TrimSpace(note)
	return len(cleanNote) >= 15 && len(strings.Fields(cleanNote)) >= 3
}
