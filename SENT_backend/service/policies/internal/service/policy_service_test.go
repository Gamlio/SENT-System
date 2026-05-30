package policies

import (
	"testing"

	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed open sqlite: %v", err)
	}
	database.DB = db
	if err := db.AutoMigrate(&models.Policy{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestApproveBaselineByAsset(t *testing.T) {
	db := setupTestDB(t)

	// create baseline entries
	p1 := models.Policy{OrgID: 1, Category: "USB_DEVICE", Value: "hash1", PolicyType: "WHITELIST", ApprovalStatus: "BASELINE", IsActive: false, AssetHWID: "HW1"}
	p2 := models.Policy{OrgID: 1, Category: "USB_DEVICE", Value: "hash2", PolicyType: "WHITELIST", ApprovalStatus: "BASELINE", IsActive: false, AssetHWID: "HW1"}
	if err := db.Create(&p1).Error; err != nil {
		t.Fatalf("create p1: %v", err)
	}
	if err := db.Create(&p2).Error; err != nil {
		t.Fatalf("create p2: %v", err)
	}

	svc := &PolicyService{}
	if err := svc.ApproveBaselineByAsset("HW1", 1, "tester"); err != nil {
		t.Fatalf("approve: %v", err)
	}

	var count int64
	db.Model(&models.Policy{}).Where("asset_hwid = ? AND approval_status = ?", "HW1", "APPROVED").Count(&count)
	if count != 2 {
		t.Fatalf("expected 2 approved, got %d", count)
	}
}

func TestBulkApprovePolicies(t *testing.T) {
	db := setupTestDB(t)
	p1 := models.Policy{OrgID: 1, Category: "PORT", Value: "22", PolicyType: "WHITELIST", ApprovalStatus: "BASELINE", IsActive: false}
	p2 := models.Policy{OrgID: 1, Category: "PORT", Value: "80", PolicyType: "WHITELIST", ApprovalStatus: "BASELINE", IsActive: false}
	if err := db.Create(&p1).Error; err != nil {
		t.Fatalf("create p1: %v", err)
	}
	if err := db.Create(&p2).Error; err != nil {
		t.Fatalf("create p2: %v", err)
	}

	svc := &PolicyService{}
	ids := []uint{p1.ID, p2.ID}
	if err := svc.BulkApprovePolicies(ids, 1, "tester"); err != nil {
		t.Fatalf("bulk approve: %v", err)
	}

	var count int64
	db.Model(&models.Policy{}).Where("approval_status = ?", "APPROVED").Count(&count)
	if count != 2 {
		t.Fatalf("expected 2 approved, got %d", count)
	}
}

func TestApprovePolicyByID(t *testing.T) {
	db := setupTestDB(t)
	p := models.Policy{OrgID: 1, Category: "PUBLISHER", Value: "Acme", PolicyType: "WHITELIST", ApprovalStatus: "BASELINE", IsActive: false}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("create p: %v", err)
	}

	svc := &PolicyService{}
	if err := svc.ApprovePolicyByID(p.ID, 1, "tester"); err != nil {
		t.Fatalf("approve by id: %v", err)
	}

	var p2 models.Policy
	if err := db.First(&p2, p.ID).Error; err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if p2.ApprovalStatus != "APPROVED" || !p2.IsActive {
		t.Fatalf("policy not approved: %#v", p2)
	}
}
