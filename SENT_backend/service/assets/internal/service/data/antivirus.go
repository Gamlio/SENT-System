package data

import (
	"SENT_backend/pkg/models"
	"encoding/json"
	"fmt"
)

func ProcessAntivirus(asset models.Asset, data interface{}) error {
	var record struct {
		HasThreat bool `json:"has_threat"`
	}

	bytesData, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu antivirus không hợp lệ: %w", err)
	}

	if err := json.Unmarshal(bytesData, &record); err == nil && record.HasThreat {
		// 1. Chuẩn bị Payload khớp với SecurityEventReq của Incident Service
		SendBehaviorLog("Malware",
			"Infection Detected",
			"[P3] Malware Detected",
			"Phát hiện mã độc trên máy trạm",
			"P3", asset)
	}
	return nil
}
