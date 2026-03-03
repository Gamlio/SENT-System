package collector

import (
	"sent_agent/internal/utils"

	psnet "github.com/shirou/gopsutil/v3/net"
)

func CollectTelemetry() interface{} {
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
