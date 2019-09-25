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
	want := PointFromCoords(0, 1, 0)
	if got, _ := semiEquator.Interpolate(0.5); !got.ApproxEqual(want) {
		t.Errorf("semiEquator.Interpolate(0.5) = %v, want %v", got, want)
	}
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
		{LatLngFromDegrees(-50, 0.5), LatLngFromDegrees(0, 0.5), 1},
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
		{r3.Vector{10, 3, 7}},
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

func TestPolylineIntersects(t *testing.T) {
	// PolylineInterectsEmpty
	empty := Polyline{}
	line := makePolyline("1:1, 4:4")
	if empty.Intersects(line) {
		t.Errorf("%v.Intersects(%v) = true, want false", empty, line)
	}

	// PolylineIntersectsOnePointPolyline
	line1 := makePolyline("1:1, 4:4")
	line2 := makePolyline("1:1")
	if line1.Intersects(line2) {
		t.Errorf("%v.Intersects(%v) = true, want false", line1, line2)
	}

	// PolylineIntersects
	line3 := makePolyline("1:1, 4:4")
	smallCrossing := makePolyline("1:2, 2:1")
	smallNoncrossing := makePolyline("1:2, 2:3")
	bigCrossing := makePolyline("1:2, 2:3, 4:3")
	if !line3.Intersects(smallCrossing) {
		t.Errorf("%v.Intersects(%v) = false, want true", line3, smallCrossing)
	}
	if line3.Intersects(smallNoncrossing) {
		t.Errorf("%v.Intersects(%v) = true, want false", line3, smallNoncrossing)
	}
	if !line3.Intersects(bigCrossing) {
		t.Errorf("%v.Intersects(%v) = false, want true", line3, bigCrossing)
	}

	// PolylineIntersectsAtVertex
	line4 := makePolyline("1:1, 4:4, 4:6")
	line5 := makePolyline("1:1, 1:2")
	line6 := makePolyline("5:1, 4:4, 2:2")
	if !line4.Intersects(line5) {
		t.Errorf("%v.Intersects(%v) = false, want true", line4, line5)
	}
	if !line4.Intersects(line6) {
		t.Errorf("%v.Intersects(%v) = false, want true", line4, line6)
	}

	// TestPolylineIntersectsVertexOnEdge
	horizontalLeftToRight := makePolyline("0:1, 0:3")
	verticalBottomToTop := makePolyline("-1:2, 0:2, 1:2")
	horizontalRightToLeft := makePolyline("0:3, 0:1")
	verticalTopToBottom := makePolyline("1:2, 0:2, -1:2")
	if !horizontalLeftToRight.Intersects(verticalBottomToTop) {
		t.Errorf("%v.Intersects(%v) = false, want true", horizontalLeftToRight, verticalBottomToTop)
	}
	if !horizontalLeftToRight.Intersects(verticalTopToBottom) {
		t.Errorf("%v.Intersects(%v) = false, want true", horizontalLeftToRight, verticalTopToBottom)
	}
	if !horizontalRightToLeft.Intersects(verticalBottomToTop) {
		t.Errorf("%v.Intersects(%v) = false, want true", horizontalRightToLeft, verticalBottomToTop)
	}
	if !horizontalRightToLeft.Intersects(verticalTopToBottom) {
		t.Errorf("%v.Intersects(%v) = false, want true", horizontalRightToLeft, verticalTopToBottom)
	}
}

func TestPolylineApproxEqual(t *testing.T) {
	degree := s1.Angle(1) * s1.Degree

	tests := []struct {
		a, b     string
		maxError s1.Angle
		want     bool
	}{
		{
			// Close lines, differences within maxError.
			a:        "0:0, 0:10, 5:5",
			b:        "0:0.1, -0.1:9.9, 5:5.2",
			maxError: 0.5 * degree,
			want:     true,
		},
		{
			// Close lines, differences outside maxError.
			a:        "0:0, 0:10, 5:5",
			b:        "0:0.1, -0.1:9.9, 5:5.2",
			maxError: 0.01 * degree,
			want:     false,
		},
		{
			// Same line, but different number of vertices.
			a:        "0:0, 0:10, 0:20",
			b:        "0:0, 0:20",
			maxError: 0.1 * degree,
			want:     false,
		},
		{
			// Same vertices, in different order.
			a:        "0:0, 5:5, 0:10",
			b:        "5:5, 0:10, 0:0",
			maxError: 0.1 * degree,
			want:     false,
		},
	}
	for _, test := range tests {
		a := makePolyline(test.a)
		b := makePolyline(test.b)
		if got := a.approxEqual(b, test.maxError); got != test.want {
			t.Errorf("%v.approxEqual(%v, %v) = %v, want %v", a, b, test.maxError, got, test.want)
		}
	}
}

