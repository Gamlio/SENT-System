package data

import (
	"SENT_backend/pkg/models"
	"encoding/json"
	"fmt"
)

func ProcessFirewall(asset models.Asset, data interface{}) error {
	var record struct {
		FirewallOff bool `json:"firewall_off"`
	}
	bytesData, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu firewall không hợp lệ: %w", err)
	}

	if err := json.Unmarshal(bytesData, &record); err == nil && record.FirewallOff {
		SendBehaviorLog(map[string]interface{}{
			"asset":    asset,
			"category": "Firewall Disabled",
			"value":    "Disabled",
			"title":    "[P1] Tường lửa bị vô hiệu hóa",
			"desc":     "Lớp phòng thủ OS Firewall đã bị tắt, nguy cơ bị tấn công mạng cao.",
			"priority": "P1",
		})
	}
	return nil
}
