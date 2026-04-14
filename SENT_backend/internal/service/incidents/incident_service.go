package incidents

import (
	"context"
	"errors"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

// isNoteComplexEnough kiểm tra xem ghi chú đóng case có đủ chi tiết không.
func isNoteComplexEnough(note string) bool {
	// Tối thiểu 15 ký tự
	if len(strings.TrimSpace(note)) < 15 {
		return false
	}
	// Tối thiểu 3 từ
	if len(strings.Fields(note)) < 3 {
		return false
	}
	// Kiểm tra các ký tự lặp lại (ví dụ: "1111111111", "aaaaaaaaaa")
	uniqueChars := make(map[rune]bool)
	for _, r := range note {
		if r != ' ' { // Bỏ qua khoảng trắng
			uniqueChars[r] = true
		}
	}
	// Nếu ghi chú dài hơn 10 ký tự nhưng có ít hơn 4 ký tự duy nhất (không phải khoảng trắng), nó có thể là rác.
	if len(note) > 10 && len(uniqueChars) < 4 {
		return false
	}
	return true
}

// CloseIncidentWorkflow: Luồng đóng sự cố bắt buộc có Audit
func (s *IncidentService) CloseIncidentWorkflow(ctx context.Context, audit *models.IncidentAudit) error {
	// Kiểm tra bằng chứng Baseline Scan trước khi cho phép đóng
	if audit.EvidenceData == "" || audit.EvidenceData == "{}" {
		return errors.New("không thể đóng: Thiếu bằng chứng quét Baseline Scan để xác minh an toàn")
	}

	// [MỚI] Kiểm tra độ phức tạp của ghi chú đóng case
	if !isNoteComplexEnough(audit.Content) {
		return errors.New("ghi chú đóng sự cố không đủ chi tiết hoặc không hợp lệ. Vui lòng cung cấp mô tả rõ ràng (tối thiểu 15 ký tự và 3 từ).")
	}

	// [MỚI] Kiểm tra bằng chứng đính kèm cho sự cố P1
	var incident models.Incident
	if err := database.DB.Select("priority").First(&incident, audit.IncidentID).Error; err != nil {
		return errors.New("không tìm thấy sự cố tương ứng để kiểm tra mức độ ưu tiên")
	}

	if incident.Priority == "P1" && audit.Images == "" {
		return errors.New("không thể đóng: Sự cố P1 bắt buộc phải có tệp bằng chứng (ảnh chụp màn hình) đính kèm")
	}

	// Sử dụng transaction để đảm bảo tính toàn vẹn giữa SQL và NoSQL
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Niêm phong bản ghi Audit
		audit.CreatedAt = time.Now()
		audit.ActionType = "CLOSE"
		audit.NewStatus = "Resolved"

		// 2. Tạo chuỗi hash, lưu log vào MongoDB, và cập nhật Golden Hash trong Postgres
		if err := s.createAndChainAudit(ctx, tx, audit); err != nil {
			return err
		}

		// 3. Cập nhật trạng thái sự cố và tóm tắt xử lý vào Postgres
		return tx.Model(&models.Incident{}).Where("id = ?", audit.IncidentID).Updates(map[string]interface{}{
			"status":             "Resolved",
			"resolution_summary": audit.Content,
		}).Error
	})
}

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
			IncidentID: int64(incidentID),
			UserID:     nil,
			ActionType: "ASSIGN",
			Content:    fmt.Sprintf("Chỉ định nhân viên (ID: %d) xử lý sự cố.", assigneeID),
			OldStatus:  oldStatus,
			NewStatus:  "Investigating",
			CreatedAt:  time.Now(),
		}
		// Tạo chuỗi hash và cập nhật Golden Hash
		return s.createAndChainAudit(ctx, tx, &audit)
	})
}

// getLastAuditHash retrieves the hash of the last audit record for a given incident.
func (s *IncidentService) getLastAuditHash(ctx context.Context, incidentID int64) (string, error) {
	var lastAudit models.IncidentAudit
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	filter := bson.M{"incident_id": incidentID}

	err := database.IncidentAuditCollection.FindOne(ctx, filter, opts).Decode(&lastAudit)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Đây là bản ghi đầu tiên, không có hash trước đó.
			return "", nil
		}
		return "", err
	}
	return lastAudit.AuditHash, nil
}

