package data

import (
	"context"
	"fmt"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ProcessInventory(asset models.Asset, data interface{}) {
	invData, ok := data.(map[string]interface{})
	if !ok || database.AssetInventoryCollection == nil {
		return
	}

	assetAssetHWID := asset.AssetHWID
	update := bson.M{"$set": bson.M{
		"asset_hwid": assetAssetHWID,
		"os_info":    fmt.Sprintf("%v", invData["os_info"]),
		"cpu_model":  fmt.Sprintf("%v", invData["cpu_model"]),
		"updated_at": time.Now(),
	}}

	if ramFloat, ok := invData["ram_total_gb"].(float64); ok {
		update["$set"].(bson.M)["ram_total_gb"] = int(ramFloat)
	}

	_, _ = database.AssetInventoryCollection.UpdateOne(context.TODO(), bson.M{"asset_hwid": assetAssetHWID}, update, options.Update().SetUpsert(true))
}
