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
	// Use DATABASE_URL from environment
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback to individual variables
		host := os.Getenv("DB_HOST")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		port := os.Getenv("DB_PORT")

		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5432"
		}
		if user == "" {
			user = "postgres"
		}
		if dbname == "" {
			dbname = "postgres"
		}
		dsn = "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=require TimeZone=UTC"
	}

	log.Println("📊 Connecting to database...")

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	log.Println("✅ Database connected successfully (PostgreSQL)")
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