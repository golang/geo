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

package s1

import (
	"github.com/golang/geo/earth"
	"github.com/google/go-units/unit"
)

// EarthAngleFromLength returns the angle subtended by a given arc length on
// the spherical earth's surface.
func EarthAngleFromLength(d unit.Length) Angle {
	return Angle(float64(d/earth.Radius)) * Radian
}

// EarthLengthFromAngle returns the arc length on the spherical earth's surface
// corresponding to a given angle.
func EarthLengthFromAngle(a Angle) unit.Length {
	return unit.Length(a.Radians()) * earth.Radius
}

// EarthAreaFromSteradians returns the surface area on the spherical earth
// corresponding to a given solid angle in steradians.
func EarthAreaFromSteradians(s float64) unit.Area {
	return unit.Area(s * earth.Radius.Meters() * earth.Radius.Meters())
}

// EarthSteradiansFromArea returns the solid angle in steradians corresponding
// to a given surface area on the spherical earth.
func EarthSteradiansFromArea(a unit.Area) float64 {
	return a.SquareMeters() / (earth.Radius.Meters() * earth.Radius.Meters())
}
