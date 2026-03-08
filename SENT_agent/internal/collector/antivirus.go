package collector

import (
	"os/exec"
	"runtime"
	"strings"
)

// CollectAntivirus: Thu thập trạng thái mã độc / trình diệt virus
func CollectAntivirus() interface{} {
	hasThreat := false

	switch runtime.GOOS {
	case "windows":
		psCmd := `Get-MpThreat | Where-Object { $_.RollupStatus -eq 1 } | Select-Object -First 1`
		cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
		out, _ := cmd.Output()
		hasThreat = len(strings.TrimSpace(string(out))) > 0
	case "darwin":
		cmd := exec.Command("spctl", "--status")
		out, _ := cmd.Output()
		hasThreat = strings.Contains(strings.ToLower(string(out)), "disabled")
	case "linux":
		cmd := exec.Command("systemctl", "is-active", "clamav-daemon")
		out, _ := cmd.Output()
		hasThreat = strings.Contains(strings.ToLower(string(out)), "inactive")
	}

	return map[string]interface{}{
		"has_threat": hasThreat,
	}
}
