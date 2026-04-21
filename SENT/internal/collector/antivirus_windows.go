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
	cmd := exec.CommandContext(ctx, "wmic", "/namespace:\\\\root\\Microsoft\\Windows\\Defender", "path", "MSFT_MpThreat", "where", "RollupStatus=1", "get", "ThreatID")

	out, err := cmd.Output()
	// Nếu output có chứa "ThreatID", tức là Defender đang báo cáo mã độc (RollupStatus=1)
	if err == nil && strings.Contains(string(out), "ThreatID") {
		return AntivirusRecord{HasThreat: true}, nil
	}
	return AntivirusRecord{HasThreat: false}, nil
}
