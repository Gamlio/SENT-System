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
	RoleLevel    int    `json:"role_level"` // 1: User (Nhân viên), 2: Admin (Quản trị)
	OrgID        *uint  `json:"org_id"`
}
type UserPermission struct {
	gorm.Model
	UserID     uint
	Permission string
}

// --- NHÓM 3: THIẾT BỊ (QUAN TRỌNG: Đã cố định tên cột) ---
type Agent struct {
	HWID      string          `gorm:"primaryKey;column:hw_id" json:"hwid"`
	OrgID     uint            `gorm:"column:org_id" json:"org_id"`
	RegionID  uint            `gorm:"column:region_id" json:"region_id"`
	Hostname  string          `gorm:"column:hostname" json:"hostname"`
	IPAddress string          `gorm:"column:ip_address" json:"ip_address"` // <--- THÊM DÒNG NÀY
	Status    string          `gorm:"column:status" json:"status"`
	LastSeen  time.Time       `gorm:"column:last_seen" json:"last_seen"`
	Inventory AgentInventory  `gorm:"foreignKey:AgentHWID;references:HWID" json:"inventory"`
	Alerts    []SecurityAlert `gorm:"foreignKey:HWID;references:HWID" json:"alerts"`
	Software  []SoftwareItem  `gorm:"foreignKey:AgentHWID;references:HWID" json:"software"`
}

// Làm tương tự cho Inventory nếu cần hiển thị chi tiết
type AgentInventory struct {
	gorm.Model
	AgentHWID string `gorm:"column:agent_hw_id;unique" json:"agent_hwid"`
	// Thêm json:"..." chữ thường vào sau
	CPUModel   string `json:"cpu_model"`
	RAMTotalGB int    `json:"ram_total_gb"`
	OSInfo     string `json:"os_info"`
}

// 2. Struct SoftwareItem
type SoftwareItem struct {
	gorm.Model
	AgentHWID    string `gorm:"column:agent_hw_id" json:"agent_hwid"`
	SoftwareName string `json:"software_name"` // Quan trọng
	Version      string `json:"version"`
}

// --- NHÓM 4: GIÁM SÁT ---
type SecurityAlert struct {
	gorm.Model
	OrgID       uint   `json:"org_id"`
	HWID        string `gorm:"column:hw_id" json:"hw_id"`
	AlertType   string `json:"alert_type"`  // Ví dụ: SOFTWARE_VIOLATION
	Title       string `json:"title"`       // Ví dụ: Phát hiện phần mềm cấm
	Description string `json:"description"` // Nội dung chi tiết
	Severity    string `json:"severity"`    // High/Medium/Low
	IsResolved  bool   `json:"is_resolved"` // Đã xử lý chưa?
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

// --- NHÓM CHÍNH SÁCH TẬP TRUNG (POLICY HUB) ---
type UniversalPolicy struct {
	gorm.Model
	OrgID uint   `json:"org_id" gorm:"index"`
	Title string `json:"title"`
	// Phân loại: "SOFTWARE", "USB", "NETWORK", "OS"
	Category string `json:"category" gorm:"index"`
	// Loại áp dụng: "BLACKLIST" (Cấm), "WHITELIST" (Cho phép)
	PolicyType string `json:"policy_type"`
	// Giá trị: "chrome.exe", "VID_0781&PID_5581", "facebook.com"
	Value       string `json:"value"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	Description string `json:"description"`
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

type PolicyDocument struct {
	gorm.Model
	Title       string `json:"title"`
	FileName    string `json:"file_name"`
	FilePath    string `json:"file_path"`
	Category    string `json:"category"`                          // Ví dụ: ISO27001, Internal, Law
	IsProcessed bool   `gorm:"default:false" json:"is_processed"` // AI đã nạp vào Vector DB chưa?
	OrgID       uint   `json:"org_id"`
}
type AgentWhitelist struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	HWID         string `json:"hwid" gorm:"index"`
	Category     string `json:"category"` // Đồng bộ Category với bảng trên
	SoftwareName string `json:"name"`
}
