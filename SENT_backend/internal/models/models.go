package models

import (
	"time"

	"gorm.io/gorm"
)

// --- NHÓM 1: TỔ CHỨC & QUẢN TRỊ GLOBAL ---
type Organization struct {
	gorm.Model
	Name              string `json:"name"`
	CompanyCode       string `gorm:"unique;not null" json:"company_code"`
	EnrollTokenPrefix string `json:"enroll_token_prefix"`
	// Relationships
	Users    []User            `gorm:"foreignKey:OrgID" json:"-"`
	Regions  []Region          `gorm:"foreignKey:OrgID" json:"-"`
	Policies []UniversalPolicy `gorm:"foreignKey:OrgID" json:"-"`
}

type Region struct {
	gorm.Model
	OrgID       uint   `json:"org_id"`
	Name        string `json:"name"`
	EnrollToken string `gorm:"unique" json:"enroll_token"`
	// Relationships
	Agents []Agent `gorm:"foreignKey:RegionID" json:"-"`
}

// --- NHÓM 2: NGƯỜI DÙNG ---
type User struct {
	gorm.Model
	Username     string `gorm:"unique;not null" json:"username"` // Bắt buộc unique để login
	PasswordHash string `gorm:"not null" json:"-"`
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	RoleLevel    int    `json:"role_level"`
	OrgID        *uint  `json:"org_id"` // Dùng pointer (*) để Level 1/2 có thể NULL

	CanManageAgents   bool `json:"can_manage_agents" gorm:"default:false"`
	CanManagePolicies bool `json:"can_manage_policies" gorm:"default:false"`
	CanManageDocs     bool `json:"can_manage_docs" gorm:"default:false"`
	CanManageUsers    bool `json:"can_manage_users" gorm:"default:false"`

	// Relationships
	Permissions []UserPermission `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"permissions"`
}

type UserPermission struct {
	gorm.Model
	UserID     uint   `json:"user_id"`
	Permission string `json:"permission"`
}

// --- NHÓM 3: THIẾT BỊ (AGENTS) ---
type Agent struct {
	HWID      string    `gorm:"primaryKey;column:hw_id" json:"hwid"`
	OrgID     uint      `gorm:"column:org_id" json:"org_id"`
	RegionID  uint      `gorm:"column:region_id" json:"region_id"`
	Hostname  string    `gorm:"column:hostname" json:"hostname"`
	IPAddress string    `gorm:"column:ip_address" json:"ip_address"`
	Status    string    `gorm:"column:status" json:"status"`
	LastSeen  time.Time `gorm:"column:last_seen" json:"last_seen"`

	// Khai báo Relationship rõ ràng: 1 Agent có 1 Inventory, nhiều Alerts, nhiều Software...
	// OnDelete:CASCADE -> Xóa Agent sẽ tự động xóa log, inventory của nó cho sạch DB.
	Inventory      AgentInventory  `gorm:"foreignKey:AgentHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"inventory"`
	Snapshot       AgentSnapshot   `gorm:"foreignKey:AgentHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"snapshot"`
	Software       []SoftwareItem  `gorm:"foreignKey:AgentHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"software"`
	Alerts         []SecurityAlert `gorm:"foreignKey:HWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"alerts"`
	SecurityEvents []SecurityEvent `gorm:"foreignKey:AgentHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"security_events"`
	OpenPorts      []OpenPort      `gorm:"foreignKey:AgentHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"open_ports"`
	USBLogs        []USBLog        `gorm:"foreignKey:AgentHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"usb_logs"`
}

type AgentInventory struct {
	gorm.Model
	AgentHWID  string `gorm:"column:agent_hw_id;uniqueIndex" json:"agent_hwid"`
	CPUModel   string `json:"cpu_model"`
	RAMTotalGB int    `json:"ram_total_gb"`
	OSInfo     string `json:"os_info"`
}

type AgentSnapshot struct {
	AgentHWID        string `gorm:"primaryKey;column:agent_hw_id"`
	LastSoftwareHash string `json:"last_software_hash"`
	LastPortHash     string `json:"last_port_hash"`
}

type SoftwareItem struct {
	gorm.Model
	AgentHWID    string `gorm:"column:agent_hw_id;index" json:"agent_hwid"`
	SoftwareName string `json:"software_name"`
	Version      string `json:"version"`
}

// --- NHÓM 4: GIÁM SÁT AN NINH & SỰ CỐ ---
type SecurityAlert struct {
	gorm.Model
	OrgID       uint   `json:"org_id" gorm:"index"`
	HWID        string `gorm:"column:hw_id;index" json:"hw_id"`
	AlertType   string `json:"alert_type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	IsResolved  bool   `json:"is_resolved" gorm:"default:false"`
}

type SecurityEvent struct {
	gorm.Model
	AgentHWID string `gorm:"column:agent_hw_id;index" json:"agent_hwid"`
	EventID   int    `json:"event_id"`
	Message   string `json:"message"`
}

type OpenPort struct {
	gorm.Model
	AgentHWID   string `gorm:"column:agent_hw_id;index" json:"agent_hwid"`
	Port        int    `json:"port"`
	ProcessName string `json:"process_name"`
}

type USBLog struct {
	gorm.Model
	AgentHWID     string `gorm:"column:agent_hw_id;index" json:"agent_hwid"`
	DeviceName    string `json:"device_name"`
	DeviceID      string `json:"device_id"`
	IsWhitelisted bool   `json:"is_whitelisted"`
	EventType     string `json:"event_type"`
}

// --- NHÓM 5: CHÍNH SÁCH TẬP TRUNG ---
type UniversalPolicy struct {
	gorm.Model
	OrgID       uint   `json:"org_id" gorm:"index"`
	Title       string `json:"title"`
	Category    string `json:"category" gorm:"index"`
	PolicyType  string `json:"policy_type"`
	Value       string `json:"value"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	Description string `json:"description"`
}

type PolicyDocument struct {
	gorm.Model
	Title       string `json:"title"`
	FileName    string `json:"file_name"`
	FilePath    string `json:"file_path"`
	Category    string `json:"category"`
	IsProcessed bool   `gorm:"default:false" json:"is_processed"`
	OrgID       uint   `json:"org_id" gorm:"index"`
}

type USBWhitelist struct {
	gorm.Model
	OrgID        uint   `json:"org_id" gorm:"index"`
	DeviceID     string `json:"device_id"`
	FriendlyName string `json:"friendly_name"`
	AssignedTo   string `json:"assigned_to"`
}

type AgentWhitelist struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	HWID         string `json:"hwid" gorm:"index"`
	Category     string `json:"category"`
	SoftwareName string `json:"name"`
}
