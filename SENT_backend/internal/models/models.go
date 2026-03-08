package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type JSONStringArray []string

// Scan: Chuyển dữ liệu từ Database (JSON/Text) thành Go Struct
func (a *JSONStringArray) Scan(value interface{}) error {
	// 1. Xử lý trường hợp NULL từ database
	if value == nil {
		*a = make([]string, 0)
		return nil
	}

	// 2. Chấp nhận cả []byte và string
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("type assertion failed: value is not []byte or string")
	}

	// 3. Parse JSON
	return json.Unmarshal(bytes, &a)
}

// Value: Chuyển Go Struct thành JSON để lưu vào Database
func (a JSONStringArray) Value() (driver.Value, error) {
	// Nếu mảng rỗng hoặc nil, lưu là "[]" thay vì NULL để tránh lỗi Scan sau này
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

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
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Bỏ gorm:"unique", thay bằng uniqueIndex kết hợp với org_id
	Username string `gorm:"uniqueIndex:idx_org_user;not null" json:"username"`
	OrgID    *uint  `gorm:"uniqueIndex:idx_org_user" json:"org_id"`

	PasswordHash string `gorm:"not null" json:"-"`
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`

	// THAY ĐỔI 1: Thay RoleLevel bằng RoleName rõ ràng
	Role string `json:"role" gorm:"default:'USER'"` // Có 2 loại: "ADMIN" (Chủ công ty) và "USER" (Nhân viên)

	RiskScore int `json:"risk_score" gorm:"default:0"`
	// THAY ĐỔI 2: Chia nhỏ quyền (Read / Write)
	CanViewAgents   bool `json:"can_view_agents" gorm:"default:false"`
	CanManageAgents bool `json:"can_manage_agents" gorm:"default:false"`

	CanViewDocs   bool `json:"can_view_docs" gorm:"default:true"`    // Mặc định ai cũng được đọc tài liệu
	CanManageDocs bool `json:"can_manage_docs" gorm:"default:false"` // Chỉ người được cấp quyền mới được thêm/sửa

	CanManagePolicies  bool `json:"can_manage_policies" gorm:"default:false"`
	CanManageIncidents bool `json:"can_manage_incidents" gorm:"default:false"`
	CanManageUsers     bool `json:"can_manage_users" gorm:"default:false"`
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
	UserID    *uint     `gorm:"column:user_id" json:"user_id"`
	Hostname  string    `gorm:"column:hostname" json:"hostname"`
	IPAddress string    `gorm:"column:ip_address" json:"ip_address"`
	Status    string    `gorm:"column:status" json:"status"`
	LastSeen  time.Time `gorm:"column:last_seen" json:"last_seen"`

	Manager *User `gorm:"foreignKey:UserID" json:"manager"`
	// Khai báo Relationship rõ ràng: 1 Agent có 1 Inventory, nhiều Alerts, nhiều Software...
	// OnDelete:CASCADE -> Xóa Agent sẽ tự động xóa log, inventory của nó cho sạch DB.
	Inventory AgentInventory  `gorm:"foreignKey:AgentHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"inventory"`
	Software  []SoftwareItem  `gorm:"foreignKey:AgentHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"software"`
	Alerts    []SecurityAlert `gorm:"foreignKey:HWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"alerts"`
	OpenPorts []OpenPort      `gorm:"foreignKey:AgentHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"open_ports"`
	USBLogs   []USBLog        `gorm:"foreignKey:AgentHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"usb_logs"`

	RiskScore int `json:"risk_score" gorm:"default:0"`

	Incidents []Incident `gorm:"foreignKey:AgentHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"incidents"`
}

type AgentInventory struct {
	gorm.Model
	AgentHWID  string `gorm:"column:agent_hw_id;uniqueIndex" json:"agent_hwid"`
	CPUModel   string `json:"cpu_model"`
	RAMTotalGB int    `json:"ram_total_gb"`
	OSInfo     string `json:"os_info"`
}

type SoftwareItem struct {
	gorm.Model
	AgentHWID    string `gorm:"column:agent_hw_id;index" json:"agent_hwid"`
	SoftwareName string `json:"software_name"`
	Version      string `json:"version"`
}

