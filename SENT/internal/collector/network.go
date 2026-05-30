package collector

import (
	"SENT/internal/policy"
	"SENT/internal/utils"
	"fmt"
	"log"

	psnet "github.com/shirou/gopsutil/v3/net"
)

// 1. Khai báo Struct cho Sensor
type PortSensor struct{}

// 2. Cấu trúc JSON chuẩn gửi về Backend
type PortRecord struct {
	OpenPorts   []OpenPortInfo `json:"open_ports"`
	IPAddress   string         `json:"ip_address"`
	FirewallOff bool           `json:"firewall_off"`
}

// 3. Khai báo tên định danh của Log
func (s *PortSensor) Name() string {
	return "port"
}

type OpenPortInfo struct {
	Port        uint32 `json:"port"`
	ProcessName string `json:"process_name"`
}

// 4. Đưa logic cũ vào hàm Collect()
func warnPortPolicy(openPort OpenPortInfo) {
	portValue := fmt.Sprintf("%d", openPort.Port)
	if violated, reason := policy.CheckPolicy("PORT", portValue); violated {
		log.Printf("[Port Sensor] CẢNH BÁO: Port %s không nằm trong whitelist - %s", portValue, reason)
	}
	if openPort.ProcessName != "" {
		if violated, reason := policy.CheckPolicy("PROCESS", openPort.ProcessName); violated {
			log.Printf("[Port Sensor] CẢNH BÁO: tiến trình lắng nghe '%s' không nằm trong whitelist PROCESS - %s", openPort.ProcessName, reason)
		}
	}
}

func (s *PortSensor) Collect() (interface{}, error) {
	// BƯỚC 1: Chụp ảnh nhanh toàn bộ tiến trình TRƯỚC (Tránh Race Condition)
	procs, _ := GetProcessesCached() // Dùng bộ nhớ đệm để tránh gọi Syscall liên tục
	procMap := make(map[int32]string)
	for _, p := range procs {
		if name, err := p.Name(); err == nil {
			procMap[p.Pid] = name
		}
	}

	// BƯỚC 2: Quét kết nối TCP
	connections, err := psnet.Connections("tcp")
	if err != nil {
		return nil, err
	}

	var openPorts []OpenPortInfo
	seenPorts := make(map[uint32]bool) // Tránh trùng lặp nếu 1 ứng dụng mở nhiều luồng trên cùng 1 Port

	for _, conn := range connections {
		// Chỉ lấy các cổng đang chờ kết nối (LISTEN)
		if conn.Status == "LISTEN" {
			if seenPorts[conn.Laddr.Port] {
				continue
			}
			seenPorts[conn.Laddr.Port] = true

			procName := "system"
			if conn.Pid > 0 {
				if name, exists := procMap[conn.Pid]; exists && name != "" {
					procName = name
				} else {
					procName = "unknown (dead/hidden)"
				}
			}

			openPort := OpenPortInfo{
				Port:        conn.Laddr.Port,
				ProcessName: procName,
			}
			openPorts = append(openPorts, openPort)
			warnPortPolicy(openPort)
		}
	}

	// Gộp log Firewall
	fwSensor := &FirewallSensor{}
	fwData, _ := fwSensor.Collect()
	isFwOff := false
	if fwRec, ok := fwData.(FirewallRecord); ok {
		isFwOff = fwRec.FirewallOff
	}

	return PortRecord{
		OpenPorts:   openPorts,
		IPAddress:   utils.GetOutboundIP(), // Gọi hàm từ package utils
		FirewallOff: isFwOff,
	}, nil
}
