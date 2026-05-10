package data

import (
	"SENT_backend/pkg/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
		payload := map[string]interface{}{
			"asset":       asset,
			"alert_type":  "Malware Detected",
			"title":       "[P1] Antivirus báo động",
			"description": "Trình diệt virus cục bộ trên máy trạm đã phát hiện mã độc.",
			"priority":    "P1",
		}

		// 2. GỌI API SANG INCIDENT SERVICE (Bất đồng bộ qua Goroutine)
		// Sử dụng DNS của Docker: incident-service trên cổng 8005
		go func() {
			jsonData, _ := json.Marshal(payload)
			url := "http://incident-service:8005/api/v1/incidents/trigger"

			resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("⚠️ Lỗi gửi sự kiện Antivirus sang Incident Service: %v\n", err)
				return
			}
			defer resp.Body.Close()
		}()
	}
	return nil
}
