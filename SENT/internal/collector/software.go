package collector

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	// [FIX-TIMEOUT] Đặt timeout 30 giây để tránh treo như Linux version (15s)
	// Windows có nhiều file hơn nên cần timeout dài hơn
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Channel để lấy kết quả từ goroutine
	resultChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	// Chạy collection trong goroutine để có thể cancel nếu timeout
	go func() {
		runningProcs := getRunningProcesses()
		// Hàm getOSSoftware sẽ tự động được Go gọi đúng file tùy theo lúc build
		softwareList, err := getOSSoftware(runningProcs)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- softwareList
	}()

	// Chờ kết quả hoặc timeout
	select {
	case <-ctx.Done():
		// Timeout xảy ra
		log.Printf("⚠️ Software collection timeout sau 30 giây")
		return nil, fmt.Errorf("software collection timeout")
	case err := <-errChan:
		return nil, err
	case result := <-resultChan:
		softwareList := result.([]SoftwareRecord)
		warnSoftwarePolicy(softwareList)
		return softwareList, nil
	}
}

func warnSoftwarePolicy(records []SoftwareRecord) {
	seen := make(map[string]bool)
	for _, rec := range records {
		if rec.FileHash != "" {
			if violated, reason := policy.CheckPolicy("SOFTWARE_HASH", rec.FileHash); violated {
				key := "hash:" + rec.FileHash
				if !seen[key] {
					log.Printf("⚠️ Phần mềm %q với hash %s không hợp lệ: %s", rec.SoftwareName, rec.FileHash, reason)
					seen[key] = true
				}
			}
		}

		if rec.Publisher != "" {
			if violated, reason := policy.CheckPolicy("PUBLISHER", rec.Publisher); violated {
				key := "publisher:" + strings.ToLower(rec.Publisher)
				if !seen[key] {
					log.Printf("⚠️ Phần mềm %q do nhà phát hành %q không hợp lệ: %s", rec.SoftwareName, rec.Publisher, reason)
					seen[key] = true
				}
			}
		}

		if rec.SoftwareName != "" {
			if violated, reason := policy.CheckPolicy("SOFTWARE", rec.SoftwareName); violated {
				key := "software:" + strings.ToLower(rec.SoftwareName)
				if !seen[key] {
					log.Printf("⚠️ Phần mềm %q không hợp lệ theo chính sách: %s", rec.SoftwareName, reason)
					seen[key] = true
				}
			}
		}
	}
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

	// [FIX-PERFORMANCE] Bỏ qua file > 100MB để tránh treo hệ thống
	// Các file lớn thường là dữ liệu hoặc video, không phải executable quan trọng
	const maxHashSize = 100 * 1024 * 1024 // 100MB
	if info.Size() > maxHashSize {
		return "", fmt.Errorf("file quá lớn (%d bytes), bỏ qua hashing", info.Size())
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
