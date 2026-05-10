package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- CÁC STRUCT GÓI GỌN DỮ LIỆU SOC CHI TIẾT ---
type DashboardSummary struct {
	// 1. Chỉ số Thiết bị (assets)
	Totalassets       int64   `json:"total_assets"`
	Onlineassets      int64   `json:"online_assets"`
	Offlineassets     int64   `json:"offline_assets"`
	ZeroTrustCoverage float64 `json:"zero_trust_coverage"`

	// 2. Chỉ số Sự cố (Incidents & Alerts)
	TotalAlerts       int64            `json:"total_alerts"`
	OpenIncidents     int64            `json:"open_incidents"`
	ResolvedIncidents int64            `json:"resolved_incidents"`
	AlertsBySeverity  map[string]int64 `json:"alerts_by_severity"`

	// 3. Chỉ số Rủi ro Tổng thể (Risk & Trust)
	HighRiskassets    int64   `json:"high_risk_assets"`
	AverageTrustScore float64 `json:"average_trust_score"`

	// 4. Danh sách Top (Actionable Data)
	TopRiskassets []assetRiskView `json:"top_risk_assets"`
	RecentAlerts  []AlertView     `json:"recent_alerts"`
}

type assetRiskView struct {
	AssetHWID     string  `json:"asset_hwid"`
	Hostname      string  `json:"hostname"`
	IPAddress     string  `json:"ip_address"`
	RiskScore     int     `json:"risk_score"`
	TrustScore    float64 `json:"trust_score"`
	DepartmentTag string  `json:"department_tag"`
}

type AlertView struct {
	AlertType string    `json:"alert_type"`
	Severity  string    `json:"severity"`
	AssetName string    `json:"asset_name"`
	CreatedAt time.Time `json:"created_at"`
}

type SecurityPosture struct {
	OverallScore       int     `json:"overall_score"` // 0-100 (Chỉ số sức khỏe hệ thống)
	ThreatLevel        string  `json:"threat_level"`  // Stable, Elevated, Critical
	ZeroTrustHealth    float64 `json:"zero_trust_health"`
	VulnerabilityTrend []int   `json:"vulnerability_trend"`
}

// Struct Controller chính
type DashboardController struct{}

func (dc *DashboardController) calculateThreatLevel(score int) string {
	if score >= 80 {
		return "Stable"
	} else if score >= 50 {
		return "Elevated"
	}
	return "Critical"
}

