package scoring

import (
	"log"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
)

// RecalculateRiskScore: Tính lại điểm rủi ro cho Agent và User sở hữu
func RecalculateRiskScore(agentHWID string) {
	var agent models.Agent
	if err := database.DB.Where("hw_id = ?", agentHWID).First(&agent).Error; err != nil {
		log.Println("❌ Lỗi tính điểm: Không tìm thấy Agent", agentHWID)
		return
	}

	// 1. Lấy các cảnh báo/sự cố CHƯA GIẢI QUYẾT của máy này
	var unresolvedAlerts []models.SecurityAlert
	database.DB.Where("hw_id = ? AND is_resolved = ?", agentHWID, false).Find(&unresolvedAlerts)

	// 2. Tính tổng điểm dựa trên Playbook Severity
	totalScore := 0
	for _, alert := range unresolvedAlerts {
		switch alert.Severity {
		case "Low":
			totalScore += 5
		case "Medium": // P3 (Ví dụ: Thiếu bản vá OS, Cắm USB lạ, Phần mềm crack)
			totalScore += 20
		case "High": // P2 (Ví dụ: Tắt Antivirus, Phát hiện Malware, Mở Port lạ)
			totalScore += 50
		case "Critical": // P1 (Ví dụ: Tắt Tường lửa hệ thống - Firewall Disabled)
			totalScore += 80
		}
	}

	// 3. Cập nhật điểm cho Agent
	database.DB.Model(&agent).Update("risk_score", totalScore)
	log.Printf("📊 Đã cập nhật điểm Agent [%s] -> %d điểm", agentHWID, totalScore)

	// 4. Cập nhật điểm cho User quản lý máy này (Nếu máy có chủ)
	if agent.UserID != nil {
		var user models.User
		if err := database.DB.First(&user, *agent.UserID).Error; err == nil {
			// Lấy TỔNG ĐIỂM của tất cả các máy mà User này đang quản lý
			var allUserAgents []models.Agent
			database.DB.Where("user_id = ?", user.ID).Find(&allUserAgents)

			userTotalScore := 0
			for _, a := range allUserAgents {
				userTotalScore += a.RiskScore
			}

			database.DB.Model(&user).Update("risk_score", userTotalScore)
			log.Printf("👤 Đã cập nhật điểm User [%s] -> %d điểm", user.Username, userTotalScore)
		}
	}
}
