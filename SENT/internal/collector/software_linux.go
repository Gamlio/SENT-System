//go:build linux

package collector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func getOSSoftware(runningProcs map[string]bool) ([]SoftwareRecord, error) {
	// [FIX-WMI-HANG] Đặt timeout 15 giây cho các lệnh OS để tránh treo Agent
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if _, err := exec.LookPath("dpkg-query"); err == nil {
		return getLinuxSoftware(ctx, "dpkg-query -W -f='${Package}|${Version}|${Maintainer}\n'", runningProcs)
	} else if _, err := exec.LookPath("rpm"); err == nil {
		return getLinuxSoftware(ctx, "rpm -qa --queryformat '%{NAME}|%{VERSION}|%{VENDOR}\n'", runningProcs)
	}
	return []SoftwareRecord{}, nil // Không có package manager nào được tìm thấy, trả về rỗng, không phải lỗi
}

func getLinuxSoftware(ctx context.Context, cmdStr string, runningProcs map[string]bool) ([]SoftwareRecord, error) {
	var list []SoftwareRecord
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return nil, fmt.Errorf("chuỗi lệnh rỗng")
	}
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	out, err := cmd.Output()
	if err != nil {
		// Lỗi thực thi lệnh của package manager là một lỗi cần báo cáo
		return nil, fmt.Errorf("lỗi thực thi '%s': %w", parts[0], err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		data := strings.Split(line, "|")
		if len(data) >= 3 {
			isRunning := isProcessRunning(data[0], runningProcs)
			list = append(list, SoftwareRecord{
				SoftwareName: data[0],
				Version:      data[1],
				Publisher:    data[2],
				Status:       "INSTALLED",
				IsRunning:    isRunning,
			})
		}
	}
	return list, nil
}
