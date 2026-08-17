package main

import (
	"log"
	"net/http"
	"security-solution/handlers"
	"security-solution/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=security_platform port=5432 sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("✅ Database migrated!")

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

	r.Static("/uploads", "./uploads") // Serve uploaded files
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
	}

	log.Println("🚀 Server running on http://localhost:8080")
	r.Run(":8080")
}