// internal/database/db.go
package database

import (
	"context"
	"fmt"
	"os"
	"sent_backend/internal/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var MongoClient *mongo.Client
var IOCollection *mongo.Collection // Collection riêng cho log truyền tải dữ liệu
var SoftwareCollection *mongo.Collection
var USBCollection *mongo.Collection
var OpenPortCollection *mongo.Collection
var AgentInventoryCollection *mongo.Collection
var AgentIOActivityCollection *mongo.Collection
var SecurityAlertCollection *mongo.Collection
var AIChatSessionCollection *mongo.Collection
var AIChatLogCollection *mongo.Collection

func InitDB() {
	// --- KẾT NỐI POSTGRESQL ---
	dsn := os.Getenv("DATABASE_URL")
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to PostgreSQL!")
	}

	// --- KẾT NỐI MONGODB ---
	mongoURI := os.Getenv("MONGODB_URI") // Thêm vào .env
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	clientOptions := options.Client().ApplyURI(mongoURI)
	MongoClient, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Printf("⚠️ Cảnh báo: Không thể kết nối MongoDB: %v\n", err)
	} else {
		db := MongoClient.Database("sent_logs")
		IOCollection = db.Collection("io_activities")
		SoftwareCollection = db.Collection("software_items")
		USBCollection = db.Collection("usb_logs")
		OpenPortCollection = db.Collection("open_ports")
		AgentInventoryCollection = db.Collection("agent_inventory")
		AgentIOActivityCollection = db.Collection("agent_io_activities")
		SecurityAlertCollection = db.Collection("security_alerts")
		AIChatSessionCollection = db.Collection("ai_chat_sessions")
		AIChatLogCollection = db.Collection("ai_chat_logs")
		fmt.Println("✅ Đã kết nối MongoDB thành công")
	}

	// --- MIGRATION (Chỉ cho Postgres, dữ liệu quan hệ) ---
	fmt.Println("⏳ Đang đồng bộ hóa PostgreSQL...")
	DB.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Agent{},
		&models.Incident{},
		&models.ApprovalTicket{},
		&models.Region{},
		&models.UniversalPolicy{},
		&models.PolicyDocument{},
	)
}
