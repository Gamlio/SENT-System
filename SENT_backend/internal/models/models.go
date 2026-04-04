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
	// Relationships (Đã bổ sung đầy đủ các liên kết Has Many)
	Users           []User            `gorm:"foreignKey:OrgID" json:"-"`
	Regions         []Region          `gorm:"foreignKey:OrgID" json:"-"`
	Policies        []UniversalPolicy `gorm:"foreignKey:OrgID" json:"-"`
	assets          []Asset           `gorm:"foreignKey:OrgID" json:"-"`
	ApprovalTickets []ApprovalTicket  `gorm:"foreignKey:OrgID" json:"-"`
	PolicyDocuments []PolicyDocument  `gorm:"foreignKey:OrgID" json:"-"`
}

type Region struct {
	gorm.Model
	OrgID       uint   `json:"org_id"`
	Name        string `json:"name"`
	EnrollToken string `gorm:"unique" json:"enroll_token"`
	// Relationships
	assets []Asset `gorm:"foreignKey:RegionID" json:"-"`
}

// --- NHÓM 2: NGƯỜI DÙNG ---
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Bỏ gorm:"unique", thay bằng uniqueIndex kết hợp với org_id
	Username     string `gorm:"uniqueIndex:idx_org_user;not null" json:"username"`
	OrgID        *uint  `gorm:"uniqueIndex:idx_org_user" json:"org_id"`
	PasswordHash string `gorm:"not null" json:"-"`
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`

	// THAY ĐỔI 1: Thay RoleLevel bằng RoleName rõ ràng

	RiskScore int `json:"risk_score" gorm:"default:0"`
	// Nhóm Assets
	PermAssetView   bool `json:"perm_asset_view" gorm:"default:false"`
	PermAssetAction bool `json:"perm_asset_action" gorm:"default:false"`
	PermAssetDelete bool `json:"perm_asset_delete" gorm:"default:false"`

	PermPolicyView   bool `json:"perm_policy_view" gorm:"default:false"`
	PermPolicyAction bool `json:"perm_policy_action" gorm:"default:false"`

	PermIncidentView   bool `json:"perm_incident_view" gorm:"default:false"`
	PermIncidentAction bool `json:"perm_incident_action" gorm:"default:false"`

	PermDocView   bool `json:"perm_doc_view" gorm:"default:false"`
	PermDocManage bool `json:"perm_doc_manage" gorm:"default:false"`

	PermUserManage     bool `json:"perm_user_manage" gorm:"default:false"`
	PermApprovalManage bool `json:"perm_approval_manage" gorm:"default:false"`

	// [MỚI] ĐỒNG BỘ: Luồng phê duyệt tài khoản mới tạo (Maker - Checker)
	ApprovalStatus string `json:"approval_status" gorm:"default:'PENDING'"` // PENDING, APPROVED, REJECTED
	ApprovedBy     string `json:"approved_by"`
}
type UserPayload struct {
	Username           string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Password           string `json:"password" binding:"required,min=8,max=128"`
	FullName           string `json:"full_name" binding:"required,min=3,max=100"`
	Phone              string `json:"phone" binding:"omitempty,len=10,numeric"`
	Email              string `json:"email" binding:"required,email"`
	PermAssetView      bool   `json:"perm_asset_view"`
	PermAssetAction    bool   `json:"perm_asset_action"`
	PermAssetDelete    bool   `json:"perm_asset_delete"`
	PermPolicyView     bool   `json:"perm_policy_view"`
	PermPolicyAction   bool   `json:"perm_policy_action"`
	PermIncidentView   bool   `json:"perm_incident_view"`
	PermIncidentAction bool   `json:"perm_incident_action"`
	PermDocView        bool   `json:"perm_doc_view"`
	PermDocManage      bool   `json:"perm_doc_manage"`
	PermUserManage     bool   `json:"perm_user_manage"`
	PermApprovalManage bool   `json:"perm_approval_manage"`
}
type UserPermission struct {
	gorm.Model
	UserID     uint   `json:"user_id"`
	Permission string `json:"permission"`
}

// --- NHÓM 3: THIẾT BỊ (assetS) ---
type Asset struct {
	HWID      string    `gorm:"primaryKey;column:hw_id" json:"hwid"`
	OrgID     uint      `gorm:"column:org_id" json:"org_id"`
	RegionID  *uint     `gorm:"column:region_id" json:"region_id"`
	UserID    *uint     `gorm:"column:user_id" json:"user_id"`
	Hostname  string    `gorm:"column:hostname" json:"hostname"`
	IPAddress string    `gorm:"column:ip_address" json:"ip_address"`
	LastSeen  time.Time `gorm:"column:last_seen" json:"last_seen"`

	Manager *User `gorm:"foreignKey:UserID" json:"manager"`
	// Khai báo Relationship rõ ràng: 1 asset có 1 Inventory, nhiều Alerts, nhiều Software...
	// Inventory đã chuyển qua MongoDB
	Inventory                  AssetInventory    `gorm:"-" json:"inventory"`
	Software                   []SoftwareItem    `gorm:"-" json:"software"`      // Virtual field, dữ liệu lấy từ MongoDB
	Alerts                     []SecurityAlert   `gorm:"-" json:"alerts"`        // Virtual field, dữ liệu lấy từ MongoDB
	OpenPorts                  []OpenPort        `gorm:"-" json:"open_ports"`    // Virtual field, dữ liệu lấy từ MongoDB
	USBLogs                    []USBLog          `gorm:"-" json:"usb_logs"`      // Virtual field, dữ liệu lấy từ MongoDB
	IOActivities               []AssetIOActivity `gorm:"-" json:"io_activities"` // Virtual field, dữ liệu lấy từ MongoDB
	RiskScore                  int               `json:"risk_score" gorm:"default:0"`
	Status                     string            `json:"status" gorm:"default:'PENDING'"`
	TrustScore                 float64           `gorm:"default:100"` // Long-term trust
	DepartmentTag              string            // Explicit Department Tag (e.g., FINANCE, DEV, PROD)
	LastIncidentAt             *time.Time        // Timestamp of the last incident for this asset
	LastTrustRecoveryAppliedAt *time.Time        // Timestamp when trust score recovery was last applied

	DeviceType     string `json:"device_type" gorm:"default:'OFFICE'"`
	IsZeroTrust    bool   `gorm:"default:false" json:"is_zero_trust"`
	BaselineStatus string `gorm:"default:'NONE'" json:"baseline_status"`

	SecretKey string `json:"-"`
	// [MỚI] ĐỒNG BỘ: Lưu người đã cấp phép máy trạm này
	ApprovedBy string `json:"approved_by"`

	Incidents []Incident `gorm:"foreignKey:assetHWID;references:HWID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"incidents"`
}
type WhitelistItem struct {
	gorm.Model
	OrgID       uint   `gorm:"index"`
	AssetHWID   string `gorm:"index" json:"asset_hwid"` // Nếu để trống thì là Global Whitelist
	Type        string `json:"type"`                    // "USB", "SOFTWARE_HASH", "PUBLISHER"
	Value       string `json:"value"`                   // Mã Hash hoặc Tên nhà phát hành (Microsoft...)
	Description string `json:"description"`
}
type EnrollmentToken struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	OrgID     uint      `json:"org_id"`
	ExpiresAt time.Time `json:"expires_at"` // Hạn sử dụng (VD: 24h)
	CreatedBy string    `json:"created_by"` // Username của Admin đã tạo mã
}
type EnrollRequest struct {
	HWID      string `json:"hwid" binding:"required,max=64,alphanum"`
	Hostname  string `json:"hostname" binding:"required,min=1,max=255"`
	IPAddress string `json:"ip_address" binding:"required,ip"`
	Token     string `json:"token" binding:"required,min=20,max=255"` // Mã cài đặt (Enrollment Token)
}

