package controllers

import (
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"

	"github.com/gin-gonic/gin"
)

type DashboardSummary struct {
	TotalAgents       int64   `json:"total_agents"`
	OnlineAgents      int64   `json:"online_agents"`
	OpenIncidents     int64   `json:"open_incidents"`
	HighRiskAgents    int64   `json:"high_risk_agents"`
	ZeroTrustCoverage float64 `json:"zero_trust_coverage"`
	// Danh sách máy rủi ro
	TopRiskAgents []AgentRiskView `json:"top_risk_agents"`
}

type AgentRiskView struct {
	HWID      string `json:"hwid"`
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip_address"`
	RiskScore int    `json:"risk_score"`
}

func GetDashboardSummary(c *gin.Context) {
	var summary DashboardSummary

	// 1. Đếm song song bằng Goroutines (Tối ưu I/O)
	errChan := make(chan error, 5)

	go func() {
		var count int64
		errChan <- database.DB.Model(&models.Agent{}).Count(&count).Error
		summary.TotalAgents = count
	}()

	go func() {
		var count int64
		errChan <- database.DB.Model(&models.Agent{}).Where("is_online = ?", true).Count(&count).Error
		summary.OnlineAgents = count
	}()

	go func() {
		var count int64
		errChan <- database.DB.Model(&models.Incident{}).Where("status = ?", "Open").Count(&count).Error
		summary.OpenIncidents = count
	}()

	go func() {
		var count int64
		// Query những máy có điểm rủi ro nguy hiểm (Dựa theo SCORING.md > 70 là Đỏ)
		errChan <- database.DB.Model(&models.Agent{}).Where("risk_score > ?", 70).Count(&count).Error
		summary.HighRiskAgents = count
	}()

	go func() {
		var agents []models.Agent
		// Lấy Top 5 máy có điểm Risk cao nhất
		errChan <- database.DB.Order("risk_score desc").Limit(5).Find(&agents).Error

		for _, a := range agents {
			summary.TopRiskAgents = append(summary.TopRiskAgents, AgentRiskView{
				HWID:      a.HWID,
				Hostname:  a.Hostname,
				IPAddress: a.IPAddress,
				RiskScore: a.RiskScore,
			})
		}
	}()

	// Chờ tất cả 5 query hoàn thành
	for i := 0; i < 5; i++ {
		if err := <-errChan; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy dữ liệu thống kê"})
			return
		}
	}

	// 2. Tính toán các chỉ số phái sinh
	if summary.TotalAgents > 0 {
		var zeroTrustCount int64
		database.DB.Model(&models.Agent{}).Where("is_zero_trust = ?", true).Count(&zeroTrustCount)
		summary.ZeroTrustCoverage = float64(zeroTrustCount) / float64(summary.TotalAgents) * 100
	}

	c.JSON(http.StatusOK, summary)
}
