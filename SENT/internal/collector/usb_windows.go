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

	// [TỐI ƯU] Thay đổi đường dẫn Registry để quét toàn bộ thiết bị USB (Chuột, phím, tai nghe, USB Flash...)
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Enum\USB`, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()

		// Tầng 1: Đọc các SubKey VID_XXXX&PID_XXXX
		subKeys, err := k.ReadSubKeyNames(-1)
		if err == nil {
			for _, subKey := range subKeys {
				sk, err := registry.OpenKey(k, subKey, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
				if err != nil {
					continue
				}

				// Tầng 2: Đọc các Instance (Serial Number hoặc ID của thiết bị)
				instanceKeys, err := sk.ReadSubKeyNames(-1)
				if err == nil {
					for _, instanceKey := range instanceKeys {
						ik, err := registry.OpenKey(sk, instanceKey, registry.QUERY_VALUE)
						if err != nil {
							continue
						}

						// Lấy tên định danh của thiết bị
						deviceName, _, err := ik.GetStringValue("FriendlyName")
						if err != nil || deviceName == "" {
							deviceName, _, err = ik.GetStringValue("DeviceDesc")
							if err != nil || deviceName == "" {
								deviceName = "Unknown USB Device"
							}
						}

						// Tạo pnpID giả lập theo chuẩn Device Path để parse bóc tách
						pnpID := fmt.Sprintf("USB\\%s\\%s", subKey, instanceKey)
						vid, pid, serial := parseWindowsUSBID(pnpID)
						hash := generateDeviceHash(vid, pid, serial)

						usbList = append(usbList, USBRecord{
							DeviceName:   deviceName,
							DeviceID:     pnpID,
							VID:          vid,
							PID:          pid,
							SerialNumber: serial,
							DeviceHash:   hash,
							EventType:    "plugged",
						})
						ik.Close()
					}
				}
				sk.Close()
			}
		}
	}
	return usbList, nil
}
