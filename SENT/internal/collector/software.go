package collector

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"SENT/internal/policy"

	"github.com/shirou/gopsutil/v3/process"
)

// --- BỘ NHỚ ĐỆM TIẾN TRÌNH (TỐI ƯU HIỆU NĂNG) ---
var (
	cachedProcs   []*process.Process
	procsMutex    sync.Mutex
	lastProcFetch time.Time
)

// GetProcessesCached lấy danh sách tiến trình và lưu cache trong 2 giây để dùng chung cho cả batch
func GetProcessesCached() ([]*process.Process, error) {
	procsMutex.Lock()
	defer procsMutex.Unlock()

	if time.Since(lastProcFetch) < 2*time.Second && cachedProcs != nil {
		return cachedProcs, nil
	}

	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}
	cachedProcs = procs
	lastProcFetch = time.Now()
	return cachedProcs, nil
}

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

	procs, err := GetProcessesCached()
	if err != nil {
		log.Printf("Lỗi lấy danh sách tiến trình để kiểm tra chính sách: %v", err)
		return
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		procNameLower := strings.ToLower(name)
		cleanName := strings.TrimSuffix(procNameLower, ".exe")

		if blacklistMap[procNameLower] || blacklistMap[cleanName] {
			// Chỉ giám sát (Audit mode), không tiêu diệt tiến trình để tránh gián đoạn hệ thống
			log.Printf("⚠️ CẢNH BÁO: Phát hiện tiến trình '%s' (PID: %d) nằm trong danh sách cấm đang hoạt động.", name, p.Pid)
			// TODO: Gọi client.SendPayload(...) để báo cáo sự kiện này về Backend để SOC xử lý
		}
	}
}

// --- CÁC HÀM DÙNG CHUNG ---

func getRunningProcesses() map[string]bool {
	procMap := make(map[string]bool)
	procs, err := GetProcessesCached()
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
	// [BẢO MẬT] Đã gỡ bỏ rào cản bỏ qua file > 50MB để ngăn chặn tấn công chèn rác (Padding Bypass).
	_ = info

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
