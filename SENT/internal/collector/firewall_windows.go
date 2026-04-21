//go:build windows

package collector

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// Collect: Thu thập trạng thái firewall on Windows
func (s *FirewallSensor) Collect() (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// [TỐI ƯU HIỆU SUẤT] Sử dụng 'netsh' thay cho PowerShell để tránh khởi tạo Engine nặng
	cmd := exec.CommandContext(ctx, "netsh", "advfirewall", "show", "allprofiles", "state")

	out, err := cmd.Output()
	outputStr := strings.ToLower(string(out))
	// Đếm số lượng profile đang có State = ON (Domain, Private, Public)
	if err == nil && strings.Count(outputStr, "on") < 3 {
		return FirewallRecord{FirewallOff: true}, nil
	}
	return FirewallRecord{FirewallOff: false}, nil
}
