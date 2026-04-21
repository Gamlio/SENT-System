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
		cmd := exec.CommandContext(ctx, "/usr/sbin/spctl", "--status")
		out, err := cmd.Output()
		// Gatekeeper bị tắt là một rủi ro bảo mật.
		if err == nil && strings.Contains(strings.ToLower(string(out)), "disabled") {
			return AntivirusRecord{HasThreat: true}, nil
		}
		return AntivirusRecord{HasThreat: false}, nil

	case "linux":
		// Dịch vụ clamav-daemon không chạy KHÔNG CÓ NGHĨA là máy đang bị nhiễm mã độc.
		// Bỏ logic báo False Positive này để tránh spam rác về SOC.
		return AntivirusRecord{HasThreat: false}, nil
	}
	return AntivirusRecord{HasThreat: false}, nil // Hệ điều hành không hỗ trợ
}
