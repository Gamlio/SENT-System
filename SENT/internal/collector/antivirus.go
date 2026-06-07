package collector

// 1. Khai báo Struct cho Sensor
type AntivirusSensor struct{}

// 2. Cấu trúc JSON chuẩn gửi về Backend
type AntivirusRecord struct {
	HasThreat   bool     `json:"has_threat"`
	ThreatNames []string `json:"threat_names"`
}

// 3. Khai báo tên định danh của Log
func (s *AntivirusSensor) Name() string {
	return "antivirus"
}
