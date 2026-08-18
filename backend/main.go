package main

import (
	"log"
	"net/http"
	"os"
	"security-solution/handlers"
	"security-solution/models"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	// Get DATABASE_URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("✅ Database connected!")

	handlers.DB = DB

	err = DB.AutoMigrate(
		&models.User{},
		&models.Case{},
		&models.Evidence{},
		&models.SOSAlert{},
		&models.Unit{},
		&models.Transaction{},
		&models.CaseProgress{},
		&models.Rating{},
		&models.Notification{},
		&models.Announcement{},
		&models.CaseOfficer{},
		&models.AIAnalysis{},
		&models.ChatMessage{},
		&models.CaseTransfer{},
		&models.OTP{},
		&models.PushSubscription{},
		&models.Report{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("✅ Database migrated!")

	if err := handlers.InitSMS(); err != nil {
		log.Println("⚠️ SMS initialization failed:", err)
	} else {
		log.Println("✅ SMS service initialized")
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.Use(func(c *gin.Context) {
		log.Printf("📩 %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/forgot-password", handlers.ForgotPassword)
			auth.POST("/reset-password", handlers.ResetPassword)
			auth.POST("/register-unit", handlers.RegisterUnit)
		}

		cases := api.Group("/cases")
		{
			cases.POST("", handlers.CreateCase)
			cases.GET("", handlers.GetCasesByUser)
			cases.POST("/:id/progress", handlers.AddProgress)
			cases.GET("/:id/progress", handlers.GetProgress)
			cases.POST("/:id/evidence", handlers.UploadEvidence)
			cases.GET("/:id/evidence", handlers.GetEvidence)
			cases.GET("/:id", handlers.GetCase)
			cases.POST("/:id/assign", handlers.AssignOfficerToCase)
			cases.DELETE("/:id/assign", handlers.RemoveOfficerFromCase)
			cases.GET("/:id/officers", handlers.GetAssignedOfficers)
			cases.POST("/:id/request-transfer", handlers.RequestCaseTransfer)
			cases.GET("/:id/transfer-status", handlers.GetCaseTransferStatus)
			cases.PUT("/:transferId/approve", handlers.ApproveCaseTransfer)
			cases.PUT("/:transferId/reject", handlers.RejectCaseTransfer)
			cases.GET("/transfers/pending", handlers.GetPendingTransfers)
		}

		sos := api.Group("/sos")
		{
			sos.POST("", handlers.SendSOS)
			sos.GET("/history", handlers.GetSOSHistory)
		}

		units := api.Group("/units")
		{
			units.GET("/nearby", handlers.GetNearbyUnits)
		}

		users := api.Group("/users")
		{
			users.PUT("/profile-picture", handlers.UploadProfilePicture)
			users.GET("/:userId", handlers.GetUserProfile)
			users.PUT("/profile", handlers.UpdateUserProfile)
			users.GET("/:userId/used-units", handlers.GetUserUsedUnits)
		}

		announcements := api.Group("/announcements")
		{
			announcements.POST("", handlers.CreateAnnouncement)
			announcements.GET("", handlers.GetAnnouncements)
			announcements.GET("/:id", handlers.GetAnnouncement)
			announcements.PUT("/:id", handlers.UpdateAnnouncement)
			announcements.DELETE("/:id", handlers.DeleteAnnouncement)
		}

		ai := api.Group("/ai")
		{
			ai.POST("/summarize", handlers.SummarizeText)
			ai.POST("/cases/:id/summarize", handlers.SummarizeCase)
			ai.GET("/analysis", handlers.GetAIAnalysis)
			ai.POST("/monitor/rss", handlers.MonitorRSSFeeds)
			ai.POST("/monitor/social", handlers.SocialMediaMonitor)
		}

		otp := api.Group("/otp")
		{
			otp.POST("/send", handlers.SendOTP)
			otp.POST("/verify", handlers.VerifyOTP)
			otp.POST("/resend", handlers.ResendOTP)
		}

		push := api.Group("/push")
		{
			push.POST("/subscribe", handlers.SubscribePush)
			push.DELETE("/unsubscribe", handlers.UnsubscribePush)
			push.GET("/vapid-public-key", handlers.GetVAPIDPublicKey)
		}

		reports := api.Group("/reports")
		{
			reports.POST("", handlers.CreateReport)
		}

		admin := api.Group("/admin")
		{
			admin.GET("/sos", handlers.AdminGetSOSAlerts)
			admin.PATCH("/sos/:id/status", handlers.AdminUpdateSOSStatus)
			admin.GET("/cases", handlers.AdminGetCases)
			admin.PATCH("/cases/:id/status", handlers.AdminUpdateCaseStatus)
			admin.GET("/units/:unitId/officers", handlers.GetUnitOfficers)
			admin.POST("/units/:unitId/officers", handlers.AddOfficer)
			admin.DELETE("/officers/:userId", handlers.RemoveOfficer)
			admin.POST("/units", handlers.CreateUnit)
			admin.GET("/units", handlers.GetUnits)
			admin.GET("/units/:unitId", handlers.GetUnit)
			admin.PUT("/units/:unitId", handlers.UpdateUnit)
			admin.DELETE("/units/:unitId", handlers.DeleteUnit)
			admin.POST("/units/:unitId/assign-admin", handlers.AssignUnitAdmin)
			admin.PUT("/units/:unitId/profile-picture", handlers.UploadUnitProfilePicture)
			admin.GET("/analytics/cases", handlers.GetCaseAnalytics)
			admin.GET("/analytics/sos", handlers.GetSOSAnalytics)
			admin.GET("/analytics/units", handlers.GetUnitAnalytics)
			admin.GET("/cases/archived", handlers.AdminGetArchivedCases)
			admin.PUT("/cases/:id/restore", handlers.RestoreArchivedCase)
			admin.DELETE("/cases/:id/permanent", handlers.AdminDeleteArchivedCase)
			admin.GET("/reports", handlers.AdminGetReports)
			admin.PUT("/reports/:id/status", handlers.AdminUpdateReportStatus)
		}

		transactions := api.Group("/transactions")
		{
			transactions.POST("", handlers.CreateTransaction)
			transactions.GET("/units/:unitId", handlers.GetUnitTransactions)
			transactions.GET("/units/:unitId/report", handlers.GetUnitReport)
			transactions.GET("/units/:unitId/statement", handlers.GetUnitStatement)
			transactions.PATCH("/:id/approve", handlers.ApproveTransaction)
		}

		ratings := api.Group("/ratings")
		{
			ratings.POST("", handlers.SubmitRating)
			ratings.GET("/units/:unitId", handlers.GetUnitRating)
			ratings.GET("/units/:unitId/reviews", handlers.GetUnitReviews)
			ratings.GET("/cases/:caseId/check", handlers.CheckUserRating)
		}

		notifications := api.Group("/notifications")
		{
			notifications.GET("/user", handlers.GetUserNotifications)
			notifications.PATCH("/:id/read", handlers.MarkNotificationRead)
			notifications.PATCH("/read-all", handlers.MarkAllNotificationsRead)
		}

		api.GET("/ws", handlers.WebSocketHandler)
		api.GET("/chat/history", handlers.GetChatHistory)
		api.GET("/officers/week", handlers.GetOfficersOfTheWeek)
	}

	go func() {
		for {
			time.Sleep(24 * time.Hour)
			if err := handlers.AutoArchiveCases(); err != nil {
				log.Println("⚠️ Auto-archive failed:", err)
			} else {
				log.Println("✅ Auto-archive completed")
			}
		}
	}()

	log.Println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}