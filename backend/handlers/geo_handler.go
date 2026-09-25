package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"security-solution/services"
)

// geoResponse is the public reverse-geocode payload. Field names match
// the frontend contract; empty strings are omitted so the UI can render
// "Location unavailable" rather than a blank field.
type geoResponse struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Address    string  `json:"address,omitempty"`
	Suburb     string  `json:"suburb,omitempty"`
	LGA        string  `json:"lga,omitempty"`
	State      string  `json:"state,omitempty"`
	Country    string  `json:"country,omitempty"`
	PostalCode string  `json:"postalCode,omitempty"`
	Landmark   string  `json:"landmark,omitempty"`
	Source     string  `json:"source"`
	FetchedAt  string  `json:"fetchedAt,omitempty"`
}

// ReverseGeocode returns the address for a coordinate.
// GET /api/v1/geo/reverse?lat=&lng=
// Public endpoint (no auth) but rate-limited.
//
// Only malformed or out-of-range coordinates produce a 4xx. A lookup
// that cannot be resolved (Nominatim unreachable, timeout, or no data
// for the point) still returns 200 with empty address fields so the
// frontend can degrade gracefully instead of seeing a server error.
func ReverseGeocode(c *gin.Context) {
	lat, lng, err := parseCoordinate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := geoResponse{
		Latitude:  lat,
		Longitude: lng,
		Source:    services.GeocodeSourceUnavailable,
	}

	rec, err := services.NewGeocodingService().ReverseGeocode(lat, lng)
	if err == nil && rec != nil {
		resp.Address = rec.Address
		resp.Suburb = rec.Suburb
		resp.LGA = rec.LGA
		resp.State = rec.State
		resp.Country = rec.Country
		resp.PostalCode = rec.PostalCode
		resp.Landmark = rec.Landmark
		if rec.Source != "" {
			resp.Source = rec.Source
		}
		if !rec.FetchedAt.IsZero() {
			resp.FetchedAt = rec.FetchedAt.UTC().Format(time.RFC3339)
		}
	} else if err != nil && !errors.Is(err, services.ErrInvalidCoordinates) {
		// Swallow upstream/DB errors — this endpoint must not 500.
		c.Error(err) //nolint:errcheck // recorded for the request log
	}

	c.JSON(http.StatusOK, resp)
}

// parseCoordinate reads and range-checks the lat/lng query parameters.
func parseCoordinate(c *gin.Context) (float64, float64, error) {
	latRaw := c.Query("lat")
	lngRaw := c.Query("lng")
	if latRaw == "" || lngRaw == "" {
		return 0, 0, errors.New("lat and lng query parameters are required")
	}

	lat, err := strconv.ParseFloat(latRaw, 64)
	if err != nil {
		return 0, 0, errors.New("lat must be a number")
	}
	lng, err := strconv.ParseFloat(lngRaw, 64)
	if err != nil {
		return 0, 0, errors.New("lng must be a number")
	}

	if lat < -90 || lat > 90 {
		return 0, 0, errors.New("lat must be between -90 and 90")
	}
	if lng < -180 || lng > 180 {
		return 0, 0, errors.New("lng must be between -180 and 180")
	}

	return lat, lng, nil
}
