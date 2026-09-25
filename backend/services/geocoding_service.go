package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"security-solution/config"
	"security-solution/models"
	"security-solution/utils"
)

const (
	// Geohash precision 6 ≈ 1.2km cell — fine enough for a city block
	// address, coarse enough that the cache stays small.
	geocodePrecision = 6

	// A resolved address stays valid for 30 days.
	geocodeCacheTTL = 30 * 24 * time.Hour

	// A failed lookup is retried much sooner so a transient Nominatim
	// outage does not poison the cache for a month.
	geocodeEmptyTTL = 24 * time.Hour

	nominatimBaseURL   = "https://nominatim.openstreetmap.org/reverse"
	nominatimUserAgent = "NativityGuard/1.0 (+https://nativityguard.app)"
	nominatimTimeout   = 10 * time.Second

	// Nominatim's usage policy is 1 request/second. 1100ms keeps a
	// safety margin over their published limit.
	nominatimMinSpacing = 1100 * time.Millisecond

	// Never sleep a caller for longer than the outbound timeout, so a
	// request blocked behind the rate limiter fails soft instead of
	// hanging the handler.
	nominatimMaxWait = 10 * time.Second

	// display_name is unbounded; fall back to the street name when the
	// full address is too long to be useful.
	maxDisplayNameLen = 300

	// Source values stored on the record.
	GeocodeSourceNominatim   = "nominatim"
	GeocodeSourceUnavailable = "unavailable"
)

// ErrInvalidCoordinates is returned when (lat, lng) cannot produce a
// geohash — either out of range, or the 0/0 unset-coords sentinel.
var ErrInvalidCoordinates = errors.New("invalid coordinates")

// nominatimMu serialises every outbound Nominatim request. The lock is
// held across the HTTP call, so concurrent lookups queue up instead of
// racing each other past the 1 req/sec policy.
var (
	nominatimMu   sync.Mutex
	lastNominatim time.Time
)

// GeocodingService reverse-geocodes coordinates via Nominatim, backed
// by the geocode_cache table.
type GeocodingService struct {
	DB      *gorm.DB
	Client  *http.Client
	BaseURL string
}

// NewGeocodingService returns a service bound to the configured DB and
// a stdlib http.Client with the required 10s timeout.
func NewGeocodingService() *GeocodingService {
	return &GeocodingService{
		DB:      config.DB,
		Client:  &http.Client{Timeout: nominatimTimeout},
		BaseURL: nominatimBaseURL,
	}
}

func (s *GeocodingService) db() *gorm.DB {
	if s != nil && s.DB != nil {
		return s.DB
	}
	return config.DB
}

func (s *GeocodingService) client() *http.Client {
	if s != nil && s.Client != nil {
		return s.Client
	}
	return &http.Client{Timeout: nominatimTimeout}
}

func (s *GeocodingService) baseURL() string {
	if s != nil && s.BaseURL != "" {
		return s.BaseURL
	}
	return nominatimBaseURL
}

