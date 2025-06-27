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

/*
Package earth implements functions for working with the planet Earth modeled as
a sphere.
*/
package earth

import (
	"math"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/google/go-units/unit"
)

const (
	// Radius is the Earth's mean radius, which is the radius of the
	// equivalent sphere with the same surface area. According to NASA,
	// this value is 6371.01 +/- 0.02 km. The equatorial radius is 6378.136
	// km, and the polar radius is 6356.752 km. They differ by one part in
	// 298.257.
	//
	// Reference: http://ssd.jpl.nasa.gov/phys_props_earth.html, which
	// quotes Yoder, C.F. 1995. "Astrometric and Geodetic Properties of
	// Earth and the Solar System" in Global Earth Physics, A Handbook of
	// Physical Constants, AGU Reference Shelf 1, American Geophysical
	// Union, Table 2.
	//
	// This value is the same as in s2earth.h and S2Earth.java in order to be
	// able to make consistent conversions across programming languages.
	Radius = 6371.01 * unit.Kilometer

	// LowestAltitude is the altitude of the lowest known point on Earth,
	// the Challenger Deep, below the surface of the spherical earth.
	LowestAltitude = -10.898 * unit.Kilometer

	// HighestAltitude is the altitude of the highest known point on Earth,
	// Mount Everest, above the surface of the spherical earth.
	HighestAltitude = 8.846 * unit.Kilometer
)

// AngleFromLength returns the angle from a given distance on the spherical
// earth's surface.
func AngleFromLength(d unit.Length) s1.Angle {
	return s1.Angle(float64(d/Radius)) * s1.Radian
}

// LengthFromAngle returns the distance on the spherical earth's surface from
// a given angle.
func LengthFromAngle(a s1.Angle) unit.Length {
	return unit.Length(a.Radians()) * Radius
}

// LengthFromPoints returns the distance between two points on the spherical
// earth's surface.
func LengthFromPoints(a, b s2.Point) unit.Length {
	return LengthFromAngle(a.Distance(b))
}

// LengthFromLatLngs returns the distance on the spherical earth's surface
// between two LatLngs.
func LengthFromLatLngs(a, b s2.LatLng) unit.Length {
	return LengthFromAngle(a.Distance(b))
}

// AreaFromSteradians returns the area on the spherical Earth's surface covered
// by s steradians, as returned by Area() methods on s2 geometry types.
func AreaFromSteradians(s float64) unit.Area {
	return unit.Area(s * Radius.Meters() * Radius.Meters())
}

// SteradiansFromArea returns the number of steradians covered by an area on the
// spherical Earth's surface.  The value will be between 0 and 4 * math.Pi if a
// does not exceed the area of the Earth.
func SteradiansFromArea(a unit.Area) float64 {
	return a.SquareMeters() / (Radius.Meters() * Radius.Meters())
}

// InitialBearingFromLatLngs computes the initial bearing from a to b.
//
// This is the bearing an observer at point a has when facing point b. A bearing
// of 0 degrees is north, and it increases clockwise (90 degrees is east, etc).
//
// If a == b, a == -b, or a is one of the Earth's poles, the return value is
// undefined.
func InitialBearingFromLatLngs(a, b s2.LatLng) s1.Angle {
	lat1 := a.Lat.Radians()
	cosLat2 := math.Cos(b.Lat.Radians())
	latDiff := b.Lat.Radians() - a.Lat.Radians()
	lngDiff := b.Lng.Radians() - a.Lng.Radians()

	x := math.Sin(latDiff) + math.Sin(lat1)*cosLat2*2*haversine(lngDiff)
	y := math.Sin(lngDiff) * cosLat2
	return s1.Angle(math.Atan2(y, x)) * s1.Radian
}

func haversine(radians float64) float64 {
	sinHalf := math.Sin(radians / 2)
	return sinHalf * sinHalf

}
