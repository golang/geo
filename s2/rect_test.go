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
		if got := test.rect.Area(); !float64Eq(got, test.want) {
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
		if got := RectFromCenterSize(test.center, test.size); !rectsApproxEqual(got, test.want, epsilon, epsilon) {
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
		if got, want := test.input.AddPoint(test.point), test.want; !rectsApproxEqual(got, want, epsilon, epsilon) {
			t.Errorf("%v.AddPoint(%v) was %v, want %v", test.input, test.point, got, want)
		}
	}
}
func TestRectVertex(t *testing.T) {
	r1 := Rect{r1.Interval{0, math.Pi / 2}, s1.IntervalFromEndpoints(-math.Pi, 0)}
	tests := []struct {
		r    Rect
		i    int
		want LatLng
	}{
		{r1, 0, LatLng{0, math.Pi}},
		{r1, 1, LatLng{0, 0}},
		{r1, 2, LatLng{math.Pi / 2, 0}},
		{r1, 3, LatLng{math.Pi / 2, math.Pi}},
	}

	for _, test := range tests {
		if got := test.r.Vertex(test.i); got != test.want {
			t.Errorf("%v.Vertex(%d) = %v, want %v", test.r, test.i, got, test.want)
		}
	}
}
func TestRectVertexCCWOrder(t *testing.T) {
	for i := 0; i < 4; i++ {
		lat := math.Pi / 4 * float64(i-2)
		lng := math.Pi/2*float64(i-2) + 0.2
		r := Rect{
			r1.Interval{lat, lat + math.Pi/4},
			s1.Interval{
				math.Remainder(lng, 2*math.Pi),
				math.Remainder(lng+math.Pi/2, 2*math.Pi),
			},
		}

		for k := 0; k < 4; k++ {
			if !Sign(PointFromLatLng(r.Vertex((k-1)&3)), PointFromLatLng(r.Vertex(k)), PointFromLatLng(r.Vertex((k+1)&3))) {
				t.Errorf("%v.Vertex(%v), vertices were not in CCW order", r, k)
			}
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
			EmptyRect(),
			LatLngFromDegrees(20, 30),
			EmptyRect(),
		},
		{
			FullRect(),
			LatLngFromDegrees(500, 500),
			FullRect(),
		},
		{
			rectFromDegrees(-90, 170, 10, 20),
			LatLngFromDegrees(30, 80),
			rectFromDegrees(-90, -180, 40, 180),
		},

		// Negative margins.
		{
			rectFromDegrees(10, -50, 60, 70),
			LatLngFromDegrees(-10, -10),
			rectFromDegrees(20, -40, 50, 60),
		},
		{
			rectFromDegrees(-20, -180, 20, 180),
			LatLngFromDegrees(-10, -10),
			rectFromDegrees(-10, -180, 10, 180),
		},
		{
			rectFromDegrees(-20, -180, 20, 180),
			LatLngFromDegrees(-30, -30),
			EmptyRect(),
		},
		{
			rectFromDegrees(-90, 10, 90, 11),
			LatLngFromDegrees(-10, -10),
			EmptyRect(),
		},
		{
			rectFromDegrees(-90, 10, 90, 100),
			LatLngFromDegrees(-10, -10),
			rectFromDegrees(-80, 20, 80, 90),
		},
		{
			EmptyRect(),
			LatLngFromDegrees(-50, -500),
			EmptyRect(),
		},
		{
			FullRect(),
			LatLngFromDegrees(-50, -50),
			rectFromDegrees(-40, -180, 40, 180),
		},

		// Mixed margins.
		{
			rectFromDegrees(10, -50, 60, 70),
			LatLngFromDegrees(-10, 30),
			rectFromDegrees(20, -80, 50, 100),
		},
		{
			rectFromDegrees(-20, -180, 20, 180),
			LatLngFromDegrees(10, -500),
			rectFromDegrees(-30, -180, 30, 180),
		},
		{
			rectFromDegrees(-90, -180, 80, 180),
			LatLngFromDegrees(-30, 500),
			rectFromDegrees(-60, -180, 50, 180),
		},
		{
			rectFromDegrees(-80, -100, 80, 150),
			LatLngFromDegrees(30, -50),
			rectFromDegrees(-90, -50, 90, 100),
		},
		{
			rectFromDegrees(0, -180, 50, 180),
			LatLngFromDegrees(-30, 500),
			EmptyRect(),
		},
		{
			rectFromDegrees(-80, 10, 70, 20),
			LatLngFromDegrees(30, -200),
			EmptyRect(),
		},
		{
			EmptyRect(),
			LatLngFromDegrees(100, -100),
			EmptyRect(),
		},
		{
			FullRect(),
			LatLngFromDegrees(100, -100),
			FullRect(),
		},
	}
	for _, test := range tests {
		if got, want := test.input.expanded(test.margin), test.want; !rectsApproxEqual(got, want, epsilon, epsilon) {
			t.Errorf("%v.Expanded(%v) was %v, want %v", test.input, test.margin, got, want)
		}
	}
}

