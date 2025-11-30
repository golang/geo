// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s2

import (
	"math"

	"github.com/golang/geo/s1"
	"github.com/google/go-units/unit"
)

// EarthLengthFromPoints returns the distance between two points on the
// spherical earth's surface.
func EarthLengthFromPoints(a, b Point) unit.Length {
	return s1.EarthLengthFromAngle(a.Distance(b))
}

// EarthLengthFromLatLngs returns the distance on the spherical earth's surface
// between two LatLngs.
func EarthLengthFromLatLngs(a, b LatLng) unit.Length {
	return s1.EarthLengthFromAngle(a.Distance(b))
}

// EarthInitialBearingFromLatLngs computes the initial bearing from a to b.
//
// This is the bearing an observer at point a has when facing point b. A bearing
// of 0 degrees is north, and it increases clockwise (90 degrees is east, etc).
//
// If a == b, a == -b, or a is one of the Earth's poles, the return value is
// undefined.
func EarthInitialBearingFromLatLngs(a, b LatLng) s1.Angle {
	latitudeA := a.Lat.Radians()
	cosLatitudeB := math.Cos(b.Lat.Radians())
	latitudeDiff := b.Lat.Radians() - a.Lat.Radians()
	longitudeDiff := b.Lng.Radians() - a.Lng.Radians()

	x := math.Sin(latitudeDiff) + math.Sin(latitudeA)*cosLatitudeB*2*haversine(longitudeDiff)
	y := math.Sin(longitudeDiff) * cosLatitudeB
	return s1.Angle(math.Atan2(y, x)) * s1.Radian
}

func haversine(radians float64) float64 {
	sinHalf := math.Sin(radians / 2)
	return sinHalf * sinHalf
}
