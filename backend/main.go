package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"security-solution/config"
	"security-solution/handlers"
	"security-solution/middleware"
	"security-solution/routes"
	"security-solution/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	config.EnforceRequiredEnv()
	config.ConnectDatabase()
	defer config.CloseDatabase()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(os.Getenv("GIN_MODE"))
	}

	router := gin.Default()

	// Structured logging middleware: assigns a request ID, echoes it back
	// in the X-Request-Id response header, and emits one structured log
	// line per completed request (JSON or text, per LOG_FORMAT).
	// Outermost of our three custom middlewares.
	router.Use(middleware.StructuredLogger())

	// Add audit middleware
	router.Use(middleware.AuditMiddleware())

	// Panic recovery middleware: innermost of our three. Recovers from
	// handler panics, reports via the ErrorReporter seam, and returns a
	// generic 500 so StructuredLogger and Audit see the final status.
	router.Use(middleware.PanicRecovery())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(router)

	// Wire out-of-band delivery for password resets. The services package
	// defines the interface; the handlers package provides the SMTP-backed
	// implementation. This keeps services free of any handlers import.
	services.SetResetNotifier(handlers.NewEmailNotifier())

	// Start background scheduler (elections, invites, expiry, account purge)
	scheduler := services.NewSchedulerService()
	scheduler.Start(1 * time.Hour)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://localhost:%s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
