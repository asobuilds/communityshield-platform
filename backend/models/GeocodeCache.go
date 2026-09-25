package models

import "time"

// GeocodeCache stores reverse-geocoded addresses keyed by geohash.
// Nominatim has a 1 req/sec usage policy — we cache every lookup
// and only refresh after CacheTTL (default 30 days).
type GeocodeCache struct {
	ID         uint      `gorm:"primaryKey"`
	Geohash    string    `gorm:"type:varchar(12);uniqueIndex;not null"`
	Latitude   float64   `gorm:"not null"`
	Longitude  float64   `gorm:"not null"`
	Address    string    `gorm:"type:text" json:"address,omitempty"`
	Suburb     string    `gorm:"type:varchar(120)" json:"suburb,omitempty"`
	LGA        string    `gorm:"type:varchar(120);index" json:"lga,omitempty"`
	State      string    `gorm:"type:varchar(120);index" json:"state,omitempty"`
	Country    string    `gorm:"type:varchar(80)" json:"country,omitempty"`
	PostalCode string    `gorm:"type:varchar(20)" json:"postalCode,omitempty"`
	Landmark   string    `gorm:"type:varchar(200)" json:"landmark,omitempty"`
	Source     string    `gorm:"type:varchar(32);default:'nominatim'" json:"source"`
	Raw        string    `gorm:"type:text" json:"-"` // full raw JSON response
	FetchedAt  time.Time `gorm:"not null;index" json:"fetchedAt"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (GeocodeCache) TableName() string { return "geocode_cache" }
