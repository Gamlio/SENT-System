package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"

	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

// Priority scores mapping
var priorityScores = map[string]float64{
	"P1": 50.0,
	"P2": 25.0,
	"P3": 10.0,
	"P4": 5.0,
}

// Department Matrix mapping SensorType and DepartmentTag to Priority
// This can be loaded from a config or database for more flexibility
var departmentMatrix = map[string]map[string]string{
	"USB_PLUG": {
		"DEV":     "P4",
		"FINANCE": "P1",
		"PROD":    "P2",
	},
	"PROC_START": {
		"DEV":     "P3",
		"FINANCE": "P1",
		"PROD":    "P1",
	},
	"PORT_OPEN": {
		"DEV":     "P2",
		"FINANCE": "P1",
		"PROD":    "P1",
	},
	// Add more sensor types and departments as needed
}

const (
	kFactor                  float64 = 50.0 // Default sensitivity factor, can be configurable (40-60)
	exponentialBase          float64 = 2.0  // Base for exponential escalation (E in E^n)
	trustScoreP1Deduction    float64 = 15.0
	trustScoreP2Deduction    float64 = 5.0
	trustScoreRecoveryAmount float64 = 2.0
	trustScoreRecoveryDays   int     = 7
	trustScoreMax            float64 = 100.0
	trustScoreMin            float64 = 0.0
	incidentLookbackDays     int     = 30 // For TrustScore deduction (30 days for P1/P2)
)

// ScoreService handles the calculation of risk and trust scores
type ScoreService struct {
	db *gorm.DB
}

// Áp dụng cơ chế Debounce đồng thời cho nhiều Assets, tránh Spam truy vấn.
var scoreTimers sync.Map

func RecalculateRiskScore(assetAssetID string) {
	if timer, ok := scoreTimers.Load(assetAssetID); ok {
		// Nếu đang trong hàng đợi tính điểm thì chỉ cần reset lại Timer
		timer.(*time.Timer).Reset(2 * time.Second)
		return
	}

	// Debounce: Chờ thêm 2 giây để đón tất cả log đồng thời (ví dụ cắm nhiều USB).
	// Gom lại xử lý 1 lần.
	timer := time.AfterFunc(2*time.Second, func() {
		scoreTimers.Delete(assetAssetID)
		processRiskScore(assetAssetID)
	})
	scoreTimers.Store(assetAssetID, timer)
}

