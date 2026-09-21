package services

import (
	"strings"
	"testing"
)

func FuzzRemoveNonDigits(f *testing.F) {
	f.Add("abc123def")                 // valid mixed
	f.Add("")                          // empty
	f.Add(strings.Repeat("!", 10000))  // very long, all non-digits
	f.Add("1a2b3c4")                   // interleaved
	f.Add("😀combining\u0301\u03A9")   // unicode edge: emoji + combining + Greek

	f.Fuzz(func(t *testing.T, s string) {
		out := removeNonDigits(s)
		if len(out) > len(s) {
			t.Errorf("removeNonDigits output length %d exceeds input length %d", len(out), len(s))
		}
		for _, r := range out {
			if r < '0' || r > '9' {
				t.Errorf("removeNonDigits output contains non-digit rune %q", r)
			}
		}
	})
}
