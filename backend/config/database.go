package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase() {
	// Get environment variables
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	// Set defaults if not provided
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "postgres"
	}
	if password == "" {
		password = "postgres"  // Your password
	}
	if dbname == "" {
		dbname = "postgres"  // Use default database
	}

	// Proper DSN format
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, user, password, dbname, port)

	log.Printf("📊 Connecting to database: %s@%s:%s/%s", user, host, port, dbname)

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