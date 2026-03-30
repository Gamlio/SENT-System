//go:build windows

package collector

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func getOSSoftware(runningProcs map[string]bool) ([]SoftwareRecord, error) {
	var softwareList []SoftwareRecord
	paths := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, path := range paths {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
		if err != nil {
			// Nếu không mở được key registry chính, có thể bỏ qua (ví dụ WOW6432Node không tồn tại trên HĐH 32-bit)
			continue
		}
		defer k.Close()

		names, err := k.ReadSubKeyNames(-1)
		if err != nil {
			log.Printf("Lỗi đọc subkey từ '%s': %v", path, err) // Ghi log và tiếp tục
			continue
		}

		for _, name := range names {
			sk, err := registry.OpenKey(k, name, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			displayName, _, _ := sk.GetStringValue("DisplayName")
			displayVersion, _, _ := sk.GetStringValue("DisplayVersion")
			publisher, _, _ := sk.GetStringValue("Publisher")
			installLocation, _, _ := sk.GetStringValue("InstallLocation")
			displayIcon, _, _ := sk.GetStringValue("DisplayIcon")

			if displayName != "" {
				status := "INSTALLED"
				fileHash := ""

				if installLocation != "" {
					cleanPath := strings.Trim(installLocation, "\"")
					if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
						status = "GHOST_REGISTRY"
					}
				}

				if displayIcon != "" {
					exePath := strings.Split(displayIcon, ",")[0]
					exePath = strings.Trim(exePath, "\"")
					if strings.HasSuffix(strings.ToLower(exePath), ".exe") {
						hash, err := calculateSHA256(exePath)
						if err == nil {
							fileHash = hash
						}
						realPublisher := getDigitalSignature(exePath)
						if realPublisher != "Unsigned" {
							publisher = realPublisher
						}
					}
				}

				isRunning := isProcessRunning(displayName, runningProcs)

				softwareList = append(softwareList, SoftwareRecord{
					SoftwareName:    displayName,
					Version:         displayVersion,
					Publisher:       publisher,
					InstallLocation: installLocation,
					FileHash:        fileHash,
					Status:          status,
					IsRunning:       isRunning,
				})
			}
			sk.Close()
		}
	}
	return softwareList, nil
}

func getDigitalSignature(exePath string) string {
	if exePath == "" {
		return "Unsigned"
	}
	cleanPath := strings.Trim(exePath, "\"")
	psCmd := fmt.Sprintf(`(Get-AuthenticodeSignature "%s").SignerCertificate.Subject`, cleanPath)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	out, err := cmd.Output()

	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		return "Unsigned"
	}
	parts := strings.Split(string(out), "CN=")
	if len(parts) > 1 {
		return strings.Split(parts[1], ",")[0]
	}
	return strings.TrimSpace(string(out))
}
