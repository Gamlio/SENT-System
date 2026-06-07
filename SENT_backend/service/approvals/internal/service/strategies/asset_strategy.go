package strategies

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type assetEnrollStrategy struct{}

type AssetEnrollPayload struct {
	AssetHWID    string `json:"asset_hwid"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	TentativeKey string `json:"tentative_key"`
}

func (s *assetEnrollStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	var payload AssetEnrollPayload
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &payload); err != nil {
		return err
	}
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	baselineEndTime := time.Now().In(loc).Add(15 * time.Minute)

	result := tx.Model(&models.Asset{}).
		Where("asset_hwid = ? AND org_id = ?", payload.AssetHWID, ticket.OrgID).
		Updates(map[string]interface{}{
			"status":         "ACTIVE",
			"secret_key":     payload.TentativeKey,
			"approved_by":    ticket.ReviewedBy,
			"last_seen":      time.Now(),
			"baseline_until": baselineEndTime,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("không tìm thấy máy trạm PENDING với HWID %s để phê duyệt", payload.AssetHWID)
	}

	// [FIX] Cập nhật ngay lập tức vào Redis Cache để tránh lỗi 403 sau khi hết hạn 1 tiếng.
	// Việc này đảm bảo SecretKey trong RAM (Cache) và Database luôn đồng nhất ngay sau khi Admin phê duyệt.
	cache.SetAssetCache(payload.AssetHWID, payload.TentativeKey, ticket.OrgID, 1*time.Hour)

	return nil
}

func (s *assetEnrollStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return nil
}

type assetDeleteStrategy struct{}

func (s *assetDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return nil
}

type AssetDeleteSnapshot struct {
	AssetHWID string `json:"asset_hwid"`
	Reason    string `json:"reason"`
}

func (s *assetDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	var snap AssetDeleteSnapshot
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &snap); err != nil {
		snap.AssetHWID = ticket.TargetName
	}

	err := tx.Model(&models.Asset{}).Where("asset_hwid = ? AND org_id = ?", snap.AssetHWID, ticket.OrgID).
		Updates(map[string]interface{}{
			"status":     "RETIRED",
			"secret_key": "",
		}).Error
	if err != nil {
		return err
	}

	err = tx.Where("asset_hwid = ? AND org_id = ? AND created_by = ?", snap.AssetHWID, ticket.OrgID, "System_Baseline_Engine").
		Delete(&models.Policy{}).Error
	if err != nil {
		return fmt.Errorf("lỗi dọn dẹp chính sách baseline của thiết bị: %w", err)
	}

	go func(hwid string, orgID uint) {
		payload, _ := json.Marshal(map[string]interface{}{
			"hwids":  []string{hwid},
			"org_id": orgID,
		})
		// Gọi đến cổng 8002 của Asset Service
		url := "http://asset-service:8000/api/v1/assets/internal/cleanup"
		http.Post(url, "application/json", bytes.NewBuffer(payload))
	}(snap.AssetHWID, ticket.OrgID)

	return nil
}

type assetBulkDeleteStrategy struct{}

func (s *assetBulkDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return nil
}

type BulkDeleteSnapshot struct {
	AssetIDs []string `json:"hwids"`
	Reason   string   `json:"reason"`
}

func (s *assetBulkDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	var snap BulkDeleteSnapshot
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &snap); err != nil {
		return err
	}

	err := tx.Model(&models.Asset{}).Where("asset_hwid IN ? AND org_id = ?", snap.AssetIDs, ticket.OrgID).
		Updates(map[string]interface{}{
			"status":     "RETIRED",
			"secret_key": "",
		}).Error
	if err != nil {
		return err
	}

	// Dọn dẹp các luật baseline của các máy bị xóa
	err = tx.Where("asset_hwid IN ? AND org_id = ? AND created_by = ?", snap.AssetIDs, ticket.OrgID, "System_Baseline_Engine").
		Delete(&models.Policy{}).Error
	if err != nil {
		return fmt.Errorf("lỗi dọn dẹp chính sách baseline của thiết bị: %w", err)
	}

	go func(ids []string, orgID uint) {
		payload, _ := json.Marshal(map[string]interface{}{
			"hwids":  ids,
			"org_id": orgID,
		})
		url := "http://asset-service:8000/api/v1/assets/internal/cleanup"
		http.Post(url, "application/json", bytes.NewBuffer(payload))
	}(snap.AssetIDs, ticket.OrgID)

	return nil
}
