package models

import (
	"time"

	"gorm.io/gorm"
)

// --- NHÓM 1: TỔ CHỨC & QUẢN TRỊ GLOBAL ---
type Organization struct {
	gorm.Model
	Name              string `gorm:"unique;not null"`
	EnrollTokenPrefix string `gorm:"unique;not null"`
}

type Region struct {
	gorm.Model
	OrgID       uint
	Name        string
	EnrollToken string `gorm:"unique"`
}

// --- NHÓM 2: NGƯỜI DÙNG ---
type User struct {
	gorm.Model
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	RoleLevel    int
	OrgID        *uint
	TOTPSecret   string
}
type UserPermission struct {
	gorm.Model
	UserID     uint
	Permission string
}

// --- NHÓM 3: THIẾT BỊ (QUAN TRỌNG: Đã cố định tên cột) ---
type Agent struct {
	// Thêm phần `json:"..."` vào sau mỗi dòng
	HWID     string    `gorm:"primaryKey;column:hw_id" json:"hwid"`
	OrgID    uint      `gorm:"column:org_id" json:"org_id"`
	RegionID uint      `gorm:"column:region_id" json:"region_id"`
	Hostname string    `gorm:"column:hostname" json:"hostname"`
	Status   string    `gorm:"column:status" json:"status"`
	LastSeen time.Time `gorm:"column:last_seen" json:"last_seen"`
}

// Làm tương tự cho Inventory nếu cần hiển thị chi tiết
type AgentInventory struct {
	gorm.Model
	AgentHWID  string `gorm:"column:agent_hw_id;unique" json:"agent_hwid"`
	CPUModel   string `json:"cpu_model"`
	RAMTotalGB int    `json:"ram_total_gb"`
	OSInfo     string `json:"os_info"`
}

type SoftwareItem struct {
	gorm.Model
	AgentHWID    string `gorm:"column:agent_hw_id"`
	SoftwareName string
	Version      string
}

// --- NHÓM 4: GIÁM SÁT ---
type SecurityAlert struct {
	gorm.Model
	OrgID       uint
	HWID        string `gorm:"column:hw_id"`
	AlertType   string
	Title       string
	Description string
	Severity    string
	IsResolved  bool `gorm:"default:false"`
}

type OpenPort struct {
	gorm.Model
	AgentHWID   string `gorm:"column:agent_hw_id"`
	Port        int
	ProcessName string
}

type USBLog struct {
	gorm.Model
	AgentHWID     string `gorm:"column:agent_hw_id"`
	DeviceName    string
	DeviceID      string
	IsWhitelisted bool
	EventType     string
}

// Các struct khác (License, Policy, etc.) giữ nguyên...
type LicensePlan struct {
	gorm.Model
	Name      string
	MaxAgents int
	Price     float64
}
type LicenseAssignment struct {
	gorm.Model
	OrgID    uint
	PlanID   uint
	ExpireAt time.Time
}
type SystemLog struct {
	gorm.Model
	AdminID uint
	Action  string
	Details string
}
type SoftwarePolicy struct {
	gorm.Model
	OrgID        uint
	SoftwareName string
	IsProhibited bool
	MinVersion   string
}
type USBWhitelist struct {
	gorm.Model
	OrgID        uint
	DeviceID     string
	FriendlyName string
	AssignedTo   string
}
type SecurityEvent struct {
	gorm.Model
	AgentHWID string `gorm:"primaryKey;column:agent_hw_id"`
	EventID   int
	Message   string
}
type AgentSnapshot struct {
	AgentHWID        string `gorm:"primaryKey;column:agent_hw_id"`
	LastSoftwareHash string
	LastPortHash     string
}
