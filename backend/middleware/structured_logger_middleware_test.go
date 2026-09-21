package middleware

import (
	"strings"
	"testing"
)

func FuzzSanitizeRequestID(f *testing.F) {
	f.Add("req-123_abc")                 // valid
	f.Add("")                            // empty
	f.Add(strings.Repeat("a", 10000))    // very long
	f.Add("bad!@#chars$%")               // malformed
	f.Add("😀combining\u0301rtl\u202E_-ab") // unicode edge: emoji + combining + RTL

	f.Fuzz(func(t *testing.T, id string) {
		out := sanitizeRequestID(id)
		if len(out) > 64 {
			t.Errorf("sanitizeRequestID output length %d exceeds 64", len(out))
		}
		for _, r := range out {
			if !((r >= 'A' && r <= 'Z') ||
				(r >= 'a' && r <= 'z') ||
				(r >= '0' && r <= '9') ||
				r == '-' || r == '_') {
				t.Errorf("sanitizeRequestID output contains invalid rune %q", r)
			}
		}
	})
}
