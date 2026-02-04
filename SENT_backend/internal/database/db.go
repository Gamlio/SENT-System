// internal/database/db.go
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

	// 1. Tự động tạo bảng (Auto Migration)
	db.AutoMigrate(&models.Organization{}, &models.Role{}, &models.User{}, &models.Region{}, &models.Agent{})

	// 2. Logic Init System (Tương tự init_db.py)
	var org models.Organization
	if err := db.Where("name = ?", "SENT Global System").First(&org).Error; err != nil {
		org = models.Organization{Name: "SENT Global System"}
		db.Create(&org)
		fmt.Println("✅ Đã tạo Org hệ thống")
	}

	var admin models.User
	if err := db.Where("username = ?", "Admin").First(&admin).Error; err != nil {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("Thanh@123"), 12)
		admin = models.User{
			Username:       "Admin",
			HashedPassword: string(hashed),
			Level:          1,
			OrgID:          &org.ID,
		}
		db.Create(&admin)
		fmt.Println("✅ Đã tạo Super Admin")
	}

	DB = db
}
