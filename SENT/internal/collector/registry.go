package collector

// Sensor là giao diện chuẩn cho mọi module thu thập
type Sensor interface {
	Name() string                  // Tên của loại Log (VD: "usb", "software", "patch")
	Collect() (interface{}, error) // Hàm chạy logic thu thập thực tế, trả về dữ liệu và lỗi
}

// Registry lưu trữ danh sách các Sensor đang hoạt động
var Registry []Sensor

// Register: Đăng ký một Sensor mới vào hệ thống
func Register(s Sensor) {
	Registry = append(Registry, s)
}

// InitCollectors khởi tạo và "cắm" toàn bộ các module vào bộ quét của Agent
func InitCollectors() {
	Register(&SoftwareSensor{})
	Register(&USBSensor{})
	Register(&PortSensor{})
	Register(&InventorySensor{})
	Register(&AntivirusSensor{})
	Register(&DataTransferSensor{})
	Register(&ETWSensor{
		Filter: EventFilter{
			EnableProcess:  true,
			EnableNetwork:  true,
			EnableRegistry: true,
			EnableFileIO:   true,
		},
	})

	// BỔ SUNG: Chốt cấu phần quản lý cấu hình và bản vá lỗi an ninh
	Register(&PatchSensor{})
}
