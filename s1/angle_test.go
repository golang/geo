package s1

import (
	"math"
	"testing"
)

func TestEmptyValue(t *testing.T) {
	var a Angle
	if rad := a.Radians(); rad != 0 {
		t.Errorf("Empty value of Angle was %v, want 0", rad)
	}
}

func TestPiRadiansExactly180Degrees(t *testing.T) {
	if rad := (math.Pi * Radian).Radians(); rad != math.Pi {
		t.Errorf("(π * Radian).Radians() was %v, want π", rad)
	}
	if deg := (math.Pi * Radian).Degrees(); deg != 180 {
		t.Errorf("(π * Radian).Degrees() was %v, want 180", deg)
	}
	if rad := (180 * Degree).Radians(); rad != math.Pi {
		t.Errorf("(180 * Degree).Radians() was %v, want π", rad)
	}
	if deg := (180 * Degree).Degrees(); deg != 180 {
		t.Errorf("(180 * Degree).Degrees() was %v, want 180", deg)
	}

	if deg := (math.Pi / 2 * Radian).Degrees(); deg != 90 {
		t.Errorf("(π/2 * Radian).Degrees() was %v, want 90", deg)
	}

	// Check negative angles.
	if deg := (-math.Pi / 2 * Radian).Degrees(); deg != -90 {
		t.Errorf("(-π/2 * Radian).Degrees() was %v, want -90", deg)
	}
	if rad := (-45 * Degree).Radians(); rad != -math.Pi/4 {
		t.Errorf("(-45 * Degree).Radians() was %v, want -π/4", rad)
	}
}

func TestE5E6E7Representation(t *testing.T) {
	// NOTE(dsymonds): This first test gives a variance in the 16th decimal place. I should track that down.
	exp, act := (-45 * Degree).Radians(), (-4500000 * E5).Radians()
	if math.Abs(exp-act) > 1e-15 {
		t.Errorf("(-4500000 * E5).Radians() was %v, want %v", act, exp)
	}
	if exp, act := (-60 * Degree).Radians(), (-60000000 * E6).Radians(); exp != act {
		t.Errorf("(-60000000 * E6).Radians() was %v, want %v", act, exp)
	}
	if exp, act := (75 * Degree).Radians(), (750000000 * E7).Radians(); exp != act {
		t.Errorf("(-750000000 * E7).Radians() was %v, want %v", act, exp)
	}

	if exp, act := int32(-17256123), (-172.56123 * Degree).E5(); exp != act {
		t.Errorf("(-172.56123°).E5() was %v, want %v", act, exp)
	}
	if exp, act := int32(12345678), (12.345678 * Degree).E6(); exp != act {
		t.Errorf("(12.345678°).E6() was %v, want %v", act, exp)
	}
	if exp, act := int32(-123456789), (-12.3456789 * Degree).E7(); exp != act {
		t.Errorf("(-12.3456789°).E7() was %v, want %v", act, exp)
	}
}

func TestNormalizeCorrectlyCanonicalizesAngles(t *testing.T) {
	tests := []struct {
		in, want float64 // both in degrees
	}{
		{360, 0},
		{-180, 180},
		{180, 180},
		{540, 180},
		{-270, 90},
	}
	for _, test := range tests {
		deg := (Angle(test.in) * Degree).Normalized().Degrees()
		if deg != test.want {
			t.Errorf("Normalized %.0f° = %v, want %v", test.in, deg, test.want)
		}
	}
}

func TestAngleString(t *testing.T) {
	if s, exp := (180 * Degree).String(), "180.0000000"; s != exp {
		t.Errorf("(180°).String() = %q, want %q", s, exp)
	}
}
