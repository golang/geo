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
Package earth provides constants for working with the planet Earth modeled as
a sphere. Functions that operate on angles are in s1 (e.g., s1.EarthAngleFromLength),
and functions that operate on s2 geometry types are in s2 (e.g., s2.EarthLengthFromLatLngs).
*/
package earth

import (
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
	// the Challenger Deep, below the surface of the spherical Earth. This value
	// is the same as the C++ and Java implementations of S2Earth. The value may
	// change as more precise measurements are made.

	LowestAltitude = -10898 * unit.Meter

	// HighestAltitude is the altitude of the highest known point on Earth,
	// Mount Everest, above the surface of the spherical Earth. This value is the
	// same as the C++ and Java implementations of S2Earth. The value may change
	// as more precise measurements are made.

	HighestAltitude = 8848 * unit.Meter
)
