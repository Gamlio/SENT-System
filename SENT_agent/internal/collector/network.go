package collector

import (
	"log"
	"sent_agent/internal/utils"

	psnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// 1. Khai báo Struct cho Sensor
type PortSensor struct{}

// 2. Cấu trúc JSON chuẩn gửi về Backend
type PortRecord struct {
	OpenPorts []OpenPortInfo `json:"open_ports"`
	IPAddress string         `json:"ip_address"`
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
func (s *PortSensor) Collect() (interface{}, error) {
	connections, err := psnet.Connections("tcp")
	if err != nil {
		return nil, err
	}
	var openPorts []OpenPortInfo
	for _, conn := range connections {
		if conn.Status == "LISTEN" {
			procName := "system" // Mặc định là system nếu không tìm thấy PID hoặc có lỗi
			if conn.Pid > 0 {
				p, err := process.NewProcess(conn.Pid)
				if err == nil {
					name, err := p.Name()
					if err == nil {
						procName = name
					}
				} else {
					// Ghi log lỗi nếu cần, ví dụ process đã kết thúc trước khi kịp lấy tên
					log.Printf("Không thể lấy thông tin process cho PID %d: %v", conn.Pid, err)
				}
			}
			openPorts = append(openPorts, OpenPortInfo{
				Port:        conn.Laddr.Port,
				ProcessName: procName,
			})
		}
	}

	return PortRecord{
		OpenPorts: openPorts,
		IPAddress: utils.GetOutboundIP(), // Gọi hàm từ package utils
	}, nil
}