func processRiskScore(assetAssetID string) {
	var asset models.Asset
	// Tối ưu DB: Select chỉ lấy trường cần thiết, Preload AssetType để tránh lỗi
	if err := database.DB.Select("id", "asset_hwid", "hostname", "risk_score", "trust_score", "asset_type_id").Preload("AssetType").Where("asset_hwid = ?", assetAssetID).First(&asset).Error; err != nil {
		return
	}

	// 1. IMPACT TỨC THỜI (R_active) - Giảm trọng số
	var activeAlerts []models.SecurityAlert
	if database.SecurityAlertCollection != nil {
		cursor, err := database.SecurityAlertCollection.Find(context.TODO(), bson.M{
			"asset_hwid": assetAssetID, "is_resolved": false,
		})
		if err == nil {
			cursor.All(context.TODO(), &activeAlerts)
		}
	}

	ti := 0.0
	hasP1 := false
	for _, alert := range activeAlerts {
		switch alert.Priority {
		case "P1":
			ti += 5.0
			hasP1 = true
		case "P2":
			ti += 2.5
		case "P4":
			ti += 0.5
		default:
			ti += 1.0
		}
	}

	var ioActivities []models.AssetIOActivity
	if database.AssetIOActivityCollection != nil {
		cursor, err := database.AssetIOActivityCollection.Find(context.TODO(), bson.M{"asset_hwid": assetAssetID})
		if err == nil {
			cursor.All(context.TODO(), &ioActivities)
		}
	}

	var openPorts []models.OpenPort
	if database.OpenPortCollection != nil {
		cursor, err := database.OpenPortCollection.Find(context.TODO(), bson.M{"asset_hwid": assetAssetID, "status": "OPEN"})
		if err == nil {
			cursor.All(context.TODO(), &openPorts)
		}
	}

	var usbLogs []models.USBLog
	if database.USBCollection != nil {
		cursor, err := database.USBCollection.Find(context.TODO(), bson.M{"asset_hwid": assetAssetID, "event_type": "CONNECTED"})
		if err == nil {
			cursor.All(context.TODO(), &usbLogs)
		}
	}

	n := float64(len(activeAlerts))

	if n == 0 && len(openPorts) == 0 && len(usbLogs) == 0 && len(ioActivities) == 0 {
		database.DB.Model(&asset).Updates(map[string]interface{}{
			"risk_score":     0.0,
			"security_grade": "A",
		})
		return
	}

	// Bổ sung telemetry khác vào ti để hệ thống vẫn phản ứng với hành vi bất thường
	if len(usbLogs) > 0 {
		ti += float64(len(usbLogs)) * 0.5
	}
	if len(ioActivities) > 0 {
		ti += float64(len(ioActivities)) * 0.2
	}

	// Sử dụng Log để n lỗi không bằng 1 lỗi nặng
	multiplier := n + 1
	if n == 0 && ti > 0 {
		multiplier = 2 // Đảm bảo math.Log2(2) = 1 để giữ nguyên điểm ti từ telemetry
	}
	rActive := ti * math.Log2(multiplier)

	// 2. IMPACT LỊCH SỬ (R_history) - Giữ vết lâu (Lookback 180 ngày)
	var pastIncidents []models.Incident
	halfYearAgo := time.Now().AddDate(0, 0, -180) // Quét lịch sử 180 ngày
	rHistory := 0.0

	if err := database.DB.Where("asset_hwid = ? AND created_at >= ?", assetAssetID, halfYearAgo).Find(&pastIncidents).Error; err == nil {
		for _, inc := range pastIncidents {
			daysOld := time.Since(inc.CreatedAt).Hours() / 24.0
			hImpact := 1.0
			switch inc.Priority {
			case "P1":
				hImpact = 10.0
			case "P2":
				hImpact = 5.0
			case "P4":
				hImpact = 0.5
			}

			// Decay cực chậm: Sau 3 tháng (90 ngày) lỗi P1 vẫn còn giữ khoảng 5 điểm rủi ro
			rHistory += hImpact / (1.0 + 0.01*daysOld)
		}
	}

	// 3. TRỌNG SỐ NGỮ CẢNH (C) VÀ HỆ SỐ PHƠI NHIỄM (V)
	cFactor := 1.0
	// Xử lý nil pointer để tránh Panic khi chạy
	if asset.AssetType != nil {
		switch asset.AssetType.Name {
		case "SERVER":
			cFactor = 2.0 // Rất cao
		case "IT_ADMIN":
			cFactor = 1.5 // Cao
		default:
			cFactor = 1.0 // Trung bình
		}
	}

	vFactor := 1.0 + (float64(len(openPorts)) * 0.05) // Mỗi port mở tăng 5%

	// 4. TỔNG HỢP & PHÂN HẠNG (0-100 scale cho UI)
	displayScore := (rActive + rHistory) * cFactor * vFactor
	if displayScore > 100.0 {
		displayScore = 100.0
	}

	grade := "A"
	switch {
	case displayScore > 60:
		grade = "F"
	case displayScore > 35:
		grade = "D"
	case displayScore > 15:
		grade = "C"
	case displayScore > 5:
		grade = "B"
	default:
		grade = "A"
	}

	// 5. CƠ CHẾ FORENSIC INTEGRITY (Niêm phong bằng chứng)
	if displayScore > 80.0 {
		log.Printf("🚨 [FORENSIC SEAL] %s: Điểm rủi ro=%.1f > 80. Hệ thống kích hoạt Snapshot Immutable!", asset.Hostname, displayScore)
		// TODO: Tích hợp gọi EventEngine đẩy lệnh thu thập Process/Port list xuống Go-SENT tại đây.
	}

	// 6. CẬP NHẬT TRẠNG THÁI
	now := time.Now()
	updates := map[string]interface{}{
		"risk_score":       displayScore,
		"security_grade":   grade,
		"last_incident_at": &now,
	}

	if hasP1 {
		newTrust := asset.TrustScore - trustScoreP1Deduction
		if newTrust < 0 {
			newTrust = 0
		}
		updates["trust_score"] = newTrust
	}

	database.DB.Model(&asset).Updates(updates)
	log.Printf("🧬 [EQRI v2] %s: R_active=%.1f | R_hist=%.1f | Grade=%s | Điểm rủi ro: %.1f/100",
		asset.Hostname, rActive, rHistory, grade, displayScore)
}

