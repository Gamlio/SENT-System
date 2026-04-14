//go:build windows

package collector

import (
	"context"
	"time"

	"github.com/StackExchange/wmi"
)

// Collect: Thu thập trạng thái firewall on Windows
func (s *FirewallSensor) Collect() (interface{}, error) {
	// [FIX-WMI-HANG] Đặt timeout 5 giây cho các lệnh OS và truy vấn WMI
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type MSFT_NetFirewallProfile struct {
		Enabled bool
	}
	var profiles []MSFT_NetFirewallProfile
	q := wmi.CreateQuery(&profiles, "")
	// Thay thế wmi.QueryWithContext bằng cơ chế timeout thủ công để tương thích ngược.
	errChan := make(chan error, 1)
	go func() {
		// Sử dụng hàm wmi.Query cũ hơn bên trong goroutine
		errChan <- wmi.Query(q, &profiles, ".", `ROOT\StandardCimv2`)
	}()

	select {
	case err := <-errChan:
		// Truy vấn WMI hoàn tất (thành công hoặc lỗi)
		if err != nil {
			// Nếu có lỗi, mặc định là an toàn (không off) để tránh báo động giả.
			return FirewallRecord{FirewallOff: false}, nil
		}
	case <-ctx.Done():
		// Truy vấn bị timeout, cũng coi như lỗi và trả về trạng thái an toàn.
		return FirewallRecord{FirewallOff: false}, nil
	}

	for _, profile := range profiles {
		if !profile.Enabled {
			return FirewallRecord{FirewallOff: true}, nil
		}
	}
	return FirewallRecord{FirewallOff: false}, nil
}
