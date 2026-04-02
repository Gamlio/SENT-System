package data // SỬA: Đổi từ agent_data sang data

import (
	"context"
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/incidents"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	if database.SoftwareCollection == nil {
		return
	}

	incSvc := &incidents.IncidentService{}

	// 1. TẠO MAP ĐỂ LƯU CÁC TIẾN TRÌNH ĐANG CHẠY (Dùng Tên + Hash làm khóa để tránh trùng)
	incomingSoftwareMap := make(map[string]bool)

	for _, rec := range records {
		uniqueKey := rec.SoftwareName + "_" + rec.FileHash
		incomingSoftwareMap[uniqueKey] = true

		filter := bson.M{"agent_hwid": agent.HWID, "software_name": rec.SoftwareName, "file_hash": rec.FileHash}
		update := bson.M{"$set": bson.M{
			"agent_hwid":       agent.HWID,
			"software_name":    rec.SoftwareName,
			"version":          rec.Version,
			"publisher":        rec.Publisher,
			"install_location": rec.InstallLocation,
			"file_hash":        rec.FileHash,
			"status":           rec.Status,
			"is_running":       rec.IsRunning,
			"updated_at":       time.Now(),
		}}

		_, _ = database.SoftwareCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))

		// --- CHỈ BÁO ĐỘNG KHI CÓ PHẦN MỀM MỚI ---
		if rec.FileHash != "" && checkMaliciousHash(rec.FileHash) {
			incSvc.TriggerSecurityEvent(agent, "Malware Detected", "[P1] Cảnh báo Mã Độc", fmt.Sprintf("Tiến trình: %s", rec.SoftwareName), "P1")
		}
		if rec.Status == "GHOST_REGISTRY" {
			incSvc.TriggerSecurityEvent(agent, "Defense Evasion", "[P2] Xóa dấu vết phần mềm", fmt.Sprintf("Phần mềm: %s", rec.SoftwareName), "P2")
		}
		if checkBannedSoftware(rec.SoftwareName) {
			incSvc.TriggerSecurityEvent(agent, "Software Violation", "[P3] Cài đặt phần mềm cấm", fmt.Sprintf("Phần mềm: %s", rec.SoftwareName), "P3")
		}
		if agent.IsZeroTrust && !VerifySoftware(rec, agent) {
			incSvc.TriggerSecurityEvent(agent, "Zero Trust Violation", "[P1] Tiến trình lạ xuất hiện", fmt.Sprintf("Chưa phê duyệt: %s", rec.SoftwareName), "P1")
		}
	}

	// 3. CHIỀU ĐÓNG (DIFFING): Tìm các phần mềm vừa bị tắt
	cursor, err := database.SoftwareCollection.Find(context.TODO(), bson.M{"agent_hwid": agent.HWID, "is_running": true})
	if err != nil {
		return
	}
	var activeSoftwares []models.SoftwareItem
	cursor.All(context.TODO(), &activeSoftwares)

	for _, dbSoft := range activeSoftwares {
		uniqueKey := dbSoft.SoftwareName + "_" + dbSoft.FileHash
		if !incomingSoftwareMap[uniqueKey] {
			_, _ = database.SoftwareCollection.UpdateOne(context.TODO(), bson.M{"_id": dbSoft.ID}, bson.M{"$set": bson.M{"is_running": false}})
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

	if database.SoftwareCollection != nil {
		_, _ = database.SoftwareCollection.DeleteMany(context.TODO(), bson.M{"agent_hwid": agent.HWID})
	}

	// Whitelist vẫn lưu ở Postgres
	tx := database.DB.Begin()
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
