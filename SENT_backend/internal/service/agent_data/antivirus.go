package agent_data

import (
	"encoding/json"
	"sent_backend/internal/models"
	"sent_backend/internal/service/security"
)

func ProcessAntivirus(agent models.Agent, data interface{}) {
	var record struct {
		HasThreat bool `json:"has_threat"`
	}
	bytes, _ := json.Marshal(data)
	if err := json.Unmarshal(bytes, &record); err == nil && record.HasThreat {
		security.TriggerSecurityEvent(agent, "Malware Detected", "[P1] Antivirus báo động", "Trình diệt virus cục bộ phát hiện mã độc.", "P1")
	}
}
