package handlers

import (
	"strings"
	"testing"
)

func FuzzValidateDOB(f *testing.F) {
	f.Add("2000-06-15")                // valid adult
	f.Add("")                          // empty
	f.Add(strings.Repeat("2", 10000))  // very long
	f.Add("not-a-date")                // malformed
	f.Add("2025-13-45")                // malformed date components

	f.Fuzz(func(t *testing.T, dobStr string) {
		parsed, status, err := validateDOB(dobStr)
		if err != nil {
			if !parsed.IsZero() {
				t.Errorf("validateDOB returned non-zero time with error=%v", err)
			}
			if status != "" {
				t.Errorf("validateDOB returned status %q with error=%v", status, err)
			}
			return
		}
		if parsed.IsZero() {
			t.Errorf("validateDOB returned zero time with nil error")
		}
		if status != "minor" && status != "adult" {
			t.Errorf("validateDOB returned unexpected status %q", status)
		}
	})
}
