package agent_data

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/security"

	"gorm.io/gorm"
)

type AgentPortPayload struct {
	OpenPorts []AgentPortRecord `json:"open_ports"`
	IPAddress string            `json:"ip_address"`
}
type AgentPortRecord struct {
	Port        int    `json:"port"`
	ProcessName string `json:"process_name"`
}

func ProcessPorts(agent models.Agent, data interface{}) {
	bytes, _ := json.Marshal(data)
	var payload AgentPortPayload
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return
	}

	database.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("agent_hw_id = ?", agent.HWID).Delete(&models.OpenPort{})
		for _, p := range payload.OpenPorts {
			tx.Create(&models.OpenPort{AgentHWID: agent.HWID, Port: p.Port, ProcessName: p.ProcessName})
			if p.Port == 3389 || p.Port == 22 {
				security.TriggerSecurityEvent(agent, "Unauthorized Port", "[P2] Mở cổng quản trị từ xa", fmt.Sprintf("Cổng %d đang mở.", p.Port), "P2")
			}
		}
		return nil
	})
}
