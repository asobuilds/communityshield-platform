//go:build integration

package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"security-solution/internal/testutil"
	"security-solution/models"
	"security-solution/services"
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

// ---------------------------------------------------------------------------
// Wave 10.2a — Nigeria geography dataset
// ---------------------------------------------------------------------------

// superAdminToken creates a super_admin user and returns a bearer token
// for it. AuthMiddleware reloads the user from the DB, so the role is
// taken from the persisted row.
func superAdminToken(t *testing.T, email, phone string) string {
	t.Helper()
	created, err := services.NewAuthService().Register(&models.User{
		Email:     email,
		Phone:     phone,
		FirstName: "Geo",
		LastName:  "Admin",
		Password:  "password-1234",
		Role:      "super_admin",
		Status:    "active",
	})
	if err != nil {
		t.Fatalf("register super_admin: %v", err)
	}
	return signedTokenFor(t, created, 0, time.Hour)
}

// postUnit creates a unit as super_admin with the given geography.
func postUnit(t *testing.T, token, name, regNumber, state, lga string) *httptest.ResponseRecorder {
	t.Helper()
	srv := FreshServer(t)

	payload := map[string]any{
		"name":               name,
		"type":               "community",
		"latitude":           7.19,
		"longitude":          8.14,
		"registrationNumber": regNumber,
	}
	if state != "" {
		payload["state"] = state
	}
	if lga != "" {
		payload["lga"] = lga
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal unit payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/units", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

// TestGeoStates_ReturnsAllStates proves the endpoint exposes 36 states
// plus the Federal Capital Territory, with the LGA array omitted.
func TestGeoStates_ReturnsAllStates(t *testing.T) {
	testutil.TruncateAll(t)

	srv := FreshServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/geo/states", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. body=%s", w.Code, w.Body.String())
	}

	body := decodeGeo(t, w)
	raw, ok := body["states"].([]any)
	if !ok {
		t.Fatalf("states is %T, want an array. body=%s", body["states"], w.Body.String())
	}
	if len(raw) != 37 {
		t.Fatalf("len(states) = %d, want 37 (36 states + FCT)", len(raw))
	}

	seen := make(map[string]bool, len(raw))
	for _, entry := range raw {
		obj, ok := entry.(map[string]any)
		if !ok {
			t.Fatalf("state entry is %T, want an object", entry)
		}
		name, _ := obj["name"].(string)
		seen[name] = true

		if _, hasLGA := obj["lgas"]; hasLGA {
			t.Errorf("state %q must not include the lgas array (payload size)", name)
		}
		for _, field := range []string{"capital", "zone", "lgaCount"} {
			if _, ok := obj[field]; !ok {
				t.Errorf("state %q missing field %q. entry=%v", name, field, obj)
			}
		}
	}

	for _, want := range []string{"Benue", "Federal Capital Territory"} {
		if !seen[want] {
			t.Errorf("states response missing %q", want)
		}
	}
}

// TestGeoLGAs_ValidState proves the LGA list for a known state.
func TestGeoLGAs_ValidState(t *testing.T) {
	testutil.TruncateAll(t)

	srv := FreshServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/geo/lgas?state=Benue", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200. body=%s", w.Code, w.Body.String())
	}

	body := decodeGeo(t, w)
	if body["state"] != "Benue" {
		t.Errorf("state = %v, want %q", body["state"], "Benue")
	}

	raw, ok := body["lgas"].([]any)
	if !ok {
		t.Fatalf("lgas is %T, want an array. body=%s", body["lgas"], w.Body.String())
	}
	got := make(map[string]bool, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			got[s] = true
		}
	}
	for _, want := range []string{"Otukpo", "Makurdi"} {
		if !got[want] {
			t.Errorf("Benue lgas missing %q", want)
		}
	}
}

// TestGeoLGAs_MissingStateParam — the state query param is required.
func TestGeoLGAs_MissingStateParam(t *testing.T) {
	testutil.TruncateAll(t)

	srv := FreshServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/geo/lgas", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body=%s", w.Code, w.Body.String())
	}
}

// TestGeoLGAs_UnknownState — a state that is not in Nigeria is a 404.
func TestGeoLGAs_UnknownState(t *testing.T) {
	testutil.TruncateAll(t)

	srv := FreshServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/geo/lgas?state=Atlantis", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404. body=%s", w.Code, w.Body.String())
	}
}

// TestCreateUnit_RejectsUnknownState — geography is validated before
// the row is written.
func TestCreateUnit_RejectsUnknownState(t *testing.T) {
	testutil.TruncateAll(t)
	token := superAdminToken(t, "geo-unknown-state@test.local", "08055550001")

	w := postUnit(t, token, "Atlantis Unit", "REG-GEO-001", "Atlantis", "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body=%s", w.Code, w.Body.String())
	}
	if body := decodeGeo(t, w); !strings.Contains(body["error"].(string), "unknown state") {
		t.Errorf("error = %v, want it to mention unknown state", body["error"])
	}

	var count int64
	testutil.GetDB().Model(&models.SecurityUnit{}).Where("name = ?", "Atlantis Unit").Count(&count)
	if count != 0 {
		t.Errorf("unit was persisted despite invalid state (rows=%d)", count)
	}
}

// TestCreateUnit_RejectsUnknownLGA — a real state with a bogus LGA.
func TestCreateUnit_RejectsUnknownLGA(t *testing.T) {
	testutil.TruncateAll(t)
	token := superAdminToken(t, "geo-unknown-lga@test.local", "08055550002")

	w := postUnit(t, token, "Benue Bad LGA Unit", "REG-GEO-002", "Benue", "NotAPlace")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body=%s", w.Code, w.Body.String())
	}
	if body := decodeGeo(t, w); !strings.Contains(body["error"].(string), "unknown LGA") {
		t.Errorf("error = %v, want it to mention unknown LGA", body["error"])
	}

	var count int64
	testutil.GetDB().Model(&models.SecurityUnit{}).Where("name = ?", "Benue Bad LGA Unit").Count(&count)
	if count != 0 {
		t.Errorf("unit was persisted despite invalid LGA (rows=%d)", count)
	}
}

// TestCreateUnit_AcceptsValidStateAndLGA — the happy path still works.
func TestCreateUnit_AcceptsValidStateAndLGA(t *testing.T) {
	testutil.TruncateAll(t)
	token := superAdminToken(t, "geo-valid@test.local", "08055550003")

	w := postUnit(t, token, "Otukpo Community Unit", "REG-GEO-003", "Benue", "Otukpo")
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201. body=%s", w.Code, w.Body.String())
	}

	var saved models.SecurityUnit
	if err := testutil.GetDB().Where("name = ?", "Otukpo Community Unit").First(&saved).Error; err != nil {
		t.Fatalf("load created unit: %v", err)
	}
	if saved.State != "Benue" {
		t.Errorf("state = %q, want %q", saved.State, "Benue")
	}
	if saved.LGA != "Otukpo" {
		t.Errorf("lga = %q, want %q", saved.LGA, "Otukpo")
	}
}
