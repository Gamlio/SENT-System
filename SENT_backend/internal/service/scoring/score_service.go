package scoring

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"sent_backend/internal/database"
	"sent_backend/internal/models"

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

func RecalculateRiskScore(assetHWID string) {
	var asset models.Asset
	if err := database.DB.Where("hw_id = ?", assetHWID).First(&asset).Error; err != nil {
		return
	}

	// 1. Lấy telemetry ở MongoDB
	var alerts []models.SecurityAlert
	if database.SecurityAlertCollection != nil {
		cursor, err := database.SecurityAlertCollection.Find(context.TODO(), bson.M{"hw_id": assetHWID, "is_resolved": false})
		if err == nil {
			cursor.All(context.TODO(), &alerts)
		}
	}

	var ioActivities []models.AssetIOActivity
	if database.AssetIOActivityCollection != nil {
		cursor, err := database.AssetIOActivityCollection.Find(context.TODO(), bson.M{"asset_hwid": assetHWID})
		if err == nil {
			cursor.All(context.TODO(), &ioActivities)
		}
	}

	var openPorts []models.OpenPort
	if database.OpenPortCollection != nil {
		cursor, err := database.OpenPortCollection.Find(context.TODO(), bson.M{"asset_hwid": assetHWID, "status": "OPEN"})
		if err == nil {
			cursor.All(context.TODO(), &openPorts)
		}
	}

	var usbLogs []models.USBLog
	if database.USBCollection != nil {
		cursor, err := database.USBCollection.Find(context.TODO(), bson.M{"asset_hwid": assetHWID, "event_type": "CONNECTED"})
		if err == nil {
			cursor.All(context.TODO(), &usbLogs)
		}
	}

	// 2. Tính điểm từ telemetry
	n := len(alerts)
	sumSi := 0.0
	hasP1 := false
	for _, alert := range alerts {
		switch alert.Priority {
		case "P1":
			sumSi += 50.0
			hasP1 = true
		case "P2":
			sumSi += 25.0
		case "P3":
			sumSi += 10.0
		case "P4":
			sumSi += 5.0
		default:
			sumSi += 5.0
		}
	}

	// Thêm điểm từ trạng thái open port nếu có
	if len(openPorts) > 0 {
		sumSi += float64(len(openPorts)) * 2.0
	}

	// Thêm điểm từ USB vi phạm (coi là mỗi USB mới là 3 điểm)
	if len(usbLogs) > 0 {
		sumSi += float64(len(usbLogs)) * 3.0
	}

	// Thêm điểm IO spike (nếu có bản ghi I/O recent)
	if len(ioActivities) > 0 {
		sumSi += float64(len(ioActivities)) * 5.0
	}

	if n == 0 && len(openPorts) == 0 && len(usbLogs) == 0 && len(ioActivities) == 0 {
		database.DB.Model(&asset).Updates(map[string]interface{}{"risk_score": 0.0})
		return
	}

	En := math.Pow(1.2, float64(n+len(openPorts)))
	rawScore := (sumSi * En) / kFactor
	currentRisk := 100.0 * (1.0 - math.Exp(-rawScore))

	if currentRisk > 100.0 {
		currentRisk = 100.0
	}

	// 3. Cập nhật thông số asset
	now := time.Now()
	updates := map[string]interface{}{
		"risk_score":       currentRisk,
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
	log.Printf("📊 [Scoring v6] asset %s: Alerts=%d openPorts=%d usb=%d io=%d | Rủi ro=%.1f | Uy tín=%.1f", asset.Hostname, n, len(openPorts), len(usbLogs), len(ioActivities), currentRisk, asset.TrustScore)
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
	if err := s.db.Where("asset_id = ? AND status = ?", assetID, "Open").Find(&incidents).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch open incidents for asset %d: %w", assetID, err)
	}

	if len(incidents) == 0 {
		// No open incidents, current risk score is 0
		// Ensure asset's risk_score is updated to 0 if it was previously higher
		if err := s.db.Model(&models.Asset{}).Where("id = ?", assetID).Update("risk_score", 0.0).Error; err != nil {
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
	if err := s.db.Where("asset_id = ? AND occurred_at >= ? AND (priority = ? OR priority = ?)",
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

	if err := s.db.Model(&models.Asset{}).Where("id = ?", assetID).Updates(updates).Error; err != nil {
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
		if err := s.db.Model(&models.Asset{}).Where("id = ?", assetID).Update("last_incident_at", incidentTime).Error; err != nil {
			return fmt.Errorf("failed to update last incident time for asset %d: %w", assetID, err)
		}
	}
	return nil
}