func TestPolylineInterpolate(t *testing.T) {
	vertices := []Point{PointFromCoords(1, 0, 0),
		PointFromCoords(0, 1, 0),
		PointFromCoords(0, 1, 1),
		PointFromCoords(0, 0, 1),
	}
	line := Polyline(vertices)

	want := vertices[0]
	point, next := line.Interpolate(-0.1)
	if point != vertices[0] {
		t.Errorf("%v.Interpolate(%v) = %v, want %v", line, -0.1, point, vertices[0])
	}
	if next != 1 {
		t.Errorf("%v.Interpolate(%v) = %v, want %v", line, -0.1, next, 1)
	}

	want = PointFromCoords(1, math.Tan(0.2*math.Pi/2.0), 0)
	if got, _ := line.Interpolate(0.1); !got.ApproxEqual(want) {
		t.Errorf("%v.Interpolate(%v) = %v, want %v", line, 0.1, got, want)
	}

	want = PointFromCoords(1, 1, 0)
	if got, _ := line.Interpolate(0.25); !got.ApproxEqual(want) {
		t.Errorf("%v.Interpolate(%v) = %v, want %v", line, 0.25, got, want)
	}

	want = vertices[1]
	if got, _ := line.Interpolate(0.5); got != want {
		t.Errorf("%v.Interpolate(%v) = %v, want %v", line, 0.5, got, want)
	}

	want = vertices[2]
	point, next = line.Interpolate(0.75)
	if !point.ApproxEqual(want) {
		t.Errorf("%v.Interpolate(%v) = %v, want %v", line, 0.75, point, want)
	}
	if next != 3 {
		t.Errorf("%v.Interpolate(%v) = %v, want %v", line, 0.75, next, 3)
	}

	point, next = line.Interpolate(1.1)
	if point != vertices[3] {
		t.Errorf("%v.Interpolate(%v) = %v, want %v", line, 1.1, point, vertices[3])
	}
	if next != 4 {
		t.Errorf("%v.Interpolate(%v) = %v, want %v", line, 1.1, next, 4)
	}

	// Check the case where the interpolation fraction is so close to 1 that
	// the interpolated point is identical to the last vertex.
	vertices2 := []Point{PointFromCoords(1, 1, 1),
		PointFromCoords(1, 1, 1+1e-15),
		PointFromCoords(1, 1, 1+2e-15),
	}
	shortLine := Polyline(vertices2)

	point, next = shortLine.Interpolate(1.0 - 2e-16)
	if point != vertices2[2] {
		t.Errorf("%v.Interpolate(%v) = %v, want %v", shortLine, 1.0-2e-16, point, vertices2[2])
	}
	if next != 3 {
		t.Errorf("%v.Interpolate(%v) = %v, want %v", shortLine, 1.0-2e-16, next, 3)
	}
}

func TestPolylineUninterpolate(t *testing.T) {
	vertices := []Point{PointFromCoords(1, 0, 0)}
	line := Polyline(vertices)
	if got, want := line.Uninterpolate(PointFromCoords(0, 1, 0), 1), 0.0; !float64Eq(got, want) {
		t.Errorf("Uninterpolate on a polyline with 2 or fewer vertices should return 0, got %v", got)
	}

	vertices = append(vertices,
		PointFromCoords(0, 1, 0),
		PointFromCoords(0, 1, 1),
		PointFromCoords(0, 0, 1),
	)
	line = Polyline(vertices)

	interpolated, nextVertex := line.Interpolate(-0.1)
	if got, want := line.Uninterpolate(interpolated, nextVertex), 0.0; !float64Eq(got, want) {
		t.Errorf("line.Uninterpolate(%v, %d) = %v, want %v", interpolated, nextVertex, got, want)
	}
	interpolated, nextVertex = line.Interpolate(0.0)
	if got, want := line.Uninterpolate(interpolated, nextVertex), 0.0; !float64Eq(got, want) {
		t.Errorf("line.Uninterpolate(%v, %d) = %v, want %v", interpolated, nextVertex, got, want)
	}
	interpolated, nextVertex = line.Interpolate(0.5)
	if got, want := line.Uninterpolate(interpolated, nextVertex), 0.5; !float64Eq(got, want) {
		t.Errorf("line.Uninterpolate(%v, %d) = %v, want %v", interpolated, nextVertex, got, want)
	}
	interpolated, nextVertex = line.Interpolate(0.75)
	if got, want := line.Uninterpolate(interpolated, nextVertex), 0.75; !float64Eq(got, want) {
		t.Errorf("line.Uninterpolate(%v, %d) = %v, want %v", interpolated, nextVertex, got, want)
	}
	interpolated, nextVertex = line.Interpolate(1.1)
	if got, want := line.Uninterpolate(interpolated, nextVertex), 1.0; !float64Eq(got, want) {
		t.Errorf("line.Uninterpolate(%v, %d) = %v, want %v", interpolated, nextVertex, got, want)
	}

	// Check that the return value is clamped to 1.0.
	if got, want := line.Uninterpolate(PointFromCoords(0, 1, 0), len(line)), 1.0; !float64Eq(got, want) {
		t.Errorf("line.Uninterpolate(%v, %d) = %v, want %v", PointFromCoords(0, 1, 0), len(line), got, want)
	}
}

// TODO(roberts): Test differences from C++:
// InitToSnapped
// InitToSimplified
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
