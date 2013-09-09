package s1

import (
	"math"
	"strconv"
)

// Angle represents a 1D angle.
type Angle float64

// Angle units.
const (
	Radian Angle = 1
	Degree       = (math.Pi / 180) * Radian

	E5 = 1e-5 * Degree
	E6 = 1e-6 * Degree
	E7 = 1e-7 * Degree
)

// Radians returns the angle in radians.
func (a Angle) Radians() float64 { return float64(a) }

// Degrees returns the angle in degrees.
func (a Angle) Degrees() float64 { return float64(a / Degree) }

// E5 returns the angle in hundred thousandths of degrees.
func (a Angle) E5() int32 { return int32(a.Degrees() * 1e5) } // TODO(dsymonds): Check rounding
// E6 returns the angle in millionths of degrees.
func (a Angle) E6() int32 { return int32(a.Degrees() * 1e6) } // TODO(dsymonds): Check rounding
// E7 returns the angle in ten millionths of degrees.
func (a Angle) E7() int32 { return int32(a.Degrees() * 1e7) } // TODO(dsymonds): Check rounding

// Abs returns the absolute value of the angle.
func (a Angle) Abs() Angle { return Angle(math.Abs(float64(a))) }

// Normalized returns an equivalent angle in [0, 2π).
func (a Angle) Normalized() Angle {
	rad := math.Mod(float64(a), 2*math.Pi)
	if rad < 0 {
		rad += 2 * math.Pi
	}
	return Angle(rad)
}

func (a Angle) String() string {
	return strconv.FormatFloat(a.Degrees(), 'f', 7, 64) // like "%.7f"
}

// BUG(dsymonds): The major differences from the C++ version are:
//   - no unsigned E5/E6/E7 methods
//   - no S2Point or S2LatLng constructors
//   - no comparison or arithmetic operators
