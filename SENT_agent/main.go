package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
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
	SERVER_URL  = "http://localhost:8000/api/v1/agents/push"
	CONFIG_FILE = "agent_config.json"
	IS_DEV_MODE = true
)

type Config struct {
	CompanyCode string `json:"company_code"`
}

var (
	lastHashes  = make(map[string]string)
	hashMutex   sync.Mutex
	agentConfig Config
)

type Payload struct {
	Type        string      `json:"type"`
	LogType     string      `json:"log_type"`
	HWID        string      `json:"hwid"`
	Hostname    string      `json:"hostname"`
	CompanyCode string      `json:"company_code"`
	Data        interface{} `json:"data"`
}

// --- CẤU HÌNH ---
func loadConfig() bool {
	file, err := os.ReadFile(CONFIG_FILE)
	if err != nil {
		return false
	}
	json.Unmarshal(file, &agentConfig)
	return agentConfig.CompanyCode != ""
}

func saveConfig(code string) {
	agentConfig.CompanyCode = strings.TrimSpace(code)
	data, _ := json.Marshal(agentConfig)
	os.WriteFile(CONFIG_FILE, data, 0644)
}

// --- THAY THẾ GUI BẰNG CLI (Không cần GCC) ---
func promptForCompanyCode() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("===========================================")
	fmt.Println("   KÍCH HOẠT SENT AGENT (NO-GUI MODE)")
	fmt.Println("===========================================")
	fmt.Println("⚠️  Chưa tìm thấy cấu hình doanh nghiệp.")

	for {
		fmt.Print("👉 Vui lòng nhập Mã Công Ty (VD: SME-XXXX): ")
		code, _ := reader.ReadString('\n')
		code = strings.TrimSpace(code)

		if len(code) >= 5 {
			fmt.Println("✅ Đang lưu cấu hình...")
			saveConfig(code)
			fmt.Println("✅ Kích hoạt thành công! Agent sẽ bắt đầu chạy...")
			time.Sleep(1 * time.Second)
			break
		} else {
			fmt.Println("❌ Mã không hợp lệ. Vui lòng nhập lại!")
		}
	}
}

// --- LOGIC THU THẬP DỮ LIỆU (Đã sửa phần Software) ---
func calculateHash(data interface{}) string {
	b, _ := json.Marshal(data)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

func collectInventory() interface{} {
	hInfo, _ := host.Info()
	cpuInfo, _ := cpu.Info()
	vMem, _ := mem.VirtualMemory()
	model := "Unknown CPU"
	if len(cpuInfo) > 0 {
		model = cpuInfo[0].ModelName
	}
	return map[string]interface{}{"os_info": runtime.GOOS + " " + hInfo.PlatformVersion, "cpu_model": model, "ram_total_gb": vMem.Total / 1024 / 1024 / 1024}
}

// SỬA: Hàm Software chuẩn trả về map (khớp models.go)
func collectSoftware() interface{} {
	if runtime.GOOS != "windows" {
		return []map[string]string{}
	}
	var softwareList []map[string]string
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		return []map[string]string{}
	}
	defer k.Close()
	names, _ := k.ReadSubKeyNames(-1)
	for _, name := range names {
		sk, _ := registry.OpenKey(k, name, registry.QUERY_VALUE)
		displayName, _, _ := sk.GetStringValue("DisplayName")
		displayVersion, _, _ := sk.GetStringValue("DisplayVersion")

		if displayName != "" {
			softwareList = append(softwareList, map[string]string{
				"software_name": displayName,
				"version":       displayVersion,
			})
		}
		sk.Close()
	}
	return softwareList
}

func collectTelemetry() interface{} {
	connections, _ := psnet.Connections("tcp")
	var openPorts []map[string]interface{}
	for _, conn := range connections {
		if conn.Status == "LISTEN" {
			openPorts = append(openPorts, map[string]interface{}{"port": conn.Laddr.Port, "process_name": "system"})
		}
	}
	return map[string]interface{}{"open_ports": openPorts}
}

func sendPayload(hwid, hostname, logType string, data interface{}, force bool) {
	hashMutex.Lock()
	currentHash := calculateHash(data)
	oldHash := lastHashes[logType]
	packetType := "HEARTBEAT"
	var dataToSend interface{} = nil
	if force || currentHash != oldHash {
		packetType = "DATA"
		dataToSend = data
		lastHashes[logType] = currentHash
		if IS_DEV_MODE {
			fmt.Printf("⚡ [%s] Gửi dữ liệu mới...\n", logType)
		}
	}
	hashMutex.Unlock()

	payload := Payload{
		Type: packetType, LogType: logType, HWID: hwid, Hostname: hostname,
		CompanyCode: agentConfig.CompanyCode, Data: dataToSend,
	}
	jsonBytes, _ := json.Marshal(payload)
	http.Post(SERVER_URL, "application/json", bytes.NewBuffer(jsonBytes))
}

func runAgentLoop() {
	hInfo, _ := host.Info()
	hwid := hInfo.HostID
	hostname, _ := os.Hostname()

	fmt.Printf("\n🛡️  AGENT ĐANG CHẠY | CTY: %s | MÁY: %s\n", agentConfig.CompanyCode, hostname)

	// Gửi lần đầu
	sendPayload(hwid, hostname, "inventory", collectInventory(), true)
	sendPayload(hwid, hostname, "software", collectSoftware(), true)
	sendPayload(hwid, hostname, "telemetry", collectTelemetry(), true)

	// Vòng lặp
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		sendPayload(hwid, hostname, "telemetry", collectTelemetry(), false)
		sendPayload(hwid, hostname, "software", collectSoftware(), false)
	}
}

func main() {
	// 1. Nếu chưa có Config -> Hỏi trực tiếp trên Terminal
	if !loadConfig() {
		promptForCompanyCode()
	}

	// 2. Chạy Agent
	runAgentLoop()
}
