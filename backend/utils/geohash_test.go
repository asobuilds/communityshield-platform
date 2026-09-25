package utils

import "testing"

func TestEncodeGeohash_LagosPrecision6(t *testing.T) {
	got := EncodeGeohash(6.5244, 3.3792, 6)
	if got != "s14mhg" {
		t.Fatalf("EncodeGeohash(6.5244,3.3792,6) = %q, want s14mhg", got)
	}
}

func TestEncodeGeohash_LagosPrecision5(t *testing.T) {
	got := EncodeGeohash(6.5244, 3.3792, 5)
	if got != "s14mh" {
		t.Fatalf("EncodeGeohash(6.5244,3.3792,5) = %q, want s14mh", got)
	}
}

func TestEncodeGeohash_LagosPrecision4(t *testing.T) {
	got := EncodeGeohash(6.5244, 3.3792, 4)
	if got != "s14m" {
		t.Fatalf("EncodeGeohash(6.5244,3.3792,4) = %q, want s14m", got)
	}
}

func TestEncodeGeohash_ZeroCoordsSentinel(t *testing.T) {
	got := EncodeGeohash(0, 0, 5)
	if got != "" {
		t.Fatalf("EncodeGeohash(0,0,5) = %q, want empty sentinel", got)
	}
}

func TestEncodeGeohash_WikipediaReference(t *testing.T) {
	got := EncodeGeohash(57.64911, 10.40744, 4)
	if got != "u4pr" {
		t.Fatalf("EncodeGeohash(57.64911,10.40744,4) = %q, want u4pr", got)
	}
}

func TestEncodeGeohash_GoldenGateReference(t *testing.T) {
	got := EncodeGeohash(37.819900, -122.478300, 9)
	if got != "9q8zhuyhb" {
		t.Fatalf("EncodeGeohash(37.819900,-122.478300,9) = %q, want 9q8zhuyhb", got)
	}
}

func TestEncodeGeohash_LagosReference(t *testing.T) {
	got := EncodeGeohash(6.454070, 3.394670, 9)
	if got != "s14ktnzvt" {
		t.Fatalf("EncodeGeohash(6.454070,3.394670,9) = %q, want s14ktnzvt", got)
	}
}

func TestEncodeGeohash_PrecisionClampLow(t *testing.T) {
	got := EncodeGeohash(6.5244, 3.3792, 0)
	if len(got) != 1 {
		t.Fatalf("precision 0 should clamp to 1 char, got %q (len %d)", got, len(got))
	}
}

func TestEncodeGeohash_PrecisionClampHigh(t *testing.T) {
	got := EncodeGeohash(6.5244, 3.3792, 99)
	if len(got) != 12 {
		t.Fatalf("precision 99 should clamp to 12 chars, got %q (len %d)", got, len(got))
	}
}

func TestDecodeGeohash_RoundTrip(t *testing.T) {
	for _, tc := range []struct {
		lat, lng float64
		prec      int
	}{
		{6.5244, 3.3792, 6},
		{37.819900, -122.478300, 9},
		{57.64911, 10.40744, 4},
	} {
		hash := EncodeGeohash(tc.lat, tc.lng, tc.prec)
		rlat, rlng, err := DecodeGeohash(hash)
		if err != nil {
			t.Fatalf("DecodeGeohash(%q) error: %v", hash, err)
		}
		// Decode returns the cell centre. The input point lies somewhere
		// inside the cell, so the centre can be up to one cell half-width
		// away. Precision 4 ≈ 20km → ~0.18°; precision 9 ≈ ~2m.
		// The true round-trip property is that re-encoding the centre
		// yields the same hash.
		if rehash := EncodeGeohash(rlat, rlng, tc.prec); rehash != hash {
			t.Fatalf("round-trip mismatch: in=(%v,%v)→%q, centre=(%v,%v)→%q",
				tc.lat, tc.lng, hash, rlat, rlng, rehash)
		}
	}
}

func TestDecodeGeohash_Empty(t *testing.T) {
	lat, lng, err := DecodeGeohash("")
	if err != nil {
		t.Fatalf("DecodeGeohash(\"\") error: %v", err)
	}
	if lat != 0 || lng != 0 {
		t.Fatalf("DecodeGeohash(\"\") = (%v,%v), want (0,0)", lat, lng)
	}
}

func TestDecodeGeohash_Invalid(t *testing.T) {
	if _, _, err := DecodeGeohash("xyz!!"); err == nil {
		t.Fatal("expected error for invalid geohash")
	}
}