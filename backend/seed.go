package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"security-solution/config"
	"security-solution/models"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found")
	}

	config.ConnectDatabase()

	// Create a test unit
	unit := models.SecurityUnit{
		ID:               uuid.New(),
		Name:             "Lagos East Security Unit",
		Type:             "community_police",
		Latitude:         6.5244,
		Longitude:        3.3792,
		OperationalRadius: 20,
		State:            "Lagos",
		City:             "Lagos",
		ContactPerson:    "Chief Security Officer",
		ContactPhone:     "08012345678",
		Status:           "active",
		IsVerified:       true,
	}
	config.DB.Create(&unit)

	// Create a test case
	caseObj := models.Case{
		ID:           uuid.New(),
		UnitID:       unit.ID,
		Title:        "Suspicious Activity at Market",
		Description:  "Unknown individuals seen loitering around the market area late at night. Residents are concerned about their safety.",
		Location:     "Lagos Mainland Market",
		Latitude:     6.5244,
		Longitude:    3.3792,
		Status:       "investigating",
		Priority:     "high",
		TrackingID:   "CS-TEST-001",
		PriorityLevel: "P2",
		IsPublic:     true,
		IncidentDate: time.Now().Add(-2 * time.Hour),
	}
	config.DB.Create(&caseObj)

	// Create a second case
	caseObj2 := models.Case{
		ID:           uuid.New(),
		UnitID:       unit.ID,
		Title:        "Theft at Local Shop",
		Description:  "A local shop was broken into last night. Items worth N500,000 were stolen.",
		Location:     "Ikeja, Lagos",
		Latitude:     6.6018,
		Longitude:    3.3515,
		Status:       "pending",
		Priority:     "medium",
		TrackingID:   "CS-TEST-002",
		PriorityLevel: "P3",
		IsPublic:     true,
		IncidentDate: time.Now().Add(-5 * time.Hour),
	}
	config.DB.Create(&caseObj2)

	// Create news articles
	news := []models.News{
		{
			ID:          uuid.New(),
			Title:       "Lagos Launches New Security Initiative",
			Content:     "The Lagos State Government has launched a new security initiative to combat rising crime rates in the state. The initiative includes increased police patrols, community engagement programs, and the deployment of new technology.",
			Source:      "Lagos State Government",
			Category:    "security",
			Location:    "Lagos",
			Sentiment:   "positive",
			ThreatLevel: "low",
			Status:      "published",
			PublishedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Title:       "Community Policing Success Story in Lagos",
			Content:     "A community policing initiative in Lagos has led to a 30% reduction in reported crimes in the area. Residents credit the success to better collaboration between citizens and security units.",
			Source:      "Community News",
			Category:    "community",
			Location:    "Lagos",
			Sentiment:   "positive",
			ThreatLevel: "low",
			Status:      "published",
			PublishedAt: time.Now().Add(-3 * time.Hour),
		},
		{
			ID:          uuid.New(),
			Title:       "Security Alert: Robbery in Ikeja",
			Content:     "A robbery occurred in Ikeja today at approximately 2 PM. Two suspects are at large. Residents are advised to be vigilant and report any suspicious activity.",
			Source:      "Local News",
			Category:    "security",
			Location:    "Ikeja, Lagos",
			Sentiment:   "negative",
			ThreatLevel: "high",
			Status:      "published",
			PublishedAt: time.Now().Add(-30 * time.Minute),
		},
	}
	for _, n := range news {
		config.DB.Create(&n)
	}

	// Create a suspect
	suspect := models.Suspect{
		ID:          uuid.New(),
		FirstName:   "John",
		LastName:    "Doe",
		Alias:       "The Shadow",
		Gender:      "Male",
		Nationality: "Nigerian",
		Description: "Tall, dark complexion, usually wears black clothing. Known to operate in Lagos Mainland area.",
		Status:      "active",
		DangerLevel: "high",
		Wanted:      true,
		Category:    "theft",
		UnitID:      &unit.ID,
	}
	config.DB.Create(&suspect)

	log.Println("✅ Sample data seeded successfully!")
}