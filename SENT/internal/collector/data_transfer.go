package collector

import (
	"sort"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
)

// 1. Khai báo Struct cho Sensor
type DataTransferSensor struct{}

// Mở rộng Record để chứa danh sách ứng dụng tiêu thụ mạng và I/O đĩa
type DataTransferRecord struct {
	TotalNetSent     uint64            `json:"total_net_sent"`
	TotalNetRecv     uint64            `json:"total_net_recv"`
	DiskBytesRead    uint64            `json:"disk_bytes_read"`
	DiskBytesWritten uint64            `json:"disk_bytes_written"`
	TopProcesses     []ProcessNetStats `json:"top_processes"`
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

	var totalDiskRead, totalDiskWritten uint64
	diskStats, err := disk.IOCounters()
	if err == nil {
		for _, stats := range diskStats {
			totalDiskRead += stats.ReadBytes
			totalDiskWritten += stats.WriteBytes
		}
	}

	// [SỬA LỖI LOGIC] Loại bỏ việc dùng p.IOCounters() vì nó trả về số liệu đọc/ghi Disk I/O,
	// dẫn đến việc báo cáo sai lệch thành hành vi mạng (Data Exfiltration).
	// Lấy danh sách tiến trình đang mở kết nối mạng để định danh ứng dụng ngốn mạng.
	var procStats []ProcessNetStats
	procs, err := GetProcessesCached()
	if err == nil {
		for _, p := range procs {
			conns, err := p.Connections()
			if err == nil && len(conns) > 0 {
				name, _ := p.Name()
				procStats = append(procStats, ProcessNetStats{
					PID:       p.Pid,
					Name:      name,
					BytesSent: uint64(len(conns)), // Dùng số lượng Socket đang mở làm trọng số mô phỏng
					BytesRecv: 0,
				})
			}
		}
	}

	sort.Slice(procStats, func(i, j int) bool {
		return procStats[i].BytesSent > procStats[j].BytesSent
	})
	if len(procStats) > 5 {
		procStats = procStats[:5]
	}

	return DataTransferRecord{
		TotalNetSent:     totalSent,
		TotalNetRecv:     totalRecv,
		DiskBytesRead:    totalDiskRead,
		DiskBytesWritten: totalDiskWritten,
		TopProcesses:     procStats,
	}, nil
}
