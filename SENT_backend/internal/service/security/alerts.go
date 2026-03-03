package security

import (
	"sent_backend/internal/database"
	"sent_backend/internal/models"
)

// CreateAlert: Hàm cốt lõi để tạo cảnh báo
// Nó tự động kiểm tra xem có cảnh báo tương tự nào đang "chưa xử lý" không để tránh spam database.
func CreateAlert(agent models.Agent, alertType, title, description, severity string) {
	var count int64

	// 1. Kiểm tra trùng lặp: Nếu máy này đã có cảnh báo cùng loại và chưa xử lý -> Bỏ qua
	database.DB.Model(&models.SecurityAlert{}).
		Where("hw_id = ? AND alert_type = ? AND is_resolved = ?", agent.HWID, alertType, false).
		Count(&count)

	if count > 0 {
		return // Đã có cảnh báo rồi, không spam thêm
	}

	// 2. Tạo cảnh báo mới
	newAlert := models.SecurityAlert{
		OrgID:       agent.OrgID,
		HWID:        agent.HWID,
		AlertType:   alertType,
		Title:       title,
		Description: description,
		Severity:    severity, // Critical, High, Medium, Low
		IsResolved:  false,
	}

	database.DB.Create(&newAlert)
}

// CreateUSBAlert: Hàm tiện ích riêng cho USB (để code bên usb.go gọi cho gọn)
func CreateUSBAlert(agent models.Agent, deviceName, deviceID string) {
	CreateAlert(
		agent,
		"USB_UNAUTHORIZED",
		"Phát hiện thiết bị ngoại vi lạ",
		"Thiết bị USB không nằm trong Whitelist: "+deviceName+" ("+deviceID+")",
		"High",
	)
}
