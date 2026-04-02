package collector

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

type SoftwareSensor struct{}

type SoftwareRecord struct {
	SoftwareName    string `json:"software_name"`
	Version         string `json:"version"`
	Publisher       string `json:"publisher"`
	InstallLocation string `json:"install_location"`
	FileHash        string `json:"file_hash"`
	Status          string `json:"status"`
	IsRunning       bool   `json:"is_running"`
}

func (s *SoftwareSensor) Name() string {
	return "software"
}

func (s *SoftwareSensor) Collect() (interface{}, error) {
	runningProcs := getRunningProcesses()
	// Hàm getOSSoftware sẽ tự động được Go gọi đúng file tùy theo lúc build
	return getOSSoftware(runningProcs) // Hàm này cần trả về (interface{}, error)
}

// --- CÁC HÀM DÙNG CHUNG ---

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

func isProcessRunning(softwareName string, runningProcs map[string]bool) bool {
	cleanSWName := strings.ToLower(softwareName)
	for procName := range runningProcs {
		if strings.Contains(cleanSWName, procName) || strings.Contains(procName, cleanSWName) {
			return true
		}
	}
	return false
}

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Lấy thông tin file để kiểm tra dung lượng
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	// Bỏ qua băm nếu file lớn hơn 50MB (50 * 1024 * 1024 bytes) để tránh nghẽn Disk I/O
	if info.Size() > 50*1024*1024 {
		return "", fmt.Errorf("file quá lớn (>50MB), bỏ qua để tối ưu hiệu suất")
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
