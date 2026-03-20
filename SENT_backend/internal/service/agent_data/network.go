package agent_data

import (
	"encoding/json"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"
)

type AgentNetworkRecord struct {
	IPAddress string `json:"ip_address"`
	// Tương lai có thể thêm: MACAddress, Gateway, Subnet...
}

func ProcessNetwork(agent models.Agent, data interface{}) {
	bytes, _ := json.Marshal(data)
	var payload AgentNetworkRecord
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return
	}

	// Cập nhật trạng thái máy và IP mới
	updateFields := map[string]interface{}{
		"status":    "online",
		"last_seen": time.Now(),
	}

	if payload.IPAddress != "" {
		updateFields["ip_address"] = payload.IPAddress
	}

	database.DB.Model(&agent).Updates(updateFields)
}
