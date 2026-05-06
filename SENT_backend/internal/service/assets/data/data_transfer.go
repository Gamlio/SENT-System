package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sent_backend/internal/cache"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func ProcessDataTransfer(asset models.Asset, data interface{}) error {
	// LOẠI BỎ CHAI CỔ CPU TỪ MARSHALLING NẾU CÓ THỂ
	bytes, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu data_transfer không hợp lệ: %w", err)
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
		return fmt.Errorf("lỗi giải mã JSON data_transfer: %w", err)
	}

	// [HIỆU NĂNG & CLUSTER CLOUD] Chuyển đổi in-memory sync.Map sang Distributed Redis Cache
	// Giải quyết triệt để rủi ro rò rỉ RAM và đồng bộ hóa state khi backend chạy đa Node (K8s).
	cacheKey := fmt.Sprintf("io_activity:%s", asset.AssetHWID)
	lastValStr, err := cache.RDB.Get(context.Background(), cacheKey).Result()

	// Cập nhật giá trị mới với TTL 5 phút để Redis tự dọn dẹp (thay thế goroutine cleanup)
	cache.RDB.Set(context.Background(), cacheKey, payload.TotalNetSent, 5*time.Minute)

	if err == redis.Nil || err != nil {
		return nil
	}

	lastNetSent, _ := strconv.ParseUint(lastValStr, 10, 64)

	// 1. TÍNH TOÁN DELTA
	var diffSent uint64
	if payload.TotalNetSent > lastNetSent {
		diffSent = payload.TotalNetSent - lastNetSent
	}

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
	return nil
}
