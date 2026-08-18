package main

import (
	"log"

	"github.com/joho/godotenv"

	"security-solution/config"
	"security-solution/models"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found")
	}

	// Connect to database
	config.ConnectDatabase()

	// Auto migrate all models
	err := config.DB.AutoMigrate(
		&models.User{},
		&models.OTP{},
		&models.SecurityUnit{},
		&models.Officer{},
		&models.Case{},
		&models.Evidence{},
		&models.Progress{},
		&models.Transaction{},
		&models.Notification{},
		&models.PushToken{},
		&models.Rating{},
		&models.SOSAlert{},
		&models.Announcement{},
		&models.CaseTransfer{},
	)

	if err != nil {
		log.Fatal("❌ Failed to migrate database:", err)
	}

	log.Println("✅ Database migration completed successfully!")
}