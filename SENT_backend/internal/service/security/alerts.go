package security

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
)

// CreateAlert: Hàm cốt lõi tạo cảnh báo và NÂNG CẤP LÊN SỰ CỐ
func CreateAlert(agent models.Agent, alertType, title, description, severity string) {
	var count int64

	// 1. Chống Spam: Nếu máy này đã có cảnh báo cùng loại chưa xử lý -> Bỏ qua
	database.DB.Model(&models.SecurityAlert{}).
		Where("hw_id = ? AND alert_type = ? AND is_resolved = ?", agent.HWID, alertType, false).
		Count(&count)

	if count > 0 {
		return
	}

	// 2. Tạo Alert mới
	newAlert := models.SecurityAlert{
		OrgID:       agent.OrgID,
		HWID:        agent.HWID,
		AlertType:   alertType,
		Title:       title,
		Description: description,
		Severity:    severity,
		IsResolved:  false,
	}

	if err := database.DB.Create(&newAlert).Error; err != nil {
		fmt.Println("Lỗi tạo Alert:", err)
		return
	}

	// 3. [QUAN TRỌNG] Tự động gom nhóm Alert này vào một Sự cố (Incident) kèm Playbook
	GroupAlertToIncident(&newAlert)
}

// GroupAlertToIncident: Logic biến Alert lẻ tẻ thành Hồ sơ sự cố có quy trình
func GroupAlertToIncident(alert *models.SecurityAlert) {
	var incident models.Incident

	// A. Tìm xem đã có Sự cố nào đang MỞ (Open/Investigating) cùng loại cho máy này chưa?
	err := database.DB.Where("agent_hw_id = ? AND type = ? AND status IN ?",
		alert.HWID, alert.AlertType, []string{"Open", "Investigating"}).
		First(&incident).Error

	if err != nil {
		// B. Nếu chưa có -> TẠO SỰ CỐ MỚI (KÈM PLAYBOOK CHECKLIST)

		// Lấy mẫu quy trình xử lý (Checklist) dựa trên loại lỗi
		playbookSteps := getPlaybookTemplate(alert.AlertType)
		progressJSON, _ := json.Marshal(map[string]interface{}{"steps": playbookSteps})

		// Mapping độ ưu tiên
		priority := "P3" // Mặc định
		if alert.Severity == "Critical" {
			priority = "P1"
		}
		if alert.Severity == "High" {
			priority = "P2"
		}

		newIncident := models.Incident{
			OrgID:            alert.OrgID,
			AgentHWID:        alert.HWID,
			Type:             alert.AlertType,
			Priority:         priority,
			Severity:         alert.Severity,
			Status:           "Open",
			Description:      fmt.Sprintf("Hệ thống phát hiện: %s", alert.Title),
			PlaybookName:     alert.AlertType,
			PlaybookProgress: string(progressJSON), // <--- LƯU JSON CHECKLIST VÀO DB
		}

		if createErr := database.DB.Create(&newIncident).Error; createErr != nil {
			fmt.Println("Lỗi tạo Incident:", createErr)
			return
		}
		incident = newIncident
	}

	// C. Gắn Alert này vào Incident (Để hiển thị trong cột trái giao diện)
	database.DB.Model(alert).Update("incident_id", incident.ID)
}

// getPlaybookTemplate: Định nghĩa các bước xử lý (Admin phải làm theo cái này)
func getPlaybookTemplate(alertType string) []map[string]interface{} {
	switch alertType {
	case "Firewall Disabled":
		return []map[string]interface{}{
			{"id": 1, "text": "Liên hệ người dùng xác minh lý do tắt Firewall", "done": false},
			{"id": 2, "text": "Kiểm tra log xem có tiến trình lạ tắt Firewall không", "done": false},
			{"id": 3, "text": "Yêu cầu bật lại Firewall hoặc đẩy GPO ép bật", "done": false},
		}
	case "Malware/AV Alert", "AI_MALWARE_SUSPICION":
		return []map[string]interface{}{
			{"id": 1, "text": "Cô lập máy trạm khỏi mạng nội bộ (Isolate)", "done": false},
			{"id": 2, "text": "Thu thập mẫu file nghi ngờ gửi Sandbox", "done": false},
			{"id": 3, "text": "Chạy quét Full Scan bằng Antivirus", "done": false},
			{"id": 4, "text": "Xác nhận máy sạch và kết nối lại mạng", "done": false},
		}
	case "Software Violation":
		return []map[string]interface{}{
			{"id": 1, "text": "Yêu cầu người dùng gỡ bỏ phần mềm không được phép", "done": false},
			{"id": 2, "text": "Kiểm tra lại Inventory xem phần mềm đã biến mất chưa", "done": false},
			{"id": 3, "text": "Nhắc nhở về chính sách an toàn thông tin", "done": false},
		}
	case "USB Violation":
		return []map[string]interface{}{
			{"id": 1, "text": "Xác minh thiết bị USB vừa cắm vào", "done": false},
			{"id": 2, "text": "Quét virus thiết bị USB (nếu được phép dùng)", "done": false},
			{"id": 3, "text": "Yêu cầu rút USB nếu không phục vụ công việc", "done": false},
		}
	default:
		return []map[string]interface{}{
			{"id": 1, "text": "Xác minh mức độ nghiêm trọng của cảnh báo", "done": false},
			{"id": 2, "text": "Thực hiện hành động khắc phục", "done": false},
			{"id": 3, "text": "Ghi chú kết quả xử lý", "done": false},
		}
	}
}
