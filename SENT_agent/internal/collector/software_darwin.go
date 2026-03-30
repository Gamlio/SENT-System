//go:build darwin

package collector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getOSSoftware(runningProcs map[string]bool) ([]SoftwareRecord, error) {
	var softwareList []SoftwareRecord
	apps, err := os.ReadDir("/Applications")
	if err != nil {
		return nil, fmt.Errorf("không thể đọc thư mục /Applications: %w", err)
	} else {
		for _, app := range apps {
			if strings.HasSuffix(app.Name(), ".app") {
				appName := strings.TrimSuffix(app.Name(), ".app")
				isRunning := isProcessRunning(appName, runningProcs)

				softwareList = append(softwareList, SoftwareRecord{
					SoftwareName:    appName,
					InstallLocation: filepath.Join("/Applications", app.Name()),
					Status:          "INSTALLED",
					IsRunning:       isRunning,
				})
			}
		}
	}
	return softwareList, nil
}
