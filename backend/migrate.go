package main

import (
	"log"

	"github.com/joho/godotenv"

	"security-solution/config"
	"security-solution/models"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found")
	}

	config.ConnectDatabase()

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
		&models.CommunityAlert{},
		&models.AlertSubscription{},
		&models.BankAccount{},
		&models.Donation{},
		&models.Suspect{},
		&models.SuspectSighting{},
		&models.SuspectCase{},
		&models.TransferRequest{},
		&models.TransferApproval{},
		&models.CommunityAnnouncement{},
		&models.CommunityEvent{},
		&models.EventAttendee{},
		&models.ForumPost{},
		&models.ForumReply{},
		&models.AuditLog{},
		&models.SystemHealth{},
		&models.ActivityLog{},
		&models.NotificationLog{},
		&models.SystemSettings{},
		&models.EmailTemplate{},
		&models.DataExport{},
		&models.UserOnboarding{},
		&models.SystemBackup{},
	)

	if err != nil {
		log.Fatal("❌ Failed to migrate database:", err)
	}

	log.Println("✅ Database migration completed successfully!")
}