// --- NHÓM 5: QUẢN LÝ SỰ CỐ & AUTOMATION ---
type Incident struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt
	// Lưu ý: ID ở đây là uint

	OrgID     uint   `json:"org_id"`
	AssetHWID string `gorm:"column:asset_hw_id" json:"asset_hw_id"`
	// Relationship
	Asset Asset `gorm:"foreignKey:AssetHWID;references:HWID" json:"asset"`

	Type         string `json:"type"`          // Loại sự cố (VD: Malware, DDoS)
	Severity     string `json:"severity"`      // Low, Medium, High, Critical
	Priority     string `json:"priority"`      // P1, P2, P3, P4
	Status       string `json:"status"`        // Open, Investigating, Resolved, False Positive
	Description  string `json:"description"`   // Mô tả ngắn gọn
	PlaybookName string `json:"playbook_name"` // Tên quy trình xử lý áp dụng
	AIAnalysis   string `json:"ai_analysis"`   // Kết quả phân tích từ AI
	// [MỚI] Thông tin người được phân công xử lý (nếu có)
	AssigneeID *uint `json:"assignee_id" gorm:"index"`
	Assignee   *User `json:"assignee" gorm:"foreignKey:AssigneeID"`

	// [MỚI THÊM] Báo cáo sau khi đóng Case
	ResolutionSummary string `json:"resolution_summary" gorm:"type:text"`
	// Alerts liên quan
	Alerts []SecurityAlert `gorm:"-" json:"alerts"` // Virtual field, dữ liệu lấy từ MongoDB

	// [MỚI] Link sang bảng hoạt động
	Activities       []IncidentActivity `gorm:"-" json:"activities"` // Virtual field, dữ liệu lấy từ MongoDB
	PlaybookProgress string             `json:"playbook_progress" gorm:"type:text"`
}

