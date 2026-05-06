package collector

import (
	"encoding/json"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

var (
	db          *bbolt.DB
	fileBucket  = []byte("FileMetadata")
	stateBucket = []byte("ModuleStates")
)

// FileMetadata stores the last known state of a file, including its hash.
type FileMetadata struct {
	ModTime time.Time `json:"mod_time"`
	Size    int64     `json:"size"`
	Hash    string    `json:"hash"`
}

func InitLocalDB(agentDataPath string) error {
	var err error
	dbPath := filepath.Join(agentDataPath, "agent_cache.db") // Renamed for clarity
	// Tăng timeout lên 3s để hệ thống kịp nhả khóa bbolt nếu lần tắt trước bị lỗi
	db, err = bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(fileBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(stateBucket); err != nil {
			return err
		}
		return nil
	})
}

// CloseLocalDB closes the database connection.
// This should be called when the agent shuts down.
func CloseLocalDB() {
	if db != nil {
		db.Close()
	}
}

// GetFileMetadata retrieves the stored metadata for a given file path.
func GetFileMetadata(filePath string) (meta FileMetadata, found bool) {
	if db == nil {
		return FileMetadata{}, false
	}

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return nil
		}

		v := b.Get([]byte(filePath))
		if v == nil {
			return nil // Not found
		}

		if err := json.Unmarshal(v, &meta); err != nil {
			return err // Corrupted data
		}
		found = true
		return nil
	})

	if err != nil {
		return FileMetadata{}, false
	}

	return meta, found
}

// UpdateFileMetadata stores the new metadata for a file path.
func UpdateFileMetadata(filePath string, meta FileMetadata) error {
	if db == nil {
		return nil
	}

	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return nil
		}

		// Marshal the struct to JSON
		buf, err := json.Marshal(meta)
		if err != nil {
			return err
		}

		return b.Put([]byte(filePath), buf)
	})
}

// UpdateFileMetadataBatch cập nhật nhiều bản ghi metadata trong một transaction duy nhất.
// Giúp giảm thiểu tối đa Disk I/O so với việc gọi db.Update lẻ tẻ.
func UpdateFileMetadataBatch(metas map[string]FileMetadata) error {
	if db == nil || len(metas) == 0 {
		return nil
	}

	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(fileBucket)
		if b == nil {
			return nil
		}
		for filePath, meta := range metas {
			if buf, err := json.Marshal(meta); err == nil {
				_ = b.Put([]byte(filePath), buf)
			}
		}
		return nil
	})
}

// IsDataNew kiểm tra xem dữ liệu của module có thay đổi so với lần gửi cuối không
func IsDataNew(moduleName string, currentHash string) bool {
	var isNew bool = true // Mặc định là mới nếu chưa có record
	if db == nil {
		return true
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(stateBucket)
		if b == nil {
			return nil
		}
		oldHash := b.Get([]byte(moduleName))
		if oldHash != nil && string(oldHash) == currentHash {
			isNew = false
		}
		return nil
	})
	return isNew
}

// UpdateModuleState lưu lại mã băm trạng thái mới nhất của module
func UpdateModuleState(moduleName string, currentHash string) error {
	if db == nil {
		return nil
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(stateBucket)
		if b == nil {
			return nil
		}
		return b.Put([]byte(moduleName), []byte(currentHash))
	})
}
