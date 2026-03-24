package agent_data

import (
	"encoding/json"
	"sent_backend/internal/models"
	"sent_backend/internal/service/security"
)

func ProcessFirewall(agent models.Agent, data interface{}) {
	var record struct {
		FirewallOff bool `json:"firewall_off"`
	}
	bytes, _ := json.Marshal(data)
	if err := json.Unmarshal(bytes, &record); err == nil && record.FirewallOff {
		security.TriggerSecurityEvent(agent, "Firewall Disabled", "[P1] Tường lửa bị vô hiệu hóa", "Nguy cơ tấn công mạng.", "P1")
	}
}
