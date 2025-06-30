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

package earth

import (
	"math"
	"testing"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/google/go-units/unit"
)

func float64Eq(x, y float64) bool {
	if x == y {
		return true
	}
	if math.Abs(x) > math.Abs(y) {
		return math.Abs(1-y/x) < 1e-14
	}
	return math.Abs(1-x/y) < 1e-14
}

var degreesToMeters = []struct {
	angle  s1.Angle
	length unit.Length
}{
	{-89.93201943346866 * s1.Degree, -1e7 * unit.Meter},
	{-30 * s1.Degree, -3335853.035324518 * unit.Meter},
	{0 * s1.Degree, 0 * unit.Meter},
	{30 * s1.Degree, 3335853.035324518 * unit.Meter},
	{89.93201943346866 * s1.Degree, 1e7 * unit.Meter},
	{90 * s1.Degree, 10007559.105973555 * unit.Meter},
	{179.86403886693734 * s1.Degree, 2e7 * unit.Meter},
	{180 * s1.Degree, 20015118.21194711 * unit.Meter},
	{359.72807773387467 * s1.Degree, 4e7 * unit.Meter},
	{360 * s1.Degree, 40030236.42389422 * unit.Meter},
	{899.3201943346867 * s1.Degree, 1e8 * unit.Meter},
}

func TestAngleFromLength(t *testing.T) {
	for _, test := range degreesToMeters {
		if got, want := AngleFromLength(test.length), test.angle; !float64Eq(got.Radians(), want.Radians()) {
			t.Errorf("AngleFromLength(%v) = %v, want %v", test.length, got, want)
		}
	}
}

func TestLengthFromAngle(t *testing.T) {
	for _, test := range degreesToMeters {
		if got, want := LengthFromAngle(test.angle), test.length; !float64Eq(got.Meters(), want.Meters()) {
			t.Errorf("LengthFromAngle(%v) = %v, want %v", test.angle, got, want)
		}
	}
}

func TestLengthFromPoints(t *testing.T) {
	tests := []struct {
		x1, y1, z1 float64
		x2, y2, z2 float64
		length     unit.Length
	}{
		{1, 0, 0, 1, 0, 0, 0 * unit.Meter},
		{1, 0, 0, 0, 1, 0, 10007559.105973555 * unit.Meter},
		{1, 0, 0, 0, 1, 1, 10007559.105973555 * unit.Meter},
		{1, 0, 0, -1, 0, 0, 20015118.21194711 * unit.Meter},
		{1, 2, 3, 2, 3, -1, 7680820.247060414 * unit.Meter},
	}
	for _, test := range tests {
		p1 := s2.PointFromCoords(test.x1, test.y1, test.z1)
		p2 := s2.PointFromCoords(test.x2, test.y2, test.z2)
		if got, want := LengthFromPoints(p1, p2), test.length; !float64Eq(got.Meters(), want.Meters()) {
			t.Errorf("LengthFromPoints(%v, %v) = %v, want %v", p1, p2, got, want)
		}
	}
}

func TestLengthFromLatLngs(t *testing.T) {
	tests := []struct {
		lat1, lng1, lat2, lng2 float64
		length                 unit.Length
	}{
		{90, 0, 90, 0, 0 * unit.Meter},
		{-37, 25, -66, -155, 8562022.790666264 * unit.Meter},
		{0, 165, 0, -80, 12787436.635410652 * unit.Meter},
		{47, -127, -47, 53, 20015118.077688109 * unit.Meter},
		{51.961951, -180.227156, 51.782383, 181.126878, 95.0783566198074 * unit.Kilometer},
	}
	for _, test := range tests {
		ll1 := s2.LatLngFromDegrees(test.lat1, test.lng1)
		ll2 := s2.LatLngFromDegrees(test.lat2, test.lng2)
		if got, want := LengthFromLatLngs(ll1, ll2), test.length; !float64Eq(got.Meters(), want.Meters()) {
			t.Errorf("LengthFromLatLngs(%v, %v) = %v, want %v", ll1, ll2, got, want)
		}
	}
}

