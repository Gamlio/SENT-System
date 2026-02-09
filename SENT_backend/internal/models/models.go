package models

import (
	"time"

	"gorm.io/gorm"
)

// --- NHÓM 1: TỔ CHỨC & QUẢN TRỊ GLOBAL (4 bảng) ---
type Organization struct {
	gorm.Model
	Name              string `gorm:"unique;not null"`
	EnrollTokenPrefix string `gorm:"unique;not null"`
}

type LicensePlan struct {
	gorm.Model
	Name      string
	MaxAgents int // Giới hạn số máy trạm
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
	AdminID uint   // ID của R1/R2 thực hiện thao tác
	Action  string // Ví dụ: "Tạo công ty mới"
	Details string
}

// --- NHÓM 2: NGƯỜI DÙNG & PHÂN QUYỀN (4 bảng) ---
type User struct {
	gorm.Model
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"`
	RoleLevel    int    // 1: Global Admin, 3: SME Admin, 4: SME Staff
	OrgID        *uint  // NULL nếu là Level 1, 2
	TOTPSecret   string
}

type UserPermission struct {
	gorm.Model
	UserID       uint
	RegionID     uint // R4 chỉ được quản lý vùng này
	CanApprove   bool
	CanViewLogs  bool
	CanManageUSB bool
}

type Region struct {
	gorm.Model
	OrgID       uint
	Name        string
	EnrollToken string `gorm:"unique"` // Token riêng cho từng vùng
}

type Department struct {
	gorm.Model
	OrgID uint
	Name  string
}

// --- NHÓM 3: THIẾT BỊ & TUÂN THỦ (6 bảng) ---
type Agent struct {
	HWID     string `gorm:"primaryKey"`
	OrgID    uint
	RegionID uint
	Hostname string
	Status   string
	LastSeen time.Time
}

type AgentInventory struct {
	gorm.Model
	AgentHWID  string `gorm:"unique"`
	CPUModel   string
	RAMTotalGB int
	OSInfo     string
}

type SoftwareItem struct {
	gorm.Model
	AgentHWID    string
	SoftwareName string
	Version      string
}

type AgentSnapshot struct {
	AgentHWID        string `gorm:"primaryKey"`
	LastSoftwareHash string // Phục vụ Differential Reporting
	LastPortHash     string
}

type USBWhitelist struct {
	gorm.Model
	OrgID        uint
	DeviceID     string `gorm:"not null"` // VID_PID_Serial thực tế
	FriendlyName string
	AssignedTo   string // Truy cứu trách nhiệm nhân viên
}

type SoftwarePolicy struct {
	gorm.Model
	OrgID        uint
	SoftwareName string `gorm:"not null"`
	IsProhibited bool   `gorm:"default:false"`
	MinVersion   string
}

// --- NHÓM 4: GIÁM SÁT & CẢNH BÁO (4 bảng) ---
type SecurityAlert struct {
	gorm.Model
	OrgID       uint
	HWID        string
	AlertType   string // USB_UNAUTHORIZED, SOFTWARE_VIOLATION
	Title       string
	Description string
	Severity    string // Low, Medium, High, Critical
	IsResolved  bool   `gorm:"default:false"`
}

type OpenPort struct {
	gorm.Model
	AgentHWID   string
	Port        int
	ProcessName string
}

type USBLog struct {
	gorm.Model
	AgentHWID     string
	DeviceName    string
	DeviceID      string
	IsWhitelisted bool
	EventType     string // "plugged" hoặc "unplugged"
}

type SecurityEvent struct {
	gorm.Model
	AgentHWID string
	EventID   int
	Message   string // Nội dung từ Windows Event Log
}
