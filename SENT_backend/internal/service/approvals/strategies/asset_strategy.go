package strategies

import (
	"encoding/json"
	"sent_backend/internal/models"
	"sent_backend/internal/service/assets"
	"time"

	"gorm.io/gorm"
)

// --- CHIẾN LƯỢC: ĐĂNG KÝ asset ---
type assetEnrollStrategy struct{}

type AssetEnrollPayload struct {
	AssetHWID string `json:"asset_hwid"`
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip_address"`
	SecretKey string `json:"secret_key"`
}

func (s *assetEnrollStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// 1. Giải mã dữ liệu từ "Đơn" (Snapshot)
	var payload AssetEnrollPayload
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &payload); err != nil {
		return err
	}

	// CẬP NHẬT bản ghi đã có từ lúc Enroll thay vì tạo mới
	return tx.Model(&models.Asset{}).
		Where("asset_hwid = ? AND org_id = ?", payload.AssetHWID, ticket.OrgID).
		Updates(map[string]interface{}{
			"status":      "ACTIVE",
			"approved_by": ticket.ReviewedBy,
			"last_seen":   time.Now(),
			"secret_key":  payload.SecretKey,
		}).Error
}

func (s *assetEnrollStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// THEO TƯ DUY MỚI: Hệ thống SẠCH, không tạo bảng Asset PENDING
	return nil
}

// ==============================================================
// --- CHIẾN LƯỢC: XÓA 1 asset (XÓA MỀM / SOFT DELETE) ---
// ==============================================================
type assetDeleteStrategy struct{}

type AssetDeleteSnapshot struct {
	AssetHWID string `json:"asset_hwid"`
	Reason    string `json:"reason"`
}

func (s *assetDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// 1. Giải mã dữ liệu từ "Đơn" (Snapshot)
	var snap AssetDeleteSnapshot
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &snap); err != nil {
		// Fallback cho các ticket cũ chưa dùng SnapshotData
		snap.AssetHWID = ticket.TargetName
	}

	// 2. Bây giờ mới thực sự thực thi thao tác XÓA MỀM trên Asset
	err := tx.Model(&models.Asset{}).Where("asset_hwid = ?", snap.AssetHWID).
		Updates(map[string]interface{}{
			"status":     "RETIRED",
			"secret_key": "", // Thu hồi khóa, vứt bỏ quyền truy cập
		}).Error
	if err != nil {
		return err
	}
	// Manual cascade cleanup dữ liệu telemetry Mongo
	lifecycleService := assets.AssetLifecycleService{}
	lifecycleService.CleanupassetTelemetry([]string{snap.AssetHWID}, ticket.OrgID)
	return nil
}

func (s *assetDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// THEO TƯ DUY MỚI: Hệ thống SẠCH.
	// Lúc tạo đơn (Ticket), ta chưa hề đụng vào bảng Asset (không đổi thành PENDING_DELETE).
	// Nên khi bị từ chối, ta KHÔNG CẦN ROLLBACK hay khôi phục gì cả.
	return nil
}

// ==============================================================
// --- CHIẾN LƯỢC: XÓA asset HÀNG LOẠT (SOFT DELETE) ---
// ==============================================================
type assetBulkDeleteStrategy struct{}

// Struct dùng để giải mã SnapshotData
type BulkDeleteSnapshot struct {
	AssetIDs []string `json:"hwids"`
	Reason   string   `json:"reason"`
}

func (s *assetBulkDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	var snap BulkDeleteSnapshot
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &snap); err != nil {
		return err
	}

	// XÓA MỀM HÀNG LOẠT: Đổi trạng thái và thu hồi khóa của hàng trăm máy trong 1 nốt nhạc
	err := tx.Model(&models.Asset{}).Where("asset_hwid IN ?", snap.AssetIDs).
		Updates(map[string]interface{}{
			"status":     "RETIRED",
			"secret_key": "",
		}).Error
	if err != nil {
		return err
	}
	lifecycleService := assets.AssetLifecycleService{}
	lifecycleService.CleanupassetTelemetry(snap.AssetIDs, ticket.OrgID)
	return nil
}

func (s *assetBulkDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// THEO TƯ DUY MỚI: Không cần rollback trạng thái về ACTIVE nữa
	// vì trong thời gian chờ duyệt, trạng thái Asset vẫn chưa bị thay đổi.
	return nil
}
