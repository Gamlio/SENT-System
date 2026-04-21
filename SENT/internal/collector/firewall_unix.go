//go:build !windows

package collector

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Collect: Thu thập trạng thái firewall on non-Windows systems
func (s *FirewallSensor) Collect() (interface{}, error) {
	// [FIX-WMI-HANG] Đặt timeout 5 giây cho các lệnh OS
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch runtime.GOOS {
	case "linux":
		cmd := exec.CommandContext(ctx, "/usr/sbin/ufw", "status")
		out, err := cmd.Output()
		if err == nil && strings.Contains(strings.ToLower(string(out)), "inactive") {
			return FirewallRecord{FirewallOff: true}, nil
		}
		return FirewallRecord{FirewallOff: false}, nil

	case "darwin":
		cmd := exec.CommandContext(ctx, "/usr/libexec/ApplicationFirewall/socketfilterfw", "--getglobalstate")
		out, err := cmd.Output()
		if err == nil && strings.Contains(strings.ToLower(string(out)), "disabled") {
			return FirewallRecord{FirewallOff: true}, nil
		}
		return FirewallRecord{FirewallOff: false}, nil
	}
	return FirewallRecord{FirewallOff: false}, nil // Hệ điều hành không hỗ trợ
}