func TestPolarClosure(t *testing.T) {
	tests := []struct {
		r    Rect
		want Rect
	}{
		{
			rectFromDegrees(-89, 0, 89, 1),
			rectFromDegrees(-89, 0, 89, 1),
		},
		{
			rectFromDegrees(-90, -30, -45, 100),
			rectFromDegrees(-90, -180, -45, 180),
		},
		{
			rectFromDegrees(89, 145, 90, 146),
			rectFromDegrees(89, -180, 90, 180),
		},
		{
			rectFromDegrees(-90, -145, 90, -144),
			FullRect(),
		},
	}
	for _, test := range tests {
		if got := test.r.PolarClosure(); !rectsApproxEqual(got, test.want, epsilon, epsilon) {
			t.Errorf("%v.PolarClosure() was %v, want %v", test.r, got, test.want)
		}
	}
}

func TestRectCapBound(t *testing.T) {
	tests := []struct {
		r    Rect
		want Cap
	}{
		{ // Bounding cap at center is smaller.
			rectFromDegrees(-45, -45, 45, 45),
			CapFromCenterHeight(PointFromCoords(1, 0, 0), 0.5),
		},
		{ // Bounding cap at north pole is smaller.
			rectFromDegrees(88, -80, 89, 80),
			CapFromCenterAngle(PointFromCoords(0, 0, 1), s1.Angle(2)*s1.Degree),
		},
		{ // Longitude span > 180 degrees.
			rectFromDegrees(-30, -150, -10, 50),
			CapFromCenterAngle(PointFromCoords(0, 0, -1), s1.Angle(80)*s1.Degree),
		},
	}
	for _, test := range tests {
		if got := test.r.CapBound(); !test.want.ApproxEqual(got) {
			t.Errorf("%v.CapBound() was %v, want %v", test.r, got, test.want)
		}
	}
}

