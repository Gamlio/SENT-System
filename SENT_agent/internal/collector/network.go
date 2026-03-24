package collector

import (
	"sent_agent/internal/utils"

	psnet "github.com/shirou/gopsutil/v3/net"
)

// 1. Khai báo Struct cho Sensor
type PortSensor struct{}

// 2. Cấu trúc JSON chuẩn gửi về Backend
type PortRecord struct {
	OpenPorts []map[string]interface{} `json:"open_ports"`
	IPAddress string                   `json:"ip_address"`
}

// 3. Khai báo tên định danh của Log
func (s *PortSensor) Name() string {
	return "port"
}

// 4. Đưa logic cũ vào hàm Collect()
func (s *PortSensor) Collect() interface{} {
	connections, _ := psnet.Connections("tcp")
	var openPorts []map[string]interface{}
	for _, conn := range connections {
		if conn.Status == "LISTEN" {
			openPorts = append(openPorts, map[string]interface{}{
				"port":         conn.Laddr.Port,
				"process_name": "system",
			})
		}
	}

	return map[string]interface{}{
		"open_ports": openPorts,
		"ip_address": utils.GetOutboundIP(), // Gọi hàm từ package utils
	}
}
