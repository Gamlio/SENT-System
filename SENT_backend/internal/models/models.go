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
	Users           []User           `gorm:"foreignKey:OrgID" json:"-"`
	Regions         []Region         `gorm:"foreignKey:OrgID" json:"-"`
	Policies        []Policy         `gorm:"foreignKey:OrgID" json:"-"`
	Assets          []Asset          `gorm:"foreignKey:OrgID" json:"-"`
	ApprovalTickets []ApprovalTicket `gorm:"foreignKey:OrgID" json:"-"`
	Documents       []Document       `gorm:"foreignKey:OrgID" json:"-"`
}

type Region struct {
	gorm.Model
	OrgID       uint   `json:"org_id"`
	Name        string `json:"name"`
	EnrollToken string `gorm:"unique" json:"enroll_token"`
	// Relationships
	Assets []Asset `gorm:"foreignKey:RegionID" json:"-"`
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
	GroupID      *uint  `json:"group_id" gorm:"index"`

	// THAY ĐỔI 1: Thay RoleLevel bằng RoleName rõ ràng

	RiskScore int `json:"risk_score" gorm:"default:0"`
	// Nhóm Assets
	PermAssetView   bool `json:"perm_asset_view" gorm:"default:false"`
	PermAssetAction bool `json:"perm_asset_action" gorm:"default:false"` // Chạy lệnh: Quét virus, Khởi động lại
	PermAssetDelete bool `json:"perm_asset_delete" gorm:"default:false"` // Xóa mềm asset
	PermAssetMove   bool `json:"perm_asset_move" gorm:"default:false"`   // Quyền chuyển máy vào các Nhóm động

	// Nhóm Policy
	PermPolicyView   bool `json:"perm_policy_view" gorm:"default:false"`
	PermPolicyManage bool `json:"perm_policy_manage" gorm:"default:false"` // Tạo đơn (Maker) thêm/xóa luật

	// Nhóm Incident (giữ nguyên)
	PermIncidentView   bool `json:"perm_incident_view" gorm:"default:false"`
	PermIncidentAction bool `json:"perm_incident_action" gorm:"default:false"`

	// Nhóm Document (giữ nguyên)
	PermDocView   bool `json:"perm_doc_view" gorm:"default:false"`
	PermDocManage bool `json:"perm_doc_manage" gorm:"default:false"`

	// Nhóm User & System
	PermUserView     bool `json:"perm_user_view" gorm:"default:false"`     // Xem danh sách nhân sự
	PermUserManage   bool `json:"perm_user_manage" gorm:"default:false"`   // Tạo đơn (Maker) thêm/sửa/xóa user
	PermSystemConfig bool `json:"perm_system_config" gorm:"default:false"` // Tạo/Xóa các Nhóm (Kế toán, IT...)
	PermGroupManage  bool `json:"perm_group_manage" gorm:"default:false"`  // Quyền quản lý Nhóm (Policy Groups)

	// Nhóm Approval
	PermApprovalView  bool `json:"perm_approval_view" gorm:"default:false"`  // Xem danh sách đơn (Ticket)
	PermApprovalFinal bool `json:"perm_approval_final" gorm:"default:false"` // [CHECKER] Quyền bấm nút "Duyệt" cuối cùng

	// [MỚI] ĐỒNG BỘ: Luồng phê duyệt tài khoản mới tạo (Maker - Checker)
	ApprovalStatus string `json:"approval_status" gorm:"default:'PENDING'"` // PENDING, APPROVED, REJECTED
	ApprovedBy     string `json:"approved_by"`
}
type UserPayload struct {
	Username           string `json:"username" binding:"required,min=3,max=50"`
	Password           string `json:"password" binding:"required,min=8,max=128"`
	FullName           string `json:"full_name" binding:"required,min=3,max=100"`
	Phone              string `json:"phone" binding:"omitempty,len=10,numeric"`
	Email              string `json:"email" binding:"required,email"`
	GroupID            *uint  `json:"group_id"`
	PermAssetView      bool   `json:"perm_asset_view"`
	PermAssetAction    bool   `json:"perm_asset_action"`
	PermAssetDelete    bool   `json:"perm_asset_delete"`
	PermAssetMove      bool   `json:"perm_asset_move"`
	PermPolicyView     bool   `json:"perm_policy_view"`
	PermPolicyManage   bool   `json:"perm_policy_manage"`
	PermIncidentView   bool   `json:"perm_incident_view"`
	PermIncidentAction bool   `json:"perm_incident_action"`
	PermDocView        bool   `json:"perm_doc_view"`
	PermDocManage      bool   `json:"perm_doc_manage"`
	PermUserView       bool   `json:"perm_user_view"`
	PermUserManage     bool   `json:"perm_user_manage"`
	PermSystemConfig   bool   `json:"perm_system_config"`
	PermGroupManage    bool   `json:"perm_group_manage"`
	PermApprovalView   bool   `json:"perm_approval_view"`
	PermApprovalFinal  bool   `json:"perm_approval_final"`
}
type UserPermission struct {
	gorm.Model
	UserID     uint   `json:"user_id"`
	Permission string `json:"permission"`
}

