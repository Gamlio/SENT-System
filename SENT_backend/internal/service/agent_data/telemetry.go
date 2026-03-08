package agent_data

import (
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"gorm.io/gorm"
)

// ProcessTelemetry xử lý danh sách cổng mạng và cập nhật IP
func ProcessTelemetry(agent models.Agent, data interface{}) {
	telemetryData, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	// 1. Cập nhật trạng thái Agent
	updateFields := map[string]interface{}{
		"status":    "online",
		"last_seen": time.Now(),
	}

	if ip, exists := telemetryData["ip_address"]; exists && ip != "" {
		updateFields["ip_address"] = fmt.Sprintf("%v", ip)
	}
	database.DB.Model(&agent).Updates(updateFields)

	// 2. Vì Agent đã băm Hash ở Local, nếu có Data gửi lên tức là Cổng mạng đã thay đổi.
	// Ta chỉ việc Parse và lưu thẳng vào Database.
	var payload struct {
		OpenPorts []models.OpenPort `json:"open_ports"`
	}

	if err := MapToStruct(data, &payload); err != nil {
		return
	}

	// 3. Xóa port cũ, thêm port mới bằng Transaction
	database.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("agent_hw_id = ?", agent.HWID).Delete(&models.OpenPort{})

		for _, p := range payload.OpenPorts {
			p.AgentHWID = agent.HWID
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
