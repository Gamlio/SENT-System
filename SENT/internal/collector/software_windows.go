//go:build windows

package collector

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows/registry"
)

func getOSSoftware(runningProcs map[string]bool) ([]SoftwareRecord, error) {
	var softwareList []SoftwareRecord
	newMetas := make(map[string]FileMetadata) // Map lưu các file cần update hash
	processedExes := make(map[string]bool)
	paths := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, path := range paths {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
		if err != nil {
			// Nếu không mở được key registry chính, có thể bỏ qua (ví dụ WOW6432Node không tồn tại trên HĐH 32-bit)
			continue
		}
		defer k.Close()

		names, err := k.ReadSubKeyNames(-1)
		if err != nil {
			log.Printf("Lỗi đọc subkey từ '%s': %v", path, err) // Ghi log và tiếp tục
			continue
		}

		for _, name := range names {
			sk, err := registry.OpenKey(k, name, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			displayName, _, _ := sk.GetStringValue("DisplayName")
			displayVersion, _, _ := sk.GetStringValue("DisplayVersion")
			publisher, _, _ := sk.GetStringValue("Publisher")
			installLocation, _, _ := sk.GetStringValue("InstallLocation")
			displayIcon, _, _ := sk.GetStringValue("DisplayIcon")

			if displayName != "" {
				status := "INSTALLED"
				fileHash := ""

				if installLocation != "" {
					cleanPath := strings.Trim(installLocation, "\"")
					if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
						status = "GHOST_REGISTRY"
					}
				}

				if displayIcon != "" {
					exePath := strings.Split(displayIcon, ",")[0]
					exePath = strings.Trim(exePath, "\"")
					if strings.HasSuffix(strings.ToLower(exePath), ".exe") {
						// [SMART-DELTA] Start of the new logic
						fileInfo, err := os.Stat(exePath)
						if err == nil {
							oldMeta, found := GetFileMetadata(exePath)

							if !found || fileInfo.ModTime() != oldMeta.ModTime || fileInfo.Size() != oldMeta.Size {
								// File is new or has changed, so re-hash it.
								hash, err := calculateSHA256(exePath)
								if err == nil {
									fileHash = hash
									// Update metadata in DB after successful hashing
									newMetas[exePath] = FileMetadata{ModTime: fileInfo.ModTime(), Size: fileInfo.Size(), Hash: hash}
								}
							} else {
								// File has not changed, use the cached hash.
								fileHash = oldMeta.Hash
							}
							processedExes[strings.ToLower(exePath)] = true
						}
					}
				}

				isRunning := isProcessRunning(displayName, runningProcs)

				// Nếu app đang chạy thực tế, không thể coi là GHOST_REGISTRY dù thiếu thư mục cài đặt
				if status == "GHOST_REGISTRY" && isRunning {
					status = "INSTALLED"
				}

				softwareList = append(softwareList, SoftwareRecord{
					SoftwareName:    displayName,
					Version:         displayVersion,
					Publisher:       publisher,
					InstallLocation: installLocation,
					FileHash:        fileHash,
					Status:          status,
					IsRunning:       isRunning,
				})
			}
			sk.Close()
		}
	}

	// Bổ sung: Bắt các Portable Binaries / Malware đang chạy ngầm nhưng trốn Registry
	procs, _ := process.Processes()
	for _, p := range procs {
		exePath, err := p.Exe()
		if err != nil || exePath == "" {
			continue
		}

		lowerExe := strings.ToLower(exePath)
		if processedExes[lowerExe] || strings.HasPrefix(lowerExe, `c:\windows\`) {
			continue // Bỏ qua file đã xử lý hoặc System Process gốc để tránh Spam
		}

		if fileInfo, err := os.Stat(exePath); err == nil {
			fileHash := ""
			oldMeta, found := GetFileMetadata(exePath)
			if !found || fileInfo.ModTime() != oldMeta.ModTime || fileInfo.Size() != oldMeta.Size {
				if hash, err := calculateSHA256(exePath); err == nil {
					fileHash = hash
					newMetas[exePath] = FileMetadata{ModTime: fileInfo.ModTime(), Size: fileInfo.Size(), Hash: hash}
				}
			} else {
				fileHash = oldMeta.Hash
			}
			name, _ := p.Name()
			softwareList = append(softwareList, SoftwareRecord{
				SoftwareName:    name,
				Version:         "Unknown",
				Publisher:       "Running Binary",
				InstallLocation: filepath.Dir(exePath),
				FileHash:        fileHash,
				Status:          "PORTABLE",
				IsRunning:       true,
			})
			processedExes[lowerExe] = true
		}
	}

	// Ghi đĩa 1 lần duy nhất cho toàn bộ các phần mềm mới/có thay đổi
	if len(newMetas) > 0 {
		UpdateFileMetadataBatch(newMetas)
	}

	return softwareList, nil
}
