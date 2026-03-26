package collector

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows/registry"
)

// 1. Khai báo Struct cho Sensor
type SoftwareSensor struct{}

// 2. Cấu trúc JSON chuẩn gửi về Backend
type SoftwareRecord struct {
	SoftwareName    string `json:"software_name"`
	Version         string `json:"version"`
	Publisher       string `json:"publisher"`
	InstallLocation string `json:"install_location"`
	FileHash        string `json:"file_hash"`  // Dùng để Backend so khớp YARA/Threat Intel
	Status          string `json:"status"`     // INSTALLED, GHOST_REGISTRY
	IsRunning       bool   `json:"is_running"` // Check xem phần mềm có đang mở không
}

// 3. Khai báo tên định danh của Log
func (s *SoftwareSensor) Name() string {
	return "software"
}

// 4. Đưa logic cũ vào hàm Collect()
func (s *SoftwareSensor) Collect() interface{} {
	var softwareList []SoftwareRecord
	runningProcs := getRunningProcesses() // Lấy danh sách tiến trình đang chạy

	switch runtime.GOOS {
	case "windows":
		softwareList = s.getWindowsSoftware(runningProcs)
	case "linux":
		// Kiểm tra hệ tư tưởng gói: Debian hay RedHat
		if _, err := exec.LookPath("dpkg-query"); err == nil {
			softwareList = s.getLinuxSoftware("dpkg-query -W -f='${Package}|${Version}|${Maintainer}\n'", runningProcs)
		} else if _, err := exec.LookPath("rpm"); err == nil {
			softwareList = s.getLinuxSoftware("rpm -qa --queryformat '%{NAME}|%{VERSION}|%{VENDOR}\n'", runningProcs)
		}
	case "darwin":
		softwareList = s.getMacOSSoftware(runningProcs)
	}

	return softwareList
}

// ==========================================
// CÁC HÀM PHỤ TRỢ (NINJA HELPERS)
// ==========================================

func (s *SoftwareSensor) getWindowsSoftware(runningProcs map[string]bool) []SoftwareRecord {
	var softwareList []SoftwareRecord
	paths := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, path := range paths {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		names, _ := k.ReadSubKeyNames(-1)
		for _, name := range names {
			sk, _ := registry.OpenKey(k, name, registry.QUERY_VALUE)

			displayName, _, _ := sk.GetStringValue("DisplayName")
			displayVersion, _, _ := sk.GetStringValue("DisplayVersion")
			publisher, _, _ := sk.GetStringValue("Publisher")
			installLocation, _, _ := sk.GetStringValue("InstallLocation")
			displayIcon, _, _ := sk.GetStringValue("DisplayIcon") // Thường chứa đường dẫn đến file .exe chính

			if displayName != "" {
				status := "INSTALLED"
				fileHash := ""

				// LỚP 2: CHECK CHÉO Ổ CỨNG (Ghost Registry)
				if installLocation != "" {
					cleanPath := strings.Trim(installLocation, "\"")
					if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
						status = "GHOST_REGISTRY"
					}
				}

				// BĂM FILE (HASHING) - Tìm file .exe chính từ DisplayIcon
				if displayIcon != "" {
					exePath := strings.Split(displayIcon, ",")[0]
					exePath = strings.Trim(exePath, "\"")
					if strings.HasSuffix(strings.ToLower(exePath), ".exe") {
						fileHash = calculateSHA256(exePath)
						// THAY THẾ: Lấy publisher từ file thay vì Registry
						realPublisher := getDigitalSignature(exePath)
						if realPublisher != "Unsigned" {
							publisher = realPublisher
						}
					}
				}

				// LỚP 3: CHECK CHÉO RAM (Đang chạy hay không?)
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
		k.Close()
	}
	return softwareList
}

func (s *SoftwareSensor) getLinuxSoftware(cmdStr string, runningProcs map[string]bool) []SoftwareRecord {
	var list []SoftwareRecord
	parts := strings.Fields(cmdStr)
	cmd := exec.Command(parts[0], parts[1:]...)
	out, _ := cmd.Output()

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
	return list
}

func (s *SoftwareSensor) getMacOSSoftware(runningProcs map[string]bool) []SoftwareRecord {
	var softwareList []SoftwareRecord
	// Quét nhanh thư mục /Applications
	apps, err := os.ReadDir("/Applications")
	if err == nil {
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
	return softwareList
}

// getRunningProcesses: Lấy danh sách tiến trình đa nền tảng (Dùng gopsutil)
func getRunningProcesses() map[string]bool {
	procMap := make(map[string]bool)
	procs, err := process.Processes()
	if err != nil {
		return procMap
	}

	for _, p := range procs {
		name, err := p.Name()
		if err == nil {
			cleanName := strings.ToLower(strings.TrimSuffix(name, ".exe"))
			procMap[cleanName] = true
		}
	}
	return procMap
}

// isProcessRunning: So khớp tên phần mềm với danh sách tiến trình
func isProcessRunning(softwareName string, runningProcs map[string]bool) bool {
	cleanSWName := strings.ToLower(softwareName)
	// Tìm kiếm tương đối (Ví dụ: "Google Chrome" sẽ match với tiến trình "chrome")
	for procName := range runningProcs {
		if strings.Contains(cleanSWName, procName) || strings.Contains(procName, cleanSWName) {
			return true
		}
	}
	return false
}

// calculateSHA256: Hàm băm file siêu nhẹ, không load toàn bộ file vào RAM
func calculateSHA256(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(nil))
}
func getDigitalSignature(exePath string) string {
	if exePath == "" {
		return "Unsigned"
	}
	cleanPath := strings.Trim(exePath, "\"")

	// Dùng PowerShell để trích xuất thông tin người ký (SignerCertificate)
	psCmd := fmt.Sprintf(`(Get-AuthenticodeSignature "%s").SignerCertificate.Subject`, cleanPath)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	out, err := cmd.Output()

	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		return "Unsigned"
	}
	// Kết quả thường có dạng: CN=Microsoft Corporation... -> Cần bóc lấy tên
	parts := strings.Split(string(out), "CN=")
	if len(parts) > 1 {
		return strings.Split(parts[1], ",")[0]
	}
	return strings.TrimSpace(string(out))
}
