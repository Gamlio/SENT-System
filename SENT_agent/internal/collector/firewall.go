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
func (s *FirewallSensor) Collect() interface{} {
	isOff := false

	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "netsh advfirewall show allprofiles state")
		out, _ := cmd.Output()
		isOff = strings.Contains(strings.ToLower(string(out)), "off")
	case "linux":
		cmd := exec.Command("ufw", "status")
		out, _ := cmd.Output()
		isOff = strings.Contains(strings.ToLower(string(out)), "inactive")
	case "darwin":
		cmd := exec.Command("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate")
		out, _ := cmd.Output()
		isOff = strings.Contains(strings.ToLower(string(out)), "disabled")
	}

	return FirewallRecord{
		FirewallOff: isOff,
	}
}
