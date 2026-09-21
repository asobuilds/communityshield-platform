//go:build integration

package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"security-solution/internal/testutil"
	"security-solution/middleware"
	"security-solution/models"
	"security-solution/routes"
	"security-solution/services"
)

// FreshServer returns a test router with the real route tree and the
// panic recovery middleware. Skips StructuredLogger and AuditMiddleware
// to keep test output clean. Does NOT start the scheduler.
func FreshServer(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.PanicRecovery())
	routes.SetupRoutes(r)
	return r
}

// signedTokenFor builds a JWT signed with JWT_SECRET for the given user
// with custom iat/exp offsets. Useful for expired-token tests.
func signedTokenFor(t *testing.T, user *models.User, iatOffset, expOffset time.Duration) string {
	t.Helper()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		t.Fatalf("JWT_SECRET is not set")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"jti":     "test-jti-" + user.ID.String()[:8],
		"iat":     now.Add(iatOffset).Unix(),
		"exp":     now.Add(expOffset).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

// TestHTTPSecurity_ForgedTokenRejected — garbage Authorization header.
func TestHTTPSecurity_ForgedTokenRejected(t *testing.T) {
	testutil.TruncateAll(t)
	r := FreshServer(t)

	req := httptest.NewRequest("GET", "/api/v1/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer this.is.garbage")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for forged token, got %d", w.Code)
	}
}

// TestHTTPSecurity_NoTokenRejected — missing Authorization header.
func TestHTTPSecurity_NoTokenRejected(t *testing.T) {
	testutil.TruncateAll(t)
	r := FreshServer(t)

	req := httptest.NewRequest("GET", "/api/v1/auth/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", w.Code)
	}
}

// TestHTTPSecurity_ExpiredTokenRejected — a well-formed token whose exp
// has already passed must be rejected.
func TestHTTPSecurity_ExpiredTokenRejected(t *testing.T) {
	testutil.TruncateAll(t)
	r := FreshServer(t)

	authSvc := services.NewAuthService()
	u := &models.User{
		Email:     "expired@test.local",
		Phone:     "08011110001",
		FirstName: "Expired",
		LastName:  "User",
		Password:  "password-1234",
		Role:      "citizen",
		Status:    "active",
	}
	created, err := authSvc.Register(u)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Sign a token that expired an hour ago.
	expired := signedTokenFor(t, created, -2*time.Hour, -1*time.Hour)

	req := httptest.NewRequest("GET", "/api/v1/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+expired)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", w.Code)
	}
}

// TestHTTPSecurity_RevokedJTIRejected — a revoked token is refused even
// though it is otherwise valid.
func TestHTTPSecurity_RevokedJTIRejected(t *testing.T) {
	testutil.TruncateAll(t)
	r := FreshServer(t)

	authSvc := services.NewAuthService()
	u := &models.User{
		Email:     "revoked@test.local",
		Phone:     "08011110002",
		FirstName: "Revoked",
		LastName:  "User",
		Password:  "password-1234",
		Role:      "citizen",
		Status:    "active",
	}
	if _, err := authSvc.Register(u); err != nil {
		t.Fatalf("register: %v", err)
	}

	token, jti, _, err := authSvc.LoginWithJTI("revoked@test.local", "password-1234")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	// Revoke the JTI before the request
	tokenSvc := services.NewTokenService()
	if err := tokenSvc.Revoke(jti, u.ID, time.Now().UTC().Add(24*time.Hour), "test"); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for revoked token, got %d", w.Code)
	}
}

// TestHTTPSecurity_CitizenCannotReachAdminRoutes — a citizen token must
// not pass the SuperAdminMiddleware.
func TestHTTPSecurity_CitizenCannotReachAdminRoutes(t *testing.T) {
	testutil.TruncateAll(t)
	r := FreshServer(t)

	citizen := testutil.MakeUser(t, "citizen")
	token := testutil.AuthHeader(t, citizen)

	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for citizen accessing admin route, got %d", w.Code)
	}
}

// TestHTTPSecurity_CrossUnitCaseReadReturns403 — an officer in Unit A
// must not be able to GET a case belonging to Unit B.
func TestHTTPSecurity_CrossUnitCaseReadReturns403(t *testing.T) {
	testutil.TruncateAll(t)
	r := FreshServer(t)

	// Unit A with an officer
	headA := testutil.MakeUser(t, "unit_admin")
	unitA := testutil.MakeUnit(t, headA)
	officerA := testutil.MakeUser(t, "officer")
	testutil.MakeMembership(t, officerA, unitA, "officer", false)

	// Unit B with a case
	headB := testutil.MakeUser(t, "unit_admin")
	unitB := testutil.MakeUnit(t, headB)
	reporterB := testutil.MakeUser(t, "citizen")
	caseB := testutil.MakeCase(t, reporterB, unitB)

	// officerA tries to read caseB via HTTP
	token := testutil.AuthHeader(t, officerA)
	req := httptest.NewRequest("GET", "/api/v1/cases/"+caseB.ID.String(), nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-unit case read, got %d", w.Code)
	}
}