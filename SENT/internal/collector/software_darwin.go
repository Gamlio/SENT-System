//go:build darwin

package collector

import (
	"os"
	"path/filepath"
	"strings"
)

func getOSSoftware(runningProcs map[string]bool) ([]SoftwareRecord, error) {
	var softwareList []SoftwareRecord

	// [MỞ RỘNG] Quét cả thư mục gốc và các binary từ Homebrew (macOS)
	searchPaths := []string{
		"/Applications",
		"/opt/homebrew/bin", // Apple Silicon brew
		"/usr/local/bin",    // Intel brew
	}

	for _, dirPath := range searchPaths {
		apps, err := os.ReadDir(dirPath)
		if err == nil {
			for _, app := range apps {
				appName := app.Name()
				// Ứng dụng GUI (.app) hoặc Executable nhị phân
				if strings.HasSuffix(appName, ".app") || !app.IsDir() {
					cleanName := strings.TrimSuffix(appName, ".app")
					isRunning := isProcessRunning(cleanName, runningProcs)

					publisher := "Unknown"
					if strings.Contains(dirPath, "homebrew") || strings.Contains(dirPath, "local") {
						publisher = "Homebrew/CLI"
					}

					softwareList = append(softwareList, SoftwareRecord{
						SoftwareName:    cleanName,
						Publisher:       publisher,
						InstallLocation: filepath.Join(dirPath, appName),
						Status:          "INSTALLED",
						IsRunning:       isRunning,
					})
				}
			}
		}
	}
	return softwareList, nil
}
