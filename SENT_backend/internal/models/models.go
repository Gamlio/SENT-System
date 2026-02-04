// internal/models/models.go
package models

import (
	"time"
)

type Organization struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"unique;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Role struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"not null"`   // Ví dụ: "Kỹ thuật viên"
	Permissions  string `gorm:"type:jsonb"` // Lưu Rules dạng JSON
	LevelContext int    `gorm:"not null"`
	OrgID        *uint  `gorm:"index"`
}

type User struct {
	ID             uint   `gorm:"primaryKey"`
	Username       string `gorm:"uniqueIndex;not null"`
	HashedPassword string `gorm:"not null"`
	Level          int    `gorm:"not null"` // 1: Master, 2: Admin, 3: Staff, 4: Viewer
	OrgID          *uint  `gorm:"index"`
	RoleID         *uint  `gorm:"index"`
}

type Region struct {
	ID          uint   `gorm:"primaryKey"`
	OrgID       uint   `gorm:"index"`
	Name        string `gorm:"not null"`
	EnrollToken string `gorm:"uniqueIndex"` // Token cho Agent cài đặt
}

type Agent struct {
	ID         uint      `gorm:"primaryKey"`
	HWID       string    `gorm:"uniqueIndex"` // ID phần cứng từ Go Agent
	Hostname   string    `gorm:"not null"`
	OrgID      uint      `gorm:"index"`
	RegionID   uint      `gorm:"index"`
	Status     string    `gorm:"default:'pending'"` // pending, approved, blocked
	OSInfo     string    `gorm:"type:varchar(100)"`
	CustomData string    `gorm:"type:jsonb"`        // Chứa IP, MAC, USB Info linh hoạt
	Manager    string    `gorm:"type:varchar(255)"` // Điền thủ công
	LastSeen   time.Time `gorm:"autoUpdateTime"`
}
