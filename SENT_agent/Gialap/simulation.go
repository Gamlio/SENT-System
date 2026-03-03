package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// --- CẤU HÌNH ---
const (
	SERVER_URL   = "http://localhost:8000/api/v1/agents/push"
	COMPANY_CODE = "SME-701D60"
	AGENT_COUNT  = 50 // Số lượng máy (Tăng lên nếu muốn test tải nặng)
)

// --- 1. KHO DỮ LIỆU ĐA HỆ ĐIỀU HÀNH ---

// Dữ liệu Windows
var winSoftware = []map[string]string{
	{"software_name": "Microsoft Office 365", "version": "16.0.17"},
	{"software_name": "Google Chrome", "version": "120.0.6099"},
	{"software_name": "Unikey", "version": "4.3 RC5"},
	{"software_name": "Kaspersky Endpoint Security", "version": "11.0"},
	{"software_name": "Zoom", "version": "5.17.0"},
}
var winProcesses = []map[string]interface{}{
	{"port": 135, "process_name": "svchost.exe"},
	{"port": 445, "process_name": "System"},       // SMB
	{"port": 3389, "process_name": "TermService"}, // RDP
}

// Dữ liệu Linux (Server/Dev)
var linuxSoftware = []map[string]string{
	{"software_name": "Docker CE", "version": "24.0.5"},
	{"software_name": "Nginx", "version": "1.18.0"},
	{"software_name": "PostgreSQL Client", "version": "14.9"},
	{"software_name": "OpenSSH Server", "version": "8.9p1"},
	{"software_name": "Vim", "version": "9.0"},
	{"software_name": "HTop", "version": "3.2.1"},
}
var linuxProcesses = []map[string]interface{}{
	{"port": 22, "process_name": "sshd"},
	{"port": 80, "process_name": "nginx"},
	{"port": 443, "process_name": "nginx"},
	{"port": 5432, "process_name": "postgres"},
	{"port": 6379, "process_name": "redis-server"},
}

// Dữ liệu macOS (Designer/Dev)
var macSoftware = []map[string]string{
	{"software_name": "Xcode", "version": "15.0"},
	{"software_name": "Slack", "version": "4.36.0"},
	{"software_name": "Adobe Photoshop 2024", "version": "25.0.0"},
	{"software_name": "Alfred 5", "version": "5.1.2"},
	{"software_name": "Homebrew", "version": "4.1.0"},
	{"software_name": "Little Snitch", "version": "5.7"},
}
var macProcesses = []map[string]interface{}{
	{"port": 88, "process_name": "kerberos"},             // macOS thường dùng
	{"port": 5000, "process_name": "ControlCenter"},      // AirPlay receiver
	{"port": 5900, "process_name": "ScreensharingAgent"}, // VNC
}

// Phần mềm Độc hại (Chung cho các OS để test Alert)
var riskySoftware = []map[string]string{
	{"software_name": "uTorrent", "version": "3.6.0"},
	{"software_name": "Tor Browser", "version": "13.0"},
	{"software_name": "CheatEngine", "version": "7.5"},   // Windows
	{"software_name": "XMRig Miner", "version": "6.2.0"}, // Linux Mining
}

// USB Devices (Trộn lẫn)
var usbDevices = []map[string]interface{}{
	{"device_name": "Logitech MX Master 3", "device_id": "USB\\VID_046D&PID_B023"},
	{"device_name": "Keychron K2 Keyboard", "device_id": "USB\\VID_3434&PID_0220"},
	{"device_name": "Samsung T7 SSD", "device_id": "USB\\VID_04E8&PID_61F5"},        // Ổ cứng ngoài
	{"device_name": "Kingston DataTraveler", "device_id": "USB\\VID_0951&PID_1666"}, // USB thường bị cấm
	{"device_name": "Yubikey 5 NFC", "device_id": "USB\\VID_1050&PID_0407"},         // Token bảo mật
}

type Payload struct {
	LogType     string      `json:"log_type"`
	HWID        string      `json:"hwid"`
	Hostname    string      `json:"hostname"`
	CompanyCode string      `json:"company_code"`
	Data        interface{} `json:"data"`
}

// Struct định nghĩa Profile cho mỗi máy giả lập
type AgentProfile struct {
	ID       int
	HWID     string
	Hostname string
	OS       string // "windows", "linux", "darwin"
	IP       string
}

func main() {
	fmt.Printf("🚀 BẮT ĐẦU GIẢ LẬP %d MÁY (WINDOWS / LINUX / MACOS)...\n", AGENT_COUNT)

	var wg sync.WaitGroup

	for i := 1; i <= AGENT_COUNT; i++ {
		wg.Add(1)

		// 1. Tạo hồ sơ máy (Random hệ điều hành)
		profile := createProfile(i)

		go func(p AgentProfile) {
			defer wg.Done()

			// Gửi gói Inventory đầu tiên để đăng ký OS với hệ thống
			sendInventory(p)

			// Vòng lặp gửi dữ liệu định kỳ
			for {
				// --- A. SOFTWARE (Biến động theo OS) ---
				sendSoftware(p)

				// --- B. TELEMETRY (Port & Resource theo OS) ---
				sendTelemetry(p)

				// --- C. USB (Random cắm rút) ---
				sendUSB(p)

				// Thời gian nghỉ (Chaos Mode: Máy 20-30 gửi nhanh hơn)
				sleepTime := time.Duration(30+rand.Intn(40)) * time.Second
				if p.ID >= 20 && p.ID <= 30 {
					sleepTime = time.Duration(10+rand.Intn(10)) * time.Second
					fmt.Printf("⚡ [%s - %s] Gửi dữ liệu dồn dập...\n", p.OS, p.Hostname)
				} else {
					fmt.Printf("DATA [%s - %s] Đã gửi log. Nghỉ %v...\n", p.OS, p.Hostname, sleepTime)
				}
				time.Sleep(sleepTime)
			}
		}(profile)

		time.Sleep(50 * time.Millisecond) // Delay khởi tạo tránh nghẽn
	}

	wg.Wait()
}

