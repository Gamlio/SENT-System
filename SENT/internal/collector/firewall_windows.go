//go:build windows

package collector

import (
	"golang.org/x/sys/windows/registry"
)

// Collect: Thu thập trạng thái firewall trên Windows không dùng netsh
func (s *FirewallSensor) Collect() (interface{}, error) {
	// 3 Profile chính của Windows Defender Firewall
	profiles := []string{"DomainProfile", "StandardProfile", "PublicProfile"}
	isOff := false

	for _, profile := range profiles {
		keyPath := `SYSTEM\CurrentControlSet\Services\SharedAccess\Parameters\FirewallPolicy\` + profile
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE)
		if err != nil {
			continue // Bỏ qua nếu không đọc được quyền
		}

		val, _, err := k.GetIntegerValue("EnableFirewall")
		k.Close()

		// val == 0 nghĩa là Tường lửa đang TẮT ở profile này
		if err == nil && val == 0 {
			isOff = true
			break // Chỉ cần 1 profile tắt là hệ thống có rủi ro
		}
	}

	return FirewallRecord{FirewallOff: isOff}, nil
}
