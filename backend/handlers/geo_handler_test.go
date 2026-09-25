//go:build integration

package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"security-solution/internal/testutil"
	"security-solution/models"
	"security-solution/utils"
)

// otsuCoords is the Wikipedia canonical Otukpo, Benue State point
// (geohash precision 6 → "s1mbcm"), used as a stable cache key.
const (
	otsuLat  = 7.19306
	otsuLng  = 8.14639
	otsuHash = "s1mbcm"
)

func getGeo(t *testing.T, query string) *httptest.ResponseRecorder {
	t.Helper()
	srv := FreshServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/geo/reverse"+query, nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func decodeGeo(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v. body=%s", err, w.Body.String())
	}
	return body
}

func TestReverseGeocode_InvalidCoords(t *testing.T) {
	testutil.TruncateAll(t)

	for _, tc := range []struct {
		name  string
		query string
	}{
		{"lat_out_of_range", "?lat=999&lng=0"},
		{"lat_above_90", "?lat=90.1&lng=0"},
		{"lng_out_of_range", "?lat=0&lng=181"},
		{"lat_not_a_number", "?lat=abc&lng=0"},
		{"lng_not_a_number", "?lat=7.19&lng=xyz"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := getGeo(t, tc.query)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400. body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestReverseGeocode_MissingCoords(t *testing.T) {
	testutil.TruncateAll(t)

	for _, tc := range []struct {
		name  string
		query string
	}{
		{"no_params", ""},
		{"missing_lng", "?lat=7.19306"},
		{"missing_lat", "?lng=8.14639"},
		{"empty_values", "?lat=&lng="},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := getGeo(t, tc.query)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400. body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestReverseGeocode_CacheHit(t *testing.T) {
	testutil.TruncateAll(t)

	// Guard the fixture against a geohash encoder regression — if the
	// hash drifts the cache row would never be found.
	if got := utils.EncodeGeohash(otsuLat, otsuLng, 6); got != otsuHash {
		t.Fatalf("EncodeGeohash(%v,%v,6) = %q, want %q", otsuLat, otsuLng, got, otsuHash)
	}

	row := models.GeocodeCache{
		Geohash:    otsuHash,
		Latitude:   otsuLat,
		Longitude:  otsuLng,
		Address:    "Otukpo, Benue State, Nigeria",
		Suburb:     "Otukpo",
		LGA:        "Otukpo",
		State:      "Benue",
		Country:    "Nigeria",
		PostalCode: "970001",
		Landmark:   "Otukpo Township Hall",
		Source:     "nominatim",
		FetchedAt:  time.Now().UTC(),
	}
	if err := testutil.GetDB().Create(&row).Error; err != nil {
		t.Fatalf("seed geocode_cache: %v", err)
	}

	w := getGeo(t, "?lat=7.19306&lng=8.14639")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. body=%s", w.Code, w.Body.String())
	}

	body := decodeGeo(t, w)
	if body["state"] != "Benue" {
		t.Errorf("state = %v, want %q", body["state"], "Benue")
	}
	if body["lga"] != "Otukpo" {
		t.Errorf("lga = %v, want %q", body["lga"], "Otukpo")
	}
	if body["source"] != "nominatim" {
		t.Errorf("source = %v, want %q", body["source"], "nominatim")
	}
	if body["latitude"] != otsuLat {
		t.Errorf("latitude = %v, want %v", body["latitude"], otsuLat)
	}
	if body["longitude"] != otsuLng {
		t.Errorf("longitude = %v, want %v", body["longitude"], otsuLng)
	}
	if body["postalCode"] != "970001" {
		t.Errorf("postalCode = %v, want %q", body["postalCode"], "970001")
	}
}

func TestReverseGeocode_NominatimFailureReturns200(t *testing.T) {
	if os.Getenv("NOMINATIM_SKIP") == "1" {
		t.Skip("NOMINATIM_SKIP=1 — skipping upstream lookup test")
	}

	testutil.TruncateAll(t)

	// Cache is empty and the point is the 0/0 Atlantic sentinel, which
	// has no geohash — the service must fail soft rather than error out.
	w := getGeo(t, "?lat=0.0&lng=0.0")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (must never 500). body=%s", w.Code, w.Body.String())
	}

	body := decodeGeo(t, w)
	for _, field := range []string{"address", "suburb", "lga", "state", "country", "postalCode", "landmark"} {
		if v, ok := body[field]; ok && v != "" {
			t.Errorf("%s = %v, want empty/absent", field, v)
		}
	}
	if body["source"] != "unavailable" {
		t.Errorf("source = %v, want %q", body["source"], "unavailable")
	}
}
