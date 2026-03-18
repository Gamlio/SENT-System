package scoring

import (
	"log"
	"math"

	"sent_backend/internal/database"
	"sent_backend/internal/models"
)

// RecalculateRiskScore: Tính lại điểm rủi ro cho Máy trạm dựa trên các CASE đang mở
func RecalculateRiskScore(agentHWID string) {
	var agent models.Agent
	if err := database.DB.Where("hw_id = ?", agentHWID).First(&agent).Error; err != nil {
		log.Println("❌ Lỗi tính điểm: Không tìm thấy Agent", agentHWID)
		return
	}

	// 1. Lấy danh sách HỒ SƠ SỰ CỐ (Incidents) CHƯA ĐÓNG của máy này
	var openIncidents []models.Incident
	database.DB.Where("agent_hw_id = ? AND status IN ('Open', 'Investigating')", agentHWID).Find(&openIncidents)

	// NẾU MÁY SẠCH BÓNG SỰ CỐ -> VỀ 0 ĐIỂM NGAY LẬP TỨC
	if len(openIncidents) == 0 {
		database.DB.Model(&agent).Update("risk_score", 0)
		updateUserScore(agent.UserID) // Đồng bộ điểm cho Dashboard
		return
	}

	// 2. XÁC ĐỊNH TRỌNG SỐ TÀI SẢN (W_asset)
	wAsset := 1.0
	switch agent.DeviceType {
	case "SERVER":
		wAsset = 2.0 // Server: Nhân đôi rủi ro
	case "IT_ADMIN":
		wAsset = 1.5 // Máy IT: Nhân 1.5 rủi ro
	case "GUEST":
		wAsset = 0.8 // Lễ tân: Rủi ro thấp
	}

	// 3. TÍNH ĐIỂM CÁC SỰ CỐ (Weighted Max-Score)
	var maxScore float64 = 0.0
	var secondaryScores []float64

	for _, incident := range openIncidents {
		baseScore := getBaseScoreBySeverity(incident.Severity)

		// Phân loại Lỗi nặng nhất làm gốc
		if baseScore > maxScore {
			if maxScore > 0 {
				secondaryScores = append(secondaryScores, maxScore)
			}
			maxScore = baseScore
		} else {
			secondaryScores = append(secondaryScores, baseScore)
		}
	}

	// 4. TÍNH TỔNG ĐIỂM (R_total)
	sumSecondary := 0.0
	for _, s := range secondaryScores {
		sumSecondary += s
	}

	// Công thức: (Lỗi nặng nhất + 15% tổng các lỗi phụ) * Hệ số máy
	rawTotal := (maxScore + 0.15*sumSecondary) * wAsset

	// Giới hạn (Cap) điểm tối đa là 100
	finalScore := int(math.Min(100, math.Round(rawTotal)))

	// 5. LƯU VÀO DATABASE
	database.DB.Model(&agent).Update("risk_score", finalScore)
	log.Printf("📊 Cập nhật điểm Agent [%s] -> %d điểm (W_asset: %.1f, Lỗi chính: %.1f, Lỗi phụ: %d)",
		agent.Hostname, finalScore, wAsset, maxScore, len(secondaryScores))

	updateUserScore(agent.UserID)
}

// Hàm phụ trợ map mức độ sang điểm gốc
func getBaseScoreBySeverity(severity string) float64 {
	switch severity {
	case "Critical":
		return 80.0
	case "High":
		return 60.0
	case "Medium":
		return 30.0
	case "Low":
		return 10.0
	default:
		return 0.0
	}
}

// Hàm phụ: Tính tổng điểm rủi ro cho User dựa trên các máy họ quản lý
func updateUserScore(userID *uint) {
	if userID == nil {
		return
	}
	var user models.User
	if err := database.DB.First(&user, *userID).Error; err == nil {
		var allUserAgents []models.Agent
		database.DB.Where("user_id = ?", user.ID).Find(&allUserAgents)

		userTotalScore := 0
		for _, a := range allUserAgents {
			userTotalScore += a.RiskScore
		}

		database.DB.Model(&user).Update("risk_score", userTotalScore)
	}
}
