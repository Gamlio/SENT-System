package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
	"golang.org/x/sys/windows/registry"
)

// Quản lý trạng thái dữ liệu cũ để so sánh sai khác
var (
	lastHashes = make(map[string]string)
	hashMutex  sync.Mutex
)

type SecurityLog struct {
	LogType        string      `json:"log_type"`
	EnrollToken    string      `json:"enroll_token"`
	HWID           string      `json:"hwid"`
	ScanTimestamp  int64       `json:"scan_timestamp"`
	IsDifferential bool        `json:"is_diff"` // Đánh dấu đây là bản gửi do có thay đổi
	Data           interface{} `json:"data"`
}

// Hàm băm dữ liệu để so sánh
func getHash(data interface{}) string {
	b, _ := json.Marshal(data)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

// Kiểm tra xem dữ liệu có thay đổi không
func hasChanged(logType string, newData interface{}) bool {
	hashMutex.Lock()
	defer hashMutex.Unlock()
	newHash := getHash(newData)
	if lastHashes[logType] == newHash {
		return false
	}
	lastHashes[logType] = newHash
	return true
}

// --- CÁC HÀM THU THẬP DỮ LIỆU (Giữ nguyên logic của bạn) ---
// Note: collectSoftware và collectEvents hiện tại chỉ chạy trên Windows

func collectInventory() interface{} {
	hInfo, _ := host.Info()
	cpuInfo, _ := cpu.Info()
	vMem, _ := mem.VirtualMemory()
	return map[string]interface{}{
		"os_info":      runtime.GOOS + " " + runtime.GOARCH,
		"cpu_model":    cpuInfo[0].ModelName,
		"ram_total_gb": vMem.Total / 1024 / 1024 / 1024,
		"kernel":       hInfo.KernelVersion,
	}
}

func collectSoftware() []string {
	if runtime.GOOS != "windows" {
		return []string{"Not implemented for " + runtime.GOOS}
	}
	var softwareList []string
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		return []string{"Error Accessing Registry"}
	}
	defer k.Close()
	names, _ := k.ReadSubKeyNames(-1)
	for _, name := range names {
		sk, _ := registry.OpenKey(k, name, registry.QUERY_VALUE)
		displayName, _, _ := sk.GetStringValue("DisplayName")
		if displayName != "" {
			softwareList = append(softwareList, displayName)
		}
		sk.Close()
	}
	return softwareList
}

func collectTelemetry() interface{} {
	vMem, _ := mem.VirtualMemory()
	connections, _ := psnet.Connections("tcp")
	// Logic lấy port... (lược bỏ cho gọn)
	return map[string]interface{}{
		"ram_used_percent": vMem.UsedPercent,
		"open_ports_count": len(connections),
	}
}

func send(logType string, token string, hwid string, payload interface{}, force bool) {
	// Chỉ gửi nếu dữ liệu thay đổi HOẶC là lần chạy đầu tiên (force=true)
	if !force && !hasChanged(logType, payload) {
		// fmt.Printf(">>> [%s] Không có thay đổi, bỏ qua gửi.\n", logType)
		return
	}

	entry := SecurityLog{
		LogType: logType, EnrollToken: token, HWID: hwid,
		ScanTimestamp: time.Now().Unix(), Data: payload,
		IsDifferential: !force,
	}
	out, _ := json.MarshalIndent(entry, "", "  ")
	fmt.Printf("\n>>> SENDING [%s] AT %s\n%s\n", logType, time.Now().Format("15:04:05"), string(out))
}

func main() {
	hInfo, _ := host.Info()
	hwid := hInfo.HostID
	token := "SENT-TOKEN-SME-01"

	fmt.Println("🛡️ SENT Agent v3.1: Chế độ Differential Reporting đã kích hoạt.")

	// Khởi tạo Tickers
	telemetryTicker := time.NewTicker(1 * time.Minute)
	softwareTicker := time.NewTicker(6 * time.Hour)
	inventoryTicker := time.NewTicker(24 * time.Hour)

	// CƠ CHẾ KHỞI ĐỘNG: Gửi toàn bộ dữ liệu lần đầu tiên (force = true)
	fmt.Println("📢 Lần đầu khởi động: Đang đẩy toàn bộ snapshot lên Server...")
	send("inventory", token, hwid, collectInventory(), true)
	send("software", token, hwid, collectSoftware(), true)
	send("telemetry", token, hwid, collectTelemetry(), true)

	for {
		select {
		case <-telemetryTicker.C:
			send("telemetry", token, hwid, collectTelemetry(), false)
		case <-softwareTicker.C:
			send("software", token, hwid, collectSoftware(), false)
		case <-inventoryTicker.C:
			send("inventory", token, hwid, collectInventory(), false)
		}
	}
}
