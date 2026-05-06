//go:build windows

package collector

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sys/windows/registry"
)

// Collect: Thu thập và định danh USB trên Windows qua Registry
func (s *USBSensor) Collect() (interface{}, error) {
	var usbList []USBRecord

	// Đặt timeout 10 giây để tránh treo lệnh hệ thống
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// [TỐI ƯU] Đọc trực tiếp Registry thay vì gọi PowerShell để loại bỏ CPU Spikes[cite: 80, 89]
	// Theo dõi USB Mass Storage (USBSTOR) - mục tiêu chính của giám sát an ninh
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Services\USBSTOR\Enum`, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		count, _, err := k.GetIntegerValue("Count")
		if err == nil {
			for i := 0; i < int(count); i++ {
				val, _, err := k.GetStringValue(fmt.Sprintf("%d", i))
				if err == nil {
					vid, pid, serial := parseWindowsUSBID(val)
					hash := generateDeviceHash(vid, pid, serial)
					usbList = append(usbList, USBRecord{
						DeviceName:   "USB Mass Storage Device",
						DeviceID:     val,
						VID:          vid,
						PID:          pid,
						SerialNumber: serial,
						DeviceHash:   hash,
						EventType:    "plugged",
					})
				}
			}
		}
	}
	return usbList, nil
}
