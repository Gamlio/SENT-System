package data // Đổi sang package data

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"

	"gorm.io/gorm"
)

type AgentTelemetryRecord struct {
	IPAddress   string            `json:"ip_address"`
	FirewallOff bool              `json:"firewall_off"`
	OpenPorts   []models.OpenPort `json:"open_ports"`
}

// ProcessPorts: Đổi tên để khớp với điều phối viên agent_service.go
func ProcessPorts(agent models.Agent, data interface{}) {
	bytes, _ := json.Marshal(data)
	var payload AgentTelemetryRecord
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return
	}

	incSvc := &incidents.IncidentService{}

	// 1. Kiểm tra trạng thái tường lửa
	if payload.FirewallOff {
		incSvc.TriggerSecurityEvent(agent,
			"Firewall Disabled",
			"[P1] Tường lửa hệ thống bị tắt",
			"Phát hiện trạng thái Firewall không hoạt động qua bản tin viễn trắc.",
			"P1",
		)
	}

	// 2. Cập nhật danh sách cổng mạng trong Transaction
	database.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("agent_hw_id = ?", agent.HWID).Delete(&models.OpenPort{})

		for _, port := range payload.OpenPorts {
			port.AgentHWID = agent.HWID
			tx.Create(&port)

			// 3. Kiểm tra các cổng quản trị từ xa nguy hiểm
			if port.Port == 3389 || port.Port == 22 {
				incSvc.TriggerSecurityEvent(agent,
					"Unauthorized Port",
					fmt.Sprintf("[P2] Mở cổng quản trị (%d) trái phép", port.Port),
					fmt.Sprintf("Phát hiện tiến trình '%s' đang mở cổng %d.", port.ProcessName, port.Port),
					"P2",
				)
			}
		}
		return nil
	})
}
