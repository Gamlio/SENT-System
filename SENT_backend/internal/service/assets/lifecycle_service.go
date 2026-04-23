package assets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"sent_backend/internal/service/scoring"
	"sent_backend/internal/websocket"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

type AssetLifecycleService struct{}

// EnrollWithKey: Xử lý đăng ký máy, tạo vé duyệt và lưu Secret Key
func (s *AssetLifecycleService) EnrollWithKey(req models.EnrollRequest, orgID uint, secretKey string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var asset models.Asset
		// Sửa lỗi Duplicate Key: Chỉ tìm theo asset_hwid vì đây là khóa chính (Primary Key).
		err := tx.Where("asset_hwid = ?", req.AssetHWID).First(&asset).Error

		if err == nil {
			// Máy cũ (có thể thuộc Org khác hoặc đã bị RETIRED): Cập nhật lại org_id, đưa về PENDING
			if err := tx.Model(&asset).Updates(map[string]interface{}{
				"org_id":     orgID,
				"status":     "PENDING",
				"hostname":   req.Hostname,
				"ip_address": req.IPAddress,
				"secret_key": secretKey,
			}).Error; err != nil {
				return err
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// Máy mới: Tạo mới hoàn toàn với Secret Key
			asset = models.Asset{
				AssetHWID: req.AssetHWID, Hostname: req.Hostname, IPAddress: req.IPAddress,
				OrgID: orgID, Status: "PENDING", LastSeen: time.Now(), SecretKey: secretKey,
			}
			if err := tx.Create(&asset).Error; err != nil {
				return err
			}
		} else {
			// Lỗi DB khác
			return err
		}

		// Tạo đơn phê duyệt
		snapData, _ := json.Marshal(map[string]interface{}{
			"hostname":     req.Hostname,
			"ip_address":   req.IPAddress,
			"asset_hwid":   req.AssetHWID, // Sửa key từ "hwid" -> "asset_hwid"
			"is_re_enroll": err == nil,    // True if the device is re-enrolling
		})

		ticket := models.ApprovalTicket{
			OrgID:         orgID,
			ModuleType:    "ASSET_ENROLL",
			ActionType:    "ENROLL",
			TargetName:    req.AssetHWID,
			Status:        "PENDING",
			RequestedBy:   "Hệ thống (Tự động)", // Tên người tạo đơn mặc định cho Enroll
			RequestReason: "Thiết bị xin gia nhập hệ thống thông qua mã cài đặt",
			SnapshotData:  string(snapData),
		}
		return tx.Create(&ticket).Error
	})
}

// CreateDeleteRequest: Logic Maker-Checker cho việc xóa 1 máy đơn lẻ
func (s *AssetLifecycleService) CreateDeleteRequest(hwid string, orgID uint, reason string, requester string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// THEO TƯ DUY MỚI: Không update status thành PENDING_DELETE để giữ hệ thống sạch
		// Chỉ tạo Ticket phê duyệt

		// 1. Tạo Snapshot cho 1 máy duy nhất
		snap, _ := json.Marshal(map[string]interface{}{
			"asset_hwid": hwid,
			"reason":     reason,
		})

		ticket := models.ApprovalTicket{
			OrgID:         orgID,
			ModuleType:    "ASSET_DELETE", // Khớp với chiến lược đơn lẻ (assetDeleteStrategy)
			ActionType:    "DELETE",
			TargetName:    hwid, // Tên máy cụ thể
			Status:        "PENDING",
			RequestedBy:   requester,
			RequestReason: reason,
			SnapshotData:  string(snap),
		}

		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}

		// 2. Thông báo Real-time cho Admin qua WebSocket
		websocket.GlobalHub.BroadcastToOrg(orgID, map[string]interface{}{"type": "REFRESH_ASSET_LIST"})
		return nil
	})
}

// CreateBulkDeleteRequest: Logic Maker-Checker cho việc xóa máy
func (s *AssetLifecycleService) CreateBulkDeleteRequest(hwids []string, orgID uint, reason string, requester string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// THEO TƯ DUY MỚI: Không update status thành PENDING_DELETE để giữ hệ thống sạch
		// Chỉ tạo Ticket phê duyệt

		// 1. Tạo Ticket
		snap, _ := json.Marshal(map[string]interface{}{"hwids": hwids, "reason": reason})
		ticket := models.ApprovalTicket{
			OrgID:         orgID,
			ModuleType:    "ASSET_BULK_DELETE",
			ActionType:    "DELETE",
			TargetName:    fmt.Sprintf("Xóa %d máy trạm", len(hwids)),
			Status:        "PENDING",
			RequestedBy:   requester,
			RequestReason: reason,
			SnapshotData:  string(snap),
		}

		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}

		// 2. Thông báo Real-time cho Admin qua WebSocket
		websocket.GlobalHub.BroadcastToOrg(orgID, map[string]interface{}{"type": "REFRESH_ASSET_LIST"})
		return nil
	})
}
func (s *AssetLifecycleService) UpdateDeviceType(hwid string, orgID uint, deviceType string) error {
	// 1. Cập nhật thông tin trong Database
	// [SECURITY] Bổ sung org_id để tránh IDOR
	err := database.DB.Model(&models.Asset{}).
		Where("asset_hwid = ? AND org_id = ?", hwid, orgID).
		Update("device_type", deviceType).Error

	if err != nil {
		return err
	}

	// 2. TỰ ĐỘNG: Tính lại điểm rủi ro ngay vì DeviceType làm thay đổi trọng số tài sản
	// Chúng ta dùng Goroutine để không làm chậm phản hồi của API
	go scoring.RecalculateRiskScore(hwid)

	return nil
}

// AssignManager: Gán nhân sự phụ trách máy
func (s *AssetLifecycleService) AssignManager(hwid string, orgID uint, userID uint) error {
	return database.DB.Model(&models.Asset{}).
		Where("asset_hwid = ? AND org_id = ?", hwid, orgID).
		Update("user_id", userID).Error

}

func (s *AssetLifecycleService) CleanupassetTelemetry(hwids []string, orgID uint) {
	ctx := context.TODO()
	// [SECURITY] Bổ sung org_id để tránh IDOR khi xóa dữ liệu trên MongoDB
	filter := bson.M{"asset_hwid": bson.M{"$in": hwids}, "org_id": int64(orgID)}

	// SOFT DELETE: Thay vì DeleteMany, ta dùng UpdateMany để đánh dấu là đã lưu trữ (Archived)
	update := bson.M{"$set": bson.M{"is_archived": true, "archived_at": time.Now()}}

	if database.SoftwareCollection != nil {
		_, _ = database.SoftwareCollection.UpdateMany(ctx, filter, update)
	}
	if database.USBCollection != nil {
		_, _ = database.USBCollection.UpdateMany(ctx, filter, update)
	}
	if database.OpenPortCollection != nil {
		_, _ = database.OpenPortCollection.UpdateMany(ctx, filter, update)
	}
	if database.AssetInventoryCollection != nil {
		_, _ = database.AssetInventoryCollection.UpdateMany(ctx, filter, update)
	}
	if database.AssetIOActivityCollection != nil {
		_, _ = database.AssetIOActivityCollection.UpdateMany(ctx, filter, update)
	}
	if database.SecurityAlertCollection != nil {
		// [FIX] Sử dụng chung filter đã được gia cố bảo mật
		_, _ = database.SecurityAlertCollection.UpdateMany(ctx, filter, update)
	}
}
