package routes

import (
	"github.com/gin-gonic/gin"
	"security-solution/handlers"
	"security-solution/middleware"
)

func SetupRoutes(router *gin.Engine) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "CommunityShield API is running",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		api.GET("/public/cases", handlers.GetPublicCases)
		api.GET("/public/units", handlers.GetPublicUnits)

		// Auth routes
		authHandler := handlers.NewAuthHandler()
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/profile", middleware.AuthMiddleware(), authHandler.GetProfile)
		}

		// OTP routes
		otp := api.Group("/otp")
		{
			otp.POST("/send", handlers.SendOTP)
			otp.POST("/verify", handlers.VerifyOTP)
			otp.POST("/resend", handlers.ResendOTP)
		}

		// Unit routes
		units := api.Group("/units")
		{
			units.GET("/nearby", handlers.GetNearbyUnits)
			units.GET("/by-location", handlers.GetUnitsByLocation)
			units.GET("", handlers.GetAllUnits)

			units.POST("/apply", middleware.AuthMiddleware(), handlers.ApplyForSecurityUnit)
			units.GET("/my-memberships", middleware.AuthMiddleware(), handlers.GetMyUnitMembership)
			units.POST("/government-id", middleware.AuthMiddleware(), handlers.SubmitGovernmentID)

			units.GET("/:id", handlers.GetUnitByID)
			units.POST("", middleware.AuthMiddleware(), handlers.CreateUnit)
			units.PUT("/:id", middleware.AuthMiddleware(), handlers.UpdateUnit)
		}

		// Push notification routes
		notify := api.Group("/notifications")
		{
			notify.POST("/register", middleware.AuthMiddleware(), handlers.RegisterDevice)
			notify.DELETE("/unregister", middleware.AuthMiddleware(), handlers.UnregisterDevice)
			notify.POST("/test", middleware.AuthMiddleware(), handlers.TestNotification)
		}

		// Case routes
		cases := api.Group("/cases")
		{
			cases.GET("", middleware.AuthMiddleware(), handlers.GetAllCases)
			cases.POST("", middleware.AuthMiddleware(), handlers.CreateCase)
			cases.GET("/analytics", middleware.AuthMiddleware(), handlers.GetCaseAnalytics)
			cases.GET("/:id", middleware.AuthMiddleware(), handlers.GetCaseByID)
			cases.PUT("/:id", middleware.AuthMiddleware(), handlers.UpdateCaseStatus)
			cases.POST("/:id/timeline", middleware.AuthMiddleware(), handlers.AddCaseTimeline)
			cases.GET("/:id/timeline", middleware.AuthMiddleware(), handlers.GetCaseTimeline)
			cases.POST("/:id/feedback", middleware.AuthMiddleware(), handlers.SubmitCaseFeedback)
			cases.POST("/:id/assign", middleware.AuthMiddleware(), handlers.AssignCase)
			cases.GET("/:id/assignments", middleware.AuthMiddleware(), handlers.GetCaseAssignments)
			cases.POST("/:id/dispatch", middleware.AuthMiddleware(), handlers.DispatchCase)
			cases.POST("/:id/arrive", middleware.AuthMiddleware(), handlers.ArriveAtCase)
			cases.POST("/:id/close", middleware.AuthMiddleware(), handlers.CloseCase)
			cases.POST("/:id/progress", middleware.AuthMiddleware(), handlers.AddCaseProgress)
			cases.GET("/:id/progress", middleware.AuthMiddleware(), handlers.GetCaseProgress)
		}

		// Evidence routes
		evidence := api.Group("/evidence")
		{
			evidence.POST("/upload", middleware.AuthMiddleware(), handlers.UploadEvidence)
			evidence.GET("/case/:caseId", middleware.AuthMiddleware(), handlers.GetEvidenceByCase)
			evidence.DELETE("/:id", middleware.AuthMiddleware(), handlers.DeleteEvidence)
			evidence.PATCH("/:id/verify", middleware.AuthMiddleware(), handlers.VerifyEvidence)
		}

		// Rating routes
		ratings := api.Group("/ratings")
		{
			ratings.GET("/units/:unitId", handlers.GetUnitRatings)
			ratings.POST("", middleware.AuthMiddleware(), handlers.SubmitRating)
		}

		// SOS routes
		sos := api.Group("/sos")
		{
			sos.POST("/send", middleware.AuthMiddleware(), handlers.SendSOSAlert)
			sos.GET("", middleware.AuthMiddleware(), handlers.GetSOSAlerts)
			sos.GET("/my", middleware.AuthMiddleware(), handlers.GetUserSOSAlerts)
			sos.GET("/:id", middleware.AuthMiddleware(), handlers.GetSOSAlertByID)
			sos.PUT("/:id/status", middleware.AuthMiddleware(), handlers.UpdateSOSAlertStatus)
		}

		// Suspect routes
		suspects := api.Group("/suspects")
		{
			suspects.POST("", middleware.AuthMiddleware(), handlers.CreateSuspect)
			suspects.GET("", middleware.AuthMiddleware(), handlers.GetAllSuspects)
			suspects.GET("/:id", middleware.AuthMiddleware(), handlers.GetSuspectByID)
			suspects.PUT("/:id", middleware.AuthMiddleware(), handlers.UpdateSuspect)
			suspects.DELETE("/:id", middleware.AuthMiddleware(), handlers.DeleteSuspect)
			suspects.POST("/:id/sighting", middleware.AuthMiddleware(), handlers.ReportSighting)
			suspects.GET("/:id/sightings", middleware.AuthMiddleware(), handlers.GetSuspectSightings)
			suspects.GET("/:id/cases", middleware.AuthMiddleware(), handlers.GetSuspectCases)
			suspects.POST("/:id/associations", middleware.AuthMiddleware(), handlers.CreateSuspectAssociation)
		}

		// Transfer routes
		transfers := api.Group("/transfers")
		{
			transfers.POST("", middleware.AuthMiddleware(), handlers.RequestTransfer)
			transfers.POST("/:id/approve", middleware.AuthMiddleware(), handlers.ApproveTransfer)
			transfers.POST("/:id/reject", middleware.AuthMiddleware(), handlers.RejectTransfer)
			transfers.GET("", middleware.AuthMiddleware(), handlers.GetTransferRequests)
			transfers.GET("/:id/approvals", middleware.AuthMiddleware(), handlers.GetTransferApprovals)
		}

		// News routes
		news := api.Group("/news")
		{
			news.POST("", middleware.AuthMiddleware(), handlers.CreateNews)
			news.GET("", middleware.AuthMiddleware(), handlers.GetAllNews)
			news.GET("/:id", middleware.AuthMiddleware(), handlers.GetNewsByID)
		}

		// AI routes
		ai := api.Group("/ai")
		{
			ai.POST("/chatbot", middleware.AuthMiddleware(), handlers.AIChatbot)
			ai.POST("/analyze-image", middleware.AuthMiddleware(), handlers.AIAnalyzeImage)
			ai.POST("/analyze-location", middleware.AuthMiddleware(), handlers.AIAnalyzeLocation)
			ai.POST("/map-insights", middleware.AuthMiddleware(), handlers.AIGetMapInsights)
			ai.POST("/security-warning", middleware.AuthMiddleware(), handlers.AIGenerateSecurityWarning)
			ai.POST("/analyze-news", middleware.AuthMiddleware(), handlers.AIAnalyzeNews)
			ai.POST("/smart-tips", middleware.AuthMiddleware(), handlers.AIGetSmartTips)
			ai.POST("/predict-hotspots", middleware.AuthMiddleware(), handlers.AIPredictHotspots)
		}

		// Bank Account routes
		bank := api.Group("/bank")
		{
			bank.POST("/accounts", middleware.AuthMiddleware(), handlers.AddBankAccount)
			bank.GET("/:unitId/accounts", middleware.AuthMiddleware(), handlers.GetBankAccounts)
			bank.PUT("/accounts/:id", middleware.AuthMiddleware(), handlers.UpdateBankAccount)
			bank.DELETE("/accounts/:id", middleware.AuthMiddleware(), handlers.DeleteBankAccount)
			bank.POST("/donations", middleware.AuthMiddleware(), handlers.RecordDonation)
			bank.POST("/donations/:id/confirm", middleware.AuthMiddleware(), handlers.ConfirmDonation)
			bank.GET("/:unitId/donations", middleware.AuthMiddleware(), handlers.GetDonations)
		}

		// Finance routes
		finance := api.Group("/finance")
		{
			finance.POST("/transactions", middleware.AuthMiddleware(), handlers.CreateTransaction)
			finance.POST("/transactions/:id/approve", middleware.AuthMiddleware(), handlers.ApproveTransaction)
			finance.POST("/transactions/:id/reject", middleware.AuthMiddleware(), handlers.RejectTransaction)
			finance.GET("/units/:unitId/transactions", middleware.AuthMiddleware(), handlers.GetTransactions)
			finance.GET("/transactions/:id", middleware.AuthMiddleware(), handlers.GetTransactionByID)
			finance.GET("/units/:unitId/summary", middleware.AuthMiddleware(), handlers.GetTransactionSummary)
			finance.POST("/budgets", middleware.AuthMiddleware(), handlers.CreateBudget)
			finance.GET("/units/:unitId/budgets", middleware.AuthMiddleware(), handlers.GetBudgets)
			finance.POST("/reports", middleware.AuthMiddleware(), handlers.GenerateFinancialReport)
			finance.GET("/units/:unitId/reports", middleware.AuthMiddleware(), handlers.GetFinancialReports)
		}

		// Community routes
		community := api.Group("/community")
		{
			community.POST("/posts", middleware.AuthMiddleware(), handlers.CreateForumPost)
			community.GET("/posts", middleware.AuthMiddleware(), handlers.GetForumPosts)
			community.GET("/posts/:id", middleware.AuthMiddleware(), handlers.GetForumPostByID)
			community.POST("/replies", middleware.AuthMiddleware(), handlers.CreateForumReply)
			community.POST("/announcements", middleware.AuthMiddleware(), handlers.CreateCommunityAnnouncement)
			community.GET("/announcements", middleware.AuthMiddleware(), handlers.GetCommunityAnnouncements)
			community.POST("/events", middleware.AuthMiddleware(), handlers.CreateCommunityEvent)
			community.GET("/events", middleware.AuthMiddleware(), handlers.GetCommunityEvents)
			community.POST("/events/:id/rsvp", middleware.AuthMiddleware(), handlers.RSVPToEvent)
		}

		// Audit routes
		audit := api.Group("/audit")
		{
			audit.POST("/activity", middleware.AuthMiddleware(), handlers.LogActivity)
			audit.GET("/activities", middleware.AuthMiddleware(), handlers.GetActivityLogs)
			audit.GET("/logs", middleware.AuthMiddleware(), handlers.GetAuditLogs)
			audit.POST("/logs", middleware.AuthMiddleware(), handlers.CreateAuditLog)
			audit.GET("/health", middleware.AuthMiddleware(), handlers.GetSystemHealth)
			audit.POST("/health", middleware.AuthMiddleware(), handlers.UpdateSystemHealth)
			audit.GET("/notifications", middleware.AuthMiddleware(), handlers.GetNotificationLogs)
		}

		// Alert routes
		// Alert routes
		alerts := api.Group("/alerts")
		{
			alerts.GET("/news", middleware.AuthMiddleware(), handlers.GetNewsAlerts)

			alerts.POST("", middleware.AuthMiddleware(), handlers.CreateCommunityAlert)
			alerts.GET("", middleware.AuthMiddleware(), handlers.GetCommunityAlerts)

			alerts.POST("/subscribe", middleware.AuthMiddleware(), handlers.SubscribeToAlerts)
			alerts.GET("/subscriptions", middleware.AuthMiddleware(), handlers.GetAlertSubscriptions)

			alerts.GET("/:id", middleware.AuthMiddleware(), handlers.GetAlertByID)
			alerts.POST("/:id/confirm", middleware.AuthMiddleware(), handlers.ConfirmAlert)
		}

		// Settings routes
		settings := api.Group("/settings")
		{
			settings.GET("/public", handlers.GetPublicSettings)

			settings.GET("/templates/:name", middleware.AuthMiddleware(), handlers.GetEmailTemplate)
			settings.PUT("/templates/:name", middleware.AuthMiddleware(), handlers.UpdateEmailTemplate)

			settings.POST("/exports", middleware.AuthMiddleware(), handlers.CreateDataExport)
			settings.GET("/exports", middleware.AuthMiddleware(), handlers.GetDataExports)

			settings.GET("/onboarding", middleware.AuthMiddleware(), handlers.GetUserOnboarding)
			settings.PUT("/onboarding", middleware.AuthMiddleware(), handlers.UpdateUserOnboarding)

			settings.GET("/:key", middleware.AuthMiddleware(), handlers.GetSystemSetting)
			settings.PUT("/:key", middleware.AuthMiddleware(), handlers.UpdateSystemSetting)
		}

		// Mobile API routes
		mobile := api.Group("/mobile")
		{
			mobile.GET("/config", middleware.AuthMiddleware(), handlers.MobileAppConfig)
			mobile.GET("/dashboard", middleware.AuthMiddleware(), handlers.MobileDashboard)
			mobile.GET("/notifications", middleware.AuthMiddleware(), handlers.MobileNotifications)
			mobile.PUT("/notifications/:id/read", middleware.AuthMiddleware(), handlers.MobileMarkNotificationRead)
			mobile.PUT("/notifications/read-all", middleware.AuthMiddleware(), handlers.MobileMarkAllRead)
			mobile.GET("/sync", middleware.AuthMiddleware(), handlers.MobileSync)
			mobile.POST("/crash-report", middleware.AuthMiddleware(), handlers.MobileCrashReport)
		}

		// Video Analytics routes
		video := api.Group("/video")
		{
			video.POST("/cameras", middleware.AuthMiddleware(), handlers.AddCamera)
			video.GET("/units/:unitId/cameras", middleware.AuthMiddleware(), handlers.GetCameras)
			video.POST("/alerts", middleware.AuthMiddleware(), handlers.GenerateVideoAlert)
			video.GET("/alerts", middleware.AuthMiddleware(), handlers.GetVideoAlerts)
			video.PUT("/alerts/:id/review", middleware.AuthMiddleware(), handlers.ReviewVideoAlert)
			video.POST("/social/monitor", middleware.AuthMiddleware(), handlers.MonitorSocialMedia)
			video.GET("/social/posts", middleware.AuthMiddleware(), handlers.GetSocialMediaPosts)
		}

		// Communication routes
		commGroup := api.Group("/communication")
		{
			commGroup.POST("/rooms", middleware.AuthMiddleware(), handlers.CreateRoom)
			commGroup.GET("/units/:unitId/rooms", middleware.AuthMiddleware(), handlers.GetRooms)
			commGroup.POST("/messages", middleware.AuthMiddleware(), handlers.SendMessage)
			commGroup.GET("/rooms/:roomId/messages", middleware.AuthMiddleware(), handlers.GetMessages)
			commGroup.POST("/calls", middleware.AuthMiddleware(), handlers.InitiateCall)
			commGroup.PUT("/calls/:id/end", middleware.AuthMiddleware(), handlers.EndCall)
			commGroup.POST("/sync", middleware.AuthMiddleware(), handlers.SyncMessages)
			commGroup.GET("/rooms/:roomId/sync-status", middleware.AuthMiddleware(), handlers.GetSyncStatus)
		}

		// SMS routes
		sms := api.Group("/sms")
		{
			sms.POST("/incoming", handlers.HandleIncomingSMS)
			sms.POST("/ussd", handlers.HandleUSSD)
		}

		// Peacebuilding routes
		peace := api.Group("/peacebuilding")
		{
			peace.POST("/committees", middleware.AuthMiddleware(), handlers.CreatePeaceCommittee)
			peace.GET("/units/:unitId/committees", middleware.AuthMiddleware(), handlers.GetPeaceCommittees)
			peace.POST("/committees/:id/members", middleware.AuthMiddleware(), handlers.AddCommitteeMember)
			peace.POST("/conflicts", middleware.AuthMiddleware(), handlers.CreateConflictResolution)
			peace.GET("/units/:unitId/conflicts", middleware.AuthMiddleware(), handlers.GetConflictResolutions)
			peace.PUT("/conflicts/:id", middleware.AuthMiddleware(), handlers.UpdateConflictResolution)
			peace.GET("/units/:unitId/peace-metrics", middleware.AuthMiddleware(), handlers.GetPeaceMetrics)
			peace.GET("/units/:unitId/trust-scores", middleware.AuthMiddleware(), handlers.GetTrustScores)
			peace.POST("/trust-scores", middleware.AuthMiddleware(), handlers.UpdateTrustScore)
		}

		// Super Admin routes
		superAdmin := api.Group("/admin")
		superAdmin.Use(middleware.AuthMiddleware(), handlers.SuperAdminMiddleware())
		{
			superAdmin.GET("/users", handlers.GetAllUsers)
			superAdmin.GET("/users/:id", handlers.GetUserByID)
			superAdmin.PUT("/users/:id/role", handlers.UpdateUserRole)
			superAdmin.POST("/users/:id/suspend", handlers.SuspendUser)
			superAdmin.POST("/users/:id/activate", handlers.ActivateUser)
			superAdmin.POST("/users/:id/impersonate", handlers.ImpersonateUser)
			superAdmin.POST("/stop-impersonate", handlers.StopImpersonation)
			superAdmin.GET("/stats", handlers.GetSystemStats)
		}
	}

	// WebSocket route (protected)
	router.GET("/ws", middleware.AuthMiddleware(), handlers.HandleWebSocket)
}
