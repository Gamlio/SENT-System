package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SoftwareItem struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AssetHWID       string             `bson:"asset_hwid" json:"asset_hwid" binding:"required,max=64"`
	OrgID           uint               `bson:"org_id" json:"org_id" binding:"required"`
	SoftwareName    string             `bson:"software_name" json:"software_name" binding:"required,min=1,max=255"`
	Version         string             `bson:"version" json:"version" binding:"max=50"`
	Publisher       string             `bson:"publisher" json:"publisher" binding:"max=255"`
	InstallLocation string             `bson:"install_location" json:"install_location" binding:"max=500"`
	FileHash        string             `bson:"file_hash" json:"file_hash" binding:"max=256"`
	Status          string             `bson:"status" json:"status" binding:"required,oneof=INSTALLED GHOST_REGISTRY"` // INSTALLED, GHOST_REGISTRY
	IsRunning       bool               `bson:"is_running" json:"is_running"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}
type OpenPort struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AssetHWID   string             `bson:"asset_hwid" json:"asset_hwid" binding:"required,max=64"`
	OrgID       uint               `bson:"org_id" json:"org_id" binding:"required"`
	Port        int                `bson:"port" json:"port" binding:"required,min=1,max=65535"`
	ProcessName string             `bson:"process_name" json:"process_name" binding:"max=255"`
	Status      string             `bson:"status" json:"status" binding:"required,oneof=OPEN CLOSED"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type USBLog struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AssetHWID     string             `bson:"asset_hwid" json:"asset_hwid" binding:"required,max=64"`
	OrgID         uint               `bson:"org_id" json:"org_id" binding:"required"`
	DeviceName    string             `bson:"device_name" json:"device_name" binding:"required,min=1,max=255"`
	DeviceID      string             `bson:"device_id" json:"device_id" binding:"max=255"`
	VID           string             `bson:"vid" json:"vid" binding:"max=10"`
	PID           string             `bson:"pid" json:"pid" binding:"max=10"`
	SerialNumber  string             `bson:"serial_number" json:"serial_number" binding:"max=255"`
	DeviceHash    string             `bson:"device_hash" json:"device_hash" binding:"max=256"`
	IsWhitelisted bool               `bson:"is_whitelisted" json:"is_whitelisted"`
	EventType     string             `bson:"event_type" json:"event_type" binding:"max=50,oneof=CONNECT DISCONNECT"`
	Timestamp     time.Time          `bson:"timestamp" json:"timestamp"`
}

type AssetIOActivity struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AssetHWID        string             `bson:"asset_hwid" json:"asset_hwid" binding:"required,max=64"`
	OrgID            uint               `bson:"org_id" json:"org_id" binding:"required"`
	NetBytesSent     uint64             `bson:"net_bytes_sent" json:"net_bytes_sent" binding:"required,gte=0"`
	NetBytesRecv     uint64             `bson:"net_bytes_recv" json:"net_bytes_recv" binding:"required,gte=0"`
	DiskBytesWritten uint64             `bson:"disk_bytes_written" json:"disk_bytes_written" binding:"required,gte=0"`
	DiskBytesRead    uint64             `bson:"disk_bytes_read" json:"disk_bytes_read" binding:"required,gte=0"`
	Timestamp        time.Time          `bson:"timestamp" json:"timestamp" binding:"required"`
}

type AssetInventory struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AssetHWID  string             `bson:"asset_hwid" json:"asset_hwid" binding:"required,max=64"`
	OrgID      uint               `bson:"org_id" json:"org_id" binding:"required"`
	CPUModel   string             `bson:"cpu_model" json:"cpu_model" binding:"required,min=1,max=255"`
	RAMTotalGB int                `bson:"ram_total_gb" json:"ram_total_gb" binding:"required,min=0,max=1000000"`
	OSInfo     string             `bson:"os_info" json:"os_info" binding:"required,min=1,max=500"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at" binding:"required"`
}
type SecurityAlert struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID       uint               `bson:"org_id" json:"org_id" binding:"required"`
	AssetHWID   string             `bson:"asset_hwid" json:"asset_hwid" binding:"required,max=64"`
	IncidentID  *uint              `bson:"incident_id,omitempty" json:"incident_id"`
	Priority    string             `bson:"priority" json:"priority" binding:"required,oneof=P1 P2 P3 P4"`
	AlertType   string             `bson:"alert_type" json:"alert_type" binding:"required,max=50"`
	Title       string             `bson:"title" json:"title" binding:"required,min=1,max=255"`
	Description string             `bson:"description" json:"description" binding:"min=1,max=5000"`
	Severity    string             `bson:"severity" json:"severity" binding:"required,oneof=Low Medium High Critical"`
	IsResolved  bool               `bson:"is_resolved" json:"is_resolved"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

type IncidentAudit struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	IncidentID  uint               `bson:"incident_id" json:"incident_id" binding:"required"`
	UserID      *uint              `bson:"user_id,omitempty" json:"user_id"`
	User        *User              `bson:"-" json:"user,omitempty"`
	UserName    string             `bson:"user_name" json:"user_name"`
	ActionType  string             `bson:"action_type" json:"action_type" binding:"required,max=50"`
	Content     string             `bson:"content" json:"content" binding:"required,min=1,max=5000"`
	OldStatus   string             `bson:"old_status" json:"old_status" binding:"max=50"`
	NewStatus   string             `bson:"new_status" json:"new_status" binding:"max=50"`
	Images      []string           `bson:"images" json:"images"`
	ImageHashes []string           `bson:"image_hashes" json:"image_hashes"`

	IPAddress    string `bson:"ip_address" json:"ip_address"`
	EvidenceData string `bson:"evidence_data" json:"evidence_data"`
	PreviousHash string `bson:"previous_hash" json:"previous_hash"`
	AuditHash    string `bson:"audit_hash" json:"audit_hash"`
	IsImmutable  bool   `bson:"is_immutable" json:"is_immutable"`
}

func (act *IncidentAudit) GenerateAuditHash() {
	var uid uint = 0
	if act.UserID != nil {
		uid = *act.UserID
	}
	imagesStr := strings.Join(act.Images, ",")
	imgHashesStr := strings.Join(act.ImageHashes, ",")
	dataStr := fmt.Sprintf("%v|%v|%s|%s|%d|%s|%s|%s|%s|%s|%s",
		act.IncidentID,
		uid,
		act.UserName,
		act.ActionType,
		act.CreatedAt.Unix(),
		act.Content,
		act.EvidenceData,
		act.PreviousHash,
		act.IPAddress,
		imagesStr,
		imgHashesStr,
	)

	hash := sha256.Sum256([]byte(dataStr))
	act.AuditHash = hex.EncodeToString(hash[:])
	act.IsImmutable = true
}

type AIChatSession struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	UserID    uint               `bson:"user_id" json:"user_id" binding:"required"`
	Title     string             `bson:"title" json:"title" binding:"required,min=1,max=255"`
}
type AIChatLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	SessionID primitive.ObjectID `bson:"session_id" json:"session_id" binding:"required"`
	UserID    uint               `bson:"user_id" json:"user_id" binding:"required"`
	Role      string             `bson:"role" json:"role" binding:"required,oneof=user assistant"`
	Content   string             `bson:"content" json:"content" binding:"required,min=1,max=10000"`
	Thought   string             `bson:"thought" json:"thought" binding:"max=10000"`
}
