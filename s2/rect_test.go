// Copyright 2014 Google Inc. All rights reserved.
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

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
	"github.com/golang/geo/r3"
	"github.com/golang/geo/s1"
)

func TestRectEmptyAndFull(t *testing.T) {
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

func TestRectArea(t *testing.T) {
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

func TestRectAddPoint(t *testing.T) {
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

func TestRectContainsLatLng(t *testing.T) {
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

func TestRectExpanded(t *testing.T) {
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

func TestRectPolarClosure(t *testing.T) {
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
			CapFromCenterHeight(Point{r3.Vector{1, 0, 0}}, 0.5),
		},
		{ // Bounding cap at north pole is smaller.
			rectFromDegrees(88, -80, 89, 80),
			CapFromCenterAngle(Point{r3.Vector{0, 0, 1}}, s1.Angle(2)*s1.Degree),
		},
		{ // Longitude span > 180 degrees.
			rectFromDegrees(-30, -150, -10, 50),
			CapFromCenterAngle(Point{r3.Vector{0, 0, -1}}, s1.Angle(80)*s1.Degree),
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

func TestRectCellOps(t *testing.T) {
	cell0 := CellFromPoint(Point{r3.Vector{1 + 1e-12, 1, 1}})
	v0 := LatLngFromPoint(cell0.Vertex(0))

	cell202 := CellFromCellID(CellIDFromFacePosLevel(2, 0, 2))
	bound202 := cell202.RectBound()

	tests := []struct {
		r          Rect
		c          Cell
		contains   bool
		intersects bool
	}{
		// Special cases
		{
			r:          EmptyRect(),
			c:          CellFromCellID(CellIDFromFacePosLevel(3, 0, 0)),
			contains:   false,
			intersects: false,
		},
		{
			r:          FullRect(),
			c:          CellFromCellID(CellIDFromFacePosLevel(2, 0, 0)),
			contains:   true,
			intersects: true,
		},
		{
			r:          FullRect(),
			c:          CellFromCellID(CellIDFromFacePosLevel(5, 0, 25)),
			contains:   true,
			intersects: true,
		},
		// This rectangle includes the first quadrant of face 0.  It's expanded
		// slightly because cell bounding rectangles are slightly conservative.
		{
			r:          rectFromDegrees(-45.1, -45.1, 0.1, 0.1),
			c:          CellFromCellID(CellIDFromFacePosLevel(0, 0, 0)),
			contains:   false,
			intersects: true,
		},
		{
			r:          rectFromDegrees(-45.1, -45.1, 0.1, 0.1),
			c:          CellFromCellID(CellIDFromFacePosLevel(0, 0, 1)),
			contains:   true,
			intersects: true,
		},
		{
			r:          rectFromDegrees(-45.1, -45.1, 0.1, 0.1),
			c:          CellFromCellID(CellIDFromFacePosLevel(1, 0, 1)),
			contains:   false,
			intersects: false,
		},
		// This rectangle intersects the first quadrant of face 0.
		{
			r:          rectFromDegrees(-10, -45, 10, 0),
			c:          CellFromCellID(CellIDFromFacePosLevel(0, 0, 0)),
			contains:   false,
			intersects: true,
		},
		{
			r:          rectFromDegrees(-10, -45, 10, 0),
			c:          CellFromCellID(CellIDFromFacePosLevel(0, 0, 1)),
			contains:   false,
			intersects: true,
		},
		{
			r:          rectFromDegrees(-10, -45, 10, 0),
			c:          CellFromCellID(CellIDFromFacePosLevel(1, 0, 1)),
			contains:   false,
			intersects: false,
		},
		// Rectangle consisting of a single point.
		{
			r:          rectFromDegrees(4, 4, 4, 4),
			c:          CellFromCellID(CellIDFromFace(0)),
			contains:   false,
			intersects: true,
		},
		// Rectangles that intersect the bounding rectangle of a face
		// but not the face itself.
		{
			r:          rectFromDegrees(41, -87, 42, -79),
			c:          CellFromCellID(CellIDFromFace(2)),
			contains:   false,
			intersects: false,
		},
		{
			r:          rectFromDegrees(-41, 160, -40, -160),
			c:          CellFromCellID(CellIDFromFace(5)),
			contains:   false,
			intersects: false,
		},
		{
			// This is the leaf cell at the top right hand corner of face 0.
			// It has two angles of 60 degrees and two of 120 degrees.
			r: rectFromDegrees(v0.Lat.Degrees()-1e-8,
				v0.Lng.Degrees()-1e-8,
				v0.Lat.Degrees()-2e-10,
				v0.Lng.Degrees()+1e-10),
			c:          cell0,
			contains:   false,
			intersects: false,
		},
		{
			// Rectangles that intersect a face but where no vertex of one region
			// is contained by the other region.  The first one passes through
			// a corner of one of the face cells.
			r:          rectFromDegrees(-37, -70, -36, -20),
			c:          CellFromCellID(CellIDFromFace(5)),
			contains:   false,
			intersects: true,
		},
		{
			// These two intersect like a diamond and a square.
			r: rectFromDegrees(bound202.Lo().Lat.Degrees()+3,
				bound202.Lo().Lng.Degrees()+3,
				bound202.Hi().Lat.Degrees()-3,
				bound202.Hi().Lng.Degrees()-3),
			c:          cell202,
			contains:   false,
			intersects: true,
		},
		{
			// from a bug report
			r:          rectFromDegrees(34.2572864, 135.2673642, 34.2707907, 135.2995742),
			c:          CellFromCellID(0x6007500000000000),
			contains:   false,
			intersects: true,
		},
	}

	for _, test := range tests {
		if got := test.r.ContainsCell(test.c); got != test.contains {
			t.Errorf("%v.ContainsCell(%v) = %t, want %t", test.r, test.c, got, test.contains)
		}

		if got := test.r.IntersectsCell(test.c); got != test.intersects {
			t.Errorf("%v.IntersectsCell(%v) = %t, want %t", test.r, test.c, got, test.intersects)
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
		{r1, Point{r3.Vector{0.5, -0.3, 0.1}}, true},
		{r1, Point{r3.Vector{0.5, 0.2, 0.1}}, false},
	}
	for _, test := range tests {
		if got, want := test.r.ContainsPoint(test.p), test.want; got != want {
			t.Errorf("%v.ContainsPoint(%v) was %v, want %v", test.r, test.p, got, want)
		}
	}
}

func TestRectIntersectsLatEdge(t *testing.T) {
	tests := []struct {
		a, b  Point
		lat   s1.Angle
		lngLo s1.Angle
		lngHi s1.Angle
		want  bool
	}{
		{
			a:     Point{r3.Vector{-1, -1, 1}},
			b:     Point{r3.Vector{1, -1, 1}},
			lat:   41 * s1.Degree,
			lngLo: -87 * s1.Degree,
			lngHi: -79 * s1.Degree,
			want:  false,
		},
		{
			a:     Point{r3.Vector{-1, -1, 1}},
			b:     Point{r3.Vector{1, -1, 1}},
			lat:   42 * s1.Degree,
			lngLo: -87 * s1.Degree,
			lngHi: -79 * s1.Degree,
			want:  false,
		},
		{
			a:     Point{r3.Vector{-1, -1, -1}},
			b:     Point{r3.Vector{1, 1, 0}},
			lat:   -3 * s1.Degree,
			lngLo: -1 * s1.Degree,
			lngHi: 23 * s1.Degree,
			want:  false,
		},
		{
			a:     Point{r3.Vector{1, 0, 1}},
			b:     Point{r3.Vector{1, -1, 0}},
			lat:   -28 * s1.Degree,
			lngLo: 69 * s1.Degree,
			lngHi: 115 * s1.Degree,
			want:  false,
		},
		{
			a:     Point{r3.Vector{0, 1, 0}},
			b:     Point{r3.Vector{1, -1, -1}},
			lat:   44 * s1.Degree,
			lngLo: 60 * s1.Degree,
			lngHi: 177 * s1.Degree,
			want:  false,
		},
		{
			a:     Point{r3.Vector{0, 1, 1}},
			b:     Point{r3.Vector{0, 1, -1}},
			lat:   -25 * s1.Degree,
			lngLo: -74 * s1.Degree,
			lngHi: -165 * s1.Degree,
			want:  true,
		},
		{
			a:     Point{r3.Vector{1, 0, 0}},
			b:     Point{r3.Vector{0, 0, -1}},
			lat:   -4 * s1.Degree,
			lngLo: -152 * s1.Degree,
			lngHi: 171 * s1.Degree,
			want:  true,
		},
		// from a bug report
		{
			a:     Point{r3.Vector{-0.589375791872893683986945, 0.583248451588733285433364, 0.558978908075738245564423}},
			b:     Point{r3.Vector{-0.587388131301997518107783, 0.581281455376392863776402, 0.563104832905072516524569}},
			lat:   34.2572864 * s1.Degree,
			lngLo: 2.3608609 * s1.Radian,
			lngHi: 2.3614230 * s1.Radian,
			want:  true,
		},
	}

	for _, test := range tests {
		if got := intersectsLatEdge(test.a, test.b, test.lat, s1.Interval{float64(test.lngLo), float64(test.lngHi)}); got != test.want {
			t.Errorf("intersectsLatEdge(%v, %v, %v, {%v, %v}) = %t, want %t",
				test.a, test.b, test.lat, test.lngLo, test.lngHi, got, test.want)
		}
	}
}

func TestRectIntersectsLngEdge(t *testing.T) {
	tests := []struct {
		a, b  Point
		latLo s1.Angle
		latHi s1.Angle
		lng   s1.Angle
		want  bool
	}{
		{
			a:     Point{r3.Vector{-1, -1, 1}},
			b:     Point{r3.Vector{1, -1, 1}},
			latLo: 41 * s1.Degree,
			latHi: 42 * s1.Degree,
			lng:   -79 * s1.Degree,
			want:  false,
		},
		{
			a:     Point{r3.Vector{-1, -1, 1}},
			b:     Point{r3.Vector{1, -1, 1}},
			latLo: 41 * s1.Degree,
			latHi: 42 * s1.Degree,
			lng:   -87 * s1.Degree,
			want:  false,
		},
		{
			a:     Point{r3.Vector{-1, -1, 1}},
			b:     Point{r3.Vector{1, -1, 1}},
			latLo: 42 * s1.Degree,
			latHi: 41 * s1.Degree,
			lng:   79 * s1.Degree,
			want:  false,
		},
		{
			a:     Point{r3.Vector{-1, -1, 1}},
			b:     Point{r3.Vector{1, -1, 1}},
			latLo: 41 * s1.Degree,
			latHi: 42 * s1.Degree,
			lng:   87 * s1.Degree,
			want:  false,
		},
		{
			a:     Point{r3.Vector{0, -1, -1}},
			b:     Point{r3.Vector{-1, 0, -1}},
			latLo: -87 * s1.Degree,
			latHi: 13 * s1.Degree,
			lng:   -143 * s1.Degree,
			want:  true,
		},
		{
			a:     Point{r3.Vector{1, 1, -1}},
			b:     Point{r3.Vector{1, -1, 1}},
			latLo: -64 * s1.Degree,
			latHi: 13 * s1.Degree,
			lng:   40 * s1.Degree,
			want:  true,
		},
		{
			a:     Point{r3.Vector{1, 1, 0}},
			b:     Point{r3.Vector{-1, 0, -1}},
			latLo: -64 * s1.Degree,
			latHi: 56 * s1.Degree,
			lng:   151 * s1.Degree,
			want:  true,
		},
		{
			a:     Point{r3.Vector{-1, -1, 0}},
			b:     Point{r3.Vector{1, -1, -1}},
			latLo: -50 * s1.Degree,
			latHi: 18 * s1.Degree,
			lng:   -84 * s1.Degree,
			want:  true,
		},
	}

	for _, test := range tests {
		if got := intersectsLngEdge(test.a, test.b, r1.Interval{float64(test.latLo), float64(test.latHi)}, test.lng); got != test.want {
			t.Errorf("intersectsLngEdge(%v, %v, {%v, %v}, %v) = %v, want %v",
				test.a, test.b, test.latLo, test.latHi, test.lng, got, test.want)
		}
	}
}

// intervalDistance returns the minimum distance (in radians) from X to the latitude
// line segment defined by the given latitude and longitude interval.
func intervalDistance(x LatLng, lat s1.Angle, iv s1.Interval) s1.Angle {
	// Is x inside the longitude interval?
	if iv.Contains(float64(x.Lng)) {
		return s1.Angle(math.Abs(float64(x.Lat - lat)))
	}

	return minAngle(
		x.Distance(LatLng{lat, s1.Angle(iv.Lo)}),
		x.Distance(LatLng{lat, s1.Angle(iv.Hi)}))
}

// Returns the minimum distance from X to the latitude line segment defined by
// the given latitude and longitude interval.
func bruteForceRectLatLngDistance(r Rect, ll LatLng) s1.Angle {
	pt := PointFromLatLng(ll)
	if r.ContainsPoint(pt) {
		return 0
	}

	loLat := intervalDistance(ll, s1.Angle(r.Lat.Lo), r.Lng)
	hiLat := intervalDistance(ll, s1.Angle(r.Lat.Hi), r.Lng)
	loLng := DistanceFromSegment(PointFromLatLng(ll),
		PointFromLatLng(LatLng{s1.Angle(r.Lat.Lo), s1.Angle(r.Lng.Lo)}),
		PointFromLatLng(LatLng{s1.Angle(r.Lat.Hi), s1.Angle(r.Lng.Lo)}))
	hiLng := DistanceFromSegment(PointFromLatLng(ll),
		PointFromLatLng(LatLng{s1.Angle(r.Lat.Lo), s1.Angle(r.Lng.Hi)}),
		PointFromLatLng(LatLng{s1.Angle(r.Lat.Hi), s1.Angle(r.Lng.Hi)}))

	return minAngle(loLat, hiLat, loLng, hiLng)
}

func TestDistanceRectFromLatLng(t *testing.T) {
	// Rect that spans 180.
	a := RectFromLatLng(LatLngFromDegrees(-1, -1)).AddPoint(LatLngFromDegrees(2, 1))
	// Rect near north pole.
	b := RectFromLatLng(LatLngFromDegrees(86, 0)).AddPoint(LatLngFromDegrees(88, 2))
	// Rect that touches north pole.
	c := RectFromLatLng(LatLngFromDegrees(88, 0)).AddPoint(LatLngFromDegrees(90, 2))

	tests := []struct {
		r        Rect
		lat, lng float64 // In degrees.
	}{
		{a, -2, -1},
		{a, 1, 2},
		{b, 87, 3},
		{b, 87, -1},
		{b, 89, 1},
		{b, 89, 181},
		{b, 85, 1},
		{b, 85, 181},
		{b, 90, 0},
		{c, 89, 3},
		{c, 89, 90},
		{c, 89, 181},
	}

	for _, test := range tests {
		ll := LatLngFromDegrees(test.lat, test.lng)
		got := test.r.DistanceToLatLng(ll)
		want := bruteForceRectLatLngDistance(test.r, ll)
		if !float64Near(float64(got), float64(want), 1e-10) {
			t.Errorf("dist from %v to %v = %v, want %v", test.r, ll, got, want)
		}
	}
}

func TestDistanceRectFromLatLngRandomPairs(t *testing.T) {
	latlng := func() LatLng { return LatLngFromPoint(randomPoint()) }

	for i := 0; i < 10000; i++ {
		r := RectFromLatLng(latlng()).AddPoint(latlng())
		ll := latlng()
		got := r.DistanceToLatLng(ll)
		want := bruteForceRectLatLngDistance(r, ll)
		if !float64Near(float64(got), float64(want), 1e-10) {
			t.Errorf("dist from %v to %v = %v, want %v", r, ll, got, want)
		}
	}
}

// This function assumes that DirectedHausdorffDistance() always returns
// a distance from some point in a to b. So the function mainly tests whether
// the returned distance is large enough, and only does a weak test on whether
// it is small enough.
func verifyDirectedHausdorffDistance(t *testing.T, a, b Rect) {
	t.Helper()

	const resolution = 0.1

	// Record the max sample distance as well as the sample point realizing the
	// max for easier debugging.
	var maxDistance s1.Angle

	sampleSizeOnLat := int(a.Lat.Length()/resolution) + 1
	sampleSizeOnLng := int(a.Lng.Length()/resolution) + 1

	deltaOnLat := s1.Angle(a.Lat.Length()) / s1.Angle(sampleSizeOnLat)
	deltaOnLng := s1.Angle(a.Lng.Length()) / s1.Angle(sampleSizeOnLng)

	ll := LatLng{Lng: s1.Angle(a.Lng.Lo)}
	for i := 0; i <= sampleSizeOnLng; i++ {
		ll.Lat = s1.Angle(a.Lat.Lo)

		for j := 0; j <= sampleSizeOnLat; j++ {
			d := b.DistanceToLatLng(ll.Normalized())
			maxDistance = maxAngle(maxDistance, d)
			ll.Lat += deltaOnLat
		}
		ll.Lng += deltaOnLng
	}

	got := a.DirectedHausdorffDistance(b)

	if got < maxDistance-1e-10 {
		t.Errorf("hausdorff(%v, %v) = %v < %v-eps, but shouldn't", a, b, got, maxDistance)
	} else if got > maxDistance+resolution {
		t.Errorf("DirectedHausdorffDistance(%v, %v) = %v > %v+resolution, but shouldn't", a, b, got, maxDistance)
	}
}

func TestRectDirectedHausdorffDistanceRandomPairs(t *testing.T) {
	// Test random pairs.
	rnd := func() LatLng { return LatLngFromPoint(randomPoint()) }
	for i := 0; i < 1000; i++ {
		a := RectFromLatLng(rnd()).AddPoint(rnd())
		b := RectFromLatLng(rnd()).AddPoint(rnd())
		// a and b are *minimum* bounding rectangles of two random points, in
		// particular, their Voronoi diagrams are always of the same topology. We
		// take the "complements" of a and b for more thorough testing.
		a2 := Rect{Lat: a.Lat, Lng: a.Lng.Complement()}
		b2 := Rect{Lat: b.Lat, Lng: b.Lng.Complement()}

		// Note that "a" and "b" come from the same distribution, so there is no
		// need to test pairs such as (b, a), (b, a2), etc.
		verifyDirectedHausdorffDistance(t, a, b)
		verifyDirectedHausdorffDistance(t, a2, b)
		verifyDirectedHausdorffDistance(t, a, b2)
		verifyDirectedHausdorffDistance(t, a2, b2)
	}
}

func TestDirectedHausdorffDistanceContained(t *testing.T) {
	// Caller rect is contained in callee rect. Should return 0.
	a := rectFromDegrees(-10, 20, -5, 90)
	tests := []Rect{
		rectFromDegrees(-10, 20, -5, 90),
		rectFromDegrees(-10, 19, -5, 91),
		rectFromDegrees(-11, 20, -4, 90),
		rectFromDegrees(-11, 19, -4, 91),
	}
	for _, test := range tests {
		got, want := a.DirectedHausdorffDistance(test), s1.Angle(0)
		if got != want {
			t.Errorf("%v.DirectedHausdorffDistance(%v) = %v, want %v", a, test, got, want)
		}
	}
}

func TestDirectHausdorffDistancePointToRect(t *testing.T) {
	// The Hausdorff distance from a point to a rect should be the same as its
	// distance to the rect.
	a1 := LatLngFromDegrees(5, 8)
	a2 := LatLngFromDegrees(90, 10) // North pole.

	tests := []struct {
		ll LatLng
		b  Rect
	}{
		{a1, rectFromDegrees(-85, -50, -80, 10)},
		{a2, rectFromDegrees(-85, -50, -80, 10)},
		{a1, rectFromDegrees(4, -10, 80, 10)},
		{a2, rectFromDegrees(4, -10, 80, 10)},
		{a1, rectFromDegrees(70, 170, 80, -170)},
		{a2, rectFromDegrees(70, 170, 80, -170)},
	}
	for _, test := range tests {
		a := RectFromLatLng(test.ll)
		got, want := a.DirectedHausdorffDistance(test.b), test.b.DistanceToLatLng(test.ll)

		if !float64Eq(float64(got), float64(want)) {
			t.Errorf("hausdorff(%v, %v) = %v, want %v, as that's the closest dist", test.b, a, got, want)
		}
	}
}

func TestDirectedHausdorffDistanceRectToPoint(t *testing.T) {
	a := rectFromDegrees(1, -8, 10, 20)
	tests := []struct {
		lat, lng float64 // Degrees.
	}{{5, 8}, {-6, -100}, {-90, -20}, {90, 0}}
	for _, test := range tests {
		verifyDirectedHausdorffDistance(t, a, RectFromLatLng(LatLngFromDegrees(test.lat, test.lng)))
	}
}

func TestDirectedHausdorffDistanceRectToRectNearPole(t *testing.T) {
	// Tests near south pole.
	a := rectFromDegrees(-87, 0, -85, 3)
	tests := []Rect{
		rectFromDegrees(-89, 1, -88, 2),
		rectFromDegrees(-84, 1, -83, 2),
		rectFromDegrees(-88, 90, -86, 91),
		rectFromDegrees(-84, -91, -83, -90),
		rectFromDegrees(-90, 181, -89, 182),
		rectFromDegrees(-84, 181, -83, 182),
	}
	for _, test := range tests {
		verifyDirectedHausdorffDistance(t, a, test)
	}
}

func TestDirectedHausdorffDistanceRectToRectDegenerateCases(t *testing.T) {
	// Rectangles that contain poles.
	verifyDirectedHausdorffDistance(t,
		rectFromDegrees(0, 10, 90, 20), rectFromDegrees(-4, -10, 4, 0))
	verifyDirectedHausdorffDistance(t,
		rectFromDegrees(-4, -10, 4, 0), rectFromDegrees(0, 10, 90, 20))

	// Two rectangles share same or complement longitudinal intervals.
	a := rectFromDegrees(-50, -10, 50, 10)
	b := rectFromDegrees(30, -10, 60, 10)
	verifyDirectedHausdorffDistance(t, a, b)

	c := Rect{Lat: a.Lat, Lng: a.Lng.Complement()}
	verifyDirectedHausdorffDistance(t, c, b)

	// Rectangle a touches b_opposite_lng.
	verifyDirectedHausdorffDistance(t,
		rectFromDegrees(10, 170, 30, 180), rectFromDegrees(-50, -10, 50, 10))
	verifyDirectedHausdorffDistance(t,
		rectFromDegrees(10, -180, 30, -170), rectFromDegrees(-50, -10, 50, 10))

	// Rectangle b's Voronoi diagram is degenerate (lng interval spans 180
	// degrees), and a touches the degenerate Voronoi vertex.
	verifyDirectedHausdorffDistance(t,
		rectFromDegrees(-30, 170, 30, 180), rectFromDegrees(-10, -90, 10, 90))
	verifyDirectedHausdorffDistance(t,
		rectFromDegrees(-30, -180, 30, -170), rectFromDegrees(-10, -90, 10, 90))

	// Rectangle a touches a voronoi vertex of rectangle b.
	verifyDirectedHausdorffDistance(t,
		rectFromDegrees(-20, 105, 20, 110), rectFromDegrees(-30, 5, 30, 15))
	verifyDirectedHausdorffDistance(t,
		rectFromDegrees(-20, 95, 20, 105), rectFromDegrees(-30, 5, 30, 15))
}

func TestRectApproxEqual(t *testing.T) {
	// s1.Interval and r1.Interval have additional testing.

	const ε = epsilon / 10
	tests := []struct {
		a, b Rect
		want bool
	}{
		{EmptyRect(), rectFromDegrees(1, 5, 1, 5), true},
		{rectFromDegrees(1, 5, 1, 5), EmptyRect(), true},

		{rectFromDegrees(1, 5, 1, 5), rectFromDegrees(2, 7, 2, 7), false},
		{rectFromDegrees(1, 5, 1, 5), rectFromDegrees(1+ε, 5+ε, 1+ε, 5+ε), true},
	}

	for _, test := range tests {
		if got := test.a.ApproxEqual(test.b); got != test.want {
			t.Errorf("%v.ApproxEquals(%v) = %t, want %t", test.a, test.b, got, test.want)
		}
	}
}

func TestRectCentroidEmptyFull(t *testing.T) {
	// Empty and full rectangles.
	if got, want := EmptyRect().Centroid(), (Point{}); !got.ApproxEqual(want) {
		t.Errorf("%v.Centroid() = %v, want %v", EmptyRect(), got, want)
	}
	if got, want := FullRect().Centroid().Norm(), epsilon; !float64Eq(got, want) {
		t.Errorf("%v.Centroid().Norm() = %v, want %v", FullRect(), got, want)
	}
}

func testRectCentroidSplitting(t *testing.T, r Rect, leftSplits int) {
	// Recursively verify that when a rectangle is split into two pieces, the
	// centroids of the children sum to give the centroid of the parent.
	var child0, child1 Rect
	if oneIn(2) {
		lat := randomUniformFloat64(r.Lat.Lo, r.Lat.Hi)
		child0 = Rect{r1.Interval{r.Lat.Lo, lat}, r.Lng}
		child1 = Rect{r1.Interval{lat, r.Lat.Hi}, r.Lng}
	} else {
		lng := randomUniformFloat64(r.Lng.Lo, r.Lng.Hi)
		child0 = Rect{r.Lat, s1.Interval{r.Lng.Lo, lng}}
		child1 = Rect{r.Lat, s1.Interval{lng, r.Lng.Hi}}
	}

	if got, want := r.Centroid().Sub(child0.Centroid().Vector).Sub(child1.Centroid().Vector).Norm(), 1e-15; got > want {
		t.Errorf("%v.Centroid() - %v.Centroid() - %v.Centroid = %v, want ~0", r, child0, child1, got)
	}
	if leftSplits > 0 {
		testRectCentroidSplitting(t, child0, leftSplits-1)
		testRectCentroidSplitting(t, child1, leftSplits-1)
	}
}

func TestRectCentroidFullRange(t *testing.T) {
	// Rectangles that cover the full longitude range.
	for i := 0; i < 100; i++ {
		lat1 := randomUniformFloat64(-math.Pi/2, math.Pi/2)
		lat2 := randomUniformFloat64(-math.Pi/2, math.Pi/2)
		r := Rect{r1.Interval{lat1, lat2}, s1.FullInterval()}
		centroid := r.Centroid()
		if want := 0.5 * (math.Sin(lat1) + math.Sin(lat2)) * r.Area(); !float64Near(want, centroid.Z, epsilon) {
			t.Errorf("%v.Centroid().Z was %v, want %v", r, centroid.Z, want)
		}
		if got := (r2.Point{centroid.X, centroid.Y}.Norm()); got > epsilon {
			t.Errorf("%v.Centroid().Norm() was %v, want > %v ", r, got, epsilon)
		}
	}

	// Rectangles that cover the full latitude range.
	for i := 0; i < 100; i++ {
		lat1 := randomUniformFloat64(-math.Pi, math.Pi)
		lat2 := randomUniformFloat64(-math.Pi, math.Pi)
		r := Rect{r1.Interval{-math.Pi / 2, math.Pi / 2}, s1.Interval{lat1, lat2}}
		centroid := r.Centroid()

		if got, want := math.Abs(centroid.Z), epsilon; got > want {
			t.Errorf("math.Abs(%v.Centroid().Z) = %v, want <= %v", r, got, want)
		}

		if got, want := LatLngFromPoint(centroid).Lng.Radians(), r.Lng.Center(); !float64Near(got, want, epsilon) {
			t.Errorf("%v.Lng.Radians() = %v, want %v", centroid, got, want)
		}

		alpha := 0.5 * r.Lng.Length()
		if got, want := (r2.Point{centroid.X, centroid.Y}.Norm()), (0.25 * math.Pi * math.Sin(alpha) / alpha * r.Area()); !float64Near(got, want, epsilon) {
			t.Errorf("%v.Centroid().Norm() = %v, want ~%v", got, want, epsilon)
		}
	}

	// Finally, verify that when a rectangle is recursively split into pieces,
	// the centroids of the pieces add to give the centroid of their parent.
	// To make the code simpler we avoid rectangles that cross the 180 degree
	// line of longitude.
	testRectCentroidSplitting(t, Rect{r1.Interval{-math.Pi / 2, math.Pi / 2}, s1.Interval{-math.Pi, math.Pi}}, 10)
}
