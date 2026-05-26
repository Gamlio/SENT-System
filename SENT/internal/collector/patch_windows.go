//go:build windows

package collector

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

func (s *PatchSensor) Collect() (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	record := PatchRecord{
		Platform:       "windows",
		MissingPatches: []string{},
		ScanTime:       time.Now(),
	}

	// Gọi Powershell để kiểm tra danh sách bản vá đã cài đặt độc lập trên Windows
	cmd := exec.CommandContext(ctx, "powershell", "-Command", "Get-HotFix | Select-Object -ExpandProperty HotFixID")
	out, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(out), "\r\n")
		installedPatches := make(map[string]bool)
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				installedPatches[trimmed] = true
			}
		}

		// Baseline đối chiếu mẫu để phát hiện thiếu bản vá an ninh tầng sâu
		if !installedPatches["KB5034765"] {
			record.MissingPatches = append(record.MissingPatches, "KB5034765 (Security Update for Windows 11)")
		}
		if !installedPatches["KB5034122"] {
			record.MissingPatches = append(record.MissingPatches, "KB5034122 (Cumulative Kernel Vulnerability Patch)")
		}
	}

	record.TotalMissing = len(record.MissingPatches)
	return record, nil
}