func TestRectIntervalOps(t *testing.T) {
	// Rectangle that covers one-quarter of the sphere.
	rect := rectFromDegrees(0, -180, 90, 0)

	// Test operations where one rectangle consists of a single point.
	rectMid := rectFromDegrees(45, -90, 45, -90)
	rect180 := rectFromDegrees(0, -180, 0, -180)
	northPole := rectFromDegrees(90, 0, 90, 0)

	tests := []struct {
		rect         Rect
		other        Rect
		contains     bool
		intersects   bool
		union        Rect
		intersection Rect
	}{
		{
			rect:         rect,
			other:        rectMid,
			contains:     true,
			intersects:   true,
			union:        rect,
			intersection: rectMid,
		},
		{
			rect:         rect,
			other:        rect180,
			contains:     true,
			intersects:   true,
			union:        rect,
			intersection: rect180,
		},
		{
			rect:         rect,
			other:        northPole,
			contains:     true,
			intersects:   true,
			union:        rect,
			intersection: northPole,
		},
		{
			rect:         rect,
			other:        rectFromDegrees(-10, -1, 1, 20),
			contains:     false,
			intersects:   true,
			union:        rectFromDegrees(-10, 180, 90, 20),
			intersection: rectFromDegrees(0, -1, 1, 0),
		},
		{
			rect:         rect,
			other:        rectFromDegrees(-10, -1, 0, 20),
			contains:     false,
			intersects:   true,
			union:        rectFromDegrees(-10, 180, 90, 20),
			intersection: rectFromDegrees(0, -1, 0, 0),
		},
		{
			rect:         rect,
			other:        rectFromDegrees(-10, 0, 1, 20),
			contains:     false,
			intersects:   true,
			union:        rectFromDegrees(-10, 180, 90, 20),
			intersection: rectFromDegrees(0, 0, 1, 0),
		},
		{
			rect:         rectFromDegrees(-15, -160, -15, -150),
			other:        rectFromDegrees(20, 145, 25, 155),
			contains:     false,
			intersects:   false,
			union:        rectFromDegrees(-15, 145, 25, -150),
			intersection: EmptyRect(),
		},
		{
			rect:         rectFromDegrees(70, -10, 90, -140),
			other:        rectFromDegrees(60, 175, 80, 5),
			contains:     false,
			intersects:   true,
			union:        rectFromDegrees(60, -180, 90, 180),
			intersection: rectFromDegrees(70, 175, 80, 5),
		},

		// Check that the intersection of two rectangles that overlap in latitude
		// but not longitude is valid, and vice versa.
		{
			rect:         rectFromDegrees(12, 30, 60, 60),
			other:        rectFromDegrees(0, 0, 30, 18),
			contains:     false,
			intersects:   false,
			union:        rectFromDegrees(0, 0, 60, 60),
			intersection: EmptyRect(),
		},
		{
			rect:         rectFromDegrees(0, 0, 18, 42),
			other:        rectFromDegrees(30, 12, 42, 60),
			contains:     false,
			intersects:   false,
			union:        rectFromDegrees(0, 0, 42, 60),
			intersection: EmptyRect(),
		},
	}
	for _, test := range tests {
		if got := test.rect.Contains(test.other); got != test.contains {
			t.Errorf("%v.Contains(%v) = %t, want %t", test.rect, test.other, got, test.contains)
		}

		if got := test.rect.Intersects(test.other); got != test.intersects {
			t.Errorf("%v.Intersects(%v) = %t, want %t", test.rect, test.other, got, test.intersects)
		}

		if got := test.rect.Union(test.other) == test.rect; test.rect.Contains(test.other) != got {
			t.Errorf("%v.Union(%v) == %v = %t, want %t",
				test.rect, test.other, test.other, got, test.rect.Contains(test.other),
			)
		}

		if got := test.rect.Intersection(test.other).IsEmpty(); test.rect.Intersects(test.other) == got {
			t.Errorf("%v.Intersection(%v).IsEmpty() = %t, want %t",
				test.rect, test.other, got, test.rect.Intersects(test.other))
		}

		if got := test.rect.Union(test.other); got != test.union {
			t.Errorf("%v.Union(%v) = %v, want %v", test.rect, test.other, got, test.union)
		}

		if got := test.rect.Intersection(test.other); got != test.intersection {
			t.Errorf("%v.Intersection(%v) = %v, want %v", test.rect, test.other, got, test.intersection)
		}
	}
}