// createAndChainAudit tạo một bản ghi audit mới, liên kết nó với bản ghi trước đó,
// và cập nhật "golden hash" trong bản ghi sự cố chính.
func (s *IncidentService) createAndChainAudit(ctx context.Context, tx *gorm.DB, audit *models.IncidentAudit) error {
	// 1. Lấy mã băm của bản ghi audit cuối cùng.
	previousHash, err := s.getLastAuditHash(ctx, audit.IncidentID)
	if err != nil {
		return fmt.Errorf("không thể lấy mã băm trước đó: %w", err)
	}
	audit.AuditHash = previousHash // Giả định models.IncidentAudit có trường `PreviousHash`

	// 2. Tạo mã băm mới cho bản ghi hiện tại.
	// QUAN TRỌNG: Phương thức GenerateAuditHash() trong model phải được sửa để bao gồm cả `PreviousHash` vào chuỗi dữ liệu cần băm.
	audit.GenerateAuditHash()

	// 3. Chèn bản ghi audit mới vào MongoDB.
	_, err = database.IncidentAuditCollection.InsertOne(ctx, audit)
	if err != nil {
		return fmt.Errorf("không thể chèn bản ghi audit mới: %w", err)
	}

	// 4. Cập nhật "Golden Hash" trong bảng sự cố PostgreSQL.
	// Giả định models.Incident có trường `LastAuditHash`
	if err := tx.Model(&models.Incident{}).Where("id = ?", audit.IncidentID).Update("last_audit_hash", audit.AuditHash).Error; err != nil {
		return fmt.Errorf("không thể cập nhật golden hash: %w", err)
	}

	return nil
}

// VerifyIncidentAuditChain kiểm tra tính toàn vẹn của toàn bộ chuỗi audit log cho một sự cố.
func (s *IncidentService) VerifyIncidentAuditChain(ctx context.Context, incidentID uint) (bool, string, error) {
	// 1. Lấy sự cố từ Postgres để truy xuất "Golden Hash".
	var incident models.Incident
	if err := database.DB.First(&incident, incidentID).Error; err != nil {
		return false, "Sự cố không tồn tại trong SQL", err
	}
	goldenHash := incident.LastAuditHash // Giả định models.Incident có trường `LastAuditHash`

	// 2. Lấy tất cả các audit log cho sự cố từ MongoDB, sắp xếp theo thứ tự thời gian.
	filter := bson.M{"incident_id": int64(incidentID)}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	cursor, err := database.IncidentAuditCollection.Find(ctx, filter, opts)
	if err != nil {
		return false, "Lỗi truy vấn audit logs từ MongoDB", err
	}
	defer cursor.Close(ctx)

	var audits []models.IncidentAudit
	if err = cursor.All(ctx, &audits); err != nil {
		return false, "Lỗi đọc danh sách audit logs", err
	}

	if len(audits) == 0 {
		if goldenHash == "" {
			return true, "Không có audit log, chuỗi hợp lệ.", nil
		}
		return false, "Golden hash tồn tại nhưng không có audit log.", nil
	}

	// 3. Lặp qua chuỗi và xác minh từng liên kết.
	var previousHash string = "" // Hash của "genesis block" là rỗng.
	for i, auditRecord := range audits {
		// Kiểm tra xem khối hiện tại có liên kết với khối trước đó không.
		if auditRecord.AuditHash != previousHash {
			return false, fmt.Sprintf("Chuỗi bị phá vỡ tại bản ghi #%d. AuditHash không khớp.", i+1), nil
		}

		// Tính toán lại hash của khối hiện tại để xác minh nó không bị giả mạo.
		// QUAN TRỌNG: Giả định GenerateAuditHash() tính toán hash dựa trên nội dung và PreviousHash.
		storedHash := auditRecord.AuditHash
		auditRecord.AuditHash = "" // Tạm thời xóa hash để tính toán lại
		auditRecord.GenerateAuditHash()
		calculatedHash := auditRecord.AuditHash
		auditRecord.AuditHash = storedHash // Khôi phục lại

		if storedHash != calculatedHash {
			return false, fmt.Sprintf("Nội dung bản ghi #%d đã bị thay đổi. Hash không khớp.", i+1), nil
		}

		// Hash hiện tại trở thành hash trước đó cho vòng lặp tiếp theo.
		previousHash = calculatedHash
	}

	// 4. Cuối cùng, kiểm tra xem hash của khối cuối cùng có khớp với golden hash không.
	if previousHash != goldenHash {
		return false, "Mã băm cuối cùng trong chuỗi không khớp với golden hash.", nil
	}

	return true, "Toàn vẹn chuỗi audit log được xác thực.", nil
}
