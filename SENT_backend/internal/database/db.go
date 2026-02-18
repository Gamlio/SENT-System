package database

import (
	"fmt"
	"os"
	"sent_backend/internal/models"

	"golang.org/x/crypto/bcrypt"
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
		// Nhóm Tổ chức & Phân quyền
		&models.Organization{},
		&models.Region{},
		&models.User{},
		&models.UserPermission{},

		// Nhóm Quản lý Thiết bị & Tuân thủ
		&models.Agent{},
		&models.AgentInventory{},
		&models.SoftwareItem{},
		&models.AgentSnapshot{},
		&models.USBWhitelist{},
		&models.SoftwarePolicy{},

		// Nhóm Log & Cảnh báo an ninh
		&models.SecurityAlert{},
		&models.OpenPort{},
		&models.USBLog{},
		&models.SecurityEvent{},
	)

	if err != nil {
		fmt.Printf("❌ Lỗi Migration: %v\n", err)
	} else {
		fmt.Println("✅ Đã đồng bộ hóa 18+ bảng dữ liệu")
	}

	// 2. Khởi tạo Tổ chức Hệ thống mặc định
	var org models.Organization
	if err := db.Where("name = ?", "SENT Global System").First(&org).Error; err != nil {
		org = models.Organization{
			Name:              "SENT Global System",
			EnrollTokenPrefix: "SENT-GLOBAL", // Dùng cho Admin hệ thống
		}
		db.Create(&org)
		fmt.Println("✅ Đã tạo Org hệ thống")
	}

	// 3. Khởi tạo Super Admin (R1)
	var admin models.User
	if err := db.Where("username = ?", "Admin").First(&admin).Error; err != nil {
		// Lưu ý: Password nên lấy từ .env, ở đây dùng mặc định của Thanh
		hashed, _ := bcrypt.GenerateFromPassword([]byte("Thanh@123"), 12)
		admin = models.User{
			Username:     "Admin",
			PasswordHash: string(hashed),
			RoleLevel:    1, // Cấp độ cao nhất (Global Admin)
			OrgID:        &org.ID,
		}
		db.Create(&admin)
		fmt.Println("✅ Đã tạo Super Admin (R1)")
	}
	var region models.Region
	if err := db.Where("enroll_token = ?", "SENT-TOKEN-SME-01").First(&region).Error; err != nil {
		region = models.Region{
			OrgID:       org.ID, // Gán vào Org hệ thống
			Name:        "Chi nhánh mặc định",
			EnrollToken: "SENT-TOKEN-SME-01", // Khớp với token trong Agent main.go
		}
		db.Create(&region)
		fmt.Println("✅ Đã tạo Vùng mặc định (Token: SENT-TOKEN-SME-01)")
	}
	DB = db
}