// --- NHÓM 5: CHÍNH SÁCH TẬP TRUNG ---
type UniversalPolicy struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	OrgID      uint   `json:"org_id" gorm:"index:idx_policy_org_status"`
	Title      string `json:"title"`
	Category   string `json:"category"` // SOFTWARE, USB, NETWORK
	Value      string `json:"value"`    // Tên exe, mã USB, Port...
	PolicyType string `json:"policy_type" gorm:"default:'BLACKLIST'"`
	IsActive   bool   `json:"is_active" gorm:"default:true"`

	TargetType  string   `json:"target_type" gorm:"default:'GLOBAL'"`
	TargetHWIDs []string `json:"target_hwids" gorm:"serializer:json"`

	IncidentID *uint `json:"incident_id"`

	// [MỚI THÊM] - LUỒNG PHÊ DUYỆT (APPROVAL WORKFLOW)
	ApprovalStatus string `json:"approval_status" gorm:"default:'PENDING';index:idx_policy_org_status"` // PENDING, APPROVED, REJECTED
	CreatedBy      string `json:"created_by"`                                                           // Username người tạo đơn
	ApprovedBy     string `json:"approved_by"`                                                          // Username người duyệt đơn
}

type PolicyDocument struct {
	gorm.Model
	Title       string `json:"title"`
	FileName    string `json:"file_name"`
	FilePath    string `json:"file_path"`
	Category    string `json:"category"`
	IsProcessed bool   `gorm:"default:false" json:"is_processed"`
	OrgID       uint   `json:"org_id" gorm:"index"`

	OriginalPath   string `json:"original_path"`    // Lưu file gốc (Word, PDF, Excel...)
	DisplayPdfPath string `json:"display_pdf_path"` // MẶC ĐỊNH LÀ PDF ĐỂ RENDER LÊN WEB
	// [MỚI] ĐỒNG BỘ: Luồng phê duyệt tài liệu (tránh up file rác)
	ApprovalStatus string `json:"approval_status" gorm:"default:'PENDING'"`
	UploadedBy     string `json:"uploaded_by"`
	ApprovedBy     string `json:"approved_by"`
}

// --- NHÓM 7: HỆ THỐNG PHÊ DUYỆT TẬP TRUNG (APPROVAL TICKETS) ---
type ApprovalTicket struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	OrgID      uint   `json:"org_id" gorm:"index"`
	ModuleType string `json:"module_type" gorm:"index"` // Loại đơn: "asset", "POLICY", "DOCUMENT", "USER"
	ActionType string `json:"action_type"`              // Hành động: "ENROLL" (đăng ký mới), "CREATE", "DELETE"
	TargetID   uint   `json:"target_id"`                // ID của bản ghi tương ứng (VD: ID của asset hoặc Policy)
	TargetName string `json:"target_name"`              // Tên để hiển thị cho dễ nhìn (VD: "PC-KETOAN-01" hoặc "Cấm USB")

	Status string `json:"status" gorm:"default:'PENDING';index"` // PENDING, APPROVED, REJECTED

	RequestedBy string `json:"requested_by"` // Tên người gửi đơn (hoặc "SYSTEM" nếu là asset tự gửi)
	ReviewedBy  string `json:"reviewed_by"`  // Tên Admin đã duyệt
	ReviewNote  string `json:"review_note"`  // Lý do từ chối (nếu có)

	// Lưu dạng JSON để admin xem trước nội dung mà không cần join bảng phức tạp
	SnapshotData string `json:"snapshot_data" gorm:"type:text"`
}
