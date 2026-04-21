package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
	"sync"
	"time"
)

// Sử dụng sync.Map thay cho Mutex để loại bỏ thắt cổ chai (Lock Contention)
var ioCache sync.Map

// init: Khởi chạy một goroutine để dọn dẹp cache định kỳ, chống rò rỉ bộ nhớ.
func init() {
	// Chạy một tiến trình dọn dẹp mỗi 10 phút.
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cleanupExpiredCache()
		}
	}()
}

// cleanupExpiredCache: Xóa các entry trong ioCache của các máy trạm không hoạt động quá 5 phút.
func cleanupExpiredCache() {
	threshold := 5 * time.Minute
	ioCache.Range(func(key, value interface{}) bool {
		record := value.(models.AssetIOActivity)
		if time.Since(record.Timestamp) > threshold {
			ioCache.Delete(key)
		}
		return true // Tiếp tục lặp
	})
}

func ProcessDataTransfer(asset models.Asset, data interface{}) {
	// LOẠI BỎ CHAI CỔ CPU TỪ MARSHALLING: Ép kiểu trực tiếp từ json.RawMessage
	var bytes []byte
	if raw, ok := data.(json.RawMessage); ok {
		bytes = raw
	} else {
		return // Dữ liệu không hợp lệ
	}

	var current models.AssetIOActivity
	if err := json.Unmarshal(bytes, &current); err != nil {
		return
	}
	current.Timestamp = time.Now()

	// Đọc từ Concurrent Map không cần Lock
	val, exists := ioCache.Load(asset.AssetHWID)
	ioCache.Store(asset.AssetHWID, current)

	if !exists {
		return // Lần đầu tiên chỉ lưu cache để lấy mốc so sánh
	}
	lastRecord := val.(models.AssetIOActivity)

	// 1. TÍNH TOÁN DELTA (Lượng dữ liệu phát sinh trong 30 giây qua)
	diffSent := current.NetBytesSent - lastRecord.NetBytesSent
	diffWrite := current.DiskBytesWritten - lastRecord.DiskBytesWritten

	// 2. NGƯỠNG CẢNH BÁO (Ví dụ: Gửi > 500MB hoặc Ghi > 1GB trong 30 giây)
	// Đây mới là con số phản ánh hoạt động "truyền tải/lưu trữ dữ liệu lớn" thực tế
	isNetworkSpike := diffSent > 500*1024*1024
	isDiskSpike := diffWrite > 1024*1024*1024

	if isNetworkSpike || isDiskSpike {
		// 3. CHỈ LƯU VÀO MongoDB KHI CÓ BẤT THƯỜNG
		current.AssetHWID = asset.AssetHWID

		// 4. KÍCH HOẠT SỰ CỐ (Chỉ hiển thị ở phần Incidents)
		incSvc := &incidents.IncidentService{}
		reason := "Phát hiện lưu lượng mạng đột biến (Nghi vấn rò rỉ dữ liệu)"
		if isDiskSpike {
			reason = "Phát hiện ghi đĩa khối lượng lớn (Nghi vấn mã hóa Ransomware hoặc Copy dữ liệu)"
		}

		incSvc.TriggerSecurityEvent(context.TODO(), asset,
			"Data Exfiltration/Intensive I/O",
			"[P1] Hoạt động I/O bất thường",
			fmt.Sprintf("%s. Lượng dữ liệu: %d MB/30s", reason, diffSent/1024/1024),
			"P1",
		)
	}
}
