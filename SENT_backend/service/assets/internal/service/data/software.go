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
		var hashes []string
		for _, r := range records {
			if r.FileHash != "" {
				hashes = append(hashes, r.FileHash)
			}
		}
		SendToPolicyBaseline(asset.OrgID, asset.AssetHWID, "SOFTWARE_HASH", hashes)
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
		var wlItems []models.WhitelistItem
		if len(hashes) > 0 {
			database.DB.Where("org_id = ? AND type = 'SOFTWARE_HASH' AND value IN ?", asset.OrgID, hashes).Find(&wlItems)
			for _, item := range wlItems {
				trustedHashes[item.Value] = true
			}
		}
		wlItems = nil
		if len(pubs) > 0 {
			database.DB.Where("org_id = ? AND type = 'PUBLISHER' AND value IN ?", asset.OrgID, pubs).Find(&wlItems)
			for _, item := range wlItems {
				trustedPubs[item.Value] = true
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

	// [GIẢM TẢI DB] Chuyển SQL Loop thành Insert Batch
	var hashItems []models.WhitelistItem
	var pubItems []models.WhitelistItem

	for _, rec := range records {
		hashItems = append(hashItems, models.WhitelistItem{
			OrgID: asset.OrgID, AssetHWID: asset.AssetHWID, Type: "SOFTWARE_HASH", Value: rec.FileHash, Description: "Baseline: " + rec.SoftwareName,
		})

		if rec.Publisher != "Unsigned" && rec.Publisher != "" {
			pubItems = append(pubItems, models.WhitelistItem{
				OrgID: asset.OrgID, AssetHWID: asset.AssetHWID, Type: "PUBLISHER", Value: rec.Publisher, Description: "Trusted Publisher",
			})
		}
	}

	tx := database.DB.Begin()
	tx.CreateInBatches(hashItems, 200) // Chunk size 200 an toàn cho Postgres
	if len(pubItems) > 0 {
		tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(pubItems, 200)
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
