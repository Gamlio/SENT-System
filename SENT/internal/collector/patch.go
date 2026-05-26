package collector

import "time"

// PatchSensor định nghĩa sensor thu thập bản vá hệ thống
type PatchSensor struct{}

// PatchRecord định nghĩa cấu trúc dữ liệu chuẩn gửi về Backend
type PatchRecord struct {
	Platform       string    `json:"platform"`
	MissingPatches []string  `json:"missing_patches"`
	TotalMissing   int       `json:"total_missing"`
	ScanTime       time.Time `json:"scan_time"`
}

func (s *PatchSensor) Name() string {
	return "patch"
}
