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

func (s *assetEnrollStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return tx.Model(&models.Asset{}).Where("hw_id = ?", ticket.TargetName).
		Updates(map[string]interface{}{
			"status":    "ACTIVE",
			"last_seen": time.Now(), // <--- BUMP NÓ LÊN TRÊN CÙNG
		}).Error
}
func (s *assetEnrollStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	return tx.Model(&models.Asset{}).Where("hw_id = ?", ticket.TargetName).
		Updates(map[string]interface{}{
			"status":    "REJECTED",
			"last_seen": time.Now(), // <--- BUMP NÓ LÊN TRÊN CÙNG
		}).Error
}

// ==============================================================
// --- CHIẾN LƯỢC: XÓA 1 asset (XÓA MỀM / SOFT DELETE) ---
// ==============================================================
type assetDeleteStrategy struct{}

func (s *assetDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// XÓA MỀM: Đổi trạng thái thành RETIRED, xóa SecretKey để chặn kết nối vĩnh viễn
	err := tx.Model(&models.Asset{}).Where("hw_id = ?", ticket.TargetName).
		Updates(map[string]interface{}{
			"status":     "RETIRED",
			"secret_key": "", // Thu hồi khóa, vứt bỏ quyền truy cập
		}).Error
	if err != nil {
		return err
	}
	// Manual cascade cleanup dữ liệu telemetry Mongo
	lifecycleService := assets.AssetLifecycleService{}
	lifecycleService.CleanupassetTelemetry([]string{ticket.TargetName})
	return nil
}

func (s *assetDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// BỊ TỪ CHỐI XÓA: Khôi phục máy trạm từ PENDING_DELETE về trạng thái ACTIVE
	return tx.Model(&models.Asset{}).Where("hw_id = ?", ticket.TargetName).Update("status", "ACTIVE").Error
}

// ==============================================================
// --- CHIẾN LƯỢC: XÓA asset HÀNG LOẠT (SOFT DELETE) ---
// ==============================================================
type assetBulkDeleteStrategy struct{}

// Struct dùng để giải mã SnapshotData
type BulkDeleteSnapshot struct {
	HWIDs  []string `json:"hwids"`
	Reason string   `json:"reason"`
}

func (s *assetBulkDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	var snap BulkDeleteSnapshot
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &snap); err != nil {
		return err
	}

	// XÓA MỀM HÀNG LOẠT: Đổi trạng thái và thu hồi khóa của hàng trăm máy trong 1 nốt nhạc
	err := tx.Model(&models.Asset{}).Where("hw_id IN ?", snap.HWIDs).
		Updates(map[string]interface{}{
			"status":     "RETIRED",
			"secret_key": "",
		}).Error
	if err != nil {
		return err
	}
	lifecycleService := assets.AssetLifecycleService{}
	lifecycleService.CleanupassetTelemetry(snap.HWIDs)
	return nil
}

func (s *assetBulkDeleteStrategy) OnReject(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	var snap BulkDeleteSnapshot
	if err := json.Unmarshal([]byte(ticket.SnapshotData), &snap); err != nil {
		return err
	}

	// BỊ TỪ CHỐI XÓA HÀNG LOẠT: Khôi phục toàn bộ danh sách máy về ACTIVE
	return tx.Model(&models.Asset{}).Where("hw_id IN ?", snap.HWIDs).Update("status", "ACTIVE").Error
}
