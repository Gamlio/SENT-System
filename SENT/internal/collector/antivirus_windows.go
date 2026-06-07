//go:build windows

package collector

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// Collect: Thu thập trạng thái mã độc / trình diệt virus on Windows
func (s *AntivirusSensor) Collect() (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// [TỐI ƯU HIỆU SUẤT] Sử dụng WMI Command-line (WMIC) nhanh hơn Powershell
	// Lấy trực tiếp tên mã độc (Name) thay vì chỉ lấy ID để hiển thị chi tiết cho SOC
	cmd := exec.CommandContext(ctx, "wmic", "/namespace:\\\\root\\Microsoft\\Windows\\Defender", "path", "MSFT_MpThreat", "where", "RollupStatus=1", "get", "Name")

	out, err := cmd.Output()
	if err != nil {
		return AntivirusRecord{HasThreat: false}, nil
	}

	lines := strings.Split(string(out), "\r\n")
	var names []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && trimmed != "Name" {
			names = append(names, trimmed)
		}
	}

	if len(names) > 0 {
		return AntivirusRecord{HasThreat: true, ThreatNames: names}, nil
	}
	return AntivirusRecord{HasThreat: false}, nil
}
