package data

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ProcessPatch(asset models.Asset, data interface{}) error {
	// 1. Kiểm tra biến toàn cục đã được nạp hay chưa
	if database.AssetPatchCollection == nil {
		return fmt.Errorf("bộ lưu trữ AssetPatchCollection chưa được khởi tạo")
	}

	bytesData, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu patch không hợp lệ: %w", err)
	}

	var agentPatch models.AssetPatchItem
	if err := json.Unmarshal(bytesData, &agentPatch); err != nil {
		return fmt.Errorf("lỗi giải mã cấu trúc dữ liệu bản vá từ Agent: %w", err)
	}

	agentPatch.AssetHWID = asset.AssetHWID
	agentPatch.OrgID = asset.OrgID
	agentPatch.UpdatedAt = time.Now()

	filter := bson.M{
		"asset_hwid": asset.AssetHWID,
		"org_id":     int64(asset.OrgID),
	}

	update := bson.M{
		"$set": bson.M{
			"platform":        agentPatch.Platform,
			"missing_patches": agentPatch.MissingPatches,
			"total_missing":   agentPatch.TotalMissing,
			"updated_at":      agentPatch.UpdatedAt,
		},
	}

	// Gọi trực tiếp thông qua biến cấu trúc toàn cục đã đồng bộ
	_, err = database.AssetPatchCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("lỗi lưu dữ liệu bản vá vào MongoDB: %w", err)
	}

	if agentPatch.TotalMissing > 0 {
		SendBehaviorLog(map[string]interface{}{
			"asset":    asset,
			"category": "Vulnerability Risk",
			"value":    fmt.Sprintf("%d patches missing", agentPatch.TotalMissing),
			"title":    "[P2] Máy trạm thiếu bản cập nhật an ninh",
			"desc":     fmt.Sprintf("Hệ thống phát hiện máy trạm đang thiếu %d bản cập nhật vá lỗi an ninh quan trọng.", agentPatch.TotalMissing),
			"priority": "P2",
		})
	}

	return nil
}
