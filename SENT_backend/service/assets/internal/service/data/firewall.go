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
		// Thay thế hàm cũ bằng cấu trúc tường minh:
		SendBehaviorLog("Firewall Disabled",
			"Disabled",
			"[P3] Tường lửa bị vô hiệu hóa",
			"Lớp phòng thủ OS Firewall đã bị tắt, nguy cơ bị tấn công mạng cao.",
			"P3", asset)
	}
	return nil
}
