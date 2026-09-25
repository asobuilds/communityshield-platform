package utils

import "errors"

var errInvalidGeohash = errors.New("invalid geohash")

const geohashBase32 = "0123456789bcdefghjkmnpqrstuvwxyz"

// EncodeGeohash returns the geohash of (lat, lng) at the given precision.
// Precision 4 ≈ 39km, 5 ≈ 4.9km, 6 ≈ 1.2km.
// If lat == 0 && lng == 0 returns "" (unset-coords sentinel).
// Precision is clamped to [1,12].
//
// Standard base-32 geohash algorithm (Gustavo Niemeyer). Longitude is the
// first bit of every pair — matches PostGIS decode_geohash_bbox and the
// geohash.org reference. Verified against:
//   - Wikipedia:  (57.64911, 10.40744, 4) → "u4pr"
//   - Golden Gate: (37.819900, -122.478300, 9) → "9q8zhuyhb"
//   - Lagos:       (6.454070, 3.394670, 9) → "s14ktnzvt"
func EncodeGeohash(lat, lng float64, precision int) string {
	if lat == 0 && lng == 0 {
		return ""
	}
	if precision < 1 {
		precision = 1
	}
	if precision > 12 {
		precision = 12
	}

	latMin, latMax := -90.0, 90.0
	lngMin, lngMax := -180.0, 180.0

	var bits [120]byte // 12 * 5
	odd := true // true → longitude bit first
	for i := 0; i < precision*5; i++ {
		if odd {
			mid := (lngMin + lngMax) / 2
			if lng >= mid {
				bits[i] = 1
				lngMin = mid
			} else {
				bits[i] = 0
				lngMax = mid
			}
		} else {
			mid := (latMin + latMax) / 2
			if lat >= mid {
				bits[i] = 1
				latMin = mid
			} else {
				bits[i] = 0
				latMax = mid
			}
		}
		odd = !odd
	}

	out := make([]byte, precision)
	for i := 0; i < precision; i++ {
		val := 0
		for j := 0; j < 5; j++ {
			val = val << 1
			if bits[i*5+j] == 1 {
				val++
			}
		}
		out[i] = geohashBase32[val]
	}
	return string(out)
}

// DecodeGeohash returns the (lat, lng) centre of the given geohash cell.
func DecodeGeohash(hash string) (float64, float64, error) {
	if hash == "" {
		return 0, 0, nil
	}
	if len(hash) < 1 || len(hash) > 12 {
		return 0, 0, errInvalidGeohash
	}

	latMin, latMax := -90.0, 90.0
	lngMin, lngMax := -180.0, 180.0
	odd := true // true → longitude bit first

	for i := 0; i < len(hash); i++ {
		idx := -1
		for j, c := range geohashBase32 {
			if byte(c) == hash[i] {
				idx = j
				break
			}
		}
		if idx < 0 {
			return 0, 0, errInvalidGeohash
		}
		for j := 4; j >= 0; j-- {
			bit := (idx >> uint(j)) & 1
			if odd {
				mid := (lngMin + lngMax) / 2
				if bit == 1 {
					lngMin = mid
				} else {
					lngMax = mid
				}
			} else {
				mid := (latMin + latMax) / 2
				if bit == 1 {
					latMin = mid
				} else {
					latMax = mid
				}
			}
			odd = !odd
		}
	}

	return (latMin + latMax) / 2, (lngMin + lngMax) / 2, nil
}