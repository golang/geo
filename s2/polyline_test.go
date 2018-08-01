// Copyright 2016 Google Inc. All rights reserved.
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
	"reflect"
	"testing"

	"github.com/golang/geo/r3"
	"github.com/golang/geo/s1"
)

func TestPolylineBasics(t *testing.T) {
	empty := Polyline{}
	if empty.RectBound() != EmptyRect() {
		t.Errorf("empty.RectBound() = %v, want %v", empty.RectBound(), EmptyRect())
	}
	if len(empty) != 0 {
		t.Errorf("empty Polyline should have no vertices")
	}
	empty.Reverse()
	if len(empty) != 0 {
		t.Errorf("reveresed empty Polyline should have no vertices")
	}

	latlngs := []LatLng{
		LatLngFromDegrees(0, 0),
		LatLngFromDegrees(0, 90),
		LatLngFromDegrees(0, 180),
	}

	semiEquator := PolylineFromLatLngs(latlngs)
	//if got, want := semiEquator.Interpolate(0.5), Point{r3.Vector{0, 1, 0}}; !got.ApproxEqual(want) {
	//	t.Errorf("semiEquator.Interpolate(0.5) = %v, want %v", got, want)
	//}
	semiEquator.Reverse()
	if got, want := (*semiEquator)[2], (Point{r3.Vector{1, 0, 0}}); !got.ApproxEqual(want) {
		t.Errorf("semiEquator[2] = %v, want %v", got, want)
	}
}

func TestPolylineShape(t *testing.T) {
	var shape Shape = makePolyline("0:0, 1:0, 1:1, 2:1")
	if got, want := shape.NumEdges(), 3; got != want {
		t.Errorf("%v.NumEdges() = %v, want %d", shape, got, want)
	}
	if got, want := shape.NumChains(), 1; got != want {
		t.Errorf("%v.NumChains() = %d, want %d", shape, got, want)
	}
	if got, want := shape.Chain(0).Start, 0; got != want {
		t.Errorf("%v.Chain(0).Start = %d, want %d", shape, got, want)
	}
	if got, want := shape.Chain(0).Length, 3; got != want {
		t.Errorf("%v.Chain(0).Length = %d, want %d", shape, got, want)
	}

	e := shape.Edge(2)
	if want := PointFromLatLng(LatLngFromDegrees(1, 1)); !e.V0.ApproxEqual(want) {
		t.Errorf("%v.Edge(%d) point A = %v  want %v", shape, 2, e.V0, want)
	}
	if want := PointFromLatLng(LatLngFromDegrees(2, 1)); !e.V1.ApproxEqual(want) {
		t.Errorf("%v.Edge(%d) point B = %v  want %v", shape, 2, e.V1, want)
	}
	if shape.ReferencePoint().Contained {
		t.Errorf("polylines should not contain their reference points")
	}
	if got, want := shape.Dimension(), 1; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
}

func TestPolylineEmpty(t *testing.T) {
	shape := &Polyline{}
	if got, want := shape.NumEdges(), 0; got != want {
		t.Errorf("%v.NumEdges() = %d, want %d", shape, got, want)
	}
	if got, want := shape.NumChains(), 0; got != want {
		t.Errorf("%v.NumChains() = %d, want %d", shape, got, want)
	}
	if !shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = false, want true")
	}
	if shape.IsFull() {
		t.Errorf("shape.IsFull() = true, want false")
	}
	if shape.ReferencePoint().Contained {
		t.Errorf("polylines should not contain their reference points")
	}
}

func TestPolylineLengthAndCentroid(t *testing.T) {
	// Construct random great circles and divide them randomly into segments.
	// Then make sure that the length and centroid are correct.  Note that
	// because of the way the centroid is computed, it does not matter how
	// we split the great circle into segments.

	for i := 0; i < 100; i++ {
		// Choose a coordinate frame for the great circle.
		f := randomFrame()

		var line Polyline
		for theta := 0.0; theta < 2*math.Pi; theta += math.Pow(randomFloat64(), 10) {
			p := Point{f.row(0).Mul(math.Cos(theta)).Add(f.row(1).Mul(math.Sin(theta)))}
			if len(line) == 0 || !p.ApproxEqual(line[len(line)-1]) {
				line = append(line, p)
			}
		}

		// Close the circle.
		line = append(line, line[0])

		length := line.Length()
		if got, want := math.Abs(length.Radians()-2*math.Pi), 2e-14; got > want {
			t.Errorf("%v.Length() = %v, want < %v", line, got, want)
		}

		centroid := line.Centroid()
		if got, want := centroid.Norm(), 2e-14; got > want {
			t.Errorf("%v.Norm() = %v, want < %v", centroid, got, want)
		}
	}
}

