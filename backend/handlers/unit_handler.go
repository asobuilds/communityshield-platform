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

// GetNearbyUnits returns units near a location with operational radius
func GetNearbyUnits(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	radiusStr := c.Query("radius")
	stateStr := c.Query("state")
	cityStr := c.Query("city")

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

	radius := 20.0
	if radiusStr != "" {
		radius, _ = parseFloat(radiusStr)
	}

	var units []models.SecurityUnit
	query := config.DB.Where("status = ?", "active")

	// Filter by state if provided
	if stateStr != "" {
		query = query.Where("state = ? OR state ILIKE ?", stateStr, "%"+stateStr+"%")
	}

	// Filter by city if provided
	if cityStr != "" {
		query = query.Where("city = ? OR city ILIKE ?", cityStr, "%"+cityStr+"%")
	}

	if err := query.Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units"})
		return
	}

	type UnitWithDistance struct {
		models.SecurityUnit
		Distance float64 `json:"distance"`
		IsInRange bool   `json:"isInRange"`
	}

	var result []UnitWithDistance
	for _, unit := range units {
		if unit.Latitude == 0 || unit.Longitude == 0 {
			continue
		}
		distance := haversine(lat, lng, unit.Latitude, unit.Longitude)
		
		// Check if within operational radius
		isInRange := distance <= unit.OperationalRadius
		
		// Also include if within the requested search radius
		if distance <= radius || isInRange {
			result = append(result, UnitWithDistance{
				SecurityUnit: unit,
				Distance:     distance,
				IsInRange:    isInRange,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"units": result,
	})
}

// GetNearbyUnitsByLocation returns units matching state/city
func GetNearbyUnitsByLocation(c *gin.Context) {
	state := c.Query("state")
	lga := c.Query("lga")
	city := c.Query("city")
	latStr := c.Query("lat")
	lngStr := c.Query("lng")

	var units []models.SecurityUnit
	query := config.DB.Where("status = ?", "active")

	if state != "" {
		query = query.Where("state ILIKE ?", "%"+state+"%")
	}
	if lga != "" {
		query = query.Where("lga ILIKE ?", "%"+lga+"%")
	}
	if city != "" {
		query = query.Where("city ILIKE ?", "%"+city+"%")
	}

	if err := query.Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units"})
		return
	}

	// Calculate distances if lat/lng provided
	type UnitWithDistance struct {
		models.SecurityUnit
		Distance float64 `json:"distance,omitempty"`
	}

	var result []UnitWithDistance
	if latStr != "" && lngStr != "" {
		lat, _ := parseFloat(latStr)
		lng, _ := parseFloat(lngStr)
		for _, unit := range units {
			if unit.Latitude != 0 && unit.Longitude != 0 {
				distance := haversine(lat, lng, unit.Latitude, unit.Longitude)
				result = append(result, UnitWithDistance{
					SecurityUnit: unit,
					Distance:     distance,
				})
			} else {
				result = append(result, UnitWithDistance{SecurityUnit: unit})
			}
		}
	} else {
		for _, unit := range units {
			result = append(result, UnitWithDistance{SecurityUnit: unit})
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

// UpdateUnit updates a unit
func UpdateUnit(c *gin.Context) {
	id := c.Param("id")
	unitID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var input struct {
		Name              string  `json:"name"`
		Type              string  `json:"type"`
		Latitude          float64 `json:"latitude"`
		Longitude         float64 `json:"longitude"`
		OperationalRadius float64 `json:"operationalRadius"`
		State             string  `json:"state"`
		LGA               string  `json:"lga"`
		City              string  `json:"city"`
		CoverageArea      string  `json:"coverageArea"`
		ContactPerson     string  `json:"contactPerson"`
		ContactPhone      string  `json:"contactPhone"`
		ContactEmail      string  `json:"contactEmail"`
		Status            string  `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update units"})
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
	if input.OperationalRadius != 0 {
		unit.OperationalRadius = input.OperationalRadius
	}
	if input.State != "" {
		unit.State = input.State
	}
	if input.LGA != "" {
		unit.LGA = input.LGA
	}
	if input.City != "" {
		unit.City = input.City
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

// GetUnitsByLocation returns units by state/LGA/city
func GetUnitsByLocation(c *gin.Context) {
	state := c.Query("state")
	lga := c.Query("lga")
	city := c.Query("city")

	var units []models.SecurityUnit
	query := config.DB.Where("status = ?", "active")

	if state != "" {
		query = query.Where("state ILIKE ?", "%"+state+"%")
	}
	if lga != "" {
		query = query.Where("lga ILIKE ?", "%"+lga+"%")
	}
	if city != "" {
		query = query.Where("city ILIKE ?", "%"+city+"%")
	}

	if err := query.Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"units": units,
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