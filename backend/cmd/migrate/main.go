package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"security-solution/models"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Using environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("database connection failed:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	log.Println("Connected to security_platform")
	log.Println("Running schema migration...")

	err = db.AutoMigrate(
		&models.User{},
		&models.SecurityUnit{},
		&models.UnitMember{},
		&models.GovernmentIDVerification{},
		&models.Officer{},

		&models.Case{},
		&models.Evidence{},
		&models.Progress{},
		&models.CaseFeedback{},
		&models.CaseOfficer{},
		&models.CaseProgress{},
		&models.CaseTemplate{},
		&models.CaseTimeline{},
		&models.CaseTransfer{},
		&models.TransferRequest{},
		&models.TransferApproval{},

		&models.Rating{},
		&models.Notification{},
		&models.OTP{},
		&models.PasswordReset{},
		&models.PushSubscription{},
		&models.PushToken{},

		&models.AIAnalysis{},
		&models.Alert{},
		&models.Announcement{},
		&models.News{},
		&models.NewsAlert{},

		&models.SOSAlert{},
		&models.BankAccount{},
		&models.Donation{},
		&models.Transaction{},
		&models.TransactionApproval{},
		&models.Budget{},
		&models.FinancialReport{},

		&models.ForumPost{},
		&models.ForumReply{},
		&models.CommunityAnnouncement{},
		&models.CommunityEvent{},
		&models.EventAttendee{},
		&models.CommunityAlert{},
		&models.AlertSubscription{},

		&models.AuditLog{},
		&models.SystemHealth{},
		&models.ActivityLog{},
		&models.NotificationLog{},

		&models.DataExport{},
		&models.EmailTemplate{},
		&models.SystemSettings{},
		&models.SystemBackup{},
		&models.UserOnboarding{},

		&models.PeaceCommittee{},
		&models.CommitteeMember{},
		&models.ConflictResolution{},
		&models.CommunityTrustScore{},
		&models.PeaceMetric{},

		&models.Suspect{},
		&models.SuspectAssociation{},
		&models.SuspectSighting{},
		&models.SuspectCase{},

		&models.Camera{},
		&models.VideoAlert{},
		&models.SocialMediaPost{},

		&models.CommunicationRoom{},
		&models.CommunicationMessage{},
		&models.VoiceCall{},
		&models.CommunicationSync{},

		&models.ChatMessage{},
		&models.Report{},
	)

	if err != nil {
		log.Fatal("schema migration failed:", err)
	}

	log.Println("Schema migration completed successfully.")
}
