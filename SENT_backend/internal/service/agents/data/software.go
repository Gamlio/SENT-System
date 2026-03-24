package data // SỬA: Đổi từ agent_data sang data

import (
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
	"strings"

	"gorm.io/gorm/clause"
)

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
	bytes, _ := json.Marshal(data)
	var records []AgentSoftwareRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
		return
	}

	incSvc := &incidents.IncidentService{}
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

		// 1. Kiểm tra Mã độc (Cập nhật gọi incSvc)
		if rec.FileHash != "" && checkMaliciousHash(rec.FileHash) {
			incSvc.TriggerSecurityEvent(agent, "Malware Detected", "[P1] Cảnh báo Mã Độc", fmt.Sprintf("Tiến trình: %s", rec.SoftwareName), "P1")
		}

		// 2. Lẩn tránh (Ghost Registry)
		if rec.Status == "GHOST_REGISTRY" {
			incSvc.TriggerSecurityEvent(agent, "Defense Evasion", "[P2] Xóa dấu vết phần mềm", fmt.Sprintf("Phần mềm: %s", rec.SoftwareName), "P2")
		}

		// 3. Phân mềm cấm
		if checkBannedSoftware(rec.SoftwareName) {
			incSvc.TriggerSecurityEvent(agent, "Software Violation", "[P3] Cài đặt phần mềm cấm", fmt.Sprintf("Phần mềm: %s", rec.SoftwareName), "P3")
		}

		// 4. [ZERO TRUST] Kiểm tra danh sách Whitelist
		if agent.IsZeroTrust {
			if !VerifySoftware(rec, agent) {
				incSvc.TriggerSecurityEvent(agent, "Zero Trust Violation", "[P1] Tiến trình lạ xuất hiện", fmt.Sprintf("Chưa phê duyệt: %s", rec.SoftwareName), "P1")
			}
		}
	}
}

// Xử lý nạp Baseline ban đầu
func HandleSoftwareBaseline(agent models.Agent, data interface{}) {
	bytes, _ := json.Marshal(data)
	var records []AgentSoftwareRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
		return
	}

	tx := database.DB.Begin()
	tx.Where("agent_hw_id = ?", agent.HWID).Delete(&models.SoftwareItem{})

	for _, rec := range records {
		tx.Create(&models.WhitelistItem{
			OrgID: agent.OrgID, AgentHWID: agent.HWID, Type: "SOFTWARE_HASH", Value: rec.FileHash, Description: "Baseline: " + rec.SoftwareName,
		})

		if rec.Publisher != "Unsigned" && rec.Publisher != "" {
			tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.WhitelistItem{
				OrgID: agent.OrgID, Type: "PUBLISHER", Value: rec.Publisher, Description: "Trusted Publisher",
			})
		}
	}

	tx.Model(&agent).Updates(map[string]interface{}{"baseline_status": "COMPLETED", "is_zero_trust": true})
	tx.Commit()
}

func VerifySoftware(rec AgentSoftwareRecord, agent models.Agent) bool {
	var trusted models.WhitelistItem
	err := database.DB.Where("org_id = ? AND type = ? AND value = ?", agent.OrgID, "PUBLISHER", rec.Publisher).First(&trusted).Error
	if err == nil {
		return true
	}
	err = database.DB.Where("org_id = ? AND type = ? AND value = ?", agent.OrgID, "SOFTWARE_HASH", rec.FileHash).First(&trusted).Error
	return err == nil
}

func checkMaliciousHash(hash string) bool {
	return map[string]bool{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855": true}[hash]
}
func checkBannedSoftware(name string) bool {
	return strings.Contains(strings.ToLower(name), "utorrent")
}
