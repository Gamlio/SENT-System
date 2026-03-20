package agent_data

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/security"

	"gorm.io/gorm"
)

// Khớp với cấu trúc trả về từ Agent (PortSensor)
type AgentPortRecord struct {
	Port        int    `json:"port"`
	ProcessName string `json:"process_name"`
}

func ProcessPorts(agent models.Agent, data interface{}) {
	bytes, _ := json.Marshal(data)
	var records []AgentPortRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
		return
	}

	// Dùng Transaction để xóa cũ, thêm mới an toàn
	database.DB.Transaction(func(tx *gorm.DB) error {
		// Xóa toàn bộ port cũ của Agent này
		tx.Where("agent_hw_id = ?", agent.HWID).Delete(&models.OpenPort{})

		for _, p := range records {
			portEntry := models.OpenPort{
				AgentHWID:   agent.HWID,
				Port:        p.Port,
				ProcessName: p.ProcessName,
			}
			tx.Create(&portEntry)

			// BẮT BỆNH BẢO MẬT: Mở cổng rủi ro cao (3389, 22)
			if p.Port == 3389 || p.Port == 22 {
				security.TriggerSecurityEvent(agent,
					"Unauthorized Port",
					"[P2] Mở cổng quản trị từ xa (RDP/SSH)",
					fmt.Sprintf("Máy trạm đang mở public cổng %d (Tiến trình: %s). Có rủi ro bị Brute-force.", p.Port, p.ProcessName),
					"P2",
				)
			}
		}
		return nil
	})
}
