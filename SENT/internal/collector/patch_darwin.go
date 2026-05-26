//go:build darwin

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
		Platform:       "darwin",
		MissingPatches: []string{},
		ScanTime:       time.Now(),
	}

	// Sử dụng lệnh kiểm toán bản cập nhật phần mềm chuẩn của hệ điều hành Apple
	cmd := exec.CommandContext(ctx, "softwareupdate", "-l")
	out, err := cmd.Output()
	if err == nil && strings.Contains(string(out), "*") {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "*") {
				record.MissingPatches = append(record.MissingPatches, strings.TrimSpace(line))
			}
		}
	}

	record.TotalMissing = len(record.MissingPatches)
	return record, nil
}
