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

	// 1. ĐỊNH NGHĨA STRUCT KHỚP VỚI AGENT MỚI
	var payload struct {
		TotalNetSent uint64 `json:"total_net_sent"`
		TotalNetRecv uint64 `json:"total_net_recv"`
		TopProcesses []struct {
			Name      string `json:"name"`
			BytesSent uint64 `json:"bytes_sent"`
		} `json:"top_processes"`
	}

	if err := json.Unmarshal(bytes, &payload); err != nil {
		return
	}

	// Tạo record hiện tại để lưu cache
	current := models.AssetIOActivity{
		Timestamp:    time.Now(),
		NetBytesSent: payload.TotalNetSent,
	}

	// Đọc từ Concurrent Map không cần Lock
	val, exists := ioCache.Load(asset.AssetHWID)
	ioCache.Store(asset.AssetHWID, current)

	if !exists {
		return
	}
	lastRecord := val.(models.AssetIOActivity)

	// 1. TÍNH TOÁN DELTA
	diffSent := current.NetBytesSent - lastRecord.NetBytesSent

	// 2. NGƯỠNG CẢNH BÁO
	isNetworkSpike := diffSent > 500*1024*1024 // 500MB

	if isNetworkSpike {
		incSvc := &incidents.IncidentService{}

		// Trích xuất tên ứng dụng ngốn mạng nhất từ TopProcesses
		topApp := "Unknown"
		if len(payload.TopProcesses) > 0 {
			topApp = payload.TopProcesses[0].Name
		}

		reason := fmt.Sprintf("Lưu lượng mạng đột biến. Ứng dụng khả nghi: %s", topApp)

		incSvc.TriggerSecurityEvent(context.TODO(), asset,
			"Data Exfiltration",
			"[P1] Hoạt động mạng bất thường",
			fmt.Sprintf("%s. Lượng dữ liệu: %d MB/30s", reason, diffSent/1024/1024),
			"P1",
		)
	}
}