// NewScoreService creates a new ScoreService instance
func NewScoreService(db *gorm.DB) *ScoreService {
	return &ScoreService{db: db}
}

// GetPriorityFromMatrix determines the priority based on sensor type and department tag
// This function is typically used by the Event Engine to assign priority to incidents/alerts.
func (s *ScoreService) GetPriorityFromMatrix(sensorType, departmentTag string) (string, error) {
	if deptMap, ok := departmentMatrix[sensorType]; ok {
		if priority, ok := deptMap[departmentTag]; ok {
			return priority, nil
		}
		// If department tag not found for a known sensor type, return a default or error
		return "P4", fmt.Errorf("department tag '%s' not found for sensor type '%s' in matrix, defaulting to P4", departmentTag, sensorType)
	}
	// If sensor type not found, return a default or error
	return "P4", fmt.Errorf("sensor type '%s' not found in department matrix, defaulting to P4", sensorType)
}

// CalculateCurrentRiskScore calculates the instantaneous risk score (R_current) for an asset.
// This score reflects the current active threats.
func (s *ScoreService) CalculateCurrentRiskScore(assetID uint) (float64, error) {
	var incidents []models.Incident
	// Fetch all OPEN incidents for the asset
	if err := s.db.Where("asset_hwid = ? AND status = ?", assetID, "Open").Find(&incidents).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch open incidents for asset %d: %w", assetID, err)
	}

	if len(incidents) == 0 {
		// No open incidents, current risk score is 0
		// Ensure asset's risk_score is updated to 0 if it was previously higher
		if err := s.db.Model(&models.Asset{}).Where("asset_hwid = ?", assetID).Update("risk_score", 0.0).Error; err != nil {
			return 0, fmt.Errorf("failed to reset risk score for asset %d: %w", assetID, err)
		}
		return 0, nil
	}

	sumSi := 0.0
	numOpenIncidents := float64(len(incidents))

	// Calculate the exponential escalation factor E^n, where n is the total number of open incidents.
	// The formula is R_current = 100 * (1 - e^(-(sum(Si * E^n) / k)))
	// This is interpreted as (sum(Si)) * E^n
	escalationFactor := math.Pow(exponentialBase, numOpenIncidents)

	for _, inc := range incidents {
		if score, ok := priorityScores[inc.Priority]; ok {
			sumSi += score
		} else {
			// Log or handle unknown priority, perhaps default to a low score
			fmt.Printf("Warning: Unknown priority '%s' for incident %d, defaulting to P4 score\n", inc.Priority, inc.ID)
			sumSi += priorityScores["P4"] // Default to P4 score for unknown priorities
		}
	}

	// Apply the escalation factor to the sum of Si
	sumSiEscalated := sumSi * escalationFactor

	// Calculate R_current using the asymptotic model
	rCurrent := 100.0 * (1 - math.Exp(-(sumSiEscalated / kFactor)))

	// Ensure R_current does not exceed 100
	if rCurrent > 100.0 {
		rCurrent = 100.0
	}

	// Update the asset's RiskScore in the database
	if err := s.db.Model(&models.Asset{}).Where("id = ?", assetID).Update("risk_score", rCurrent).Error; err != nil {
		return rCurrent, fmt.Errorf("failed to update risk score for asset %d: %w", assetID, err)
	}

	return rCurrent, nil
}

