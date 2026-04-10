// internal/database/db.go
package database

import (
	"context"
	"fmt"
	"os"
	"sent_backend/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
var AssetInventoryCollection *mongo.Collection
var AssetIOActivityCollection *mongo.Collection
var SecurityAlertCollection *mongo.Collection
var IncidentAuditCollection *mongo.Collection
var AIChatSessionCollection *mongo.Collection
var AIChatLogCollection *mongo.Collection

// createMongoIndexes: [TỐI ƯU HIỆU NĂNG] Tạo các chỉ mục (index) cho MongoDB
// Việc này giúp tăng tốc độ truy vấn trên các trường hay được filter, đặc biệt là org_id và asset_hwid
func createMongoIndexes(ctx context.Context) {
	fmt.Println("⏳ Đang tạo chỉ mục cho MongoDB để tối ưu hiệu năng...")

	// Helper function to create index and log result
	createIndex := func(collection *mongo.Collection, keys bson.D, name string) {
		if collection == nil {
			return
		}
		_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: keys})
		if err != nil {
			// Không panic, chỉ cảnh báo, vì index có thể đã tồn tại hoặc có lỗi kết nối
			fmt.Printf("  ⚠️ Cảnh báo khi tạo index '%s': %v\n", name, err)
		} else {
			fmt.Printf("  ✅ Đã tạo/xác thực index '%s'\n", name)
		}
	}

	// 1. SecurityAlerts: Collection quan trọng nhất cho Dashboard và Incident
	createIndex(SecurityAlertCollection, bson.D{{Key: "org_id", Value: 1}, {Key: "created_at", Value: -1}}, "alerts_org_created")
	createIndex(SecurityAlertCollection, bson.D{{Key: "asset_hwid", Value: 1}}, "alerts_hwid")

	// 2. IncidentAudits: Tối ưu timeline của sự cố
	createIndex(IncidentAuditCollection, bson.D{{Key: "incident_id", Value: 1}, {Key: "created_at", Value: 1}}, "audits_incident_created")

	// 3. Các collection log theo từng máy (asset-centric)
	createIndex(SoftwareCollection, bson.D{{Key: "asset_hwid", Value: 1}}, "software_hwid")
	createIndex(USBCollection, bson.D{{Key: "asset_hwid", Value: 1}}, "usb_hwid")
	createIndex(OpenPortCollection, bson.D{{Key: "asset_hwid", Value: 1}}, "port_hwid")
	createIndex(AssetInventoryCollection, bson.D{{Key: "asset_hwid", Value: 1}}, "inventory_hwid")
	createIndex(AssetIOActivityCollection, bson.D{{Key: "asset_hwid", Value: 1}}, "io_hwid")
	fmt.Println("✅ Hoàn tất tạo chỉ mục MongoDB.")
}

func InitDB() {
	// --- KẾT NỐI POSTGRESQL ---
	dsn := os.Getenv("DATABASE_URL")
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	for i := 0; i < 10; i++ {
		fmt.Printf("⏳ [%d/10] Đang thử kết nối PostgreSQL...\n", i+1)
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second) // Đợi 2 giây trước khi thử lại
	}
	if err != nil {
		panic("Failed to connect to PostgreSQL!")
	}

	// --- KẾT NỐI MONGODB ---
	mongoURI := os.Getenv("MONGODB_URI")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second) // Context cho cả connect và tạo index
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	MongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Printf("⚠️ Cảnh báo: Không thể kết nối MongoDB: %v\n", err)
	} else {
		db := MongoClient.Database("sent_logs")
		IOCollection = db.Collection("io_activities")
		SoftwareCollection = db.Collection("software_items")
		USBCollection = db.Collection("usb_logs")
		OpenPortCollection = db.Collection("open_ports")
		AssetInventoryCollection = db.Collection("asset_inventory")
		AssetIOActivityCollection = db.Collection("asset_io_activities")
		SecurityAlertCollection = db.Collection("security_alerts")
		IncidentAuditCollection = db.Collection("incident_audits")

		AIChatSessionCollection = db.Collection("ai_chat_sessions")
		AIChatLogCollection = db.Collection("ai_chat_logs")
		fmt.Println("✅ Đã kết nối MongoDB thành công.")

		// [MỚI] Chạy tạo Index sau khi đã có các collection
		// Chạy trong goroutine để không block luồng khởi động chính
		go createMongoIndexes(context.Background()) // Dùng context mới để không bị cancel cùng với context connect
	}

	// --- MIGRATION (Chỉ cho Postgres, dữ liệu quan hệ) ---
	fmt.Println("⏳ Đang đồng bộ hóa PostgreSQL...")
	DB.Config.DisableForeignKeyConstraintWhenMigrating = true
	err = DB.AutoMigrate(
		&models.Organization{},
		&models.Region{},
		&models.User{},
		&models.Asset{}, // Asset tạo trước để Incident có cái mà tham chiếu
		&models.EnrollmentToken{},
		&models.Incident{},
		&models.ApprovalTicket{},
		&models.Policy{},
		&models.Document{},
		&models.UserPermission{},
		&models.WhitelistItem{},
	)
	if err != nil {
		panic("🔥 LỖI MIGRATION: " + err.Error())
	}
	fmt.Println("✅ Đã đồng bộ hóa PostgreSQL thành công!")
}
