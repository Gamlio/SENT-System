//go:build linux

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
		Platform:       "linux",
		MissingPatches: []string{},
		ScanTime:       time.Now(),
	}

	var cmd *exec.Cmd
	// Kiểm tra xem hệ thống dùng apt (Debian/Ubuntu) hay yum (CentOS/RHEL) để quét bản vá an ninh ngầm
	if _, err := exec.LookPath("apt-get"); err == nil {
		cmd = exec.CommandContext(ctx, "sh", "-c", "apt-get -s upgrade | grep -i security")
	} else if _, err := exec.LookPath("yum"); err == nil {
		cmd = exec.CommandContext(ctx, "sh", "-c", "yum check-update --security | grep -i security")
	}

	if cmd != nil {
		out, _ := cmd.Output()
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.Contains(trimmed, "Inst") {
				record.MissingPatches = append(record.MissingPatches, trimmed)
			}
		}
	}

	record.TotalMissing = len(record.MissingPatches)
	return record, nil
}
