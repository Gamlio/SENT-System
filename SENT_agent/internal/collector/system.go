package collector

import (
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func CollectInventory() interface{} {
	hInfo, _ := host.Info()
	cpuInfo, _ := cpu.Info()
	vMem, _ := mem.VirtualMemory()

	model := "Unknown CPU"
	if len(cpuInfo) > 0 {
		model = cpuInfo[0].ModelName
	}

	displayOS := hInfo.Platform + " " + hInfo.PlatformVersion
	if runtime.GOOS == "windows" {
		displayOS = "Windows " + hInfo.PlatformVersion // Logic giản lược
	}

	return map[string]interface{}{
		"os_info":      displayOS,
		"cpu_model":    model,
		"ram_total_gb": vMem.Total / 1024 / 1024 / 1024,
	}
}
