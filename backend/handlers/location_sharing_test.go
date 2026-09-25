//go:build integration

package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"security-solution/config"
	"security-solution/internal/testutil"
	"security-solution/models"
	"security-solution/utils"
)

// TestEncodeGeohash_ReferenceValues pins the encoder against
// independently-verified reference points. If any of these fail,
// the algorithm has drifted.
func TestEncodeGeohash_ReferenceValues(t *testing.T) {
	cases := []struct {
		name      string
		lat, lng  float64
		precision int
		want      string
	}{
		// Wikipedia classic example.
		{"wiki", 57.64911, 10.40744, 4, "u4pr"},
		// Golden Gate Bridge (whatismylocation.ai reference).
		{"goldengate", 37.819900, -122.478300, 9, "9q8zhuyhb"},
		// Lagos, Nigeria (whatismylocation.ai reference).
		{"lagos", 6.454070, 3.394670, 9, "s14ktnzvt"},
		// Otukpo, Benue State, Nigeria — Wikipedia canonical.
		{"otukpo_p4", 7.19306, 8.14639, 4, "s1mb"},
		{"otukpo_p5", 7.19306, 8.14639, 5, "s1mbc"},
		{"otukpo_p6", 7.19306, 8.14639, 6, "s1mbcm"},
		{"otukpo_p9", 7.19306, 8.14639, 9, "s1mbcmkn8"},
		// Sentinel: unset coords.
		{"zero", 0, 0, 5, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.EncodeGeohash(tc.lat, tc.lng, tc.precision)
			if got != tc.want {
				t.Errorf("EncodeGeohash(%v,%v,%d) = %q, want %q",
					tc.lat, tc.lng, tc.precision, got, tc.want)
			}
		})
	}
}

func TestLocationSharing_ToggleRoundTrip(t *testing.T) {
	testutil.TruncateAll(t)

	user := testutil.MakeUser(t, "citizen")
	hdr := testutil.AuthHeader(t, user)
	srv := FreshServer(t)

	// Initial: disabled.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/location/sharing", nil)
	req.Header.Set("Authorization", hdr)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200. body=%s", w.Code, w.Body.String())
	}
	var got map[string]bool
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if got["enabled"] {
		t.Errorf("initial enabled = true, want false")
	}

	// Enable.
	body := bytes.NewBufferString(`{"enabled":true}`)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/location/sharing", body)
	req.Header.Set("Authorization", hdr)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200. body=%s", w.Code, w.Body.String())
	}

	// Reload from DB and confirm.
	var fresh models.User
	if err := config.DB.First(&fresh, "id = ?", user.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if !fresh.LocationSharingEnabled {
		t.Errorf("LocationSharingEnabled = false, want true")
	}
}

func TestCreateCase_AnonymousStoresGeohashOnly(t *testing.T) {
	testutil.TruncateAll(t)

	user := testutil.MakeUser(t, "citizen")
	unit := testutil.MakeUnit(t, user)
	hdr := testutil.AuthHeader(t, user)
	srv := FreshServer(t)

	payload := map[string]interface{}{
		"unitId":       unit.ID.String(),
		"title":        "Anonymous test",
		"description":  "Testing anonymous geohash storage",
		"latitude":     7.19306,
		"longitude":    8.14639,
		"location":     "Otukpo Market",
		"hideLocation": true,
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases", bytes.NewReader(raw))
	req.Header.Set("Authorization", hdr)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("POST status = %d. body=%s", w.Code, w.Body.String())
	}

	var got models.Case
	if err := config.DB.Order("created_at desc").First(&got).Error; err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if !got.IsAnonymous {
		t.Errorf("IsAnonymous = false, want true")
	}
	if got.Latitude != 0 || got.Longitude != 0 {
		t.Errorf("coords not stripped: lat=%v lng=%v", got.Latitude, got.Longitude)
	}
	if got.Location != "" {
		t.Errorf("Location = %q, want empty", got.Location)
	}
	wantHash := utils.EncodeGeohash(7.19306, 8.14639, 5)
	if got.LocationGeohash != wantHash {
		t.Errorf("LocationGeohash = %q, want %q", got.LocationGeohash, wantHash)
	}
}

func TestCreateCase_PreciseLocationWhenSharingEnabled(t *testing.T) {
	testutil.TruncateAll(t)

	user := testutil.MakeUser(t, "citizen")
	user.LocationSharingEnabled = true
	if err := config.DB.Save(user).Error; err != nil {
		t.Fatalf("save user: %v", err)
	}
	unit := testutil.MakeUnit(t, user)
	hdr := testutil.AuthHeader(t, user)
	srv := FreshServer(t)

	payload := map[string]interface{}{
		"unitId":      unit.ID.String(),
		"title":       "Precise test",
		"description": "Testing precise storage",
		"latitude":    7.19306,
		"longitude":   8.14639,
		"location":    "Otukpo Market",
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases", bytes.NewReader(raw))
	req.Header.Set("Authorization", hdr)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("POST status = %d. body=%s", w.Code, w.Body.String())
	}

	var got models.Case
	if err := config.DB.Order("created_at desc").First(&got).Error; err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if got.Latitude != 7.19306 {
		t.Errorf("Latitude = %v, want 7.19306", got.Latitude)
	}
	if got.IsAnonymous {
		t.Errorf("IsAnonymous = true, want false")
	}
	wantHash := utils.EncodeGeohash(7.19306, 8.14639, 5)
	if got.LocationGeohash != wantHash {
		t.Errorf("LocationGeohash = %q, want %q", got.LocationGeohash, wantHash)
	}
}

func TestSendSOS_AnonymousKeepsPreciseCoords(t *testing.T) {
	testutil.TruncateAll(t)

	user := testutil.MakeUser(t, "citizen")
	// default LocationSharingEnabled = false
	hdr := testutil.AuthHeader(t, user)
	srv := FreshServer(t)

	payload := map[string]interface{}{
		"latitude":     7.19306,
		"longitude":    8.14639,
		"description":  "Anonymous SOS test",
		"hideLocation": true,
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sos/send", bytes.NewReader(raw))
	req.Header.Set("Authorization", hdr)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("POST status = %d. body=%s", w.Code, w.Body.String())
	}

	var got models.SOSAlert
	if err := config.DB.Order("created_at desc").First(&got).Error; err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if !got.IsAnonymous {
		t.Errorf("IsAnonymous = false, want true")
	}
	if got.Latitude != 7.19306 {
		t.Errorf("SOS Latitude = %v, want 7.19306 (SOS keeps precise coords)", got.Latitude)
	}
	wantHash := utils.EncodeGeohash(7.19306, 8.14639, 5)
	if got.LocationGeohash != wantHash {
		t.Errorf("SOS LocationGeohash = %q, want %q", got.LocationGeohash, wantHash)
	}
}
