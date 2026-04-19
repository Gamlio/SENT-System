package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	RDB *redis.Client
	Ctx = context.Background()
)

// InitRedis khởi tạo kết nối Redis
func InitRedis(addr string, password string, db int) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	// Kiểm tra kết nối thực tế
	return RDB.Ping(Ctx).Err()
}

// ============================= BLACKLIST HELPERS =============================

// SetBlacklist chặn một IP hoặc HWID với thời gian hết hạn (TTL)
func SetBlacklist(key string, duration time.Duration) error {
	return RDB.Set(Ctx, fmt.Sprintf("blacklist:%s", key), "blocked", duration).Err()
}

// IsBlacklisted kiểm tra xem đối tượng có nằm trong danh sách chặn không
func IsBlacklisted(key string) bool {
	val, err := RDB.Exists(Ctx, fmt.Sprintf("blacklist:%s", key)).Result()
	if err != nil {
		return false
	}
	return val > 0
}

// ============================= DISTRIBUTED ASSET CACHE =============================

type AssetCacheData struct {
	SecretKey string `json:"secret_key"`
	OrgID     uint   `json:"org_id"`
}

// SetAssetCache lưu Key của Asset vào Redis để đồng bộ toàn bộ cluster
func SetAssetCache(hwid string, secretKey string, orgID uint, ttl time.Duration) error {
	data := AssetCacheData{SecretKey: secretKey, OrgID: orgID}
	bytes, _ := json.Marshal(data)
	return RDB.Set(Ctx, fmt.Sprintf("asset_cache:%s", hwid), bytes, ttl).Err()
}

// GetAssetCache lấy Key của Asset từ Redis
func GetAssetCache(hwid string) (string, uint, error) {
	val, err := RDB.Get(Ctx, fmt.Sprintf("asset_cache:%s", hwid)).Result()
	if err != nil {
		return "", 0, err
	}
	var data AssetCacheData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return "", 0, err
	}
	return data.SecretKey, data.OrgID, nil
}

// ============================= DISTRIBUTED RATE LIMIT =============================

// AllowRequest thực hiện Rate Limit bằng Redis INCR + EXPIRE (Fixed Window)
func AllowRequest(key string, limit int64, window time.Duration) bool {
	count, err := RDB.Incr(Ctx, key).Result()
	if err != nil {
		return true // Fail-open: Nếu Redis lỗi tạm thời, không chặn request hợp lệ
	}
	if count == 1 {
		RDB.Expire(Ctx, key, window)
	}
	return count <= limit
}

// RecordAssetFlood đếm số lượng request và băng thông của Asset qua Redis Pipeline
func RecordAssetFlood(hwid string, bodySize int64) (bool, string) {
	pipe := RDB.Pipeline()
	reqKey := fmt.Sprintf("flood:req:%s", hwid)
	byteKey := fmt.Sprintf("flood:byte:%s", hwid)

	reqIncr := pipe.Incr(Ctx, reqKey)
	byteIncr := pipe.IncrBy(Ctx, byteKey, bodySize)

	_, err := pipe.Exec(Ctx)
	if err != nil {
		return true, "" // Fail-open
	}

	if reqIncr.Val() == 1 {
		RDB.Expire(Ctx, reqKey, time.Minute)
	}
	if byteIncr.Val() == bodySize {
		RDB.Expire(Ctx, byteKey, time.Minute)
	}

	// Giới hạn cấu hình: 100 req/phút hoặc 50MB/phút
	if reqIncr.Val() > 100 {
		SetBlacklist(hwid, 10*time.Minute)
		return false, "Vượt quá giới hạn request/phút"
	}
	if byteIncr.Val() > 50*1024*1024 {
		SetBlacklist(hwid, 10*time.Minute)
		return false, "Vượt quá giới hạn băng thông/phút"
	}
	return true, ""
}

// ============================= SEQUENCE NUMBER (ANTI-REPLAY) =============================

// CheckSequenceNumber kiểm tra số thứ tự gói tin để chống Replay Attack tuyệt đối
// Trả về true nếu sequence hợp lệ (lớn hơn số cũ)
func CheckSequenceNumber(hwid string, newSeq int64) bool {
	key := fmt.Sprintf("seq:%s", hwid)

	// Lấy số thứ tự cũ từ Redis
	lastSeqStr, err := RDB.Get(Ctx, key).Result()
	if err == redis.Nil {
		// Lần đầu tiên agent gửi request, chấp nhận và lưu lại
		RDB.Set(Ctx, key, newSeq, 180*24*time.Hour) // Tăng TTL lên 6 tháng (Chống Replay dài hạn)
		return true
	}

	var lastSeq int64
	fmt.Sscanf(lastSeqStr, "%d", &lastSeq)

	// Nếu số mới nhỏ hơn hoặc bằng số cũ -> Replay Attack Detected
	if newSeq <= lastSeq {
		return false
	}

	// Cập nhật số thứ tự mới nhất vào Redis
	RDB.Set(Ctx, key, newSeq, 180*24*time.Hour)
	return true
}

// ============================= TOKEN REVOCATION =============================

// RevokeToken đưa JWT vào danh sách thu hồi cho đến khi nó hết hạn tự nhiên
func RevokeToken(token string, expiration time.Duration) error {
	return RDB.Set(Ctx, fmt.Sprintf("revoked:%s", token), "1", expiration).Err()
}

// IsTokenRevoked kiểm tra token đã bị vô hiệu hóa chưa
func IsTokenRevoked(token string) bool {
	val, err := RDB.Exists(Ctx, fmt.Sprintf("revoked:%s", token)).Result()
	if err != nil {
		return false
	}
	return val > 0
}