func TestPolylineIntersectsCell(t *testing.T) {
	pline := Polyline{
		Point{r3.Vector{1, -1.1, 0.8}.Normalize()},
		Point{r3.Vector{1, -0.8, 1.1}.Normalize()},
	}

	for face := 0; face < 6; face++ {
		cell := CellFromCellID(CellIDFromFace(face))
		if got, want := pline.IntersectsCell(cell), face&1 == 0; got != want {
			t.Errorf("%v.IntersectsCell(%v) = %v, want %v", pline, cell, got, want)
		}
	}
}

func TestPolylineSubsample(t *testing.T) {
	const polyStr = "0:0, 0:1, -1:2, 0:3, 0:4, 1:4, 2:4.5, 3:4, 3.5:4, 4:4"

	tests := []struct {
		have      string
		tolerance float64
		want      []int
	}{
		{
			// No vertices.
			tolerance: 1.0,
		},
		{
			// One vertex.
			have:      "0:1",
			tolerance: 1.0,
			want:      []int{0},
		},
		{
			// Two vertices.
			have:      "10:10, 11:11",
			tolerance: 5.0,
			want:      []int{0, 1},
		},
		{
			// Three points on a straight line. In theory, zero tolerance
			// should work, but in practice there are floating point errors.
			have:      "-1:0, 0:0, 1:0",
			tolerance: 1e-15,
			want:      []int{0, 2},
		},
		{
			// Zero tolerance on a non-straight line.
			have:      "-1:0, 0:0, 1:1",
			tolerance: 0.0,
			want:      []int{0, 1, 2},
		},
		{
			// Negative tolerance should return all vertices.
			have:      "-1:0, 0:0, 1:1",
			tolerance: -1.0,
			want:      []int{0, 1, 2},
		},
		{
			// Non-zero tolerance with a straight line.
			have:      "0:1, 0:2, 0:3, 0:4, 0:5",
			tolerance: 1.0,
			want:      []int{0, 4},
		},
		{
			// And finally, verify that we still do something
			// reasonable if the client passes in an invalid polyline
			// with two or more adjacent vertices.
			have:      "0:1, 0:1, 0:1, 0:2",
			tolerance: 0.0,
			want:      []int{0, 3},
		},

		// Simple examples
		{
			have:      polyStr,
			tolerance: 3.0,
			want:      []int{0, 9},
		},
		{
			have:      polyStr,
			tolerance: 2.0,
			want:      []int{0, 6, 9},
		},
		{
			have:      polyStr,
			tolerance: 0.9,
			want:      []int{0, 2, 6, 9},
		},
		{
			have:      polyStr,
			tolerance: 0.4,
			want:      []int{0, 1, 2, 3, 4, 6, 9},
		},
		{
			have:      polyStr,
			tolerance: 0,
			want:      []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},

		// Check that duplicate vertices are never generated.
		{
			have:      "10:10, 12:12, 10:10",
			tolerance: 5.0,
			want:      []int{0},
		},
		{
			have:      "0:0, 1:1, 0:0, 0:120, 0:130",
			tolerance: 5.0,
			want:      []int{0, 3, 4},
		},

		// Check that points are not collapsed if they would create a line segment
		// longer than 90 degrees, and also that the code handles original polyline
		// segments longer than 90 degrees.
		{
			have:      "90:0, 50:180, 20:180, -20:180, -50:180, -90:0, 30:0, 90:0",
			tolerance: 5.0,
			want:      []int{0, 2, 4, 5, 6, 7},
		},

		// Check that the output polyline is parametrically equivalent and not just
		// geometrically equivalent, i.e. that backtracking is preserved.  The
		// algorithm achieves this by requiring that the points must be encountered
		// in increasing order of distance along each output segment, except for
		// points that are within "tolerance" of the first vertex of each segment.
		{
			have:      "10:10, 10:20, 10:30, 10:15, 10:40",
			tolerance: 5.0,
			want:      []int{0, 2, 3, 4},
		},
		{
			have:      "10:10, 10:20, 10:30, 10:10, 10:30, 10:40",
			tolerance: 5.0,
			want:      []int{0, 2, 3, 5},
		},
		{
			have:      "10:10, 12:12, 9:9, 10:20, 10:30",
			tolerance: 5.0,
			want:      []int{0, 4},
		},
	}

	for _, test := range tests {
		p := makePolyline(test.have)
		got := p.SubsampleVertices(s1.Angle(test.tolerance) * s1.Degree)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%q.SubsampleVertices(%v°) = %v, want %v", test.have, test.tolerance, got, test.want)
		}
	}
}

