package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// RiskPredictionService handles AI predictions
type RiskPredictionService struct{}

// GenerateLocationRisk calculates risk score for a location asynchronously
func GenerateLocationRisk(db *gorm.DB, lat, lng float64, location string) string {
	// Simulate AI processing (in reality, this would call OpenRouter or a ML model)
	// We are adding a sleep to simulate an "AI" call, but in the future, this will be a real API call.
	time.Sleep(1 * time.Second) // Simulate "Think" time

	var count int64
	// Query to find similar cases in the last 30 days
	db.Model(&models.Case{}).
		Where("created_at > ? AND (latitude BETWEEN ? AND ?) AND (longitude BETWEEN ? AND ?)",
			time.Now().AddDate(0, 0, -30),
			lat-0.1, lat+0.1,
			lng-0.1, lng+0.1,
		).
		Count(&count)

	var riskLevel string
	var riskScore int
	if count > 10 {
		riskLevel = "High"
		riskScore = 80
	} else if count > 3 {
		riskLevel = "Medium"
		riskScore = 50
	} else {
		riskLevel = "Low"
		riskScore = 10
	}

	return fmt.Sprintf("Risk Level: %s (Score: %d/100) based on %d recent incidents", riskLevel, riskScore, count)
}