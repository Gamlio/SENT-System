package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
	"golang.org/x/sys/windows/registry"
)

// --- CẤU HÌNH ---
const (
	SERVER_URL   = "http://localhost:8000/api/v1/agents/push"
	ENROLL_TOKEN = "SENT-TOKEN-SME-01"
)

var (
	lastHashes = make(map[string]string)
	hashMutex  sync.Mutex
)

type Payload struct {
	Type     string      `json:"type"`     // "DATA" hoặc "HEARTBEAT"
	LogType  string      `json:"log_type"` // "telemetry", "software", "inventory"
	HWID     string      `json:"hwid"`
	Hostname string      `json:"hostname"`
	Data     interface{} `json:"data"` // Dữ liệu thật (nếu có)
}

// Hàm tính Hash
func calculateHash(data interface{}) string {
	b, _ := json.Marshal(data)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

// Hàm gửi dữ liệu
func sendPayload(hwid, hostname, logType string, data interface{}) {
	hashMutex.Lock()
	currentHash := calculateHash(data)
	oldHash := lastHashes[logType]

	packetType := "HEARTBEAT"
	var dataToSend interface{} = nil

	// LOGIC SÀNG LỌC CỦA AGENT:
	if currentHash != oldHash {
		// 1. Nếu dữ liệu thay đổi -> Gửi DATA thật
		packetType = "DATA"
		dataToSend = data
		lastHashes[logType] = currentHash // Cập nhật bộ nhớ
		fmt.Printf("⚡ [%s] Có thay đổi -> Gửi dữ liệu mới.\n", logType)
	} else {
		// 2. Nếu y nguyên -> Chỉ gửi Heartbeat để báo Online
		// fmt.Printf("💤 [%s] Không đổi -> Gửi Heartbeat.\n", logType)
	}
	hashMutex.Unlock()

	// Đóng gói
	payload := Payload{
		Type:     packetType,
		LogType:  logType,
		HWID:     hwid,
		Hostname: hostname,
		Data:     dataToSend, // Heartbeat thì cái này là null, rất nhẹ
	}

	jsonBytes, _ := json.Marshal(payload)
	http.Post(SERVER_URL, "application/json", bytes.NewBuffer(jsonBytes))
}

// SỬA 1: Thêm trường Hostname vào struct để Backend nhận diện
type SecurityLog struct {
	LogType        string      `json:"log_type"`
	EnrollToken    string      `json:"enroll_token"`
	HWID           string      `json:"hwid"`
	Hostname       string      `json:"hostname"` // Quan trọng!
	ScanTimestamp  int64       `json:"scan_timestamp"`
	IsDifferential bool        `json:"is_diff"`
	Data           interface{} `json:"data"`
}

func getHash(data interface{}) string {
	b, _ := json.Marshal(data)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

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

func collectInventory() interface{} {
	hInfo, _ := host.Info()
	cpuInfo, _ := cpu.Info()
	vMem, _ := mem.VirtualMemory()

	model := "Unknown CPU"
	if len(cpuInfo) > 0 {
		model = cpuInfo[0].ModelName
	}

	return map[string]interface{}{
		"os_info":      runtime.GOOS + " " + hInfo.PlatformVersion,
		"cpu_model":    model,
		"ram_total_gb": vMem.Total / 1024 / 1024 / 1024,
	}
}

func collectSoftware() []string {
	if runtime.GOOS != "windows" {
		return []string{"Not implemented for " + runtime.GOOS}
	}
	var softwareList []string
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		return []string{}
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

// SỬA 2: Telemetry trả về danh sách Port chi tiết thay vì chỉ đếm số lượng
func collectTelemetry() interface{} {
	connections, _ := psnet.Connections("tcp")

	var openPorts []map[string]interface{}
	for _, conn := range connections {
		// Chỉ lấy các cổng đang lắng nghe (LISTEN)
		if conn.Status == "LISTEN" {
			openPorts = append(openPorts, map[string]interface{}{
				"port":         conn.Laddr.Port,
				"process_name": "system", // Tạm thời để system (lấy tên process cần quyền Admin cao hơn)
			})
		}
	}

	// Trả về đúng cấu trúc Backend mong đợi
	return map[string]interface{}{
		"open_ports": openPorts,
	}
}

// Hàm gửi dữ liệu
func send(logType string, hwid string, hostname string, payload interface{}, force bool) {
	// Kiểm tra thay đổi trước khi gửi
	if !force && !hasChanged(logType, payload) {
		return
	}

	entry := SecurityLog{
		LogType:        logType,
		EnrollToken:    ENROLL_TOKEN,
		HWID:           hwid,
		Hostname:       hostname, // Gửi kèm tên máy
		ScanTimestamp:  time.Now().Unix(),
		Data:           payload,
		IsDifferential: !force,
	}

	jsonData, _ := json.Marshal(entry)
	resp, err := http.Post(SERVER_URL, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		fmt.Printf("❌ Lỗi kết nối Server: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("🚀 [%s] Đã gửi thành công (Size: %d bytes)\n", logType, len(jsonData))
	} else {
		fmt.Printf("⚠️ Server trả về lỗi: %d\n", resp.StatusCode)
	}
}

func main() {
	hInfo, _ := host.Info()
	hwid := hInfo.HostID
	hostname, _ := os.Hostname() // Lấy tên máy từ OS

	fmt.Printf("🛡️ SMART AGENT v4.0 - Edge Processing\n")
	fmt.Printf("Máy trạm: %s (%s)\n", hostname, hwid)

	// 1. Gửi Inventory ngay lập tức (Để đăng ký máy)
	fmt.Println("📢 Đang gửi thông tin đăng ký máy...")
	send("inventory", hwid, hostname, collectInventory(), true)

	// 2. Gửi dữ liệu lần đầu (Force = true)
	send("software", hwid, hostname, collectSoftware(), true)
	send("telemetry", hwid, hostname, collectTelemetry(), true)

	// 3. Vòng lặp gửi định kỳ
	ticker := time.NewTicker(10 * time.Second) // Check 10 giây/lần
	for range ticker.C {
		// Thu thập dữ liệu
		telemetry := collectTelemetry() // Hàm cũ của bạn
		software := collectSoftware()   // Hàm cũ của bạn

		// Để Agent tự quyết định có gửi hay không
		sendPayload(hwid, hostname, "telemetry", telemetry)
		sendPayload(hwid, hostname, "software", software)
	}
}
