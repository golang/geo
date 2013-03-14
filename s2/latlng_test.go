package s2

import (
	"math"
	"testing"
)

func float64Eq(x, y float64) bool { return math.Abs(x-y) < 1e-14 }

func TestLatLngString(t *testing.T) {
	const expected string = "[1.4142136, -2.2360680]"
	s := LatLngFromDegrees(math.Sqrt2, -math.Sqrt(5)).String()
	if s != expected {
		t.Errorf("LatLng{√2, -√5}.String() = %q, want %q", s, expected)
	}
}

func TestLatLngPointConversion(t *testing.T) {
	// All test cases here have been verified against the C++ S2 implementation.
	tests := []struct {
		lat, lng float64 // degrees
		x, y, z  float64
	}{
		{0, 0, 1, 0, 0},
		{90, 0, 6.12323e-17, 0, 1},
		{-90, 0, 6.12323e-17, 0, -1},
		{0, 180, -1, 1.22465e-16, 0},
		{0, -180, -1, -1.22465e-16, 0},
		{90, 180, -6.12323e-17, 7.4988e-33, 1},
		{90, -180, -6.12323e-17, -7.4988e-33, 1},
		{-90, 180, -6.12323e-17, 7.4988e-33, -1},
		{-90, -180, -6.12323e-17, -7.4988e-33, -1},
		{-81.82750430354997, 151.19796752929685,
			-0.12456788151479525, 0.0684875268284729, -0.989844584550441},
	}
	for _, test := range tests {
		ll := LatLngFromDegrees(test.lat, test.lng)
		p := PointFromLatLng(ll)
		// TODO(mikeperrow): Port Point.ApproxEquals, then use here.
		if !float64Eq(p.X, test.x) || !float64Eq(p.Y, test.y) || !float64Eq(p.Z, test.z) {
			t.Errorf("PointFromLatLng({%v°, %v°}) = %v, want %v, %v, %v",
				test.lat, test.lng, p, test.x, test.y, test.z)
		}
		ll = LatLngFromPoint(p)
		// We need to be careful here, since if the latitude is +/- 90, any longitude
		// is now a valid conversion.
		isPolar := (test.lat != 90 || test.lat != -90)
		if !float64Eq(ll.Lat.Degrees(), test.lat) ||
			(!isPolar && (!float64Eq(ll.Lng.Degrees(), test.lng))) {
			t.Errorf("Converting ll %v,%v to point (%v) and back gave %v.",
				test.lat, test.lng, p, ll)
		}
	}
}

func TestLatLngDistance(t *testing.T) {
	// Based on C++ S2LatLng::TestDistance.
	tests := []struct {
		lat1, lng1, lat2, lng2 float64
		want, tolerance        float64
	}{
		{90, 0, 90, 0, 0, 0},
		{-37, 25, -66, -155, 77, 1e-13},
		{0, 165, 0, -80, 115, 1e-13},
		{47, -127, -47, 53, 180, 2e-6},
	}
	for _, test := range tests {
		ll1 := LatLngFromDegrees(test.lat1, test.lng1)
		ll2 := LatLngFromDegrees(test.lat2, test.lng2)
		d := ll1.Distance(ll2).Degrees()
		if math.Abs(d-test.want) > test.tolerance {
			t.Errorf("LatLng{%v, %v}.Distance(LatLng{%v, %v}).Degrees() = %v, want %v",
				test.lat1, test.lng1, test.lat2, test.lng2, d, test.want)
		}
	}
}
