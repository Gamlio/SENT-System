//go:build linux

package collector

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

func getOSSoftware(runningProcs map[string]bool) ([]SoftwareRecord, error) {
	// [FIX-WMI-HANG] Đặt timeout 15 giây cho các lệnh OS để tránh treo SENT
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var allSoftware []SoftwareRecord

	if _, err := exec.LookPath("/usr/bin/dpkg-query"); err == nil {
		allSoftware, _ = getLinuxSoftware(ctx, "/usr/bin/dpkg-query -W -f='${Package}|${Version}|${Maintainer}\n'", runningProcs)
	} else if _, err := exec.LookPath("/usr/bin/rpm"); err == nil {
		allSoftware, _ = getLinuxSoftware(ctx, "/usr/bin/rpm -qa --queryformat '%{NAME}|%{VERSION}|%{VENDOR}\n'", runningProcs)
	}

	// Bổ sung quét các ứng dụng Flatpak
	flatpakCmd := exec.CommandContext(ctx, "flatpak", "list", "--columns=name,version,application")
	if out, err := flatpakCmd.Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			parts := strings.Split(line, "\t") // Flatpak phân tách bằng tab
			if len(parts) >= 2 {
				allSoftware = append(allSoftware, SoftwareRecord{
					SoftwareName: strings.TrimSpace(parts[0]),
					Version:      strings.TrimSpace(parts[1]),
					Publisher:    "Flatpak",
					Status:       "INSTALLED",
					IsRunning:    isProcessRunning(parts[0], runningProcs),
				})
			}
		}
	}

	// Bổ sung: Bắt các App qua Snap/Flatpak hoặc Binary chạy trực tiếp
	procs, _ := process.Processes()
	processedNames := make(map[string]bool)
	for _, sw := range allSoftware {
		processedNames[strings.ToLower(sw.SoftwareName)] = true
	}

	for _, p := range procs {
		exePath, err := p.Exe()
		if err != nil || exePath == "" || strings.HasPrefix(exePath, "/usr/lib/") || strings.HasPrefix(exePath, "/sbin/") {
			continue
		}

		name, _ := p.Name()
		lowerName := strings.ToLower(name)
		if processedNames[lowerName] {
			continue
		}

		allSoftware = append(allSoftware, SoftwareRecord{
			SoftwareName:    name,
			Version:         "Unknown",
			Publisher:       "Running Binary",
			InstallLocation: filepath.Dir(exePath),
			Status:          "PORTABLE",
			IsRunning:       true,
		})
		processedNames[lowerName] = true
	}

	return allSoftware, nil
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
