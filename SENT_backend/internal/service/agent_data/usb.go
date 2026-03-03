package agent_data

import (
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/security" // Import security
)

func ProcessUSB(agent models.Agent, data interface{}) {
	usbData, ok := data.([]interface{})
	if !ok {
		return
	}

	// Xóa log thiết bị cũ
	database.DB.Where("agent_hw_id = ?", agent.HWID).Delete(&models.USBLog{})

	for _, item := range usbData {
		u, _ := item.(map[string]interface{})

		// Chỉ lưu log thô, không xét whitelist/blacklist ở đây nữa
		newLog := models.USBLog{
			AgentHWID:  agent.HWID,
			DeviceName: fmt.Sprintf("%v", u["device_name"]),
			DeviceID:   fmt.Sprintf("%v", u["device_id"]),
			EventType:  "active",
		}

		// Lưu vào DB
		database.DB.Create(&newLog)
	}

	// QUAN TRỌNG: Gọi hàm kiểm tra chính sách USB mới
	go security.CheckUSBCompliance(agent, data)
}
