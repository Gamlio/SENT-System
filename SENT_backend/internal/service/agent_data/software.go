package agent_data

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/security"
	"strings"
)

// Khớp 100% với Struct gửi từ Agent
type AgentSoftwareRecord struct {
	SoftwareName    string `json:"software_name"`
	Version         string `json:"version"`
	Publisher       string `json:"publisher"`
	InstallLocation string `json:"install_location"`
	FileHash        string `json:"file_hash"`
	Status          string `json:"status"`
	IsRunning       bool   `json:"is_running"`
}

func ProcessSoftware(agent models.Agent, data interface{}) {
	// Ép kiểu an toàn từ interface{} sang Struct
	bytes, _ := json.Marshal(data)
	var records []AgentSoftwareRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
		return
	}

	// Xóa dữ liệu cũ để cập nhật mới
	database.DB.Where("agent_hw_id = ?", agent.HWID).Delete(&models.SoftwareItem{})

	for _, rec := range records {
		dbItem := models.SoftwareItem{
			AgentHWID:       agent.HWID,
			SoftwareName:    rec.SoftwareName,
			Version:         rec.Version,
			Publisher:       rec.Publisher,
			InstallLocation: rec.InstallLocation,
			FileHash:        rec.FileHash,
			Status:          rec.Status,
			IsRunning:       rec.IsRunning,
		}
		database.DB.Create(&dbItem)

		// --- LUỒNG TRIAGE TỰ ĐỘNG BẮT BỆNH ---

		// 1. MÃ ĐỘC (YARA/Threat Intel Match)
		if rec.FileHash != "" && checkMaliciousHash(rec.FileHash) {
			security.TriggerSecurityEvent(agent,
				"Malware Detected",
				"[P1] Cảnh báo Mã Độc",
				fmt.Sprintf("Tiến trình '%s' chứa mã băm độc hại: %s", rec.SoftwareName, rec.FileHash),
				"P1",
			)
		}

		// 2. LẨN TRÁNH (Ghost Registry)
		if rec.Status == "GHOST_REGISTRY" {
			security.TriggerSecurityEvent(agent,
				"Defense Evasion",
				"[P2] Xóa dấu vết phần mềm",
				fmt.Sprintf("Phần mềm '%s' bị xóa vật lý nhưng vẫn giữ Registry để lẩn tránh.", rec.SoftwareName),
				"P2",
			)
		}

		// 3. PHẦN MỀM CẤM
		if checkBannedSoftware(rec.SoftwareName) {
			security.TriggerSecurityEvent(agent,
				"Software Violation",
				"[P3] Cài đặt phần mềm bị cấm",
				fmt.Sprintf("Phát hiện phần mềm vi phạm nội quy: %s", rec.SoftwareName),
				"P3",
			)
		}
	}

	// Vẫn giữ lại hàm check Policy động (nếu có)
	go security.CheckSoftwareCompliance(agent, data)
}

// Mock kiểm tra Hash và tên
func checkMaliciousHash(hash string) bool {
	badHashes := map[string]bool{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855": true}
	return badHashes[hash]
}

func checkBannedSoftware(name string) bool {
	badNames := []string{"utorrent", "cheat engine"}
	nameLower := strings.ToLower(name)
	for _, b := range badNames {
		if strings.Contains(nameLower, b) {
			return true
		}
	}
	return false
}
