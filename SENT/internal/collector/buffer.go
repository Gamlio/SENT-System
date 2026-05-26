package collector

import (
	"encoding/json"
	"fmt"
	"time"

	"go.etcd.io/bbolt"
)

var bufferBucket = []byte("OfflineTelemetryBuffer")

// InitBufferBucket khởi tạo phân vùng chứa dữ liệu đệm ngoại tuyến
func InitBufferBucket() error {
	if db == nil {
		return fmt.Errorf("cơ sở dữ liệu localdb chưa được khởi tạo")
	}
	return db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bufferBucket)
		return err
	})
}

type OfflineLog struct {
	LogType   string      `json:"log_type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// PushToOfflineBuffer đưa bản tin dữ liệu vào vùng đệm khi mất mạng
func PushToOfflineBuffer(logType string, data interface{}) error {
	if db == nil {
		return nil
	}

	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bufferBucket)
		if b == nil {
			return fmt.Errorf("buffer bucket không tồn tại")
		}

		// Kiểm tra kích thước hoặc số lượng phần tử để khống chế dưới 50MB theo đặc tả phi chức năng
		stats := b.Stats()
		if stats.LeafPageN > 2000 { // Giới hạn tương đối vùng nhớ đệm bảo vệ RAM/Disk
			return fmt.Errorf("vùng đệm ngoại tuyến đạt ngưỡng giới hạn an toàn (50MB)")
		}

		logItem := OfflineLog{
			LogType:   logType,
			Timestamp: time.Now(),
			Data:      data,
		}

		buf, err := json.Marshal(logItem)
		if err != nil {
			return err
		}

		// Khóa key tuần tự theo thời gian Nano để tạo hàng đợi FIFO
		key := []byte(fmt.Sprintf("%d", time.Now().UnixNano()))
		return b.Put(key, buf)
	})
}

// PopAllOfflineBuffer lấy toàn bộ dữ liệu tích lũy ra để đẩy tuần tự lên Backend khi có mạng lại
func PopAllOfflineBuffer() (map[string]interface{}, [][]byte) {
	var logsToFlush = make(map[string]interface{})
	var keysToDelete [][]byte

	if db == nil {
		return logsToFlush, nil
	}

	_ = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bufferBucket)
		if b == nil {
			return nil
		}

		_ = b.ForEach(func(k, v []byte) error {
			var logItem OfflineLog
			if err := json.Unmarshal(v, &logItem); err == nil {
				logsToFlush[fmt.Sprintf("%s_%s", logItem.LogType, string(k))] = logItem
				keysToDelete = append(keysToDelete, k)
			}
			return nil
		})
		return nil
	})

	return logsToFlush, keysToDelete
}

// ClearFlushedLogs Xóa các bản ghi đã gửi thành công khỏi DB cứng máy trạm
func ClearFlushedLogs(keys [][]byte) error {
	if db == nil || len(keys) == 0 {
		return nil
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bufferBucket)
		if b == nil {
			return nil
		}
		for _, key := range keys {
			_ = b.Delete(key)
		}
		return nil
	})
}
