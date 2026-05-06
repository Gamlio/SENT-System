//go:build darwin

package collector

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Collect: Thu thập và định danh USB trên macOS qua system_profiler
func (s *USBSensor) Collect() (interface{}, error) {
	var usbList []USBRecord

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// macOS: Dùng system_profiler xuất ra JSON để thu thập thông tin chi tiết
	cmd := exec.CommandContext(ctx, "/usr/sbin/system_profiler", "SPUSBDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("lỗi thực thi system_profiler: %w", err)
	}

	// TODO: Parse output JSON tùy theo nhu cầu thu thập chi tiết
	_ = output

	return usbList, nil
}
