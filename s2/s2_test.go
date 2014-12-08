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

func float64Eq(x, y float64) bool { return math.Abs(x-y) < 1e-14 }

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

// Return a uniformly distributed 64-bit unsigned integer.
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

// randomUniformDouble returns a uniformly distributed value in the range [min, max).
func randomUniformDouble(min, max float64) float64 {
	return min + randomFloat64()*(max-min)
}

// randomPoint returns a random unit-length vector.
func randomPoint() Point {
	return Point{PointFromCoords(randomUniformDouble(-1, 1),
		randomUniformDouble(-1, 1), randomUniformDouble(-1, 1)).Normalize()}
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
		// Skip empty strings.
		if len(strings.TrimSpace(piece)) == 0 {
			continue
		}

		p := strings.Split(strings.TrimSpace(piece), ":")
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
