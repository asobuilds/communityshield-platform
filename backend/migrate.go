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

	// Handle NULL tracking_id values
	config.DB.Exec("UPDATE cases SET tracking_id = CONCAT('CS-', EXTRACT(YEAR FROM created_at), EXTRACT(MONTH FROM created_at), EXTRACT(DAY FROM created_at), '-', id::text) WHERE tracking_id IS NULL OR tracking_id = ''")
	config.DB.Exec("ALTER TABLE cases ADD COLUMN IF NOT EXISTS tracking_id text")
	config.DB.Exec("UPDATE cases SET tracking_id = CONCAT('CS-', EXTRACT(YEAR FROM created_at), EXTRACT(MONTH FROM created_at), EXTRACT(DAY FROM created_at), '-', id::text) WHERE tracking_id IS NULL OR tracking_id = ''")

	// Handle NULL entity_type in audit_logs - set default value
	config.DB.Exec("ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS entity_type text DEFAULT 'unknown'")
	config.DB.Exec("UPDATE audit_logs SET entity_type = 'unknown' WHERE entity_type IS NULL OR entity_type = ''")

	// Handle NULL entity_id in audit_logs
	config.DB.Exec("ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS entity_id text DEFAULT ''")
	config.DB.Exec("UPDATE audit_logs SET entity_id = '' WHERE entity_id IS NULL")

	// Handle NULL old_value and new_value
	config.DB.Exec("ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS old_value text DEFAULT ''")
	config.DB.Exec("ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS new_value text DEFAULT ''")

	// Now run AutoMigrate
	err := config.DB.AutoMigrate(
		&models.User{},
		&models.OTP{},
		&models.SecurityUnit{},
		&models.Officer{},
		&models.Case{},
		&models.Evidence{},
		&models.Progress{},
		&models.Transaction{},
		&models.TransactionApproval{},
		&models.Budget{},
		&models.FinancialReport{},
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
		&models.CommunicationRoom{},
		&models.CommunicationMessage{},
		&models.VoiceCall{},
		&models.CommunicationSync{},
		&models.PeaceCommittee{},
		&models.CommitteeMember{},
		&models.ConflictResolution{},
		&models.CommunityTrustScore{},
		&models.PeaceMetric{},
		&models.Camera{},
		&models.VideoAlert{},
		&models.SocialMediaPost{},
		&models.CaseTemplate{},
		&models.CaseTimeline{},
		&models.CaseFeedback{},
		&models.PasswordReset{},
	)

	if err != nil {
		log.Fatal("❌ Failed to migrate database:", err)
	}

	log.Println("✅ Database migration completed successfully!")
}