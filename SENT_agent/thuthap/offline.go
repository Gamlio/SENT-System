package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
	"golang.org/x/sys/windows/registry"
)

// Cấu trúc dữ liệu
type SystemDump struct {
	Timestamp      string      `json:"timestamp"`
	Hostname       string      `json:"hostname"`
	HWID           string      `json:"hwid"`
	Inventory      interface{} `json:"inventory"`
	Software       interface{} `json:"software"`
	OpenPorts      interface{} `json:"open_ports"`
	SecurityEvents interface{} `json:"security_events"`
}

// 1. Tự động khởi chạy cùng Windows (Persistence)
func ensurePersistence() {
	if runtime.GOOS != "windows" {
		return
	}
	// Lấy đường dẫn tuyệt đối của file exe đang chạy
	exePath, err := os.Executable()
	if err != nil {
		return
	}

	// Ghi vào Registry: Software\Microsoft\Windows\CurrentVersion\Run
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()

	// Đặt tên ứng dụng là "SENT_Agent_Service"
	_ = k.SetStringValue("SENT_Agent_Service", exePath)
}

func isAdmin() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	cmd := exec.Command("cmd", "/C", "net session >nul 2>&1")
	return cmd.Run() == nil
}

// --- CÁC HÀM THU THẬP GIỮ NGUYÊN ---
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

func collectSoftware() interface{} {
	var softwareList []map[string]string
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		return softwareList
	}
	defer k.Close()
	names, _ := k.ReadSubKeyNames(-1)
	for _, name := range names {
		sk, _ := registry.OpenKey(k, name, registry.QUERY_VALUE)
		displayName, _, _ := sk.GetStringValue("DisplayName")
		displayVersion, _, _ := sk.GetStringValue("DisplayVersion")
		if displayName != "" {
			softwareList = append(softwareList, map[string]string{
				"software_name": strings.TrimSpace(displayName),
				"version":       strings.TrimSpace(displayVersion),
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
			openPorts = append(openPorts, map[string]interface{}{
				"port": conn.Laddr.Port, "local_ip": conn.Laddr.IP, "pid": conn.Pid,
			})
		}
	}
	return openPorts
}

func collectSecurityLogs() interface{} {
	cmd := exec.Command("wevtutil", "qe", "Security", "/q:*[System[(EventID=4625)]]", "/f:text", "/c:10", "/rd:true")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil || out.String() == "" {
		return []string{"No logs found"}
	}
	logs := strings.Split(out.String(), "Event[")
	var parsedLogs []string
	for _, l := range logs {
		if strings.TrimSpace(l) != "" {
			parsedLogs = append(parsedLogs, "Event["+strings.TrimSpace(l))
		}
	}
	return parsedLogs
}

func runAgentTask() {
	hInfo, _ := host.Info()
	hostname, _ := os.Hostname()
	// Tên file có cả Phút và Giây
	timestamp := time.Now().Format("20060102_15h04m05s")

	dump := SystemDump{
		Timestamp:      time.Now().Format(time.RFC3339),
		Hostname:       hostname,
		HWID:           hInfo.HostID,
		Inventory:      collectInventory(),
		Software:       collectSoftware(),
		OpenPorts:      collectTelemetry(),
		SecurityEvents: collectSecurityLogs(),
	}

	jsonData, _ := json.MarshalIndent(dump, "", "  ")
	fileName := fmt.Sprintf("sent_data_%s_%s.json", hostname, timestamp)
	_ = os.WriteFile(fileName, jsonData, 0644)

	webhookURL := "https://script.google.com/macros/s/AKfycbx9O8Qno_06KmtpBI7Y7_4weS1Z4r2vy-M8FJdxXDacctgKPX9YM10eiB4HTouzsWadYQ/exec"
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))

	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 || resp.StatusCode == 302 {
			_ = os.Remove(fileName) // Xóa file JSON khi thành công
		}
	}
}

func main() {
	if !isAdmin() {
		return
	}

	// 1. Kích hoạt tự động khởi chạy cùng Win
	ensurePersistence()

	// 2. Chạy ngay lập tức lần đầu
	runAgentTask()

	// 3. Cập nhật mỗi 10 PHÚT
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		runAgentTask()
	}
}
