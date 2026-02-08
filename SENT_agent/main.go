// SENT_agent/main.go
package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

type TelemetryData struct {
	EnrollToken string                   `json:"enroll_token"` // Dùng cho lần đầu đăng ký
	HWID        string                   `json:"hwid"`
	Hostname    string                   `json:"hostname"`
	OSInfo      string                   `json:"os_info"`
	NetworkInfo map[string]string        `json:"network_info"`
	USBDevices  []map[string]interface{} `json:"usb_devices"`
}

func collectInfo() TelemetryData {
	hostname, _ := os.Hostname()

	// Thu thập IP/MAC
	netInfo := make(map[string]string)
	interfaces, _ := net.Interfaces()
	for _, i := range interfaces {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					netInfo["ip"] = ipnet.IP.String()
					netInfo["mac"] = i.HardwareAddr.String()
				}
			}
		}
	}

	return TelemetryData{
		EnrollToken: "SENT-TOKEN-SME-01", // Đọc từ file config.json
		HWID:        "Mã_Phần_Cứng_Duy_Nhất",
		Hostname:    hostname,
		OSInfo:      runtime.GOOS + " " + runtime.GOARCH,
		NetworkInfo: netInfo,
		USBDevices:  []map[string]interface{}{{"name": "Kingston", "id": "vid_0951&pid_1666"}}, // Giả lập quét USB
	}
}

func main() {
	for {
		data := collectInfo()
		jsonData, _ := json.Marshal(data)

		// Đẩy dữ liệu lên Mediator (Backend Go)
		http.Post("http://localhost:8000/api/v1/assets/push", "application/json", bytes.NewBuffer(jsonData))

		time.Sleep(1 * time.Minute) // Chu kỳ giám sát
	}
}
