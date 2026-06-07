package data

import (
	"SENT_backend/pkg/models"
	"encoding/json"
	"fmt"
	"strings"
)

func ProcessAntivirus(asset models.Asset, data interface{}) error {
	var record struct {
		HasThreat   bool     `json:"has_threat"`
		ThreatNames []string `json:"threat_names"`
	}

	bytesData, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu antivirus không hợp lệ: %w", err)
	}

	if err := json.Unmarshal(bytesData, &record); err == nil && record.HasThreat {
		// Tổng hợp danh sách mã độc để đưa vào chi tiết cảnh báo
		threatList := "Unknown Malware"
		if len(record.ThreatNames) > 0 {
			threatList = strings.Join(record.ThreatNames, ", ")
		}

		// Nâng cấp Priority lên P1 (Chí mạng) theo đúng SCORING.md cho mã độc
		SendBehaviorLog("Malware Detected",
			threatList,
			"[P1] Phát hiện mã độc đang hoạt động",
			fmt.Sprintf("Hệ thống phòng thủ trên máy trạm phát hiện các biến thể mã độc: %s. Cần thực hiện quy trình ứng cứu sự cố ngay lập tức.", threatList),
			"P1", asset)
	}
	return nil
}
