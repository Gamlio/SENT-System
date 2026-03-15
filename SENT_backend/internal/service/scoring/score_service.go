package scoring

import (
	"log"
	"math"
	"time"

	"sent_backend/internal/database"
	"sent_backend/internal/models"
)

// RecalculateRiskScore: Tính lại điểm rủi ro cho Agent bằng mô hình Toán học SOC
func RecalculateRiskScore(agentHWID string) {
	var agent models.Agent
	if err := database.DB.Where("hw_id = ?", agentHWID).First(&agent).Error; err != nil {
		log.Println("❌ Lỗi tính điểm: Không tìm thấy Agent", agentHWID)
		return
	}

	// 1. Lấy các cảnh báo CHƯA GIẢI QUYẾT của máy này
	var unresolvedAlerts []models.SecurityAlert
	database.DB.Where("hw_id = ? AND is_resolved = ?", agentHWID, false).Find(&unresolvedAlerts)

	// NẾU MÁY SẠCH BÓNG -> VỀ 0 ĐIỂM NGAY LẬP TỨC
	if len(unresolvedAlerts) == 0 {
		database.DB.Model(&agent).Update("risk_score", 0)
		updateUserScore(agent.UserID) // Đồng bộ điểm cho người quản lý
		return
	}

	// 2. XÁC ĐỊNH TRỌNG SỐ TÀI SẢN (W_asset)
	wAsset := 1.0 // Mặc định là máy Văn phòng (Tier 3)
	switch agent.DeviceType {
	case "SERVER":
		wAsset = 1.5 // Tier 1: Cực kỳ quan trọng
	case "IT_ADMIN":
		wAsset = 1.2 // Tier 2: Rủi ro cao do có đặc quyền
	case "OFFICE", "":
		wAsset = 1.0 // Tier 3: Tiêu chuẩn
	case "GUEST":
		wAsset = 0.8 // Tier 4: Hạn chế, ít quan trọng
	}

	var maxE float64 = 0
	var secondaryE []float64
	now := time.Now()

	// 3. TÍNH ĐIỂM CHO TỪNG SỰ KIỆN (Event Score - Ei)
	for _, alert := range unresolvedAlerts {
		// A. Trọng số cơ bản (W_base)
		var wBase float64
		switch alert.Severity {
		case "Critical":
			wBase = 80 // Rất nghiêm trọng (Tắt Tường lửa, Malware)
		case "High":
			wBase = 60 // Nghiêm trọng (Mở port lạ, USB lạ)
		case "Medium":
			wBase = 30 // Trung bình (Cài phần mềm trái phép)
		case "Low":
			wBase = 10 // Thấp (Thiếu update vặt)
		default:
			wBase = 10
		}

		// B. Thang độ thời gian (W_time) - Phạt nếu Helpdesk chây ì
		hoursOpen := now.Sub(alert.CreatedAt).Hours()
		wTime := 1.0
		if hoursOpen >= 24 {
			wTime = 1.5 // Quá 1 ngày chưa xử lý -> x1.5 điểm
		} else if hoursOpen >= 4 {
			wTime = 1.2 // Quá 4 tiếng chưa xử lý -> x1.2 điểm
		}

		// C. Tần suất vi phạm (W_freq) - Đánh giá tính cố tình
		var historyCount int64
		database.DB.Model(&models.SecurityAlert{}).
			Where("hw_id = ? AND alert_type = ?", agentHWID, alert.AlertType).
			Count(&historyCount)

		wFreq := 1.0
		if historyCount > 1 {
			wFreq = 1.3 // Cố tình lặp lại cùng 1 lỗi nhiều lần (Ngoan cố)
		}

		// Tính điểm sự kiện: E = Base * Time * Freq
		eventScore := wBase * wTime * wFreq

		// Phân loại: Tìm ra Lỗi có điểm cao nhất để làm Gốc
		if eventScore > maxE {
			if maxE > 0 {
				secondaryE = append(secondaryE, maxE)
			}
			maxE = eventScore
		} else {
			secondaryE = append(secondaryE, eventScore)
		}
	}

	// 4. TÍNH TỔNG ĐIỂM (R_total)
	// Áp dụng công thức: Tổng = (Max + 15% các lỗi phụ) * W_asset
	sumSecondary := 0.0
	for _, e := range secondaryE {
		sumSecondary += e
	}

	rawTotal := (maxE + 0.15*sumSecondary) * wAsset

	// Giới hạn (Cap) điểm tối đa là 100
	finalScore := int(math.Min(100, math.Round(rawTotal)))

	// 5. LƯU VÀO DATABASE
	database.DB.Model(&agent).Update("risk_score", finalScore)
	log.Printf("📊 Cập nhật điểm Agent [%s] -> %d điểm (W_asset: %.1f, MaxE: %.1f, Lỗi phụ: %d)", agent.Hostname, finalScore, wAsset, maxE, len(secondaryE))

	// 6. ĐỒNG BỘ ĐIỂM NGƯỜI QUẢN LÝ
	updateUserScore(agent.UserID)
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