func TestRectContainsCell(t *testing.T) {
	cell0 := CellFromPoint(PointFromCoords(1+1e-12, 1, 1))
	vertex0 := LatLngFromPoint(cell0.Vertex(0))

	cell202 := CellFromCellID(CellIDFromFacePosLevel(2, 0, 2))
	bound202 := cell202.RectBound()

	tests := []struct {
		r    Rect
		c    Cell
		want bool
	}{
		{
			EmptyRect(),
			CellFromCellID(CellIDFromFacePosLevel(3, 0, 0)),
			false,
		},
		{
			FullRect(),
			CellFromCellID(CellIDFromFacePosLevel(2, 0, 0)),
			true,
		},
		{
			FullRect(),
			CellFromCellID(CellIDFromFacePosLevel(5, 0, 25)),
			true,
		},
		// This rectangle includes the first quadrant of face 0.  It's expanded
		// slightly because cell bounding rectangles are slightly conservative.
		{
			rectFromDegrees(-45.1, -45.1, 0.1, 0.1),
			CellFromCellID(CellIDFromFacePosLevel(0, 0, 0)),
			false,
		},
		{
			rectFromDegrees(-45.1, -45.1, 0.1, 0.1),
			CellFromCellID(CellIDFromFacePosLevel(0, 0, 1)),
			true,
		},
		{
			rectFromDegrees(-45.1, -45.1, 0.1, 0.1),
			CellFromCellID(CellIDFromFacePosLevel(1, 0, 1)),
			false,
		},
		// This rectangle intersects the first quadrant of face 0.
		{
			rectFromDegrees(-10, -45, 10, 0),
			CellFromCellID(CellIDFromFacePosLevel(0, 0, 0)),
			false,
		},
		{
			rectFromDegrees(-10, -45, 10, 0),
			CellFromCellID(CellIDFromFacePosLevel(0, 0, 1)),
			false,
		},
		{
			rectFromDegrees(-10, -45, 10, 0),
			CellFromCellID(CellIDFromFacePosLevel(1, 0, 1)),
			false,
		},
		// Rectangle consisting of a single point.
		{
			rectFromDegrees(4, 4, 4, 4),
			CellFromCellID(CellIDFromFace(0)),
			false,
		},
		// Rectangles that intersect the bounding rectangle of a face
		// but not the face itself.
		{
			rectFromDegrees(41, -87, 42, -79),
			CellFromCellID(CellIDFromFace(2)),
			false,
		},
		{
			rectFromDegrees(-41, 160, -40, -160),
			CellFromCellID(CellIDFromFace(5)),
			false,
		},
		{
			// This is the leaf cell at the top right hand corner of face 0.
			// It has two angles of 60 degrees and two of 120 degrees.
			rectFromDegrees(vertex0.Lat.Degrees()-1e-8,
				vertex0.Lng.Degrees()-1e-8,
				vertex0.Lat.Degrees()-2e-10,
				vertex0.Lng.Degrees()+1e-10),
			cell0,
			false,
		},
		{
			// Rectangles that intersect a face but where no vertex of one region
			// is contained by the other region.  The first one passes through
			// a corner of one of the face cells.
			rectFromDegrees(-37, -70, -36, -20),
			CellFromCellID(CellIDFromFace(5)),
			false,
		},
		{
			// These two intersect like a diamond and a square.
			rectFromDegrees(bound202.Lo().Lat.Degrees()+3,
				bound202.Lo().Lng.Degrees()+3,
				bound202.Hi().Lat.Degrees()-3,
				bound202.Hi().Lng.Degrees()-3),
			cell202,
			false,
		},
	}

	for _, test := range tests {
		if got := test.r.ContainsCell(test.c); got != test.want {
			t.Errorf("%v.ContainsCell(%v) = %v, want %v", test.r, test.c, got, test.want)
		}
	}

}
func TestRectContainsPoint(t *testing.T) {
	r1 := rectFromDegrees(0, -180, 90, 0)

	tests := []struct {
		r    Rect
		p    Point
		want bool
	}{
		{r1, PointFromCoords(0.5, -0.3, 0.1), true},
		{r1, PointFromCoords(0.5, 0.2, 0.1), false},
	}
	for _, test := range tests {
		if got, want := test.r.ContainsPoint(test.p), test.want; got != want {
			t.Errorf("%v.ContainsPoint(%v) was %v, want %v", test.r, test.p, got, want)
		}
	}
}
