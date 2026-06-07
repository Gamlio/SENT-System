package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"SENT_backend/pkg/websocket"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/gorm"
)

type AssetLifecycleService struct{}

func (s *AssetLifecycleService) EnrollWithKey(req models.EnrollRequest, orgID uint, secretKey string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var asset models.Asset
		err := tx.Where("asset_hwid = ?", req.AssetHWID).First(&asset).Error

		if err == nil {
			if asset.OrgID != orgID {
				return fmt.Errorf("Asset %s đã thuộc về tổ chức khác và không thể đăng ký lại", req.AssetHWID)
			}
			if err := tx.Model(&asset).Updates(map[string]interface{}{
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

		snapData, _ := json.Marshal(map[string]interface{}{
			"hostname":      req.Hostname,
			"ip_address":    req.IPAddress,
			"asset_hwid":    req.AssetHWID,
			"is_re_enroll":  err == nil,
			"tentative_key": secretKey,
		})

		ticket := models.ApprovalTicket{
			OrgID:         orgID,
			ModuleType:    "ASSET_ENROLL",
			ActionType:    "ENROLL",
			TargetName:    req.AssetHWID,
			Status:        "PENDING",
			RequestedBy:   "Hệ thống (Tự động)",
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

	// 2. TỰ ĐỘNG: Thông báo Behavior Service; Behavior sẽ chịu trách nhiệm kích hoạt Scoring
	go func(id string) {
		payload := map[string]interface{}{
			"org_id":        orgID,
			"asset_hwid":    id,
			"category":      "DeviceTypeChange",
			"value":         "",
			"title":         "Device type changed",
			"desc":          "Device type update; request centralized scoring",
			"base_priority": "P3",
		}
		b, _ := json.Marshal(payload)
		url := "http://behavior-service:8000/api/v1/behaviors/log"
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
		if err != nil {
			fmt.Printf("⚠️ Lỗi gọi Behavior Service cho máy %s: %v\n", id, err)
			return
		}
		resp.Body.Close()
	}(hwid)

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
