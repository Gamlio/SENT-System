package data // SỬA: Đổi từ asset_data sang data

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

func ProcessSoftware(asset models.Asset, data interface{}) error {
	bytes, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu đầu vào không hợp lệ: %w", err)
	}
	log.Printf("[DEBUG] HWID: %s | Bytes: %s", asset.AssetHWID, string(bytes))
	var records []assetSoftwareRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
		return fmt.Errorf("lỗi giải mã JSON software: %w", err)
	}
	if database.SoftwareCollection == nil {
		return fmt.Errorf("database SoftwareCollection chưa sẵn sàng")
	}

	incSvc := &incidents.IncidentService{}

	var violations []softwareViolation // Slice để thu thập tất cả vi phạm từ payload này
	var activeHashes []string

	// [TỐI ƯU HIỆU NĂNG] Xử lý N+1 Query: Truy vấn PostgreSQL 1 lần duy nhất bằng IN clause
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
		// Tra cứu từ Hash Map siêu tốc (O(1)) thay vì truy vấn SQL
		if asset.IsZeroTrust && !trustedPubs[rec.Publisher] && !trustedHashes[rec.FileHash] {
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
	bytes, err := GetBytesFromData(data)
	if err != nil {
		return err
	}

	var records []assetSoftwareRecord
	if err := json.Unmarshal(bytes, &records); err != nil {
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
