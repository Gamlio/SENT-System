package collector

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// 1. Khai báo Struct cho Sensor
type InventorySensor struct{}

// 2. Cấu trúc JSON chuẩn gửi về Backend
type InventoryRecord struct {
	OSInfo     string `json:"os_info"`
	CPUModel   string `json:"cpu_model"`
	RAMTotalGB uint64 `json:"ram_total_gb"`
}

// 3. Khai báo tên định danh của Log
func (s *InventorySensor) Name() string {
	return "inventory"
}

// 4. Đưa logic cũ vào hàm Collect()
func (s *InventorySensor) Collect() (interface{}, error) {
	hInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy thông tin host: %w", err)
	}
	cpuInfo, err := cpu.Info()
	if err != nil {
		// Có thể không phải lỗi nghiêm trọng, vẫn tiếp tục với CPU Unknown
		cpuInfo = []cpu.InfoStat{}
	}
	vMem, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy thông tin RAM: %w", err)
	}

	model := "Unknown CPU"
	if len(cpuInfo) > 0 {
		model = cpuInfo[0].ModelName
	}

	displayOS := hInfo.Platform + " " + hInfo.PlatformVersion
	if runtime.GOOS == "windows" {
		displayOS = "Windows " + hInfo.PlatformVersion // Logic giản lược
	}

	return InventoryRecord{
		OSInfo:     displayOS,
		CPUModel:   model,
		RAMTotalGB: vMem.Total / 1024 / 1024 / 1024,
	}, nil
}