var (
	earthArea        = unit.Area(Radius.Meters()*Radius.Meters()) * math.Pi * 4
	steradiansToArea = []struct {
		steradians float64
		area       unit.Area
	}{
		{1, earthArea / 4 / math.Pi},
		{4 * math.Pi, earthArea},
		{s2.PolygonFromLoops([]*s2.Loop{s2.FullLoop()}).Area(), earthArea},
		{s2.PolygonFromLoops([]*s2.Loop{s2.LoopFromPoints([]s2.Point{
			s2.PointFromLatLng(s2.LatLngFromDegrees(-90, 0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(0, 0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(90, 0)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(0, -90)),
		})}).Area(), earthArea / 4},
		{s2.CellFromCellID(s2.CellIDFromFace(2)).ExactArea(), earthArea / 6},
		{s2.AvgAreaMetric.Value(10), 81.07281893380302 * unit.SquareKilometer},  // average area of level 10 cells
		{s2.AvgAreaMetric.Value(20), 77.31706517582228 * unit.SquareMeter},      // average area of level 20 cells
		{s2.AvgAreaMetric.Value(30), 73.73529927808979 * unit.SquareMillimeter}, // average area of level 30 cells
		{s2.PolygonFromLoops([]*s2.Loop{s2.EmptyLoop()}).Area(), 0 * unit.SquareMeter},
	}
)

func TestAreaFromSteradians(t *testing.T) {
	for _, test := range steradiansToArea {
		if got, want := AreaFromSteradians(test.steradians), test.area; !float64Eq(got.SquareMeters(), want.SquareMeters()) {
			t.Errorf("AreaFromSteradians(%v) = %v, want %v", test.steradians, got, want)
		}
	}
}

func TestSteradiansFromArea(t *testing.T) {
	for _, test := range steradiansToArea {
		if got, want := SteradiansFromArea(test.area), test.steradians; !float64Eq(got, want) {
			t.Errorf("SteradiansFromArea(%v) = %v, want %v", test.area, got, want)
		}
	}
}

func TestInitialBearingFromLatLngs(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b s2.LatLng
		want s1.Angle
	}{
		{"Westward on equator", s2.LatLngFromDegrees(0, 50),
			s2.LatLngFromDegrees(0, 100), s1.Degree * 90},
		{"Eastward on equator", s2.LatLngFromDegrees(0, 50),
			s2.LatLngFromDegrees(0, 0), s1.Degree * -90},
		{"Northward on meridian", s2.LatLngFromDegrees(16, 28),
			s2.LatLngFromDegrees(81, 28), s1.Degree * 0},
		{"Southward on meridian", s2.LatLngFromDegrees(24, 64),
			s2.LatLngFromDegrees(-27, 64), s1.Degree * 180},
		{"Towards north pole", s2.LatLngFromDegrees(12, 76),
			s2.LatLngFromDegrees(90, 50), s1.Degree * 0},
		{"Towards south pole", s2.LatLngFromDegrees(-35, 105),
			s2.LatLngFromDegrees(-90, -120), s1.Degree * 180},
		{"Spain to Japan", s2.LatLngFromDegrees(40.4379332, -3.749576),
			s2.LatLngFromDegrees(35.6733227, 139.6403486), s1.Degree * 29.2},
		{"Japan to Spain", s2.LatLngFromDegrees(35.6733227, 139.6403486),
			s2.LatLngFromDegrees(40.4379332, -3.749576), s1.Degree * -27.2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := InitialBearingFromLatLngs(tc.a, tc.b)
			if diff := (got - tc.want).Abs(); diff > 0.01 {
				t.Errorf("InitialBearingFromLatLngs(%s, %s): got %s, want %s, diff %s", tc.a, tc.b, got, tc.want, diff)
			}
		})
	}
}

func TestInitialBearingFromLatLngsUndefinedResultDoesNotCrash(t *testing.T) {
	// InitialBearingFromLatLngs says if a == b, a == -b, or a is one of Earth's
	// poles, the return value is undefined.  Make sure it returns a real value
	// (but don't assert what it is) rather than panicking or NaN.
	// Bearing from a pole is undefined because 0° is north, but the observer
	// can't face north from the north pole, so the calculation depends on the
	// latitude value at the pole, even though 90°N 123°E and 90°N 45°W represent
	// the same point.  Bearing is undefined when a == b because the observer can
	// point any direction and still be present.  Bearing is undefined when
	// a == -b (two antipodal points) because there are two possible paths.
	for _, tc := range []struct {
		name string
		a, b s2.LatLng
	}{
		{"North pole prime meridian to Null Island", s2.LatLngFromDegrees(90, 0), s2.LatLngFromDegrees(0, 0)},
		{"North pole facing east to Guatemala", s2.LatLngFromDegrees(90, 90), s2.LatLngFromDegrees(15, -90)},
		{"South pole facing west to McMurdo", s2.LatLngFromDegrees(-90, -90), s2.LatLngFromDegrees(-78, 166)},
		{"South pole anti-prime meridian to Null Island", s2.LatLngFromDegrees(-90, -180), s2.LatLngFromDegrees(0, 0)},
		{"Jakarta and antipode", s2.LatLngFromDegrees(-6.109, 106.668), s2.LatLngFromDegrees(6.109, -180+106.668)},
		{"Alert and antipode", s2.LatLngFromDegrees(82.499, -62.350), s2.LatLngFromDegrees(-82.499, 180-62.350)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := InitialBearingFromLatLngs(tc.a, tc.b)
			if math.IsNaN(got.Radians()) || math.IsInf(got.Radians(), 0) {
				t.Errorf("InitialBearingFromLatLngs(%s, %s): got %s, want a real value", tc.a, tc.b, got)
			}
		})
	}
}
