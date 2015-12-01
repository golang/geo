/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package s2

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/golang/geo/s1"
)

// float64Eq reports whether the two values are within the default epsilon.
func float64Eq(x, y float64) bool { return float64Near(x, y, 1e-14) }

// float64Near reports whether the two values are within the given epsilon.
func float64Near(x, y, ε float64) bool {
	return math.Abs(x-y) <= ε
}

// TODO(roberts): Add in flag to allow specifying the random seed for repeatable tests.

// kmToAngle converts a distance on the Earth's surface to an angle.
func kmToAngle(km float64) s1.Angle {
	// The Earth's mean radius in kilometers (according to NASA).
	const earthRadiusKm = 6371.01
	return s1.Angle(km / earthRadiusKm)
}

// randomBits returns a 64-bit random unsigned integer whose lowest "num" are random, and
// whose other bits are zero.
func randomBits(num uint32) uint64 {
	// Make sure the request is for not more than 63 bits.
	if num > 63 {
		num = 63
	}
	return uint64(rand.Int63()) & ((1 << num) - 1)
}

// Return a uniformly distributed 64-bit unsigned integer.
func randomUint64() uint64 {
	return uint64(rand.Int63() | (rand.Int63() << 63))
}

// Return a uniformly distributed 32-bit unsigned integer.
func randomUint32() uint32 {
	return uint32(randomBits(32))
}

// randomFloat64 returns a uniformly distributed value in the range [0,1).
// Note that the values returned are all multiples of 2**-53, which means that
// not all possible values in this range are returned.
func randomFloat64() float64 {
	const randomFloatBits = 53
	return math.Ldexp(float64(randomBits(randomFloatBits)), -randomFloatBits)
}

// randomUniformInt returns a uniformly distributed integer in the range [0,n).
// NOTE: This is replicated here to stay in sync with how the C++ code generates
// uniform randoms. (instead of using Go's math/rand package directly).
func randomUniformInt(n int) int {
	return int(randomFloat64() * float64(n))
}

// randomUniformFloat64 returns a uniformly distributed value in the range [min, max).
func randomUniformFloat64(min, max float64) float64 {
	return min + randomFloat64()*(max-min)
}

// randomPoint returns a random unit-length vector.
func randomPoint() Point {
	return Point{PointFromCoords(randomUniformFloat64(-1, 1),
		randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1)).Normalize()}
}

// randomCellIDForLevel returns a random CellID at the given level.
// The distribution is uniform over the space of cell ids, but only
// approximately uniform over the surface of the sphere.
func randomCellIDForLevel(level int) CellID {
	face := randomUniformInt(numFaces)
	pos := randomUint64() & uint64((1<<posBits)-1)
	return CellIDFromFacePosLevel(face, pos, level)
}

// randomCellID returns a random CellID at a randomly chosen
// level. The distribution is uniform over the space of cell ids,
// but only approximately uniform over the surface of the sphere.
func randomCellID() CellID {
	return randomCellIDForLevel(randomUniformInt(maxLevel + 1))
}

// parsePoint returns an Point from the latitude-longitude coordinate in degrees
// in the given string, or the origin if the string was invalid.
// e.g., "-20:150"
func parsePoint(s string) Point {
	p := parsePoints(s)
	if len(p) > 0 {
		return p[0]
	}

	return PointFromCoords(0, 0, 0)
}

// parseRect returns the minimal bounding Rect that contains the one or more
// latitude-longitude coordinates in degrees in the given string.
// Examples of input:
//   "-20:150"                     // one point
//   "-20:150, -20:151, -19:150"   // three points
func parseRect(s string) Rect {
	var rect Rect
	lls := parseLatLngs(s)
	if len(lls) > 0 {
		rect = RectFromLatLng(lls[0])
	}

	for _, ll := range lls[1:] {
		rect = rect.AddPoint(ll)
	}

	return rect
}

// parseLatLngs splits up a string of lat:lng points and returns the list of parsed
// entries.
func parseLatLngs(s string) []LatLng {
	pieces := strings.Split(s, ",")
	var lls []LatLng
	for _, piece := range pieces {
		piece = strings.TrimSpace(piece)

		// Skip empty strings.
		if piece == "" {
			continue
		}

		p := strings.Split(piece, ":")
		if len(p) != 2 {
			panic(fmt.Sprintf("invalid input string for parseLatLngs: %q", piece))
		}

		lat, err := strconv.ParseFloat(p[0], 64)
		if err != nil {
			panic(fmt.Sprintf("invalid float in parseLatLngs: %q, err: %v", p[0], err))
		}

		lng, err := strconv.ParseFloat(p[1], 64)
		if err != nil {
			panic(fmt.Sprintf("invalid float in parseLatLngs: %q, err: %v", p[1], err))
		}

		lls = append(lls, LatLngFromDegrees(lat, lng))
	}
	return lls
}

// parsePoints takes a string of lat:lng points and returns the set of Points it defines.
func parsePoints(s string) []Point {
	lls := parseLatLngs(s)
	points := make([]Point, len(lls))
	for i, ll := range lls {
		points[i] = PointFromLatLng(ll)
	}
	return points
}

// skewedInt returns a number in the range [0,2^max_log-1] with bias towards smaller numbers.
func skewedInt(maxLog int) int {
	base := uint32(rand.Int31n(int32(maxLog + 1)))
	return int(randomBits(31) & ((1 << base) - 1))
}

// randomCap returns a cap with a random axis such that the log of its area is
// uniformly distributed between the logs of the two given values. The log of
// the cap angle is also approximately uniformly distributed.
func randomCap(minArea, maxArea float64) Cap {
	capArea := maxArea * math.Pow(minArea/maxArea, randomFloat64())
	return CapFromCenterArea(randomPoint(), capArea)
}

// pointsApproxEquals reports whether the two points are within the given distance
// of each other. This is the same as Point.ApproxEquals but permits specifying
// the epsilon.
func pointsApproxEquals(a, b Point, epsilon float64) bool {
	return float64(a.Vector.Angle(b.Vector)) <= epsilon
}

// samplePointFromRect returns a point chosen uniformly at random (with respect
// to area on the sphere) from the given rectangle.
func samplePointFromRect(rect Rect) Point {
	// First choose a latitude uniformly with respect to area on the sphere.
	sinLo := math.Sin(rect.Lat.Lo)
	sinHi := math.Sin(rect.Lat.Hi)
	lat := math.Asin(randomUniformFloat64(sinLo, sinHi))

	// Now choose longitude uniformly within the given range.
	lng := rect.Lng.Lo + randomFloat64()*rect.Lng.Length()

	return PointFromLatLng(LatLng{s1.Angle(lat), s1.Angle(lng)}.Normalized())
}

// TODO:
// Most of the other s2 testing methods.
