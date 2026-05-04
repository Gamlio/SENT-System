package data // SỬA: Đổi từ asset_data sang data

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

type assetSoftwareRecord struct {
	SoftwareName    string `json:"software_name"`
	Version         string `json:"version"`
	Publisher       string `json:"publisher"`
	InstallLocation string `json:"install_location"`
	FileHash        string `json:"file_hash"`
	Status          string `json:"status"`
	IsRunning       bool   `json:"is_running"`
}

// softwareViolation: Struct để giữ chi tiết vi phạm, phục vụ cho việc gom nhóm (batching).
type softwareViolation struct {
	Name     string
	Type     string // Ví dụ: "Malware Detected", "Software Violation"
	Priority string
}

func ProcessSoftware(asset models.Asset, data interface{}) {
	bytes, err := GetBytesFromData(data)
	if err != nil {
		return
	}

	var records []assetSoftwareRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
		return
	}
	if database.SoftwareCollection == nil {
		return
	}

	incSvc := &incidents.IncidentService{}

	var violations []softwareViolation // Slice để thu thập tất cả vi phạm từ payload này
	var activeHashes []string

	for _, rec := range records {
		activeHashes = append(activeHashes, rec.FileHash)

		// TÍCH HỢP ORG_ID CHỐNG GHI ĐÈ CHÉO TỔ CHỨC
		filter := bson.M{"asset_hwid": asset.AssetHWID, "org_id": asset.OrgID, "software_name": rec.SoftwareName, "file_hash": rec.FileHash}
		update := bson.M{"$set": bson.M{
			"asset_hwid":       asset.AssetHWID,
			"org_id":           asset.OrgID,
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

		// [FIX-EVENT-STORM] Thay vì gọi TriggerSecurityEvent ngay lập tức, chúng ta thu thập các vi phạm.
		if rec.FileHash != "" && checkMaliciousHash(rec.FileHash) {
			violations = append(violations, softwareViolation{Name: rec.SoftwareName, Type: "Malware Detected", Priority: "P1"})
		}
		if rec.Status == "GHOST_REGISTRY" {
			violations = append(violations, softwareViolation{Name: rec.SoftwareName, Type: "Defense Evasion", Priority: "P2"})
		}
		if checkBannedSoftware(rec.SoftwareName) {
			violations = append(violations, softwareViolation{Name: rec.SoftwareName, Type: "Software Violation", Priority: "P3"})
		}
		if asset.IsZeroTrust && !VerifySoftware(rec, asset) {
			violations = append(violations, softwareViolation{Name: rec.SoftwareName, Type: "Zero Trust Violation", Priority: "P1"})
		}
	}

	// [FIX-EVENT-STORM] Nếu có bất kỳ vi phạm nào, xử lý chúng như một sự kiện duy nhất.
	if len(violations) > 0 {
		highestPriority := "P4"
		highestAlertType := "Software Violation" // Mặc định
		priorityMap := map[string]int{"P1": 4, "P2": 3, "P3": 2, "P4": 1}

		groupedViolations := make(map[string][]string)
		for _, v := range violations {
			groupedViolations[v.Type] = append(groupedViolations[v.Type], v.Name)
			// Tìm ra vi phạm có mức độ ưu tiên cao nhất để làm "mồi" cho việc gom nhóm Incident
			if priorityMap[v.Priority] > priorityMap[highestPriority] {
				highestPriority = v.Priority
				highestAlertType = v.Type
			}
		}

		var descriptionBuilder strings.Builder
		descriptionBuilder.WriteString(fmt.Sprintf("Phát hiện %d vi phạm phần mềm trên máy trạm. ", len(violations)))

		// Xây dựng một chuỗi mô tả chi tiết, tổng hợp tất cả các lỗi
		if names, ok := groupedViolations["Malware Detected"]; ok {
			descriptionBuilder.WriteString(fmt.Sprintf("Mã độc (%d): %s. ", len(names), strings.Join(names, ", ")))
		}
		if names, ok := groupedViolations["Zero Trust Violation"]; ok {
			descriptionBuilder.WriteString(fmt.Sprintf("Tiến trình lạ (%d): %s. ", len(names), strings.Join(names, ", ")))
		}
		if names, ok := groupedViolations["Defense Evasion"]; ok {
			descriptionBuilder.WriteString(fmt.Sprintf("Xóa dấu vết (%d): %s. ", len(names), strings.Join(names, ", ")))
		}
		if names, ok := groupedViolations["Software Violation"]; ok {
			descriptionBuilder.WriteString(fmt.Sprintf("Phần mềm cấm (%d): %s. ", len(names), strings.Join(names, ", ")))
		}

		title := fmt.Sprintf("[%s] Phát hiện nhiều vi phạm phần mềm", highestPriority)

		// Chỉ gọi TriggerSecurityEvent một lần duy nhất với dữ liệu đã được tổng hợp.
		incSvc.TriggerSecurityEvent(context.TODO(), asset, highestAlertType, title, descriptionBuilder.String(), highestPriority)
	}

	// [TỐI ƯU DIFFING] Thay vì kéo toàn bộ DB lên RAM rồi lặp (N+1 Query), sử dụng Bulk Update với $nin
	if len(activeHashes) > 0 {
		_, _ = database.SoftwareCollection.UpdateMany(context.TODO(), bson.M{
			"asset_hwid": asset.AssetHWID,
			"org_id":     asset.OrgID,
			"is_running": true,
			"file_hash":  bson.M{"$nin": activeHashes},
		}, bson.M{"$set": bson.M{"is_running": false}})
	}
}

// Xử lý nạp Baseline ban đầu
func HandleSoftwareBaseline(asset models.Asset, data interface{}) {
	bytes, err := GetBytesFromData(data)
	if err != nil {
		return
	}

	var records []assetSoftwareRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
		return
	}

	if database.SoftwareCollection != nil {
		_, _ = database.SoftwareCollection.DeleteMany(context.TODO(), bson.M{"asset_hwid": asset.AssetHWID, "org_id": asset.OrgID})
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
}

func VerifySoftware(rec assetSoftwareRecord, asset models.Asset) bool {
	var trusted models.WhitelistItem
	err := database.DB.Where("org_id = ? AND type = ? AND value = ?", asset.OrgID, "PUBLISHER", rec.Publisher).First(&trusted).Error
	if err == nil {
		return true
	}
	err = database.DB.Where("org_id = ? AND type = ? AND value = ?", asset.OrgID, "SOFTWARE_HASH", rec.FileHash).First(&trusted).Error
	return err == nil
}

func checkMaliciousHash(hash string) bool {
	return map[string]bool{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855": true}[hash]
}
func checkBannedSoftware(name string) bool {
	return strings.Contains(strings.ToLower(name), "utorrent")
}
