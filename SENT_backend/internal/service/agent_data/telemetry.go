package agent_data

import (
	"encoding/json"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/security"
	"time"

	"gorm.io/gorm"
)

type AgentTelemetryRecord struct {
	IPAddress   string            `json:"ip_address"`
	FirewallOff bool              `json:"firewall_off"` // Hứng thêm trạng thái Tường lửa
	OpenPorts   []models.OpenPort `json:"open_ports"`
}

func ProcessTelemetry(agent models.Agent, data interface{}) {
	bytes, _ := json.Marshal(data)
	var payload AgentTelemetryRecord
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return
	}

	// 1. Cập nhật trạng thái máy
	updateFields := map[string]interface{}{
		"status":    "online",
		"last_seen": time.Now(),
	}
	if payload.IPAddress != "" {
		updateFields["ip_address"] = payload.IPAddress
	}
	database.DB.Model(&agent).Updates(updateFields)

	// 2. BẮT CASE: TẮT TƯỜNG LỬA
	if payload.FirewallOff {
		security.TriggerSecurityEvent(agent,
			"Firewall Disabled",
			"[P1] Tường lửa hệ thống bị tắt",
			"Lớp phòng thủ Tường lửa (OS Firewall) đã bị vô hiệu hóa.",
			"P1",
		)
	}

	// 3. Xóa port cũ, lưu port mới
	database.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("agent_hw_id = ?", agent.HWID).Delete(&models.OpenPort{})
		for _, port := range payload.OpenPorts {
			port.AgentHWID = agent.HWID
			tx.Create(&port)

			// BẮT CASE: RDP / Telnet
			if port.Port == 3389 {
				security.TriggerSecurityEvent(agent,
					"Unauthorized Port",
					"[P2] Mở cổng Remote Desktop trái phép",
					"Phát hiện cổng 3389 (RDP) đang mở public, nguy cơ bị brute-force.",
					"P2",
				)
			}
		}
		return nil
	})
}
