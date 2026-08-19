package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"security-solution/config"
	"security-solution/middleware"
	"security-solution/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, using system environment variables")
	}

	config.ConnectDatabase()
	defer config.CloseDatabase()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(os.Getenv("GIN_MODE"))
	}

	router := gin.Default()

	// Add audit middleware
	router.Use(middleware.AuditMiddleware())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on http://localhost:%s", port)
	log.Printf("📝 Environment: %s", os.Getenv("GIN_MODE"))

	if err := router.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}