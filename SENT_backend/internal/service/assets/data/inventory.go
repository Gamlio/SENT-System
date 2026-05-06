package data

import (
	"context"
	"encoding/json"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ProcessInventory(asset models.Asset, data interface{}) error {
	if database.AssetInventoryCollection == nil {
		return fmt.Errorf("database AssetInventoryCollection chưa sẵn sàng")
	}

	bytes, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu inventory không hợp lệ: %w", err)
	}

	var invData map[string]interface{}
	if err := json.Unmarshal(bytes, &invData); err != nil {
		return fmt.Errorf("lỗi giải mã JSON inventory: %w", err)
	}

	assetAssetHWID := asset.AssetHWID
	update := bson.M{"$set": bson.M{
		"asset_hwid": assetAssetHWID,
		"org_id":     int64(asset.OrgID), // TÍCH HỢP ORG_ID CHỐNG IDOR MONGODB
		"os_info":    fmt.Sprintf("%v", invData["os_info"]),
		"cpu_model":  fmt.Sprintf("%v", invData["cpu_model"]),
		"updated_at": time.Now(),
	}}

	if ramFloat, ok := invData["ram_total_gb"].(float64); ok {
		update["$set"].(bson.M)["ram_total_gb"] = int(ramFloat)
	}

	if _, err := database.AssetInventoryCollection.UpdateOne(context.TODO(), bson.M{"asset_hwid": assetAssetHWID, "org_id": int64(asset.OrgID)}, update, options.Update().SetUpsert(true)); err != nil {
		return fmt.Errorf("lỗi cập nhật inventory MongoDB: %w", err)
	}
	return nil
}
