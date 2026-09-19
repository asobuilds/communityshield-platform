package config

import (
	"log"
	"os"
	"strings"
)

// requiredEnvVars lists env vars the backend cannot safely run without.
// If any is missing, startup halts — fail-closed, never fail-open.
var requiredEnvVars = []string{
	"DATABASE_URL",
	"JWT_SECRET",
	"ENCRYPTION_KEY",
}

// forbiddenDefaults are values that must never appear in production.
var forbiddenDefaults = map[string]string{
	"JWT_SECRET": "your-secret-key-change-in-production",
}

// EnforceRequiredEnv exits the process if any required env var is missing
// or set to a known-insecure default. Call at the very start of main().
func EnforceRequiredEnv() {
	var missing []string
	for _, key := range requiredEnvVars {
		val := strings.TrimSpace(os.Getenv(key))
		if val == "" {
			missing = append(missing, key)
			continue
		}
		if bad, ok := forbiddenDefaults[key]; ok && val == bad {
			log.Fatalf("FATAL: %s is set to a known-insecure default. Generate a real value.", key)
		}
	}

	if len(missing) > 0 {
		log.Fatalf("FATAL: missing required environment variables: %s", strings.Join(missing, ", "))
	}

	// ENCRYPTION_KEY must be 64 hex chars (32 bytes).
	enc := strings.TrimSpace(os.Getenv("ENCRYPTION_KEY"))
	if len(enc) != 64 {
		log.Fatalf("FATAL: ENCRYPTION_KEY must be exactly 64 hex characters (32 bytes), got %d", len(enc))
	}

	// JWT_SECRET must be reasonably strong.
	jwt := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if len(jwt) < 32 {
		log.Fatalf("FATAL: JWT_SECRET must be at least 32 characters, got %d", len(jwt))
	}
}