// FetchDetailedSummary: Xử lý mọi logic nghiệp vụ, tính toán phân tán
func (dc *DashboardController) FetchDetailedSummary(orgID uint) (map[string]interface{}, error) {
	var summary DashboardSummary
	summary.AlertsBySeverity = make(map[string]int64)

	// Dùng channel để hứng lỗi từ các Goroutines
	errChan := make(chan error, 8)

	// Ngưỡng thời gian định nghĩa Online (2 phút)
	onlineThreshold := time.Now().Add(-2 * time.Minute)

	var totalRisk float64
	// Luồng 1: Đếm tổng máy & Điểm Trust trung bình
	go func() {
		var assets []models.Asset
		err := database.DB.Where("org_id = ? AND status != ?", orgID, "RETIRED").Find(&assets).Error
		if err == nil {
			summary.Totalassets = int64(len(assets))
			var totalTrust float64
			var zeroTrustCount int64
			var highRiskCount int64

			for _, a := range assets {
				totalTrust += a.TrustScore
				totalRisk += float64(a.RiskScore)
				if a.IsZeroTrust {
					zeroTrustCount++
				}
				if a.RiskScore > 70 {
					highRiskCount++
				}
			}

			summary.HighRiskassets = highRiskCount
			if summary.Totalassets > 0 {
				summary.ZeroTrustCoverage = (float64(zeroTrustCount) / float64(summary.Totalassets)) * 100
				summary.AverageTrustScore = totalTrust / float64(summary.Totalassets)
			}
		}
		errChan <- err
	}()

	// Luồng 2: Đếm máy Online
	go func() {
		errChan <- database.DB.Model(&models.Asset{}).
			Where("org_id = ? AND status = ? AND last_seen >= ?", orgID, "ACTIVE", onlineThreshold).
			Count(&summary.Onlineassets).Error
	}()

	// Luồng 3: Sự cố đang mở
	go func() {
		errChan <- database.DB.Model(&models.Incident{}).
			Where("org_id = ? AND status IN ?", orgID, []string{"Open", "Investigating"}).
			Count(&summary.OpenIncidents).Error
	}()

	// Luồng 4: Sự cố đã đóng
	go func() {
		errChan <- database.DB.Model(&models.Incident{}).
			Where("org_id = ? AND status = ?", orgID, "Resolved").
			Count(&summary.ResolvedIncidents).Error
	}()

	// Luồng 5: Phân bổ Alerts theo mức độ
	go func() {
		var totalAlerts int64
		if database.SecurityAlertCollection != nil {
			cursor, err := database.SecurityAlertCollection.Find(context.TODO(), bson.M{"org_id": orgID})
			if err != nil {
				errChan <- err
				return
			}
			var alerts []models.SecurityAlert
			cursor.All(context.TODO(), &alerts)
			for _, al := range alerts {
				summary.AlertsBySeverity[al.Severity]++
				totalAlerts++
			}
		}
		summary.TotalAlerts = totalAlerts
		errChan <- nil
	}()
	// Luồng 6: Top 5 máy rủi ro cao nhất (Actionable Insight)
	go func() {
		var assets []models.Asset
		err := database.DB.Where("org_id = ?", orgID).Order("risk_score desc").Limit(5).Find(&assets).Error
		for _, a := range assets {
			summary.TopRiskassets = append(summary.TopRiskassets, assetRiskView{
				AssetHWID:     a.AssetHWID,
				Hostname:      a.Hostname,
				IPAddress:     a.IPAddress,
				RiskScore:     a.RiskScore,
				TrustScore:    a.TrustScore,
				DepartmentTag: a.DepartmentTag,
			})
		}
		errChan <- err
	}()

	// Luồng 7: 5 Cảnh báo mới nhất (Live Feed)
	go func() {
		if database.SecurityAlertCollection != nil {
			opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(5)
			cursor, err := database.SecurityAlertCollection.Find(context.TODO(), bson.M{"org_id": orgID}, opts)
			if err != nil {
				errChan <- err
				return
			}
			var alerts []models.SecurityAlert
			cursor.All(context.TODO(), &alerts)
			for _, al := range alerts {
				summary.RecentAlerts = append(summary.RecentAlerts, AlertView{
					AlertType: al.AlertType,
					Severity:  al.Severity,
					AssetName: al.AssetHWID,
					CreatedAt: al.CreatedAt,
				})
			}
		}
		errChan <- nil
	}()

	// Đợi 7 luồng hoàn tất
	for i := 0; i < 7; i++ {
		if err := <-errChan; err != nil {
			return nil, err
		}
	}

	// Xử lý các phép toán phụ
	summary.Offlineassets = summary.Totalassets - summary.Onlineassets

	// Tính toán OverallScore dựa trên RiskScore trung bình và số Incident đang mở
	avgRisk := 0.0
	if summary.Totalassets > 0 {
		avgRisk = totalRisk / float64(summary.Totalassets)
	}

	postureScore := 100.0 - (avgRisk * 0.5) - (float64(summary.OpenIncidents) * 2.0)
	if postureScore < 0 {
		postureScore = 0
	}

	posture := SecurityPosture{
		OverallScore:       int(postureScore),
		ThreatLevel:        dc.calculateThreatLevel(int(postureScore)),
		ZeroTrustHealth:    summary.ZeroTrustCoverage,
		VulnerabilityTrend: []int{0, 0, 0, 0, 0}, // Có thể fill mảng dữ liệu thật từ DB ở các bản cập nhật sau
	}

	return map[string]interface{}{
		"summary": summary,
		"posture": posture,
	}, nil
}
