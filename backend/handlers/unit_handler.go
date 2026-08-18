package handlers

import (
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// GetNearbyUnits returns units near a location
func GetNearbyUnits(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	radiusStr := c.Query("radius")

	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Latitude and longitude required"})
		return
	}

	lat, err := parseFloat(latStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude"})
		return
	}

	lng, err := parseFloat(lngStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude"})
		return
	}

	radius := 20.0 // Default 20km
	if radiusStr != "" {
		radius, _ = parseFloat(radiusStr)
	}

	var units []models.SecurityUnit
	if err := config.DB.Where("status = ?", "active").Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units"})
		return
	}

	// Calculate distance for each unit
	type UnitWithDistance struct {
		models.SecurityUnit
		Distance float64 `json:"distance"`
	}

	var result []UnitWithDistance
	for _, unit := range units {
		if unit.Latitude == 0 || unit.Longitude == 0 {
			continue
		}
		distance := haversine(lat, lng, unit.Latitude, unit.Longitude)
		if distance <= radius {
			result = append(result, UnitWithDistance{
				SecurityUnit: unit,
				Distance:     distance,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"units": result,
	})
}

// GetAllUnits returns all units
func GetAllUnits(c *gin.Context) {
	var units []models.SecurityUnit
	if err := config.DB.Where("status = ?", "active").Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"units": units,
	})
}

// GetUnitByID returns a specific unit
func GetUnitByID(c *gin.Context) {
	id := c.Param("id")
	unitID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var unit models.SecurityUnit
	if err := config.DB.First(&unit, "id = ?", unitID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unit": unit,
	})
}

// CreateUnit creates a new unit (admin only)
func CreateUnit(c *gin.Context) {
	var input struct {
		Name             string  `json:"name" binding:"required"`
		Type             string  `json:"type" binding:"required"`
		Latitude         float64 `json:"latitude"`
		Longitude        float64 `json:"longitude"`
		CoverageArea     string  `json:"coverageArea"`
		ContactPerson    string  `json:"contactPerson"`
		ContactPhone     string  `json:"contactPhone"`
		ContactEmail     string  `json:"contactEmail"`
		RegistrationNumber string `json:"registrationNumber"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unit := models.SecurityUnit{
		Name:             input.Name,
		Type:             input.Type,
		Latitude:         input.Latitude,
		Longitude:        input.Longitude,
		CoverageArea:     input.CoverageArea,
		ContactPerson:    input.ContactPerson,
		ContactPhone:     input.ContactPhone,
		ContactEmail:     input.ContactEmail,
		RegistrationNumber: input.RegistrationNumber,
		Status:           "active",
	}

	if err := config.DB.Create(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create unit"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Unit created successfully",
		"unit":    unit,
	})
}

// UpdateUnit updates a unit (admin only)
func UpdateUnit(c *gin.Context) {
	id := c.Param("id")
	unitID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var input struct {
		Name          string  `json:"name"`
		Type          string  `json:"type"`
		Latitude      float64 `json:"latitude"`
		Longitude     float64 `json:"longitude"`
		CoverageArea  string  `json:"coverageArea"`
		ContactPerson string  `json:"contactPerson"`
		ContactPhone  string  `json:"contactPhone"`
		ContactEmail  string  `json:"contactEmail"`
		Status        string  `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var unit models.SecurityUnit
	if err := config.DB.First(&unit, "id = ?", unitID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
		return
	}

	if input.Name != "" {
		unit.Name = input.Name
	}
	if input.Type != "" {
		unit.Type = input.Type
	}
	if input.Latitude != 0 {
		unit.Latitude = input.Latitude
	}
	if input.Longitude != 0 {
		unit.Longitude = input.Longitude
	}
	if input.CoverageArea != "" {
		unit.CoverageArea = input.CoverageArea
	}
	if input.ContactPerson != "" {
		unit.ContactPerson = input.ContactPerson
	}
	if input.ContactPhone != "" {
		unit.ContactPhone = input.ContactPhone
	}
	if input.ContactEmail != "" {
		unit.ContactEmail = input.ContactEmail
	}
	if input.Status != "" {
		unit.Status = input.Status
	}

	if err := config.DB.Save(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update unit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Unit updated successfully",
		"unit":    unit,
	})
}

// Helper functions
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscan(s, &f)
	return f, err
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}