// PolicyGroup định nghĩa một "Hồ sơ bảo mật" (VD: Kế toán, Máy chủ, Dev)
type PolicyGroup struct {
	gorm.Model
	OrgID       uint     `json:"org_id" gorm:"index"`
	Name        string   `json:"name"` // "High Security Profile"
	Description string   `json:"description"`
	CreatedBy   string   `json:"created_by"` // [AUDIT] Lưu tên người tạo
	Policies    []Policy `json:"policies" gorm:"foreignKey:GroupID"`
	Assets      []Asset  `json:"assets" gorm:"foreignKey:GroupID"`
}

// PolicyGroupPayload định nghĩa dữ liệu đầu vào cho việc tạo/sửa nhóm
type PolicyGroupPayload struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// --- NHÓM 3: THIẾT BỊ (assetS) ---
type Asset struct {
	AssetHWID string    `gorm:"primaryKey;column:asset_hwid" json:"asset_hwid"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	OrgID     uint      `gorm:"column:org_id" json:"org_id" binding:"required"`
	RegionID  *uint     `gorm:"column:region_id" json:"region_id"`
	UserID    *uint     `gorm:"column:user_id" json:"user_id"`
	// [MỚI] Liên kết với PolicyGroup
	GroupID   *uint     `json:"group_id" gorm:"column:group_id;index"`
	Hostname  string    `gorm:"column:hostname" json:"hostname"`
	IPAddress string    `gorm:"column:ip_address" json:"ip_address"`
	LastSeen  time.Time `gorm:"column:last_seen" json:"last_seen"`

	Manager                    *User             `gorm:"foreignKey:UserID" json:"manager"`
	Inventory                  AssetInventory    `gorm:"-" json:"inventory"`
	Software                   []SoftwareItem    `gorm:"-" json:"software"`
	Alerts                     []SecurityAlert   `gorm:"-" json:"alerts"`
	OpenPorts                  []OpenPort        `gorm:"-" json:"open_ports"`
	USBLogs                    []USBLog          `gorm:"-" json:"usb_logs"`
	IOActivities               []AssetIOActivity `gorm:"-" json:"io_activities"`
	RiskScore                  int               `json:"risk_score" gorm:"default:0"`
	Status                     string            `json:"status" gorm:"default:'PENDING'"`
	AssetTypeID                *uint             `json:"asset_type_id" gorm:"column:asset_type_id"`
	AssetType                  *AssetType        `json:"asset_type" gorm:"foreignKey:AssetTypeID"`
	TrustScore                 float64           `gorm:"default:100"`
	DepartmentTag              string            `json:"department_tag"`
	LastIncidentAt             *time.Time        `json:"last_incident_at"`
	LastTrustRecoveryAppliedAt *time.Time        `json:"last_trust_recovery_applied_at"`
	IsZeroTrust                bool              `gorm:"default:false" json:"is_zero_trust"`
	SecretKey                  string            `json:"-"`
	// [MỚI] ĐỒNG BỘ: Lưu người đã cấp phép máy trạm này
	ApprovedBy string `json:"approved_by"`

	Incidents []Incident `gorm:"foreignKey:AssetHWID;" json:"incidents"`
}
type AssetType struct {
	gorm.Model
	OrgID       uint    `json:"org_id" gorm:"index"`
	Name        string  `json:"name"`        // VD: "Máy chủ nghiệp vụ"
	Icon        string  `json:"icon"`        // Tên icon (Server, Laptop...)
	RiskWeight  float64 `json:"risk_weight"` // Hệ số rủi ro: 1.5, 1.2...
	Description string  `json:"description"`
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
	AssetHWID string `json:"asset_hwid" binding:"required,max=64"`
	Hostname  string `json:"hostname" binding:"required,min=1,max=255"`
	IPAddress string `json:"ip_address" binding:"required,ip"`
	Token     string `json:"token" binding:"required,min=10,max=255"` // Mã cài đặt (Enrollment Token)
}

// --- NHÓM 5: QUẢN LÝ SỰ CỐ & AUTOMATION ---
type Incident struct {
	gorm.Model // ID, CreatedAt, UpdatedAt, DeletedAt
	// Lưu ý: ID ở đây là uint

	OrgID       uint   `json:"org_id" binding:"required"`
	AssetHWID   string `gorm:"type:varchar(64);column:asset_hwid;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"asset_hwid"`
	Asset       Asset  `gorm:"foreignKey:AssetHWID;references:AssetHWID" json:"asset"`
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	Description string `json:"description"`
	AIAnalysis  string `json:"ai_analysis"`
	AssigneeID  *uint  `json:"assignee_id" gorm:"index"`
	Assignee    *User  `json:"assignee" gorm:"foreignKey:AssigneeID"`

	// [MỚI THÊM] Báo cáo sau khi đóng Case
	ResolutionSummary string `json:"resolution_summary" gorm:"type:text"`
	// Alerts liên quan
	Alerts           []SecurityAlert `gorm:"-" json:"alerts"`
	LastAuditHash    string          `json:"last_audit_hash"`
	Activities       []IncidentAudit `gorm:"-" json:"activities"`
	PlaybookProgress string          `json:"playbook_progress" gorm:"type:text"`
}

