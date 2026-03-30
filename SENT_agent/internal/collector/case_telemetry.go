package collector

import (
	"log"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
)

// 1. Khai báo Struct cho Sensor
type TelemetrySensor struct{}

// 2. Cấu trúc JSON chuẩn gửi về Backend
type TelemetryRecord struct {
	NetBytesSent     uint64 `json:"net_bytes_sent"`
	NetBytesRecv     uint64 `json:"net_bytes_recv"`
	DiskBytesWritten uint64 `json:"disk_bytes_written"`
	DiskBytesRead    uint64 `json:"disk_bytes_read"`
}

// 3. Khai báo tên định danh của Log
func (s *TelemetrySensor) Name() string {
	return "telemetry"
}

// 4. Logic thu thập dữ liệu
func (s *TelemetrySensor) Collect() (interface{}, error) {
	var totalNetSent, totalNetRecv uint64
	var totalDiskWrite, totalDiskRead uint64

	// A. Thu thập I/O Mạng (Network)
	// Tham số 'false' gộp tổng lưu lượng của TẤT CẢ các card mạng (Wifi, LAN, VPN...)
	netStats, err := net.IOCounters(false)
	if err != nil {
		// Ghi lại cảnh báo thay vì im lặng bỏ qua.
		log.Printf("Telemetry Warning: Không thể lấy thông tin Network I/O: %v", err)
	} else if len(netStats) > 0 {
		totalNetSent = netStats[0].BytesSent
		totalNetRecv = netStats[0].BytesRecv
	}

	// B. Thu thập I/O Ổ cứng (Disk)
	// Hàm này trả về map thông số của từng phân vùng (C:, D: hoặc /dev/sda1), ta cần cộng dồn lại
	diskStats, err := disk.IOCounters()
	if err != nil {
		// Ghi lại cảnh báo thay vì im lặng bỏ qua.
		log.Printf("Telemetry Warning: Không thể lấy thông tin Disk I/O: %v", err)
	} else {
		for _, stat := range diskStats {
			totalDiskWrite += stat.WriteBytes
			totalDiskRead += stat.ReadBytes
		}
	}

	return TelemetryRecord{
		NetBytesSent:     totalNetSent,
		NetBytesRecv:     totalNetRecv,
		DiskBytesWritten: totalDiskWrite,
		DiskBytesRead:    totalDiskRead,
	}, nil // Telemetry thường không nên trả về lỗi làm dừng agent, chỉ trả về giá trị 0 nếu không thu thập được
}