// ReverseGeocode returns the address for (lat, lng).
// It checks the cache first (keyed by geohash precision 6 ≈ 1.2km).
// On cache miss, calls Nominatim and stores the result.
// Returns a GeocodeCache record.
//
// A Nominatim failure is not an error: the geohash is returned with
// empty address fields and cached for the short TTL so we retry sooner.
func (s *GeocodingService) ReverseGeocode(lat, lng float64) (*models.GeocodeCache, error) {
	geohash := utils.EncodeGeohash(lat, lng, geocodePrecision)
	if geohash == "" {
		return nil, ErrInvalidCoordinates
	}

	var existing models.GeocodeCache
	err := s.db().Where("geohash = ?", geohash).First(&existing).Error
	switch {
	case err == nil:
		ttl := geocodeCacheTTL
		if existing.Address == "" {
			// Previously-unresolved cell: honour the short TTL.
			ttl = geocodeEmptyTTL
		}
		if time.Since(existing.FetchedAt.UTC()) < ttl {
			return &existing, nil
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Cache miss — fall through to Nominatim.
	default:
		return nil, err
	}

	rec, raw, lookupErr := s.lookupNominatim(lat, lng)

	fetchedAt := time.Now().UTC()
	record := &models.GeocodeCache{
		Geohash:   geohash,
		Latitude:  lat,
		Longitude: lng,
		Source:    GeocodeSourceNominatim,
		Raw:       raw,
		FetchedAt: fetchedAt,
	}

	if lookupErr != nil {
		// Fail soft: no usable address, short-lived negative cache.
		record.Source = GeocodeSourceUnavailable
	} else {
		record.Address = pickAddress(rec)
		record.Suburb = firstNonEmpty(rec.Address.Suburb, rec.Address.City, rec.Address.Town, rec.Address.Village)
		record.LGA = firstNonEmpty(rec.Address.County, rec.Address.StateDistrict)
		record.State = rec.Address.State
		record.Country = rec.Address.Country
		record.PostalCode = rec.Address.Postcode
		record.Landmark = firstNonEmpty(rec.Address.Amenity, rec.Address.Shop, rec.Address.Neighbourhood, rec.Address.Hamlet)
	}

	if err := s.store(record); err != nil {
		return nil, err
	}
	return record, nil
}

// store upserts the record on the geohash unique index.
func (s *GeocodingService) store(rec *models.GeocodeCache) error {
	db := s.db()
	if db == nil {
		return errors.New("database not configured")
	}
	return db.Where("geohash = ?", rec.Geohash).
		Assign(map[string]any{
			"latitude":    rec.Latitude,
			"longitude":   rec.Longitude,
			"address":     rec.Address,
			"suburb":      rec.Suburb,
			"lga":         rec.LGA,
			"state":       rec.State,
			"country":     rec.Country,
			"postal_code": rec.PostalCode,
			"landmark":    rec.Landmark,
			"source":      rec.Source,
			"raw":         rec.Raw,
			"fetched_at":  rec.FetchedAt,
		}).
		FirstOrCreate(rec).Error
}

// nominatimResponse mirrors the subset of the Nominatim reverse-geocode
// payload we consume.
type nominatimResponse struct {
	DisplayName string `json:"display_name"`
	Error       string `json:"error"`
	Address     struct {
		Road          string `json:"road"`
		Suburb        string `json:"suburb"`
		City          string `json:"city"`
		Town          string `json:"town"`
		Village       string `json:"village"`
		State         string `json:"state"`
		Country       string `json:"country"`
		Postcode      string `json:"postcode"`
		County        string `json:"county"`
		StateDistrict string `json:"state_district"`
		Neighbourhood string `json:"neighbourhood"`
		Hamlet        string `json:"hamlet"`
		Amenity       string `json:"amenity"`
		Shop          string `json:"shop"`
	} `json:"address"`
}

// lookupNominatim performs the outbound call and parses the result.
// The returned raw string is the response body, preserved on the record.
func (s *GeocodingService) lookupNominatim(lat, lng float64) (nominatimResponse, string, error) {
	var out nominatimResponse

	body, err := s.fetchNominatim(lat, lng)
	if err != nil {
		return out, "", err
	}

	if err := json.Unmarshal(body, &out); err != nil {
		return out, string(body), fmt.Errorf("decode nominatim response: %w", err)
	}
	if out.Error != "" {
		return out, string(body), fmt.Errorf("nominatim: %s", out.Error)
	}
	if out.DisplayName == "" && out.Address.Road == "" {
		return out, string(body), errors.New("nominatim: no address for coordinate")
	}

	return out, string(body), nil
}

// fetchNominatim issues the reverse-geocode request, holding
// nominatimMu for the whole call so requests can never overlap and the
// 1 req/sec policy is never violated.
func (s *GeocodingService) fetchNominatim(lat, lng float64) ([]byte, error) {
	nominatimMu.Lock()
	defer nominatimMu.Unlock()

	if !lastNominatim.IsZero() {
		if wait := nominatimMinSpacing - time.Since(lastNominatim); wait > 0 {
			if wait > nominatimMaxWait {
				wait = nominatimMaxWait
			}
			time.Sleep(wait)
		}
	}
	lastNominatim = time.Now()

	url := fmt.Sprintf(
		"%s?format=json&lat=%s&lon=%s&zoom=16&addressdetails=1",
		s.baseURL(),
		strconv.FormatFloat(lat, 'f', -1, 64),
		strconv.FormatFloat(lng, 'f', -1, 64),
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build nominatim request: %w", err)
	}
	req.Header.Set("User-Agent", nominatimUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("nominatim request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read nominatim body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nominatim status %d", resp.StatusCode)
	}

	return body, nil
}

// pickAddress prefers the full display_name, falling back to the street
// name when display_name is unwieldy.
func pickAddress(rec nominatimResponse) string {
	name := strings.TrimSpace(rec.DisplayName)
	if name == "" {
		name = strings.TrimSpace(rec.Address.Road)
	}
	if len(name) > maxDisplayNameLen {
		if road := strings.TrimSpace(rec.Address.Road); road != "" {
			name = road
		} else {
			name = name[:maxDisplayNameLen]
		}
	}
	return name
}

// firstNonEmpty returns the first value that is not blank after trim.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
