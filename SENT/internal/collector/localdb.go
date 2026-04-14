package collector

import (
	"encoding/json"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

var (
	db         *bbolt.DB
	fileBucket = []byte("FileMetadata")
)

// FileMetadata stores the last known state of a file, including its hash.
type FileMetadata struct {
	ModTime time.Time `json:"mod_time"`
	Size    int64     `json:"size"`
	Hash    string    `json:"hash"`
}

// InitLocalDB initializes the local BoltDB database.
// This should be called once when the agent starts.
func InitLocalDB(agentDataPath string) error {
	var err error
	dbPath := filepath.Join(agentDataPath, "agent_cache.db") // Renamed for clarity
	db, err = bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(fileBucket)
		return err
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
