package controllers

import (
	"context"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- CÁC STRUCT GÓI GỌN DỮ LIỆU SOC CHI TIẾT ---
type DashboardSummary struct {
	// 1. Chỉ số Thiết bị (Agents)
	TotalAgents       int64   `json:"total_agents"`
	OnlineAgents      int64   `json:"online_agents"`
	OfflineAgents     int64   `json:"offline_agents"`
	ZeroTrustCoverage float64 `json:"zero_trust_coverage"`

	// 2. Chỉ số Sự cố (Incidents & Alerts)
	TotalAlerts       int64            `json:"total_alerts"`
	OpenIncidents     int64            `json:"open_incidents"`
	ResolvedIncidents int64            `json:"resolved_incidents"`
	AlertsBySeverity  map[string]int64 `json:"alerts_by_severity"`

	// 3. Chỉ số Rủi ro Tổng thể (Risk & Trust)
	HighRiskAgents    int64   `json:"high_risk_agents"`
	AverageTrustScore float64 `json:"average_trust_score"`

	// 4. Danh sách Top (Actionable Data)
	TopRiskAgents []AgentRiskView `json:"top_risk_agents"`
	RecentAlerts  []AlertView     `json:"recent_alerts"`
}

type AgentRiskView struct {
	HWID          string  `json:"hwid"`
	Hostname      string  `json:"hostname"`
	IPAddress     string  `json:"ip_address"`
	RiskScore     int     `json:"risk_score"`
	TrustScore    float64 `json:"trust_score"`
	DepartmentTag string  `json:"department_tag"`
}

type AlertView struct {
	AlertType string    `json:"alert_type"`
	Severity  string    `json:"severity"`
	AgentName string    `json:"agent_name"`
	CreatedAt time.Time `json:"created_at"`
}

// Struct Controller chính
type DashboardController struct{}

// FetchDetailedSummary: Xử lý mọi logic nghiệp vụ, tính toán phân tán
func (dc *DashboardController) FetchDetailedSummary(orgID uint) (*DashboardSummary, error) {
	var summary DashboardSummary
	summary.AlertsBySeverity = make(map[string]int64)

	// Dùng channel để hứng lỗi từ các Goroutines
	errChan := make(chan error, 8)

	// Ngưỡng thời gian định nghĩa Online (2 phút)
	onlineThreshold := time.Now().Add(-2 * time.Minute)

	// Luồng 1: Đếm tổng máy & Điểm Trust trung bình
	go func() {
		var agents []models.Agent
		err := database.DB.Where("org_id = ? AND status != ?", orgID, "RETIRED").Find(&agents).Error
		if err == nil {
			summary.TotalAgents = int64(len(agents))
			var totalTrust float64
			var zeroTrustCount int64
			var highRiskCount int64

			for _, a := range agents {
				totalTrust += a.TrustScore
				if a.IsZeroTrust {
					zeroTrustCount++
				}
				if a.RiskScore > 70 {
					highRiskCount++
				}
			}

			summary.HighRiskAgents = highRiskCount
			if summary.TotalAgents > 0 {
				summary.ZeroTrustCoverage = (float64(zeroTrustCount) / float64(summary.TotalAgents)) * 100
				summary.AverageTrustScore = totalTrust / float64(summary.TotalAgents)
			}
		}
		errChan <- err
	}()

	// Luồng 2: Đếm máy Online
	go func() {
		errChan <- database.DB.Model(&models.Agent{}).
			Where("org_id = ? AND status = ? AND last_seen >= ?", orgID, "ACTIVE", onlineThreshold).
			Count(&summary.OnlineAgents).Error
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
		var agents []models.Agent
		err := database.DB.Where("org_id = ?", orgID).Order("risk_score desc").Limit(5).Find(&agents).Error
		for _, a := range agents {
			summary.TopRiskAgents = append(summary.TopRiskAgents, AgentRiskView{
				HWID:          a.HWID,
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
					AgentName: al.HWID,
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
	summary.OfflineAgents = summary.TotalAgents - summary.OnlineAgents

	return &summary, nil
}
