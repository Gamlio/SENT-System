package collector

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"SENT/internal/policy"

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

// EnforceSoftwarePolicy kiểm tra các tiến trình đang chạy với blacklist và chấm dứt chúng.
// Hàm này nên được gọi trong một goroutine riêng, chạy thường xuyên (ví dụ: mỗi 10 giây).
func EnforceSoftwarePolicy() {
	blacklist := policy.GetBlacklistedSoftware()
	if len(blacklist) == 0 {
		return // Không có chính sách nào để thực thi
	}

	// Tạo một map để tra cứu nhanh hơn
	blacklistMap := make(map[string]bool)
	for _, name := range blacklist {
		blacklistMap[name] = true
	}

	procs, err := process.Processes()
	if err != nil {
		log.Printf("Lỗi lấy danh sách tiến trình để kiểm tra chính sách: %v", err)
		return
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		// Kiểm tra xem tên tiến trình có trong blacklist không
		// So sánh cả tên gốc (ví dụ: utorrent.exe) và tên đã loại bỏ .exe (ví dụ: utorrent)
		procNameLower := strings.ToLower(name)
		if blacklistMap[procNameLower] || blacklistMap[strings.TrimSuffix(procNameLower, ".exe")] {
			// Chỉ giám sát (Audit mode), không tiêu diệt tiến trình để tránh gián đoạn hệ thống
			log.Printf("⚠️ CẢNH BÁO: Phát hiện tiến trình '%s' (PID: %d) nằm trong danh sách cấm đang hoạt động.", name, p.Pid)
			// Ghi chú: Có thể mở rộng để gọi API gửi alert P1 về backend tại đây.
		}
	}
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
