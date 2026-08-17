package handlers

import (
	"math"
	"net/http"
	"os"
	"path/filepath"
	"security-solution/models"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Haversine distance in km
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func CreateUnit(c *gin.Context) {
	var input struct {
		Name               string  `json:"name" binding:"required"`
		Type               string  `json:"type" binding:"required"`
		Latitude           float64 `json:"latitude"`
		Longitude          float64 `json:"longitude"`
		CoverageArea       string  `json:"coverageArea"`
		ContactPerson      string  `json:"contactPerson"`
		ContactPhone       string  `json:"contactPhone"`
		ContactEmail       string  `json:"contactEmail"`
		RegistrationNumber string  `json:"registrationNumber"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unit := models.Unit{
		Name:               input.Name,
		Type:               input.Type,
		Latitude:           input.Latitude,
		Longitude:          input.Longitude,
		CoverageArea:       input.CoverageArea,
		ContactPerson:      input.ContactPerson,
		ContactPhone:       input.ContactPhone,
		ContactEmail:       input.ContactEmail,
		RegistrationNumber: input.RegistrationNumber,
		Status:             "active",
	}

	if err := DB.Create(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create unit"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Unit created successfully",
		"unit":    unit,
	})
}

func GetUnits(c *gin.Context) {
	var units []models.Unit
	if err := DB.Order("created_at desc").Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"units": units})
}

func GetUnit(c *gin.Context) {
	unitID := c.Param("unitId")
	parsedID, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var unit models.Unit
	if err := DB.First(&unit, "id = ?", parsedID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unit": unit})
}

func UpdateUnit(c *gin.Context) {
	unitID := c.Param("unitId")
	parsedID, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var input struct {
		Name               string  `json:"name"`
		Type               string  `json:"type"`
		Latitude           float64 `json:"latitude"`
		Longitude          float64 `json:"longitude"`
		CoverageArea       string  `json:"coverageArea"`
		ContactPerson      string  `json:"contactPerson"`
		ContactPhone       string  `json:"contactPhone"`
		ContactEmail       string  `json:"contactEmail"`
		RegistrationNumber string  `json:"registrationNumber"`
		Status             string  `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var unit models.Unit
	if err := DB.First(&unit, "id = ?", parsedID).Error; err != nil {
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
	if input.RegistrationNumber != "" {
		unit.RegistrationNumber = input.RegistrationNumber
	}
	if input.Status != "" {
		unit.Status = input.Status
	}

	if err := DB.Save(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update unit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Unit updated successfully",
		"unit":    unit,
	})
}

func DeleteUnit(c *gin.Context) {
	unitID := c.Param("unitId")
	parsedID, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	if err := DB.Delete(&models.Unit{}, "id = ?", parsedID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete unit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unit deleted successfully"})
}

func AssignUnitAdmin(c *gin.Context) {
	unitID := c.Param("unitId")
	parsedUnitID, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var input struct {
		UserID string `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var unit models.Unit
	if err := DB.First(&unit, "id = ?", parsedUnitID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
		return
	}

	var user models.User
	if err := DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Role = "unit_admin"
	user.UnitID = &parsedUnitID
	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign admin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin assigned successfully",
		"user": gin.H{
			"id":     user.ID,
			"email":  user.Email,
			"role":   user.Role,
			"unitId": user.UnitID,
		},
	})
}

func GetAdminUnits(c *gin.Context) {
	adminID := c.Param("adminId")
	parsedAdminID, err := uuid.Parse(adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin ID"})
		return
	}

	var user models.User
	if err := DB.First(&user, "id = ?", parsedAdminID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var units []models.Unit
	if user.UnitID != nil {
		if err := DB.First(&units, "id = ?", *user.UnitID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
			return
		}
	} else {
		units = []models.Unit{}
	}

	c.JSON(http.StatusOK, gin.H{"units": units})
}

// GetNearbyUnits returns nearby units based on latitude/longitude and radius
func GetNearbyUnits(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lng are required"})
		return
	}
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lat"})
		return
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lng"})
		return
	}

	radius := 10.0 // default 10 km
	if r := c.Query("radius"); r != "" {
		if parsed, err := strconv.ParseFloat(r, 64); err == nil && parsed > 0 {
			radius = parsed
		}
	}

	var units []models.Unit
	if err := DB.Where("status = ?", "active").Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units"})
		return
	}

	type NearbyUnit struct {
		models.Unit
		Distance float64 `json:"distance"`
	}
	var nearby []NearbyUnit

	for _, u := range units {
		dist := haversine(lat, lng, u.Latitude, u.Longitude)
		if dist <= radius {
			nearby = append(nearby, NearbyUnit{Unit: u, Distance: dist})
		}
	}

	for i := 0; i < len(nearby); i++ {
		for j := i + 1; j < len(nearby); j++ {
			if nearby[i].Distance > nearby[j].Distance {
				nearby[i], nearby[j] = nearby[j], nearby[i]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"units":    nearby,
		"count":    len(nearby),
		"radius":   radius,
		"location": gin.H{"lat": lat, "lng": lng},
	})
}

// 🔥 NEW: UploadUnitProfilePicture uploads a profile picture for a unit
func UploadUnitProfilePicture(c *gin.Context) {
	unitIDStr := c.Param("unitId")
	if unitIDStr == "" {
		unitIDStr = c.PostForm("unitId")
	}
	if unitIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unitId is required"})
		return
	}

	unitID, err := uuid.Parse(unitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unitId"})
		return
	}

	var unit models.Unit
	if err := DB.First(&unit, "id = ?", unitID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
		return
	}

	file, err := c.FormFile("profilePicture")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "profilePicture file is required"})
		return
	}

	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Max size is 2MB"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: jpg, jpeg, png, gif, webp"})
		return
	}

	uploadDir := "./uploads/unit-logos"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	filename := unitID.String() + ext
	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	profileImageURL := "/uploads/unit-logos/" + filename
	unit.ProfileImage = profileImageURL
	if err := DB.Save(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update unit profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Unit profile picture uploaded successfully",
		"profileImage": profileImageURL,
	})
}