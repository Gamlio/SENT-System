package data // SỬA: Đổi từ asset_data sang data

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm/clause"
)

type assetSoftwareRecord struct {
	SoftwareName    string `json:"software_name"`
	Version         string `json:"version"`
	Publisher       string `json:"publisher"`
	InstallLocation string `json:"install_location"`
	FileHash        string `json:"file_hash"`
	Status          string `json:"status"`
	IsRunning       bool   `json:"is_running"`
}

type softwareViolation struct {
	Name     string
	Type     string // Ví dụ: "Malware Detected", "Software Violation"
	Priority string
}

func ProcessSoftware(asset models.Asset, data interface{}) error {
	bytesData, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu đầu vào không hợp lệ: %w", err)
	}
	log.Printf("[DEBUG] HWID: %s | Bytes: %s", asset.AssetHWID, string(bytesData))
	var records []assetSoftwareRecord
	if err := json.Unmarshal(bytesData, &records); err != nil {
		return fmt.Errorf("lỗi giải mã JSON software: %w", err)
	}

	if asset.BaselineUntil != nil && time.Now().Before(*asset.BaselineUntil) {
		var entries []map[string]interface{}
		for _, r := range records {
			if r.FileHash != "" {
				entries = append(entries, map[string]interface{}{
					"file_hash":     r.FileHash,
					"software_name": r.SoftwareName,
					"publisher":     r.Publisher,
					"version":       r.Version,
				})
			}
		}
		SendToPolicyBaseline(asset.OrgID, asset.AssetHWID, "SOFTWARE_HASH", entries)
	}

	if database.SoftwareCollection == nil {
		return fmt.Errorf("database SoftwareCollection chưa sẵn sàng")
	}

	var violations []softwareViolation
	var activeHashes []string

	trustedHashes := make(map[string]bool)
	trustedPubs := make(map[string]bool)
	if asset.IsZeroTrust {
		var hashes, pubs []string
		for _, rec := range records {
			if rec.FileHash != "" {
				hashes = append(hashes, rec.FileHash)
			}
			if rec.Publisher != "" {
				pubs = append(pubs, rec.Publisher)
			}
		}

		// Query approved WHITELIST policies instead of legacy WhitelistItem
		if len(hashes) > 0 {
			var policies []models.Policy
			database.DB.Where("org_id = ? AND policy_type = 'WHITELIST' AND category = ? AND value IN ? AND approval_status = ?", asset.OrgID, "SOFTWARE_HASH", hashes, "APPROVED").Find(&policies)
			for _, p := range policies {
				trustedHashes[p.Value] = true
			}
		}

		if len(pubs) > 0 {
			var policies []models.Policy
			database.DB.Where("org_id = ? AND policy_type = 'WHITELIST' AND category = ? AND value IN ? AND approval_status = ?", asset.OrgID, "PUBLISHER", pubs, "APPROVED").Find(&policies)
			for _, p := range policies {
				trustedPubs[p.Value] = true
			}
		}
	}

	for _, rec := range records {
		activeHashes = append(activeHashes, rec.FileHash)

		// TÍCH HỢP ORG_ID CHỐNG GHI ĐÈ CHÉO TỔ CHỨC
		filter := bson.M{"asset_hwid": asset.AssetHWID, "org_id": int64(asset.OrgID), "software_name": rec.SoftwareName, "file_hash": rec.FileHash}
		update := bson.M{"$set": bson.M{
			"asset_hwid":       asset.AssetHWID,
			"org_id":           int64(asset.OrgID),
			"software_name":    rec.SoftwareName,
			"version":          rec.Version,
			"publisher":        rec.Publisher,
			"install_location": rec.InstallLocation,
			"file_hash":        rec.FileHash,
			"status":           rec.Status,
			"is_running":       rec.IsRunning,
			"updated_at":       time.Now(),
		}}

		if _, err := database.SoftwareCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true)); err != nil {
			return fmt.Errorf("lỗi cập nhật MongoDB (UpdateOne): %w", err)
		}

		if rec.FileHash != "" && checkMaliciousHash(rec.FileHash) {
			violations = append(violations, softwareViolation{Name: rec.SoftwareName, Type: "Malware Detected", Priority: "P1"})
		}
		if rec.Status == "GHOST_REGISTRY" {
			violations = append(violations, softwareViolation{Name: rec.SoftwareName, Type: "Defense Evasion", Priority: "P2"})
		}
		if checkBannedSoftware(rec.SoftwareName) {
			violations = append(violations, softwareViolation{Name: rec.SoftwareName, Type: "Software Violation", Priority: "P3"})
		}
		if asset.IsZeroTrust && !trustedPubs[rec.Publisher] && !trustedHashes[rec.FileHash] {
			violations = append(violations, softwareViolation{Name: rec.SoftwareName, Type: "Zero Trust Violation", Priority: "P1"})
		}
	}

	if len(violations) > 0 {
		highestPriority := "P3"
		highestAlertType := "Software Violation"
		priorityMap := map[string]int{"P1": 3, "P2": 2, "P3": 1}

		groupedViolations := make(map[string][]string)
		for _, v := range violations {
			groupedViolations[v.Type] = append(groupedViolations[v.Type], v.Name)
			if priorityMap[v.Priority] > priorityMap[highestPriority] {
				highestPriority = v.Priority
				highestAlertType = v.Type
			}
		}

		var descBuilder strings.Builder
		descBuilder.WriteString(fmt.Sprintf("Phát hiện %d vi phạm phần mềm. ", len(violations)))
		for vType, names := range groupedViolations {
			descBuilder.WriteString(fmt.Sprintf("%s: %s. ", vType, strings.Join(names, ", ")))
		}

		SendBehaviorLog(map[string]interface{}{
			"asset":    asset,
			"category": highestAlertType,
			"value":    fmt.Sprintf("%d violations", len(violations)),
			"title":    fmt.Sprintf("[%s] Phát hiện vi phạm phần mềm tổng hợp", highestPriority),
			"desc":     descBuilder.String(),
			"priority": highestPriority,
		})
	}

	if len(activeHashes) > 0 {
		if _, err := database.SoftwareCollection.UpdateMany(context.TODO(), bson.M{
			"asset_hwid": asset.AssetHWID,
			"org_id":     int64(asset.OrgID),
			"is_running": true,
			"file_hash":  bson.M{"$nin": activeHashes},
		}, bson.M{"$set": bson.M{"is_running": false}}); err != nil {
			return fmt.Errorf("lỗi cập nhật trạng thái is_running (UpdateMany): %w", err)
		}
	}

	return nil
}

