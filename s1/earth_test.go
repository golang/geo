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
	"math"
	"testing"

	"github.com/golang/geo/earth"
	"github.com/google/go-units/unit"
)

func earthFloat64Eq(x, y float64) bool {
	if x == y {
		return true
	}
	if math.Abs(x) > math.Abs(y) {
		return math.Abs(1-y/x) < 1e-14
	}
	return math.Abs(1-x/y) < 1e-14
}

var degreesToMeters = []struct {
	angle  Angle
	length unit.Length
}{
	{-89.93201943346866 * Degree, -1e7 * unit.Meter},
	{-30 * Degree, -3335853.035324518 * unit.Meter},
	{0 * Degree, 0 * unit.Meter},
	{30 * Degree, 3335853.035324518 * unit.Meter},
	{89.93201943346866 * Degree, 1e7 * unit.Meter},
	{90 * Degree, 10007559.105973555 * unit.Meter},
	{179.86403886693734 * Degree, 2e7 * unit.Meter},
	{180 * Degree, 20015118.21194711 * unit.Meter},
	{359.72807773387467 * Degree, 4e7 * unit.Meter},
	{360 * Degree, 40030236.42389422 * unit.Meter},
	{899.3201943346867 * Degree, 1e8 * unit.Meter},
}

func TestEarthAngleFromLength(t *testing.T) {
	for _, test := range degreesToMeters {
		got := EarthAngleFromLength(test.length)
		want := test.angle
		if !earthFloat64Eq(got.Radians(), want.Radians()) {
			t.Errorf("EarthAngleFromLength(%v) = %v, want %v", test.length, got, want)
		}
	}

	// Verify the fundamental identity: earth.Radius maps to exactly 1 radian.
	if got := EarthAngleFromLength(earth.Radius); got != 1*Radian {
		t.Errorf("EarthAngleFromLength(earth.Radius) = %v, want %v", got, 1*Radian)
	}
}

func TestEarthLengthFromAngle(t *testing.T) {
	for _, test := range degreesToMeters {
		got := EarthLengthFromAngle(test.angle)
		want := test.length
		if !earthFloat64Eq(got.Meters(), want.Meters()) {
			t.Errorf("EarthLengthFromAngle(%v) = %v, want %v", test.angle, got, want)
		}
	}

	// Verify the fundamental identity: 1 radian maps to exactly earth.Radius.
	if got := EarthLengthFromAngle(1 * Radian); got != earth.Radius {
		t.Errorf("EarthLengthFromAngle(1*Radian) = %v, want %v", got, earth.Radius)
	}
}

var (
	earthArea = unit.Area(earth.Radius.Meters()*earth.Radius.Meters()) * math.Pi * 4

	steradiansToArea = []struct {
		steradians float64
		area       unit.Area
	}{
		{0, 0 * unit.SquareMeter},
		{1, earthArea / 4 / math.Pi},
		{math.Pi, earthArea / 4},
		{2 * math.Pi, earthArea / 2},
		{4 * math.Pi, earthArea},
	}
)

func TestEarthAreaFromSteradians(t *testing.T) {
	for _, test := range steradiansToArea {
		got := EarthAreaFromSteradians(test.steradians)
		want := test.area
		if !earthFloat64Eq(got.SquareMeters(), want.SquareMeters()) {
			t.Errorf("EarthAreaFromSteradians(%v) = %v, want %v", test.steradians, got, want)
		}
	}
}

func TestEarthSteradiansFromArea(t *testing.T) {
	for _, test := range steradiansToArea {
		got := EarthSteradiansFromArea(test.area)
		want := test.steradians
		if !earthFloat64Eq(got, want) {
			t.Errorf("EarthSteradiansFromArea(%v) = %v, want %v", test.area, got, want)
		}
	}
}
