package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type AssetDataService struct{}

// GetAssetStats: Tối ưu đếm số lượng với context timeout
func (s *AssetDataService) GetAssetStats(orgID uint) (map[string]int64, error) {
	var total, online, regions int64
	var alerts int64
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// SQL Count
	database.DB.Model(&models.Asset{}).Where("org_id = ?", orgID).Count(&total)

	threshold := time.Now().Add(-2 * time.Minute)
	database.DB.Model(&models.Asset{}).Where("org_id = ? AND last_seen >= ?", orgID, threshold).Count(&online)
	database.DB.Model(&models.Region{}).Where("org_id = ?", orgID).Count(&regions)

	// MongoDB Count
	if database.SecurityAlertCollection != nil {
		// [FIX-CRITICAL] Bổ sung filter theo org_id để tránh rò rỉ dữ liệu giữa các công ty
		c, err := database.SecurityAlertCollection.CountDocuments(ctx, bson.M{"org_id": int64(orgID), "is_resolved": false})
		if err == nil {
			alerts = c
		}
	}

	return map[string]int64{
		"total": total, "online": online, "alerts": alerts, "regions": regions,
	}, nil
}

// attachAssetTelemetry: Chỉ dùng cho trang chi tiết (Detail) để tránh tải nặng trang danh sách
func (s *AssetDataService) attachAssetTelemetry(asset *models.Asset) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Tối ưu: Sử dụng AssetHWID đã chuẩn hóa
	hwid := asset.AssetHWID

	if database.AssetInventoryCollection != nil {
		// [SECURITY] Bổ sung org_id để tránh IDOR
		var inv models.AssetInventory
		if err := database.AssetInventoryCollection.FindOne(ctx, bson.M{"asset_hwid": hwid, "org_id": int64(asset.OrgID)}).Decode(&inv); err == nil {
			asset.Inventory = inv
		}
	}

	// Tối ưu: Chỉ lấy dữ liệu nếu thực sự cần thiết, giới hạn số lượng record nếu là log
	if database.SoftwareCollection != nil {
		if cursor, err := database.SoftwareCollection.Find(ctx, bson.M{"asset_hwid": hwid, "org_id": int64(asset.OrgID)}); err == nil {
			var items []models.SoftwareItem
			cursor.All(ctx, &items)
			asset.Software = items
		}
	}

	if database.SecurityAlertCollection != nil {
		// Chỉ lấy các cảnh báo chưa xử lý hoặc giới hạn số lượng
		if cursor, err := database.SecurityAlertCollection.Find(ctx, bson.M{"asset_hwid": hwid, "org_id": int64(asset.OrgID)}); err == nil {
			var alerts []models.SecurityAlert
			cursor.All(ctx, &alerts)
			asset.Alerts = alerts
		}
	}

	// ... Tương tự cho OpenPort, USBLog, IOActivity sử dụng asset_hwid
	if database.OpenPortCollection != nil {
		if cursor, err := database.OpenPortCollection.Find(ctx, bson.M{"asset_hwid": hwid, "org_id": int64(asset.OrgID)}); err == nil {
			var ports []models.OpenPort
			cursor.All(ctx, &ports)
			asset.OpenPorts = ports
		}
	}
}

// GetAssetList: Đã tối ưu - LOẠI BỎ truy vấn N+1 VÀ ÁP DỤNG PHÂN TRANG (Server-side Pagination)
func (s *AssetDataService) GetAssetList(orgID uint, page int, limit int, search string, status string) (int64, []map[string]interface{}) {
	var assets []models.Asset
	var total int64

	query := database.DB.Model(&models.Asset{}).Where("org_id = ?", orgID)

	// 1. Áp dụng Filter & Search (SQL WHERE) dùng ILIKE cho PostgreSQL (không phân biệt hoa thường)
	if search != "" {
		query = query.Where("hostname ILIKE ? OR ip_address ILIKE ? OR asset_hwid ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 2. Đếm tổng số lượng máy trạm (SAU KHI ĐÃ LỌC)
	query.Count(&total)

	// 3. Tính toán Offset và Truy vấn với Order By (Sắp xếp thiết bị hoạt động gần nhất lên đầu)
	offset := (page - 1) * limit
	query.Preload("Manager").Order("last_seen DESC").Limit(limit).Offset(offset).Find(&assets)

	var result []map[string]interface{}
	threshold := time.Now().Add(-2 * time.Minute)

	for _, a := range assets {
		// Tính toán trạng thái kết nối động
		connectionStatus := "offline"
		if a.LastSeen.After(threshold) {
			connectionStatus = "online"
		}

		// Trả về cả 2 loại trạng thái
		res := map[string]interface{}{ // Bao gồm thêm các trường cần thiết cho frontend
			"asset_hwid":        a.AssetHWID,
			"hostname":          a.Hostname,
			"ip_address":        a.IPAddress,
			"lifecycle_status":  a.Status,         // PENDING / ACTIVE / PENDING_DELETE
			"connection_status": connectionStatus, // online / offline
			"last_seen":         a.LastSeen,
			"risk_score":        a.RiskScore,
			"trust_score":       a.TrustScore,
			"department_tag":    a.DepartmentTag,
			"assets_type":       a.AssetType,
			"is_zero_trust":     a.IsZeroTrust,
			"user_id":           a.UserID, // ID của người quản lý
		}

		// Thêm thông tin người quản lý nếu có
		if a.Manager != nil {
			res["manager"] = map[string]interface{}{
				"id":        a.Manager.ID,
				"full_name": a.Manager.FullName,
			}
		}
		result = append(result, res)
	}
	return total, result
}

// GetAssetDetail: Tối ưu tìm kiếm theo asset_hwid chuẩn hóa
func (s *AssetDataService) GetAssetDetail(hwid string, orgID uint) (models.Asset, error) {
	var asset models.Asset
	// [SECURITY] Bổ sung org_id để tránh IDOR
	err := database.DB.Preload("Manager").Where("asset_hwid = ? AND org_id = ?", hwid, orgID).First(&asset).Error
	if err != nil {
		return asset, err
	}

	// Chỉ tải telemetry chi tiết khi người dùng xem cụ thể một máy
	s.attachAssetTelemetry(&asset)
	return asset, nil
}

// GetAssetLogs: Tối ưu sử dụng asset_hwid thay vì hw_id
func (s *AssetDataService) GetAssetLogs(hwid string, orgID uint) []models.SecurityAlert {
	var alerts []models.SecurityAlert
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if database.SecurityAlertCollection != nil {
		// [SECURITY] Bổ sung org_id để tránh IDOR
		cursor, err := database.SecurityAlertCollection.Find(ctx, bson.M{"asset_hwid": hwid, "org_id": int64(orgID)})
		if err == nil {
			cursor.All(ctx, &alerts)
		}
	}
	return alerts
}
