package data

import (
	"SENT_backend/pkg/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

		go func() {
			payload := map[string]interface{}{
				"asset":       asset,
				"alert_type":  "Firewall Disabled",
				"title":       "[P1] Tường lửa bị vô hiệu hóa",
				"description": "Lớp phòng thủ OS Firewall đã bị tắt, nguy cơ bị tấn công mạng cao.",
				"priority":    "P1",
			}
			jsonData, _ := json.Marshal(payload)
			http.Post("http://incident-service:8005/api/v1/incidents/trigger", "application/json", bytes.NewBuffer(jsonData))
		}()
	}
	return nil
}
