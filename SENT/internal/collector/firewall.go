package collector

// 1. Khai báo Struct cho Sensor
type FirewallSensor struct{}

// 2. Cấu trúc JSON chuẩn gửi về Backend
type FirewallRecord struct {
	FirewallOff bool `json:"firewall_off"`
}

// 3. Khai báo tên định danh của Log
func (s *FirewallSensor) Name() string {
	return "firewall"
}
