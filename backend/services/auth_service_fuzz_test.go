package services

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func FuzzValidateToken(f *testing.F) {
	secret := "fuzz-secret-do-not-use-in-prod-min-32-chars-long"
	f.Setenv("JWT_SECRET", secret)

	valid, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		f.Fatalf("failed to mint seed token: %v", err)
	}

	f.Add(valid)                         // valid, well-formed token
	f.Add("")                            // empty string
	f.Add(strings.Repeat("a", 10000))    // very long input
	f.Add("not.a.jwt")                   // malformed: 3 segments, bad content
	f.Add("emóji😀combining\u0301rtl\u202E") // unicode edge: emoji + combining marks + RTL

	f.Fuzz(func(t *testing.T, tokenString string) {
		s := NewAuthService()
		tok, err := s.ValidateToken(tokenString)
		if err == nil && tok == nil {
			t.Errorf("ValidateToken returned nil token with nil error")
		}
	})
}