// Xử lý nạp Baseline ban đầu
func HandleSoftwareBaseline(asset models.Asset, data interface{}) error {
	bytesData, err := GetBytesFromData(data)
	if err != nil {
		return err
	}

	var records []assetSoftwareRecord
	if err := json.Unmarshal(bytesData, &records); err != nil {
		return err
	}

	// [GIẢM TẢI DB] Chuyển SQL Loop thành Insert Batch vào bảng policies (BASELINE)
	var hashPolicies []models.Policy
	var pubPolicies []models.Policy

	for _, rec := range records {
		if rec.FileHash != "" {
			hashPolicies = append(hashPolicies, models.Policy{
				OrgID:          asset.OrgID,
				Title:          fmt.Sprintf("Baseline [%s]: %s", asset.AssetHWID, rec.FileHash),
				Category:       "SOFTWARE_HASH",
				Value:          strings.ToLower(rec.FileHash),
				PolicyType:     "WHITELIST",
				ApprovalStatus: "BASELINE",
				IsActive:       false,
				CreatedBy:      "System_Baseline_Engine",
				AssetHWID:      asset.AssetHWID,
				GroupID:        asset.GroupID,
			})
		}

		if rec.Publisher != "Unsigned" && rec.Publisher != "" {
			pubPolicies = append(pubPolicies, models.Policy{
				OrgID:          asset.OrgID,
				Title:          fmt.Sprintf("Baseline Publisher [%s]: %s", asset.AssetHWID, rec.Publisher),
				Category:       "PUBLISHER",
				Value:          rec.Publisher,
				PolicyType:     "WHITELIST",
				ApprovalStatus: "BASELINE",
				IsActive:       false,
				CreatedBy:      "System_Baseline_Engine",
				AssetHWID:      asset.AssetHWID,
				GroupID:        asset.GroupID,
			})
		}
	}

	tx := database.DB.Begin()
	if len(hashPolicies) > 0 {
		tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "org_id"}, {Name: "category"}, {Name: "policy_type"}, {Name: "value"}, {Name: "asset_hwid"}}, DoNothing: true}).CreateInBatches(hashPolicies, 200)
	}
	if len(pubPolicies) > 0 {
		tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "org_id"}, {Name: "category"}, {Name: "policy_type"}, {Name: "value"}, {Name: "asset_hwid"}}, DoNothing: true}).CreateInBatches(pubPolicies, 200)
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// [ĐỒNG BỘ DỮ LIỆU] Chỉ xóa cấu hình cũ trên MongoDB KHI Postgres đã lưu an toàn cấu hình mới
	if database.SoftwareCollection != nil {
		_, _ = database.SoftwareCollection.DeleteMany(context.TODO(), bson.M{"asset_hwid": asset.AssetHWID, "org_id": int64(asset.OrgID)})
	}
	return nil
}

func checkMaliciousHash(hash string) bool {
	return map[string]bool{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855": true}[hash]
}
func checkBannedSoftware(name string) bool {
	return strings.Contains(strings.ToLower(name), "utorrent")
}