// --- NHÓM 5: CHÍNH SÁCH TẬP TRUNG ---
type Policy struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	OrgID      uint   `json:"org_id" binding:"required" gorm:"index:idx_policy_org_status"`
	Title      string `json:"title"`
	Category   string `json:"category"` // SOFTWARE, USB, NETWORK
	Value      string `json:"value"`    // Tên exe, mã USB, Port...
	PolicyType string `json:"policy_type" gorm:"default:'BLACKLIST'"`
	IsActive   bool   `json:"is_active" gorm:"default:true"`

	GroupID *uint `json:"group_id" gorm:"index"` // Nếu NULL = Global Policy

	IncidentID *uint `json:"incident_id"`

	// [MỚI THÊM] - LUỒNG PHÊ DUYỆT (APPROVAL WORKFLOW)
	ApprovalStatus string `json:"approval_status" gorm:"default:'PENDING';index:idx_policy_org_status"` // PENDING, APPROVED, REJECTED
	CreatedBy      string `json:"created_by"`                                                           // Username người tạo đơn
	ApprovedBy     string `json:"approved_by"`                                                          // Username người duyệt đơn
}

type Document struct {
	gorm.Model
	OrgID          uint   `json:"org_id" gorm:"index"`
	Title          string `json:"title"`
	FileName       string `json:"file_name"`
	FilePath       string `json:"file_path"`
	Category       string `json:"category" gorm:"index"`
	OriginalName   string `json:"original_name"`
	ApprovedBy     string `json:"approved_by"`
	DisplayPdfPath string `json:"display_pdf_path"`

	ApprovalStatus string `json:"approval_status" gorm:"default:'PENDING'"`
	UploadedBy     string `json:"uploaded_by"`
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

	RequestedBy    string     `json:"requested_by"`              // Tên người gửi đơn (hoặc "SYSTEM" nếu là asset tự gửi)
	RequestReason  string     `json:"request_reason"`            // Nguyên nhân/lý do làm đơn
	ReviewedBy     string     `json:"reviewed_by"`               // Tên Admin đã duyệt
	ReviewNote     string     `json:"review_note"`               // Lý do duyệt/từ chối (nếu có)
	Priority       int        `json:"priority" gorm:"default:1"` // 1: Thấp, 2: Trung bình, 3: Cao
	LastNotifiedAt *time.Time `json:"last_notified_at"`
	// Lưu dạng JSON để admin xem trước nội dung mà không cần join bảng phức tạp
	SnapshotData string `json:"snapshot_data" gorm:"type:jsonb"`
}
type SentVersion struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Tag         string    `gorm:"index" json:"tag"`      // v4.0, v4.1...
	Platform    string    `gorm:"index" json:"platform"` // windows, linux, mac
	DownloadURL string    `json:"download_url"`          // Link từ GitHub Release
	Checksum    string    `json:"checksum"`              // SHA256 để đảm bảo an toàn
	IsLatest    bool      `gorm:"default:false" json:"is_latest"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	ReleaseNote string    `json:"release_note"`
}
