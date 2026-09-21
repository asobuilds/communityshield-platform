//go:build integration

package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func buildRouterWithCORS() *gin.Engine {
	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     parseAllowedOriginsForTest(os.Getenv("ALLOWED_ORIGINS")),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"pong": true})
	})
	return router
}

func parseAllowedOriginsForTest(raw string) []string {
	if raw == "" {
		return []string{"http://configured.example"}
	}
	return []string{raw}
}

func TestCORS_AllowsConfiguredOrigin(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "https://allowed.example")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	router := buildRouterWithCORS()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "https://allowed.example")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://allowed.example" {
		t.Fatalf("expected allowed origin header, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("expected no credentials header, got %q", got)
	}
}

func TestCORS_RejectsUnconfiguredOrigin(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "https://allowed.example")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	router := buildRouterWithCORS()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no allow-origin header for evil origin, got %q", got)
	}
}