// UpdateTrustScore updates the long-term trust score (D_debt) for an asset.
// This function should be called periodically (e.g., daily via a cron job)
// and also when an incident is created or resolved to ensure immediate reflection of changes.
func (s *ScoreService) UpdateTrustScore(assetID uint) (float64, error) {
	var asset models.Asset
	if err := s.db.First(&asset, assetID).Error; err != nil {
		return 0, fmt.Errorf("asset not found: %w", err)
	}

	currentTrustScore := asset.TrustScore
	if currentTrustScore < trustScoreMin {
		currentTrustScore = trustScoreMin // Cap at min before deductions
	}

	// --- Deduct points for P1 and P2 incidents in the last 30 days ---
	var recentIncidents []models.Incident
	thirtyDaysAgo := time.Now().AddDate(0, 0, -incidentLookbackDays)

	// Fetch incidents that occurred within the last 30 days and are P1 or P2.
	// The document implies "lỗi P1 trong 30 ngày qua" (P1 errors in the past 30 days)
	// should cause deduction, regardless of their current status (Open/Resolved).
	if err := s.db.Where("asset_hwid = ? AND occurred_at >= ? AND (priority = ? OR priority = ?)",
		assetID, thirtyDaysAgo, "P1", "P2").Find(&recentIncidents).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch recent incidents for trust score deduction for asset %d: %w", assetID, err)
	}

	// Calculate total deduction from recent incidents
	totalDeduction := 0.0
	for _, inc := range recentIncidents {
		if inc.Priority == "P1" {
			totalDeduction += trustScoreP1Deduction
		} else if inc.Priority == "P2" {
			totalDeduction += trustScoreP2Deduction
		}
	}
	currentTrustScore -= totalDeduction

	// --- Trust Score Recovery: "Sau mỗi 7 ngày "sạch" (không có lỗi mới), máy được cộng lại 2đ uy tín." ---
	isCleanForRecoveryPeriod := false
	if asset.LastIncidentAt == nil {
		isCleanForRecoveryPeriod = true // asset has never had an incident
	} else if time.Since(*asset.LastIncidentAt).Hours() >= float64(trustScoreRecoveryDays*24) {
		isCleanForRecoveryPeriod = true // Last incident was more than 'trustScoreRecoveryDays' ago
	}

	shouldApplyRecovery := false
	if isCleanForRecoveryPeriod {
		if asset.LastTrustRecoveryAppliedAt == nil {
			shouldApplyRecovery = true // Never applied recovery, so apply it
		} else if time.Since(*asset.LastTrustRecoveryAppliedAt).Hours() >= float64(trustScoreRecoveryDays*24) {
			shouldApplyRecovery = true // Last recovery was applied more than 'trustScoreRecoveryDays' ago, apply again
		}
	}

	if shouldApplyRecovery {
		currentTrustScore += trustScoreRecoveryAmount
		now := time.Now()
		asset.LastTrustRecoveryAppliedAt = &now // Update the timestamp of last recovery
	}

	// Ensure TrustScore stays within bounds [0, 100]
	if currentTrustScore > trustScoreMax {
		currentTrustScore = trustScoreMax
	}
	if currentTrustScore < trustScoreMin {
		currentTrustScore = trustScoreMin
	}

	// Update the asset's TrustScore and LastTrustRecoveryAppliedAt in the database
	updates := map[string]interface{}{
		"trust_score": currentTrustScore,
	}
	if asset.LastTrustRecoveryAppliedAt != nil {
		updates["last_trust_recovery_applied_at"] = asset.LastTrustRecoveryAppliedAt
	}

	if err := s.db.Model(&models.Asset{}).Where("asset_hwid = ?", assetID).Updates(updates).Error; err != nil {
		return currentTrustScore, fmt.Errorf("failed to update trust score for asset %d: %w", assetID, err)
	}

	return currentTrustScore, nil
}

// UpdateassetLastIncidentTime updates the LastIncidentAt field for an asset.
// This should be called whenever a new incident is created for an asset.
func (s *ScoreService) UpdateassetLastIncidentTime(assetID uint, incidentTime time.Time) error {
	var asset models.Asset
	if err := s.db.First(&asset, assetID).Error; err != nil {
		return fmt.Errorf("asset not found: %w", err)
	}

	// Only update if the new incident time is more recent than the current LastIncidentAt
	// or if LastIncidentAt is nil.
	if asset.LastIncidentAt == nil || incidentTime.After(*asset.LastIncidentAt) {
		if err := s.db.Model(&models.Asset{}).Where("asset_hwid = ?", assetID).Update("last_incident_at", incidentTime).Error; err != nil {
			return fmt.Errorf("failed to update last incident time for asset %d: %w", assetID, err)
		}
	}
	return nil
}
