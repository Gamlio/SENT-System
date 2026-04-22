package incidents

import (
	"context"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/scoring"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

const (
	DefaultCorrelationWindow = 24 * time.Hour // Cửa sổ thời gian gom nhóm sự cố
)

func (s *IncidentService) applyContextualMatrix(alertType, basePriority, departmentTag string) string {
	tag := strings.ToUpper(departmentTag)
	switch alertType {
	case "USB Violation":
		if tag == "DEV" {
			return "P4"
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
		OrgID:       int64(asset.OrgID), // Ép kiểu cho MongoDB
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
		// Dùng ctx truyền vào thay vì context.TODO()
		res, err := database.SecurityAlertCollection.InsertOne(ctx, alert)
		if err == nil {
			if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
				alert.ID = oid
			}
		}
	}

	// [CHUYỂN DỊCH BEHAVIOR-CENTRIC]
	// Đã TẮT tính năng tự động tạo Incident trong PostgreSQL.
	// Giờ đây mọi vi phạm chỉ được lưu dưới dạng Alert (Hành vi) trong MongoDB.
	// Việc tạo Sự cố (Incident) sẽ được thực hiện thông qua API Escalate (Nâng cấp) do Admin/AI quyết định.

	// 4. Gọi Engine tính điểm v6.0[cite: 9, 10]
	scoring.RecalculateRiskScore(asset.AssetHWID)
}

// AutoResolveIncident: Tự động đóng Case nếu asset báo cáo trạng thái đã an toàn
func (s *IncidentService) AutoResolveIncident(hwid string, alertType string) {
	var incident models.Incident
	err := database.DB.Where("asset_hwid = ? AND type = ? AND status != ?", hwid, alertType, "Resolved").First(&incident).Error

	if err == nil {
		database.DB.Transaction(func(tx *gorm.DB) error {
			tx.Model(&incident).Updates(map[string]interface{}{
				"status":      "Resolved",
				"description": incident.Description + " [Auto-Resolved]",
			})

			// [SỬA LỖI] Ghi log vào MongoDB, không dùng tx.Create vì IncidentAudit là NoSQL
			audit := models.IncidentAudit{
				IncidentID:   int64(incident.ID),
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

// Helpers nội bộ
func (s *IncidentService) getSeverityByPriority(p string) string {
	mapping := map[string]string{"P1": "Critical", "P2": "High", "P3": "Medium", "P4": "Low"}
	if val, ok := mapping[p]; ok {
		return val
	}
	return "Low"
}

func (s *IncidentService) shouldUpgradePriority(current, new string) bool {
	levels := map[string]int{"P1": 4, "P2": 3, "P3": 2, "P4": 1}
	return levels[new] > levels[current]
}
