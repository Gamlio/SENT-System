//go:build windows

package collector

import (
	"context"
	"time"

	"github.com/StackExchange/wmi"
)

// Collect: Thu thập trạng thái mã độc / trình diệt virus on Windows
func (s *AntivirusSensor) Collect() (interface{}, error) {
	// [FIX-WMI-HANG] Đặt timeout 5 giây cho các lệnh OS và truy vấn WMI
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type MSFT_MpThreat struct {
		RollupStatus uint32
	}
	var threats []MSFT_MpThreat
	q := wmi.CreateQuery(&threats, "WHERE RollupStatus = 1")
	// Thay thế wmi.QueryWithContext bằng cơ chế timeout thủ công để tương thích ngược.
	errChan := make(chan error, 1)
	go func() {
		// Sử dụng hàm wmi.Query cũ hơn bên trong goroutine
		errChan <- wmi.Query(q, &threats, ".", `ROOT\Microsoft\Windows\Defender`)
	}()

	select {
	case err := <-errChan:
		// Truy vấn WMI hoàn tất (thành công hoặc lỗi)
		if err == nil && len(threats) > 0 {
			return AntivirusRecord{HasThreat: true}, nil
		}
		// Nếu có lỗi hoặc không có threat, sẽ rơi xuống return mặc định bên dưới.
	case <-ctx.Done():
		// Truy vấn bị timeout, cũng rơi xuống return mặc định.
	}
	return AntivirusRecord{HasThreat: false}, nil
}
