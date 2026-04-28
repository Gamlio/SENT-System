package collector

import (
	"log"
	"sort"

	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// 1. Khai báo Struct cho Sensor
type DataTransferSensor struct{}

// Mở rộng Record để chứa danh sách ứng dụng tiêu thụ mạng
type DataTransferRecord struct {
	TotalNetSent uint64            `json:"total_net_sent"`
	TotalNetRecv uint64            `json:"total_net_recv"`
	TopProcesses []ProcessNetStats `json:"top_processes"`
}

type ProcessNetStats struct {
	PID       int32  `json:"pid"`
	Name      string `json:"name"`
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
}

// 3. Khai báo tên định danh của Log
func (s *DataTransferSensor) Name() string {
	return "data_transfer"
}

func (s *DataTransferSensor) Collect() (interface{}, error) {
	// Lấy tổng quan (Tổng lưu lượng mạng)
	var totalSent, totalRecv uint64
	netStats, err := net.IOCounters(false)
	if err == nil && len(netStats) > 0 {
		totalSent = netStats[0].BytesSent
		totalRecv = netStats[0].BytesRecv
	}

	// Lấy thông tin I/O theo từng PID
	procs, err := process.Processes()
	if err != nil {
		log.Printf("Telemetry lỗi: Không thể đọc processes: %v", err)
		return DataTransferRecord{TotalNetSent: totalSent, TotalNetRecv: totalRecv}, nil
	}

	var procStats []ProcessNetStats
	for _, p := range procs {
		ioc, err := p.IOCounters()
		if err == nil && (ioc.WriteBytes > 0 || ioc.ReadBytes > 0) {
			name, _ := p.Name()
			procStats = append(procStats, ProcessNetStats{
				PID:       p.Pid,
				Name:      name,
				BytesSent: ioc.WriteBytes, // Trên một số OS, IO Bytes có thể map mạng hoặc đĩa
				BytesRecv: ioc.ReadBytes,
			})
		}
	}

	// Sắp xếp giảm dần theo lưu lượng gửi đi (Ưu tiên phát hiện Data Exfiltration)
	sort.Slice(procStats, func(i, j int) bool {
		return procStats[i].BytesSent > procStats[j].BytesSent
	})

	// Chỉ lấy Top 5 tiến trình ăn mạng nhiều nhất để giảm kích thước Payload
	if len(procStats) > 5 {
		procStats = procStats[:5]
	}

	return DataTransferRecord{
		TotalNetSent: totalSent,
		TotalNetRecv: totalRecv,
		TopProcesses: procStats,
	}, nil
}
