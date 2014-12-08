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
	"math"
	"testing"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/s1"
)

func TestEmptyAndFullRects(t *testing.T) {
	tests := []struct {
		rect  Rect
		valid bool
		empty bool
		full  bool
		point bool
	}{
		{EmptyRect(), true, true, false, false},
		{FullRect(), true, false, true, false},
	}

	for _, test := range tests {
		if got := test.rect.IsValid(); got != test.valid {
			t.Errorf("%v.IsValid() = %v, want %v", test.rect, got, test.valid)
		}
		if got := test.rect.IsEmpty(); got != test.empty {
			t.Errorf("%v.IsEmpty() = %v, want %v", test.rect, got, test.empty)
		}
		if got := test.rect.IsFull(); got != test.full {
			t.Errorf("%v.IsFull() = %v, want %v", test.rect, got, test.full)
		}
		if got := test.rect.IsPoint(); got != test.point {
			t.Errorf("%v.IsPoint() = %v, want %v", test.rect, got, test.point)
		}
	}
}

func TestArea(t *testing.T) {
	tests := []struct {
		rect Rect
		want float64
	}{
		{Rect{}, 0},
		{FullRect(), 4 * math.Pi},
		{Rect{r1.Interval{0, math.Pi / 2}, s1.Interval{0, math.Pi / 2}}, math.Pi / 2},
	}
	for _, test := range tests {
		got := test.rect.Area()
		if math.Abs(got-test.want) > 1e-14 {
			t.Errorf("%v.Area() = %v, want %v", test.rect, got, test.want)
		}
	}
}

func TestRectString(t *testing.T) {
	const want = "[Lo[-90.0000000, -180.0000000], Hi[90.0000000, 180.0000000]]"
	if s := FullRect().String(); s != want {
		t.Errorf("FullRect().String() = %q, want %q", s, want)
	}
}

func TestRectFromLatLng(t *testing.T) {
	ll := LatLngFromDegrees(23, 47)
	got := RectFromLatLng(ll)
	if got.Center() != ll {
		t.Errorf("RectFromLatLng(%v).Center() = %v, want %v", ll, got.Center(), ll)
	}
	if !got.IsPoint() {
		t.Errorf("RectFromLatLng(%v) = %v, want a point", ll, got)
	}
}

func rectFromDegrees(latLo, lngLo, latHi, lngHi float64) Rect {
	// Convenience method to construct a rectangle. This method is
	// intentionally *not* in the S2LatLngRect interface because the
	// argument order is ambiguous, but is fine for the test.
	return Rect{
		Lat: r1.Interval{
			Lo: (s1.Angle(latLo) * s1.Degree).Radians(),
			Hi: (s1.Angle(latHi) * s1.Degree).Radians(),
		},
		Lng: s1.IntervalFromEndpoints(
			(s1.Angle(lngLo) * s1.Degree).Radians(),
			(s1.Angle(lngHi) * s1.Degree).Radians(),
		),
	}
}

func rectApproxEqual(a, b Rect) bool {
	const epsilon = 1e-15
	return math.Abs(a.Lat.Lo-b.Lat.Lo) < epsilon &&
		math.Abs(a.Lat.Hi-b.Lat.Hi) < epsilon &&
		math.Abs(a.Lng.Lo-b.Lng.Lo) < epsilon &&
		math.Abs(a.Lng.Hi-b.Lng.Hi) < epsilon
}

func TestRectFromCenterSize(t *testing.T) {
	tests := []struct {
		center, size LatLng
		want         Rect
	}{
		{
			LatLngFromDegrees(80, 170),
			LatLngFromDegrees(40, 60),
			rectFromDegrees(60, 140, 90, -160),
		},
		{
			LatLngFromDegrees(10, 40),
			LatLngFromDegrees(210, 400),
			FullRect(),
		},
		{
			LatLngFromDegrees(-90, 180),
			LatLngFromDegrees(20, 50),
			rectFromDegrees(-90, 155, -80, -155),
		},
	}
	for _, test := range tests {
		if got := RectFromCenterSize(test.center, test.size); !rectApproxEqual(got, test.want) {
			t.Errorf("RectFromCenterSize(%v,%v) was %v, want %v", test.center, test.size, got, test.want)
		}
	}
}

func TestAddPoint(t *testing.T) {
	tests := []struct {
		input Rect
		point LatLng
		want  Rect
	}{
		{
			Rect{r1.EmptyInterval(), s1.EmptyInterval()},
			LatLngFromDegrees(0, 0),
			rectFromDegrees(0, 0, 0, 0),
		},
		{
			rectFromDegrees(0, 0, 0, 0),
			LatLng{0 * s1.Radian, (-math.Pi / 2) * s1.Radian},
			rectFromDegrees(0, -90, 0, 0),
		},
		{
			rectFromDegrees(0, -90, 0, 0),
			LatLng{(math.Pi / 4) * s1.Radian, (-math.Pi) * s1.Radian},
			rectFromDegrees(0, -180, 45, 0),
		},
		{
			rectFromDegrees(0, -180, 45, 0),
			LatLng{(math.Pi / 2) * s1.Radian, 0 * s1.Radian},
			rectFromDegrees(0, -180, 90, 0),
		},
	}
	for _, test := range tests {
		if got, want := test.input.AddPoint(test.point), test.want; !rectApproxEqual(got, want) {
			t.Errorf("%v.AddPoint(%v) was %v, want %v", test.input, test.point, got, want)
		}
	}
}

func TestContainsLatLng(t *testing.T) {
	tests := []struct {
		input Rect
		ll    LatLng
		want  bool
	}{
		{
			rectFromDegrees(0, -180, 90, 0),
			LatLngFromDegrees(30, -45),
			true,
		},
		{
			rectFromDegrees(0, -180, 90, 0),
			LatLngFromDegrees(30, 45),
			false,
		},
		{
			rectFromDegrees(0, -180, 90, 0),
			LatLngFromDegrees(0, -180),
			true,
		},
		{
			rectFromDegrees(0, -180, 90, 0),
			LatLngFromDegrees(90, 0),
			true,
		},
	}
	for _, test := range tests {
		if got, want := test.input.ContainsLatLng(test.ll), test.want; got != want {
			t.Errorf("%v.ContainsLatLng(%v) was %v, want %v", test.input, test.ll, got, want)
		}
	}
}

func TestExpanded(t *testing.T) {
	empty := Rect{FullRect().Lat, s1.EmptyInterval()}
	tests := []struct {
		input  Rect
		margin LatLng
		want   Rect
	}{
		{
			rectFromDegrees(70, 150, 80, 170),
			LatLngFromDegrees(20, 30),
			rectFromDegrees(50, 120, 90, -160),
		},
		{
			empty,
			LatLngFromDegrees(20, 30),
			empty,
		},
		{
			FullRect(),
			LatLngFromDegrees(20, 30),
			FullRect(),
		},
		{
			rectFromDegrees(-90, 170, 10, 20),
			LatLngFromDegrees(30, 80),
			rectFromDegrees(-90, -180, 40, 180),
		},
	}
	for _, test := range tests {
		if got, want := test.input.expanded(test.margin), test.want; !rectApproxEqual(got, want) {
			t.Errorf("%v.Expanded(%v) was %v, want %v", test.input, test.margin, got, want)
		}
	}
}
