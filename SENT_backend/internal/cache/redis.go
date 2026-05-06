package cache

import (
	"context"
	"fmt"
	"strconv"
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

// ============================= LUA SCRIPTS (ATOMIC OPERATIONS) =============================
var (
	// Sliding Window Rate Limiting (Chặn Burst Traffic triệt để)
	rateLimitScript = redis.NewScript(`
		redis.call('zremrangebyscore', KEYS[1], 0, tonumber(ARGV[3]) - tonumber(ARGV[2]))
		local count = redis.call('zcard', KEYS[1])
		if count < tonumber(ARGV[1]) then
			redis.call('zadd', KEYS[1], tonumber(ARGV[3]), tonumber(ARGV[3]))
			redis.call('pexpire', KEYS[1], tonumber(ARGV[2]))
			return 1
		end
		return 0
	`)

	// Gộp Incr, Expire và Check Limit vào một thao tác nguyên tử
	floodScript = redis.NewScript(`
		local reqKey = KEYS[1]; local byteKey = KEYS[2]; local blKey = KEYS[3]
		local bodySize = tonumber(ARGV[1])
		local reqs = redis.call('incr', reqKey); if reqs == 1 then redis.call('expire', reqKey, 60) end
		local bytes = redis.call('incrby', byteKey, bodySize); if bytes == bodySize then redis.call('expire', byteKey, 60) end
		if reqs > 100 then redis.call('set', blKey, 'blocked', 'EX', 600); return 1
		elseif bytes > 52428800 then redis.call('set', blKey, 'blocked', 'EX', 600); return 2 end
		return 0
	`)

	// Chống Race Condition tuyệt đối khi cập nhật Sequence
	seqScript = redis.NewScript(`
		local current = redis.call('get', KEYS[1]); if current and tonumber(ARGV[1]) <= tonumber(current) then return 0 end
		redis.call('set', KEYS[1], ARGV[1], 'EX', tonumber(ARGV[2])); return 1
	`)
)

// ============================= BLACKLIST HELPERS =============================

// SetBlacklist chặn một IP hoặc HWID với thời gian hết hạn (TTL)
func SetBlacklist(key string, duration time.Duration) error {
	return RDB.Set(Ctx, fmt.Sprintf("blacklist:%s", key), "blocked", duration).Err()
}

// IsBlacklisted kiểm tra xem đối tượng có nằm trong danh sách chặn không
func IsBlacklisted(key string) bool {
	val, err := RDB.Exists(Ctx, fmt.Sprintf("blacklist:%s", key)).Result()
	if err != nil {
		return true // [BẢO MẬT] Fail-closed: Coi như bị chặn nếu Redis sập
	}
	return val > 0
}

// ============================= DISTRIBUTED ASSET CACHE =============================

// SetAssetCache lưu Key của Asset vào Redis để đồng bộ toàn bộ cluster
func SetAssetCache(hwid string, secretKey string, orgID uint, ttl time.Duration) error {
	// [HIỆU NĂNG] Thay vì JSON chuỗi, dùng Hash Map siêu tốc (HSET)
	key := fmt.Sprintf("asset_cache:%s", hwid)
	pipe := RDB.Pipeline()
	pipe.HSet(Ctx, key, "secret_key", secretKey, "org_id", orgID)
	pipe.Expire(Ctx, key, ttl)
	_, err := pipe.Exec(Ctx)
	return err
}

// GetAssetCache lấy Key của Asset từ Redis
func GetAssetCache(hwid string) (string, uint, error) {
	key := fmt.Sprintf("asset_cache:%s", hwid)
	res, err := RDB.HGetAll(Ctx, key).Result()
	if err != nil || len(res) == 0 {
		return "", 0, fmt.Errorf("cache miss")
	}
	orgID, _ := strconv.ParseUint(res["org_id"], 10, 32)
	return res["secret_key"], uint(orgID), nil
}

// ============================= DISTRIBUTED RATE LIMIT =============================

// AllowRequest thực hiện Rate Limit bằng Redis Sorted Set (Sliding Window Log)
func AllowRequest(key string, limit int64, window time.Duration) bool {
	now := time.Now().UnixMilli()
	windowMs := window.Milliseconds()
	res, err := rateLimitScript.Run(Ctx, RDB, []string{key}, limit, windowMs, now).Result()
	if err != nil {
		return false // [BẢO MẬT] Đóng chặt (Fail-closed) khi Redis có lỗi
	}
	return res.(int64) == 1
}

// RecordAssetFlood đếm số lượng request và băng thông của Asset qua Redis Pipeline
func RecordAssetFlood(hwid string, bodySize int64) (bool, string) {
	reqKey := fmt.Sprintf("flood:req:%s", hwid)
	byteKey := fmt.Sprintf("flood:byte:%s", hwid)
	blKey := fmt.Sprintf("blacklist:%s", hwid)

	res, err := floodScript.Run(Ctx, RDB, []string{reqKey, byteKey, blKey}, bodySize).Result()
	if err != nil {
		return false, "Lỗi hệ thống kiểm tra Rate Limit" // [BẢO MẬT] Fail-closed
	}

	status := res.(int64)
	if status == 1 {
		return false, "Vượt quá giới hạn request/phút"
	} else if status == 2 {
		return false, "Vượt quá giới hạn băng thông/phút"
	}
	return true, ""
}

// ============================= SEQUENCE NUMBER (ANTI-REPLAY) =============================

// CheckSequenceNumber kiểm tra số thứ tự gói tin để chống Replay Attack tuyệt đối
func CheckSequenceNumber(hwid string, newSeq int64) bool {
	key := fmt.Sprintf("seq:%s", hwid)

	// Lệnh Run trả về 1 nếu hợp lệ và được lưu mới, 0 nếu không.
	// 15552000 là số giây cho TTL 180 ngày (180 * 24 * 60 * 60)
	res, err := seqScript.Run(Ctx, RDB, []string{key}, newSeq, 15552000).Result()
	if err != nil {
		return false
	}
	return res.(int64) == 1
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
		return true // [BẢO MẬT] Fail-closed: Hủy mọi phiên nếu Redis sập
	}
	return val > 0
}
