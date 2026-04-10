package data // Đổi sang package data

import (
	"context"
	"encoding/json"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents" // Import service sự cố mới
)

func ProcessAntivirus(asset models.Asset, data interface{}) {
	var record struct {
		HasThreat bool `json:"has_threat"`
	}
	bytes, _ := json.Marshal(data)

	if err := json.Unmarshal(bytes, &record); err == nil && record.HasThreat {
		// Khởi tạo service sự cố để xử lý gom nhóm tự động
		incSvc := &incidents.IncidentService{}
		incSvc.TriggerSecurityEvent(context.TODO(), asset,
			"Malware Detected",
			"[P1] Antivirus báo động",
			"Trình diệt virus cục bộ trên máy trạm đã phát hiện mã độc.",
			"P1",
		)
	}
}