// --- NHÓM 4: GIÁM SÁT AN NINH & SỰ CỐ (ĐÃ REFACTOR THEO PLAYBOOK) ---
// --- NHÓM 5: QUẢN LÝ SỰ CỐ & AUTOMATION ---
type Incident struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt
	// Lưu ý: ID ở đây là uint

	OrgID     uint   `json:"org_id"`
	AgentHWID string `json:"agent_hw_id"`
	// Relationship
	Agent Agent `gorm:"foreignKey:AgentHWID;references:HWID" json:"agent"`

	Type         string `json:"type"`          // Loại sự cố (VD: Malware, DDoS)
	Severity     string `json:"severity"`      // Low, Medium, High, Critical
	Priority     string `json:"priority"`      // P1, P2, P3, P4
	Status       string `json:"status"`        // Open, Investigating, Resolved, False Positive
	Description  string `json:"description"`   // Mô tả ngắn gọn
	PlaybookName string `json:"playbook_name"` // Tên quy trình xử lý áp dụng
	AIAnalysis   string `json:"ai_analysis"`   // Kết quả phân tích từ AI

	// Alerts liên quan
	Alerts []SecurityAlert `gorm:"foreignKey:IncidentID" json:"alerts"`

	// [MỚI] Link sang bảng hoạt động
	Activities       []IncidentActivity `json:"activities" gorm:"foreignKey:IncidentID"`
	PlaybookProgress string             `json:"playbook_progress" gorm:"type:text"`
}

// [MỚI] Bảng lưu lịch sử xử lý (Timeline)
type IncidentActivity struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	IncidentID uint `json:"incident_id" gorm:"index"`
	UserID     uint `json:"user_id"` // Người thực hiện (0 nếu là System/AI)
	User       User `json:"user" gorm:"foreignKey:UserID"`

	ActionType string          `json:"action_type"` // COMMENT, STATUS_CHANGE, AI_ANALYSIS
	Content    string          `json:"content"`     // Nội dung chi tiết
	OldStatus  string          `json:"old_status"`  // Trạng thái cũ
	NewStatus  string          `json:"new_status"`  // Trạng thái mới
	Images     JSONStringArray `json:"images" gorm:"type:text"`
}
type SecurityAlert struct {
	gorm.Model
	OrgID      uint   `json:"org_id" gorm:"index"`
	HWID       string `gorm:"column:hw_id;index" json:"hw_id"`
	IncidentID *uint  `json:"incident_id" gorm:"index"` // Dùng pointer (*) vì Alert có thể chưa bị gom vào Incident

	// [THÊM MỚI] - Để nhận diện cấp độ khẩn cấp ngay từ Agent gửi lên
	Priority string `json:"priority"` // "P1", "P2", "P3"

	AlertType   string `json:"alert_type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	IsResolved  bool   `json:"is_resolved" gorm:"default:false"`
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
	Category    string `json:"category" gorm:"index"` // SOFTWARE, USB, NETWORK
	PolicyType  string `json:"policy_type"`           // BLACKLIST, WHITELIST
	Value       string `json:"value"`                 // Giá trị (tên process, ID USB...)
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	Description string `json:"description"`

	// --- CÁC TRƯỜNG PHÂN CẤP ---
	TargetType  string          `json:"target_type" gorm:"default:'GLOBAL'"` // GLOBAL hoặc SPECIFIC
	TargetHWIDs JSONStringArray `json:"target_hwids" gorm:"type:json"`       // Danh sách HWID áp dụng

	// --- TRUY VẾT ---
	IncidentID *uint `json:"incident_id"` // Chính sách này được tạo ra từ sự cố nào?
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

// --- NHÓM 6: AI CHAT HISTORY ---
type AIChatSession struct {
	// Thay gorm.Model bằng các trường cụ thể có tag json:"id"
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uint   `json:"user_id" gorm:"index"`
	Title  string `json:"title"`
}

type AIChatLog struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	SessionID uint   `json:"session_id" gorm:"index"`
	UserID    uint   `json:"user_id" gorm:"index"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Thought   string `json:"thought"`
}
