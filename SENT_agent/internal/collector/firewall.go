package collector

import (
	"os/exec"
	"runtime"
	"strings"
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
	isOff := false
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "netsh advfirewall show allprofiles state")
	case "linux":
		// Cảnh báo: Logic này chỉ check 'ufw'. Linux có thể dùng firewalld, iptables.
		cmd = exec.Command("ufw", "status")
	case "darwin":
		cmd = exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate")
	default:
		return FirewallRecord{FirewallOff: false}, nil // Hệ điều hành không hỗ trợ
	}

	out, err := cmd.Output()
	// Tương tự Antivirus, lỗi command không tồn tại không nên làm dừng agent.
	// Sẽ log trong tương lai, hiện tại chỉ kiểm tra output nếu không có lỗi.
	if err == nil {
		output := strings.ToLower(string(out))
		isOff = strings.Contains(output, "off") || strings.Contains(output, "inactive") || strings.Contains(output, "disabled")
	}

	return FirewallRecord{FirewallOff: isOff}, nil
}
