package config

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("❌ DATABASE_URL environment variable not set")
	}

	log.Println("📊 Connecting to database...")

	var err error
	// Disable prepared statements completely
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
		PreferSimpleProtocol: true, // This forces simple protocol, no prepared statements
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: false,
		SkipDefaultTransaction: true, // Skip default transaction
	})

	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("❌ Failed to get database instance:", err)
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("❌ Database ping failed:", err)
	}

	log.Println("✅ Database connected successfully")
}

func CloseDatabase() {
	if DB == nil {
		return
	}
	sqlDB, err := DB.DB()
	if err != nil {
		log.Println("⚠️ Error getting database instance:", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Println("⚠️ Error closing database:", err)
	}
}