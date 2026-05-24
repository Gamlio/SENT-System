package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"

	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

// Department Matrix mapping SensorType and DepartmentTag to Priority
// This can be loaded from a config or database for more flexibility
var departmentMatrix = map[string]map[string]string{
	"USB_PLUG": {
		"DEV":     "P3",
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

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

var scoreTimers sync.Map

func RecalculateRiskScore(assetAssetID string) {
	if timer, ok := scoreTimers.Load(assetAssetID); ok {
		timer.(*time.Timer).Reset(2 * time.Second)
		return
	}

	timer := time.AfterFunc(2*time.Second, func() {
		scoreTimers.Delete(assetAssetID)
		processRiskScore(assetAssetID)
	})
	scoreTimers.Store(assetAssetID, timer)
}

func processRiskScore(assetAssetID string) {
	var asset models.Asset
	if err := database.DB.Select("id", "asset_hwid", "hostname", "risk_score", "trust_score", "asset_type_id").Preload("AssetType").Where("asset_hwid = ?", assetAssetID).First(&asset).Error; err != nil {
		return
	}
	var activeAlerts []models.SecurityAlert
	if database.SecurityAlertCollection != nil {
		cursor, err := database.SecurityAlertCollection.Find(context.TODO(), bson.M{
			"asset_hwid": assetAssetID, "is_resolved": false,
		})
		if err == nil {
			cursor.All(context.TODO(), &activeAlerts)
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

	if len(activeAlerts) == 0 && len(openPorts) == 0 && len(usbLogs) == 0 && len(ioActivities) == 0 {
		database.DB.Model(&asset).Updates(map[string]interface{}{
			"risk_score":     0.0,
			"security_grade": "A",
		})
		return
	}

	priorityCounts := make(map[string]int)
	hasP1 := false
	for _, alert := range activeAlerts {
		priorityCounts[alert.Priority]++
		if alert.Priority == "P1" {
			hasP1 = true
		}
	}

	// --- Tính toán R_active theo công thức logarit từ SCORING.md ---
	rActive := 0.0
	alertWeights := map[string]float64{
		"P1": 5.0,
		"P2": 2.5,
		"P3": 1.0,
	}
	for priority, count := range priorityCounts {
		if weight, ok := alertWeights[priority]; ok {
			rActive += weight * math.Log2(float64(count)+1.0)
		}
	}

	// --- Tính toán R_history với hàm suy giảm theo thời gian ---
	var pastIncidents []models.Incident
	halfYearAgo := time.Now().AddDate(0, 0, -180)
	rHistory := 0.0

	if err := database.DB.Where("asset_hwid = ? AND created_at >= ?", assetAssetID, halfYearAgo).Find(&pastIncidents).Error; err == nil {
		for _, inc := range pastIncidents {
			daysOld := time.Since(inc.CreatedAt).Hours() / 24.0
			hImpact := 1.0
			switch inc.Priority {
			case "P1":
				hImpact = 5.0
			case "P2":
				hImpact = 2.5
			case "P3":
				hImpact = 1.0
			}

			rHistory += hImpact / (1.0 + 0.01*daysOld)
		}
	}

	// --- Tính toán các hệ số ngữ cảnh C và V theo SCORING.md ---
	cFactor := 1.0
	if asset.AssetType != nil && asset.AssetType.RiskWeight > 0 {
		cFactor = asset.AssetType.RiskWeight
	}

	// V = 1.0 + (Số cổng mạng mở * 0.05) + log10(n_usb_unknown + 1) + log10((Delta Disk I/O / 10^6) + 1)
	var totalDiskIO uint64
	for _, activity := range ioActivities {

		totalDiskIO += activity.DiskBytesRead + activity.DiskBytesWritten
	}

	vFactor := 1.0 +
		(float64(len(openPorts)) * 0.05) +
		math.Log10(float64(len(usbLogs))+1.0) +
		math.Log10((float64(totalDiskIO)/1_000_000)+1.0)

	// --- Tính điểm tổng hợp ---
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

	if displayScore > 80.0 {
		log.Printf("🚨 [FORENSIC SEAL] %s: Điểm rủi ro=%.1f > 80. Hệ thống kích hoạt Snapshot Immutable!", asset.Hostname, displayScore)
		sendCommandToBehaviorService(map[string]interface{}{
			"asset":    asset,
			"category": "Forensic Snapshot",
			"value":    "High Risk Score",
			"title":    "[P1] Kích hoạt thu thập dữ liệu pháp y",
			"desc":     fmt.Sprintf("Điểm rủi ro của máy trạm '%s' đã vượt ngưỡng nguy hiểm (%.1f/100). Hệ thống tự động yêu cầu thu thập danh sách tiến trình và cổng mạng đang mở để phân tích sâu.", asset.Hostname, displayScore),
			"priority": "P1",
			"action":   "COLLECT_FORENSICS",
		})
	}

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
		return "P3", fmt.Errorf("department tag '%s' not found for sensor type '%s' in matrix, defaulting to P3", departmentTag, sensorType)
	}
	// If sensor type not found, return a default or error
	return "P3", fmt.Errorf("sensor type '%s' not found in department matrix, defaulting to P3", sensorType)
}

// CalculateCurrentRiskScore calculates the instantaneous risk score (R_current) for an asset.
// This score reflects the current active threats.
func (s *ScoreService) CalculateCurrentRiskScore(assetID uint) (float64, error) {
	// 1. Lấy thông tin thiết bị và cấu hình loại máy trạm (AssetType) để lấy RiskWeight (cFactor)
	var asset models.Asset
	if err := s.db.Preload("AssetType").First(&asset, assetID).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch asset info: %w", err)
	}

	var incidents []models.Incident
	// Lấy các sự cố đang mở (Open) thuộc về máy trạm này
	if err := s.db.Where("asset_hwid = ? AND status = ?", asset.AssetHWID, "Open").Find(&incidents).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch open incidents for asset %s: %w", asset.AssetHWID, err)
	}

	// Nếu không có sự cố nào, reset điểm rủi ro tức thời về 0
	if len(incidents) == 0 {
		if err := s.db.Model(&models.Asset{}).Where("asset_hwid = ?", asset.AssetHWID).Update("risk_score", 0.0).Error; err != nil {
			return 0, fmt.Errorf("failed to reset risk score for asset %s: %w", asset.AssetHWID, err)
		}
		return 0, nil
	}

	// 2. Đếm số lượng sự cố theo từng mức độ ưu tiên
	priorityCounts := make(map[string]int)
	for _, inc := range incidents {
		priorityCounts[inc.Priority]++
	}

	// 3. Áp dụng công thức R_active từ SCORING.md: sum(T_i * log2(n_i + 1))
	rActive := 0.0
	incidentWeights := map[string]float64{
		"P1": 5.0,
		"P2": 2.5,
		"P3": 1.0,
	}
	for priority, count := range priorityCounts {
		if weight, ok := incidentWeights[priority]; ok {
			rActive += weight * math.Log2(float64(count)+1.0)
		}
	}

	// 4. Lấy hệ số loại thiết bị cFactor (C) từ cấu hình AssetType đã gán
	cFactor := 1.0
	if asset.AssetType != nil && asset.AssetType.RiskWeight > 0 {
		cFactor = asset.AssetType.RiskWeight
	}

	// 5. Tính điểm rủi ro tức thời (chỉ gồm R_active * C, không có R_history và V)
	rCurrent := rActive * cFactor

	// Giới hạn trần rủi ro tuyệt đối không vượt quá 100
	if rCurrent > 100.0 {
		rCurrent = 100.0
	}

	// 6. Cập nhật điểm số chuẩn hóa vào cơ sở dữ liệu Postgres
	if err := s.db.Model(&models.Asset{}).Where("asset_hwid = ?", asset.AssetHWID).Update("risk_score", rCurrent).Error; err != nil {
		return rCurrent, fmt.Errorf("failed to update risk score for asset %s: %w", asset.AssetHWID, err)
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

func sendCommandToBehaviorService(payload map[string]interface{}) {
	go func() {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Printf("🚨 [SCORING] Lỗi tạo JSON để gửi lệnh: %v", err)
			return
		}

		url := "http://behavior-service:8000/api/v1/behaviors/log"

		resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("🚨 [SCORING] Lỗi gửi lệnh sang Behavior Service: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			log.Printf("🚨 [SCORING] Behavior Service phản hồi lỗi: %s", resp.Status)
		}
	}()
}
