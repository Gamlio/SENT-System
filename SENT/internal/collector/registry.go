package collector

// Sensor là giao diện chuẩn cho mọi module thu thập
type Sensor interface {
	Name() string                  // Tên của loại Log (VD: "usb", "software", "telemetry")
	Collect() (interface{}, error) // Hàm chạy logic thu thập thực tế, trả về dữ liệu và lỗi (nếu có)
}

// Registry lưu trữ danh sách các Sensor đang hoạt động
var Registry []Sensor

// Register: Đăng ký một Sensor mới vào hệ thống
func Register(s Sensor) {
	Registry = append(Registry, s)
}

// Hàm khởi tạo: Nơi bạn "cắm" các module vào asset
func InitCollectors() {
	Register(&SoftwareSensor{})
	Register(&USBSensor{})
	Register(&PortSensor{})
	Register(&InventorySensor{})
	Register(&FirewallSensor{})
	Register(&AntivirusSensor{})
	Register(&DataTransferSensor{})
}
