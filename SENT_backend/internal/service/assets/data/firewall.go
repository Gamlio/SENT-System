package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
)

func ProcessFirewall(asset models.Asset, data interface{}) error {
	var record struct {
		FirewallOff bool `json:"firewall_off"`
	}
	bytes, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu firewall không hợp lệ: %w", err)
	}

	if err := json.Unmarshal(bytes, &record); err == nil && record.FirewallOff {
		incSvc := &incidents.IncidentService{}
		incSvc.TriggerSecurityEvent(context.TODO(), asset,
			"Firewall Disabled",
			"[P1] Tường lửa bị vô hiệu hóa",
			"Lớp phòng thủ OS Firewall đã bị tắt, nguy cơ bị tấn công mạng cao.",
			"P1",
		)
	}
	return nil
}
