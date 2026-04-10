package collector

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/StackExchange/wmi"
)

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

// 4. Đưa logic cũ vào hàm Collect()
func (s *FirewallSensor) Collect() (interface{}, error) {
	switch runtime.GOOS {
	case "windows":
		type MSFT_NetFirewallProfile struct {
			Enabled bool
		}
		var profiles []MSFT_NetFirewallProfile
		q := wmi.CreateQuery(&profiles, "")
		// Truy vấn WMI. Nếu lỗi (VD: service WMI tắt), mặc định là an toàn (không off) để tránh báo động giả.
		if err := wmi.Query(q, &profiles, nil, `ROOT\StandardCimv2`); err != nil {
			return FirewallRecord{FirewallOff: false}, nil
		}

		for _, profile := range profiles {
			if !profile.Enabled {
				// Nếu bất kỳ profile nào bị tắt, coi như firewall đang off.
				return FirewallRecord{FirewallOff: true}, nil
			}
		}
		// Tất cả profile đều đang bật.
		return FirewallRecord{FirewallOff: false}, nil

	case "linux":
		// Cảnh báo: Logic này chỉ check 'ufw'. Linux có thể dùng firewalld, iptables.
		cmd := exec.Command("ufw", "status")
		out, err := cmd.Output()
		if err == nil && strings.Contains(strings.ToLower(string(out)), "inactive") {
			return FirewallRecord{FirewallOff: true}, nil
		}
		return FirewallRecord{FirewallOff: false}, nil

	case "darwin":
		cmd := exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate")
		out, err := cmd.Output()
		if err == nil && strings.Contains(strings.ToLower(string(out)), "disabled") {
			return FirewallRecord{FirewallOff: true}, nil
		}
		return FirewallRecord{FirewallOff: false}, nil

	default:
		return FirewallRecord{FirewallOff: false}, nil // Hệ điều hành không hỗ trợ
	}
}
