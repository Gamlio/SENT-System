package collector

import (
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func CollectSoftware() interface{} {
	if runtime.GOOS == "windows" {
		var list []map[string]string
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
		if err != nil {
			return list
		}
		defer k.Close()

		names, _ := k.ReadSubKeyNames(-1)
		for _, name := range names {
			sk, _ := registry.OpenKey(k, name, registry.QUERY_VALUE)
			displayName, _, _ := sk.GetStringValue("DisplayName")
			displayVersion, _, _ := sk.GetStringValue("DisplayVersion")
			if displayName != "" {
				list = append(list, map[string]string{"software_name": displayName, "version": displayVersion})
			}
			sk.Close()
		}
		return list
	}

	if runtime.GOOS == "linux" {
		cmd := exec.Command("dpkg-query", "-W", "-f=${Package};${Version}\n")
		out, err := cmd.Output()
		if err != nil {
			return []map[string]string{}
		}

		var list []map[string]string
		for _, line := range strings.Split(string(out), "\n") {
			parts := strings.Split(line, ";")
			if len(parts) == 2 {
				list = append(list, map[string]string{"software_name": parts[0], "version": parts[1]})
			}
		}
		return list
	}
	return []map[string]string{}
}