func TestProject(t *testing.T) {
	latlngs := []LatLng{
		LatLngFromDegrees(0, 0),
		LatLngFromDegrees(0, 1),
		LatLngFromDegrees(0, 2),
		LatLngFromDegrees(1, 2),
	}
	line := PolylineFromLatLngs(latlngs)
	tests := []struct {
		haveLatLng     LatLng
		wantProjection LatLng
		wantNext       int
	}{
		{LatLngFromDegrees(0.5, -0.5), LatLngFromDegrees(0, 0), 1},
		{LatLngFromDegrees(0.5, 0.5), LatLngFromDegrees(0, 0.5), 1},
		{LatLngFromDegrees(0.5, 1), LatLngFromDegrees(0, 1), 2},
		{LatLngFromDegrees(-0.5, 2.5), LatLngFromDegrees(0, 2), 3},
		{LatLngFromDegrees(2, 2), LatLngFromDegrees(1, 2), 4},
	}
	for _, test := range tests {
		projection, next := line.Project(PointFromLatLng(test.haveLatLng))
		if !PointFromLatLng(test.wantProjection).ApproxEqual(projection) {
			t.Errorf("line.Project(%v) = %v, want %v", test.haveLatLng, projection, test.wantProjection)
		}
		if next != test.wantNext {
			t.Errorf("line.Project(%v) = %v, want %v", test.haveLatLng, next, test.wantNext)
		}
	}
}

func TestIsOnRight(t *testing.T) {
	latlngs := []LatLng{
		LatLngFromDegrees(0, 0),
		LatLngFromDegrees(0, 1),
		LatLngFromDegrees(0, 2),
		LatLngFromDegrees(1, 2),
	}
	line1 := PolylineFromLatLngs(latlngs)
	latlngs = []LatLng{
		LatLngFromDegrees(0, 0),
		LatLngFromDegrees(0, 1),
		LatLngFromDegrees(-1, 0),
	}
	line2 := PolylineFromLatLngs(latlngs)
	tests := []struct {
		line        *Polyline
		haveLatLng  LatLng
		wantOnRight bool
	}{
		{line1, LatLngFromDegrees(-0.5, 0.5), true},
		{line1, LatLngFromDegrees(0.5, -0.5), false},
		{line1, LatLngFromDegrees(0.5, 0.5), false},
		{line1, LatLngFromDegrees(0.5, 1.0), false},
		{line1, LatLngFromDegrees(-0.5, 2.5), true},
		{line1, LatLngFromDegrees(1.5, 2.5), true},
		// Explicitly test the case where the closest point is an interior vertex.
		// The points are chosen such that they are on different sides of the two
		// edges that the interior vertex is on.
		{line2, LatLngFromDegrees(-0.5, 5.0), false},
		{line2, LatLngFromDegrees(5.5, 5.0), false},
	}
	for _, test := range tests {
		onRight := test.line.IsOnRight(PointFromLatLng(test.haveLatLng))
		if onRight != test.wantOnRight {
			t.Errorf("line.IsOnRight(%v) = %v, want %v", test.haveLatLng, onRight, test.wantOnRight)
		}
	}
}

func TestPolylineValidate(t *testing.T) {
	p := makePolyline("0:0, 2:1, 0:2, 2:3, 0:4, 2:5, 0:6")
	if p.Validate() != nil {
		t.Errorf("valid polygon should Validate")
	}

	p1 := Polyline([]Point{
		PointFromCoords(0, 1, 0),
		Point{r3.Vector{10, 3, 7}},
		PointFromCoords(0, 0, 1),
	})

	if err := p1.Validate(); err == nil {
		t.Errorf("polyline with non-normalized points shouldn't validate")
	}

	// adjacent equal vertices
	p2 := Polyline([]Point{
		PointFromCoords(0, 1, 0),
		PointFromCoords(0, 0, 1),
		PointFromCoords(0, 0, 1),
		PointFromCoords(1, 0, 0),
	})

	if err := p2.Validate(); err == nil {
		t.Errorf("polyline with repeated vertices shouldn't validate")
	}

	pt := PointFromCoords(1, 1, 0)
	antiPt := Point{pt.Mul(-1)}

	// non-adjacent antipodal points.
	p3 := Polyline([]Point{
		PointFromCoords(0, 1, 0),
		pt,
		PointFromCoords(0, 0, 1),
		antiPt,
	})

	if err := p3.Validate(); err != nil {
		t.Errorf("polyline with non-adjacent antipodal points should validate")
	}

	// non-adjacent antipodal points.
	p4 := Polyline([]Point{
		PointFromCoords(0, 1, 0),
		PointFromCoords(0, 0, 1),
		pt,
		antiPt,
	})

	if err := p4.Validate(); err == nil {
		t.Errorf("polyline with antipodal adjacent points shouldn't validate")
	}
}

// TODO(roberts): Test differences from C++:
// Interpolate
// UnInterpolate
// IntersectsEmptyPolyline
// IntersectsOnePointPolyline
// Intersects
// IntersectsAtVertex
// IntersectsVertexOnEdge
// ApproxEquals
//
// PolylineCoveringTest
//    PolylineOverlapsSelf
//    PolylineDoesNotOverlapReverse
//    PolylineOverlapsEquivalent
//    ShortCoveredByLong
//    PartialOverlapOnly
//    ShortBacktracking
//    LongBacktracking
//    IsResilientToDuplicatePoints
//    CanChooseBetweenTwoPotentialStartingPoints
//    StraightAndWigglyPolylinesCoverEachOther
//    MatchStartsAtLastVertex
//    MatchStartsAtDuplicatedLastVertex
//    EmptyPolylines
