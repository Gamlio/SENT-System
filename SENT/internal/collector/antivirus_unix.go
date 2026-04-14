//go:build !windows

package collector

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Collect: Thu thập trạng thái mã độc / trình diệt virus on non-Windows systems
func (s *AntivirusSensor) Collect() (interface{}, error) {
	// [FIX-WMI-HANG] Đặt timeout 5 giây cho các lệnh OS
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.CommandContext(ctx, "spctl", "--status")
		out, err := cmd.Output()
		// Gatekeeper bị tắt là một rủi ro bảo mật.
		if err == nil && strings.Contains(strings.ToLower(string(out)), "disabled") {
			return AntivirusRecord{HasThreat: true}, nil
		}
		return AntivirusRecord{HasThreat: false}, nil

	case "linux":
		cmd := exec.CommandContext(ctx, "systemctl", "is-active", "clamav-daemon")
		out, err := cmd.Output()
		if err != nil || !strings.Contains(string(out), "active") {
			return AntivirusRecord{HasThreat: true}, nil
		}
		return AntivirusRecord{HasThreat: false}, nil
	}
	return AntivirusRecord{HasThreat: false}, nil // Hệ điều hành không hỗ trợ
}
