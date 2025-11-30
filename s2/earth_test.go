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
	"testing"

	"github.com/golang/geo/s1"
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

func TestEarthLengthFromPoints(t *testing.T) {
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
		p1 := PointFromCoords(test.x1, test.y1, test.z1)
		p2 := PointFromCoords(test.x2, test.y2, test.z2)
		got := EarthLengthFromPoints(p1, p2)
		want := test.length
		if !earthFloat64Eq(got.Meters(), want.Meters()) {
			t.Errorf("EarthLengthFromPoints(%v, %v) = %v, want %v", p1, p2, got, want)
		}
	}
}

func TestEarthLengthFromLatLngs(t *testing.T) {
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
		ll1 := LatLngFromDegrees(test.lat1, test.lng1)
		ll2 := LatLngFromDegrees(test.lat2, test.lng2)
		got := EarthLengthFromLatLngs(ll1, ll2)
		want := test.length
		if !earthFloat64Eq(got.Meters(), want.Meters()) {
			t.Errorf("EarthLengthFromLatLngs(%v, %v) = %v, want %v", ll1, ll2, got, want)
		}
	}
}

func TestEarthInitialBearingFromLatLngs(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b LatLng
		want s1.Angle
	}{
		{
			"Westward on equator",
			LatLngFromDegrees(0, 50),
			LatLngFromDegrees(0, 100),
			s1.Degree * 90,
		},
		{
			"Eastward on equator",
			LatLngFromDegrees(0, 50),
			LatLngFromDegrees(0, 0),
			s1.Degree * -90,
		},
		{
			"Northward on meridian",
			LatLngFromDegrees(16, 28),
			LatLngFromDegrees(81, 28),
			s1.Degree * 0,
		},
		{
			"Southward on meridian",
			LatLngFromDegrees(24, 64),
			LatLngFromDegrees(-27, 64),
			s1.Degree * 180,
		},
		{
			"Towards north pole",
			LatLngFromDegrees(12, 76),
			LatLngFromDegrees(90, 50),
			s1.Degree * 0,
		},
		{
			"Towards south pole",
			LatLngFromDegrees(-35, 105),
			LatLngFromDegrees(-90, -120),
			s1.Degree * 180,
		},
		{
			"Spain to Japan",
			LatLngFromDegrees(40.4379332, -3.749576),
			LatLngFromDegrees(35.6733227, 139.6403486),
			s1.Degree * 29.2,
		},
		{
			"Japan to Spain",
			LatLngFromDegrees(35.6733227, 139.6403486),
			LatLngFromDegrees(40.4379332, -3.749576),
			s1.Degree * -27.2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := EarthInitialBearingFromLatLngs(tc.a, tc.b)
			if diff := (got - tc.want).Abs(); diff > 0.01 {
				t.Errorf(
					"EarthInitialBearingFromLatLngs(%s, %s): got %s, want %s, diff %s",
					tc.a,
					tc.b,
					got,
					tc.want,
					diff,
				)
			}
		})
	}
}

func TestEarthInitialBearingFromLatLngsUndefinedResultDoesNotCrash(t *testing.T) {
	// EarthInitialBearingFromLatLngs says if a == b, a == -b, or a is one of
	// Earth's poles, the return value is undefined. Make sure it returns a real
	// value (but don't assert what it is) rather than panicking or NaN.
	for _, tc := range []struct {
		name string
		a, b LatLng
	}{
		{
			"North pole prime meridian to Null Island",
			LatLngFromDegrees(90, 0),
			LatLngFromDegrees(0, 0),
		},
		{
			"North pole facing east to Guatemala",
			LatLngFromDegrees(90, 90),
			LatLngFromDegrees(15, -90),
		},
		{
			"South pole facing west to McMurdo",
			LatLngFromDegrees(-90, -90),
			LatLngFromDegrees(-78, 166),
		},
		{
			"South pole anti-prime meridian to Null Island",
			LatLngFromDegrees(-90, -180),
			LatLngFromDegrees(0, 0),
		},
		{
			"Jakarta and antipode",
			LatLngFromDegrees(-6.109, 106.668),
			LatLngFromDegrees(6.109, -180+106.668),
		},
		{
			"Alert and antipode",
			LatLngFromDegrees(82.499, -62.350),
			LatLngFromDegrees(-82.499, 180-62.350),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := EarthInitialBearingFromLatLngs(tc.a, tc.b)
			if math.IsNaN(got.Radians()) || math.IsInf(got.Radians(), 0) {
				t.Errorf(
					"EarthInitialBearingFromLatLngs(%s, %s): got %s, want a real value",
					tc.a,
					tc.b,
					got,
				)
			}
		})
	}
}
