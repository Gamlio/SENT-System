package collector

import (
	"os/exec"
	"runtime"
	"strings"
)

// CollectFirewall: Thu thập trạng thái tường lửa đa nền tảng
func CollectFirewall() interface{} {
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

	return map[string]interface{}{
		"firewall_off": isOff,
	}
}
