package database

import (
	"fmt"
	"os"
	"sent_backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB // Biến toàn cục (viết hoa)

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")

	// 1. Khai báo biến err riêng ra trước
	var err error

	// 2. Ép dùng biến DB toàn cục bằng dấu = (KHÔNG dùng :=)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	fmt.Println("⏳ Đang đồng bộ hóa cơ sở dữ liệu...")

	// 3. Đổi db (viết thường) thành DB (viết hoa) ở đây
	err = DB.AutoMigrate(
		// 1. Nhóm Core (Cha)
		&models.Organization{},
		&models.Region{},
		&models.User{},
		&models.UserPermission{},

		// 2. Nhóm Policies
		&models.UniversalPolicy{},
		&models.PolicyDocument{},

		// 3. Nhóm Thiết bị (Con của Org/Region)
		&models.Agent{},
		&models.AgentInventory{},
		&models.SoftwareItem{},

		// 4. Nhóm Log & Alerts (Con của Agent)
		&models.SecurityAlert{},
		&models.USBLog{},
		&models.OpenPort{},

		// 5. Nhóm Sự cố (Con của Agent)
		&models.Incident{},
		&models.IncidentActivity{},

		// 6. Nhóm AI Chat History
		&models.AIChatLog{},
		&models.AIChatSession{},
		// 7. Nhóm Phê duyệt
		&models.ApprovalTicket{},
	)
	if !DB.Migrator().HasTable(&models.IncidentActivity{}) {
		fmt.Println("⚠️ Bảng IncidentActivity CHƯA CÓ. Đang ép tạo bảng...")
		err := DB.Migrator().CreateTable(&models.IncidentActivity{})
		if err != nil {
			fmt.Printf("🚨 LỖI NGHIÊM TRỌNG TỪ POSTGRES KHI TẠO BẢNG: %v\n", err)
		} else {
			fmt.Println("✅ Đã ép tạo bảng IncidentActivity thành công!")
		}
	}
	if err != nil {
		fmt.Printf("❌ Lỗi Migration: %v\n", err)
	} else {
		fmt.Println("✅ Đã đồng bộ hóa các bảng dữ liệu")
	}
}
