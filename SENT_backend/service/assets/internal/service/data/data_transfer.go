package data

import (
	"SENT_backend/pkg/cache"
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

func ProcessDataTransfer(asset models.Asset, data interface{}) error {
	bytesData, err := GetBytesFromData(data)
	if err != nil {
		return fmt.Errorf("dữ liệu data_transfer không hợp lệ: %w", err)
	}

	var payload struct {
		TotalNetSent uint64 `json:"total_net_sent"`
		TotalNetRecv uint64 `json:"total_net_recv"`
		TopProcesses []struct {
			Name      string `json:"name"`
			BytesSent uint64 `json:"bytes_sent"`
		} `json:"top_processes"`
	}

	if err := json.Unmarshal(bytesData, &payload); err != nil {
		log.Printf("[DataTransfer] Lỗi giải mã JSON từ Agent (HWID: %s): %v", asset.AssetHWID, err)
		return fmt.Errorf("lỗi giải mã JSON data_transfer: %w", err)
	}

	if database.AssetIOActivityCollection != nil {
		ioRecord := bson.M{
			"asset_hwid":     asset.AssetHWID,
			"org_id":         int64(asset.OrgID),
			"timestamp":      time.Now(),
			"total_net_sent": payload.TotalNetSent,
			"total_net_recv": payload.TotalNetRecv,
			"top_processes":  payload.TopProcesses,
		}
		_, _ = database.AssetIOActivityCollection.InsertOne(context.TODO(), ioRecord)
	}

	cacheKey := fmt.Sprintf("io_activity:%s", asset.AssetHWID)
	lastValStr, err := cache.RDB.Get(context.Background(), cacheKey).Result()
	cache.RDB.Set(context.Background(), cacheKey, payload.TotalNetSent, 5*time.Minute)

	if err == redis.Nil || err != nil {
		return nil
	}

	lastNetSent, _ := strconv.ParseUint(lastValStr, 10, 64)
	var diffSent uint64
	if payload.TotalNetSent > lastNetSent {
		diffSent = payload.TotalNetSent - lastNetSent
	}
	// if diffSent > 1*1024*1024*1024 { //if need 1Gb for demo
	if diffSent > 15*1024*1024*1024 {
		topApp := "Unknown"
		if len(payload.TopProcesses) > 0 {
			topApp = payload.TopProcesses[0].Name
		}
		reason := fmt.Sprintf("Lưu lượng mạng đột biến. Ứng dụng khả nghi: %s", topApp)

		SendBehaviorLog(map[string]interface{}{
			"asset":    asset,
			"category": "Data Exfiltration",
			"value":    fmt.Sprintf("%d MB", diffSent/1024/1024),
			"title":    "[P1] Hoạt động mạng bất thường",
			"desc":     fmt.Sprintf("%s. Lượng dữ liệu: %d MB/30s", reason, diffSent/1024/1024),
			"priority": "P1",
		})
	}
	return nil
}
