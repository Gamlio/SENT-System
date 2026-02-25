package database

import (
	"fmt"
	"os"
	"sent_backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	// 1. Tự động tạo toàn bộ hệ thống bảng (Auto Migration)
	fmt.Println("⏳ Đang đồng bộ hóa cơ sở dữ liệu...")
	err = db.AutoMigrate(
		// 1. Nhóm Core (Cha)
		&models.Organization{},
		&models.Region{},
		&models.User{},
		&models.UserPermission{},

		// 2. Nhóm Policies
		&models.UniversalPolicy{},
		&models.PolicyDocument{},
		&models.USBWhitelist{},
		&models.AgentWhitelist{},

		// 3. Nhóm Thiết bị (Con của Org/Region)
		&models.Agent{},
		&models.AgentInventory{},
		&models.AgentSnapshot{},
		&models.SoftwareItem{},

		// 4. Nhóm Log & Alerts (Con của Agent)
		&models.SecurityAlert{},
		&models.SecurityEvent{},
		&models.USBLog{},
		&models.OpenPort{},
	)
	if err != nil {
		fmt.Printf("❌ Lỗi Migration: %v\n", err)
	} else {
		fmt.Println("✅ Đã đồng bộ hóa 18+ bảng dữ liệu")
	}

}
