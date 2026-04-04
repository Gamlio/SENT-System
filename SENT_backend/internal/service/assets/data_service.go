package assets

import (
	"context"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type AssetDataService struct{}

// GetassetStats: Tính toán số liệu tổng quan cho Dashboard
func (s *AssetDataService) GetassetStats() (map[string]int64, error) {
	var total, online, regions int64
	var alerts int64

	database.DB.Model(&models.Asset{}).Count(&total)

	threshold := time.Now().Add(-2 * time.Minute)
	database.DB.Model(&models.Asset{}).Where("last_seen >= ?", threshold).Count(&online)

	if database.SecurityAlertCollection != nil {
		c, err := database.SecurityAlertCollection.CountDocuments(context.TODO(), bson.M{"is_resolved": false})
		if err != nil {
			return nil, err
		}
		alerts = c
	}

	database.DB.Model(&models.Region{}).Count(&regions)

	return map[string]int64{
		"total": total, "online": online, "alerts": alerts, "regions": regions,
	}, nil
}

func (s *AssetDataService) attachassetTelemetry(asset *models.Asset) {
	ctx := context.TODO()

	if database.AssetInventoryCollection != nil {
		var inv models.AssetInventory
		err := database.AssetInventoryCollection.FindOne(ctx, bson.M{"asset_hwid": asset.HWID}).Decode(&inv)
		if err == nil {
			asset.Inventory = inv
		}
	}

	if database.SoftwareCollection != nil {
		cursor, err := database.SoftwareCollection.Find(ctx, bson.M{"asset_hwid": asset.HWID})
		if err == nil {
			var items []models.SoftwareItem
			cursor.All(ctx, &items)
			asset.Software = items
		}
	}

	if database.SecurityAlertCollection != nil {
		cursor, err := database.SecurityAlertCollection.Find(ctx, bson.M{"hw_id": asset.HWID})
		if err == nil {
			var alerts []models.SecurityAlert
			cursor.All(ctx, &alerts)
			asset.Alerts = alerts
		}
	}

	if database.OpenPortCollection != nil {
		cursor, err := database.OpenPortCollection.Find(ctx, bson.M{"asset_hwid": asset.HWID})
		if err == nil {
			var ports []models.OpenPort
			cursor.All(ctx, &ports)
			asset.OpenPorts = ports
		}
	}

	if database.USBCollection != nil {
		cursor, err := database.USBCollection.Find(ctx, bson.M{"asset_hwid": asset.HWID})
		if err == nil {
			var usb []models.USBLog
			cursor.All(ctx, &usb)
			asset.USBLogs = usb
		}
	}

	if database.AssetIOActivityCollection != nil {
		cursor, err := database.AssetIOActivityCollection.Find(ctx, bson.M{"asset_hwid": asset.HWID})
		if err == nil {
			var ios []models.AssetIOActivity
			cursor.All(ctx, &ios)
			asset.IOActivities = ios
		}
	}
}

// GetassetList: Lấy danh sách máy kèm logic Online/Offline ảo
func (s *AssetDataService) GetassetList(orgID uint) []models.Asset {
	var assets []models.Asset
	database.DB.Preload("Manager").
		Where("org_id = ? AND status != ?", orgID, "RETIRED").
		Order("last_seen desc").Find(&assets)

	threshold := time.Now().Add(-2 * time.Minute)
	for i := range assets {
		if assets[i].Status == "ACTIVE" {
			if assets[i].LastSeen.After(threshold) {
				assets[i].Status = "online"
			} else {
				assets[i].Status = "offline"
			}
		}
		s.attachassetTelemetry(&assets[i])
	}
	return assets
}

// GetassetDetail: Truy vấn sâu 1 máy trạm
func (s *AssetDataService) GetassetDetail(hwid string) (models.Asset, error) {
	var asset models.Asset
	err := database.DB.Preload("Manager").Where("hw_id = ?", hwid).First(&asset).Error
	if err != nil {
		return asset, err
	}

	s.attachassetTelemetry(&asset)
	return asset, nil
}

// GetassetLogs: Lấy lịch sử cảnh báo của 1 máy
func (s *AssetDataService) GetassetLogs(hwid string) []models.SecurityAlert {
	var alerts []models.SecurityAlert
	if database.SecurityAlertCollection != nil {
		cursor, err := database.SecurityAlertCollection.Find(context.TODO(), bson.M{"hw_id": hwid})
		if err == nil {
			cursor.All(context.TODO(), &alerts)
		}
	}
	return alerts
}