// --- CÁC HÀM LOGIC ---

func createProfile(id int) AgentProfile {
	// Chia tỷ lệ: 50% Windows, 30% Linux, 20% Mac
	randNum := rand.Intn(100)
	var osType, hostPrefix, ipSegment string

	if randNum < 50 {
		osType = "windows"
		hostPrefix = "DESKTOP-"
		ipSegment = "10" // Mạng văn phòng 192.168.10.x
	} else if randNum < 80 {
		osType = "linux"
		hostPrefix = "SRV-LNX-"
		ipSegment = "20" // Mạng Server 192.168.20.x
	} else {
		osType = "darwin" // MacOS identifies as darwin
		hostPrefix = "MACBOOK-PRO-"
		ipSegment = "30" // Mạng Design 192.168.30.x
	}

	return AgentProfile{
		ID:       id,
		HWID:     fmt.Sprintf("SIM-%s-%04d", osType, id),
		Hostname: fmt.Sprintf("%s%03d", hostPrefix, id),
		OS:       osType,
		IP:       fmt.Sprintf("192.168.%s.%d", ipSegment, 100+(id%150)),
	}
}

func sendInventory(p AgentProfile) {
	var osInfo, cpuModel string
	var ram int

	switch p.OS {
	case "windows":
		osInfo = "Windows 11 Pro 23H2"
		cpuModel = "Intel Core i5-12400"
		ram = 16
	case "linux":
		distros := []string{"Ubuntu 22.04 LTS", "CentOS 9 Stream", "Debian 12"}
		osInfo = distros[rand.Intn(len(distros))]
		cpuModel = "AMD EPYC 7763 (Virtual)"
		ram = 32 // Server nhiều RAM
	case "darwin":
		osInfo = "macOS Sonoma 14.2"
		cpuModel = "Apple M2 Pro"
		ram = 16
	}

	data := map[string]interface{}{
		"os_info":      osInfo,
		"cpu_model":    cpuModel,
		"ram_total_gb": float64(ram),
	}
	sendData(p, "inventory", data)
}

func sendSoftware(p AgentProfile) {
	// Lấy list base theo OS
	var baseList []map[string]string
	switch p.OS {
	case "windows":
		baseList = winSoftware
	case "linux":
		baseList = linuxSoftware
	case "darwin":
		baseList = macSoftware
	}

	// Copy để không ảnh hưởng list gốc
	currentSofts := make([]map[string]string, len(baseList))
	copy(currentSofts, baseList)

	// RANDOM: Thêm phần mềm độc hại (Cho máy Chaos ID 20-30)
	if p.ID >= 20 && p.ID <= 30 {
		if rand.Float32() < 0.6 { // 60% khả năng cài phần mềm lạ
			badApp := riskySoftware[rand.Intn(len(riskySoftware))]
			currentSofts = append(currentSofts, badApp)
		}
	}

	sendData(p, "software", currentSofts)
}

func sendTelemetry(p AgentProfile) {
	// Lấy port base theo OS
	var basePorts []map[string]interface{}
	switch p.OS {
	case "windows":
		basePorts = winProcesses
	case "linux":
		basePorts = linuxProcesses
	case "darwin":
		basePorts = macProcesses
	}

	// Copy list
	currentPorts := make([]map[string]interface{}, len(basePorts))
	copy(currentPorts, basePorts)

	// Random thêm port lạ cho máy Chaos
	if p.ID >= 20 && p.ID <= 30 && rand.Float32() < 0.5 {
		currentPorts = append(currentPorts, map[string]interface{}{
			"port": 6660 + rand.Intn(100), "process_name": "unknown_miner",
		})
	}

	// CPU/RAM Simulation
	cpuLoad := rand.Intn(30) + 5
	ramLoad := rand.Intn(40) + 20

	// Linux Server thường chạy nặng hơn chút
	if p.OS == "linux" {
		ramLoad += 20
	}

	data := map[string]interface{}{
		"open_ports": currentPorts,
		"ip_address": p.IP,
		"cpu_usage":  cpuLoad,
		"ram_usage":  ramLoad,
	}
	sendData(p, "telemetry", data)
}

func sendUSB(p AgentProfile) {
	var currentUSBs []map[string]interface{}

	// Mặc định luôn có chuột phím (Trừ server linux headless ít khi cắm)
	if p.OS != "linux" {
		currentUSBs = append(currentUSBs, usbDevices[0], usbDevices[1])
	}

	// Random cắm USB lạ
	if rand.Float32() < 0.3 { // 30% máy đang cắm USB storage
		usb := usbDevices[2+rand.Intn(3)] // Lấy từ index 2 trở đi (Storage)
		currentUSBs = append(currentUSBs, usb)
	}

	sendData(p, "usb", currentUSBs)
}

// Hàm gửi request HTTP chung
func sendData(p AgentProfile, logType string, data interface{}) {
	payload := Payload{
		LogType:     logType,
		HWID:        p.HWID,
		Hostname:    p.Hostname,
		CompanyCode: COMPANY_CODE,
		Data:        data,
	}

	jsonBytes, _ := json.Marshal(payload)
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(SERVER_URL, "application/json", bytes.NewBuffer(jsonBytes))

	if err != nil {
		// fmt.Println("Lỗi gửi:", err) // Bật lên nếu cần debug
		return
	}
	defer resp.Body.Close()
}
