/*
Copyright 2016 Google Inc. All rights reserved.

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
	"fmt"
	"math"
	"strconv"
	"strings"
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

	if got, want := shape.numChains(), 1; got != want {
		t.Errorf("%v.numChains() = %d, want %d", shape, got, want)
	}
	if got, want := shape.chainStart(0), 0; got != want {
		t.Errorf("%v.chainStart(0) = %d, want %d", shape, got, want)
	}
	if got, want := shape.chainStart(1), 3; got != want {
		t.Errorf("%v.chainStart(1) = %d, want %d", shape, got, want)
	}

	v2, v3 := shape.Edge(2)
	if want := PointFromLatLng(LatLngFromDegrees(1, 1)); !v2.ApproxEqual(want) {
		t.Errorf("%v.Edge(%d) point A = %v  want %v", shape, 2, v2, want)
	}
	if want := PointFromLatLng(LatLngFromDegrees(2, 1)); !v3.ApproxEqual(want) {
		t.Errorf("%v.Edge(%d) point B = %v  want %v", shape, 2, v3, want)
	}

	if shape.HasInterior() {
		t.Errorf("polylines should not have an interior")
	}
	if shape.ContainsOrigin() {
		t.Errorf("polylines should not contain the origin")
	}

	if shape.dimension() != polylineGeometry {
		t.Errorf("polylines should have PolylineGeometry")
	}

	empty := &Polyline{}
	if got, want := empty.NumEdges(), 0; got != want {
		t.Errorf("%v.NumEdges() = %d, want %d", empty, got, want)
	}
	if got, want := empty.numChains(), 0; got != want {
		t.Errorf("%v.numChains() = %d, want %d", empty, got, want)
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

func makePolylineFromString(input string) (*Polyline, error) {
	var latLngs []LatLng
	if len(input) > 0 {
		latLngStrs := strings.Split(input, ", ")
		for _, latLngStr := range latLngStrs {
			pair := strings.Split(latLngStr, ":")
			if len(pair) != 2 {
				return nil, fmt.Errorf("bad polyline string: \"%s\" \"%s\"", latLngStr, pair)
			}

			lat, err := strconv.ParseFloat(pair[0], 64)
			if err != nil {
				return nil, err
			}

			lng, err := strconv.ParseFloat(pair[1], 64)
			if err != nil {
				return nil, err
			}

			latLngs = append(latLngs, LatLngFromDegrees(lat, lng))
		}
	}
	return PolylineFromLatLngs(latLngs), nil
}

func checkSubsample(t *testing.T, inputStr string, tolerance s1.Angle, expectedStr string) {
	polyline, err := makePolylineFromString(inputStr)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := makePolylineFromString(expectedStr)
	if err != nil {
		t.Fatal(err)
	}
	actual := polyline.SubsampleVertices(tolerance)
	if len(*expected) != len(*actual) {
		t.Errorf("polylines not equal: expected=%#v, actual=%#v", expected, actual)
	}

	for i, pExpected := range *expected {
		pActual := (*actual)[i]
		if pActual != pExpected {
			t.Errorf("polylines not equal: expected=%#v, actual=%#v", expected, actual)
		}
	}
}

func TestPolylineSubsampleVerticesTrivialInputs(t *testing.T) {
	// No vertices.
	checkSubsample(t, "", s1.Degree*1.0, "")
	// One vertex.
	checkSubsample(t, "0:1", s1.Degree*1.0, "0:1")
	// Two vertices.
	checkSubsample(t, "10:10, 11:11", s1.Degree*5.0, "10:10, 11:11")
	// Three points on a straight line.
	// In theory, zero tolerance should work, but in practice there are floating
	// point errors.
	checkSubsample(t, "-1:0, 0:0, 1:0", s1.Degree*1e-15, "-1:0, 1:0")
	// Zero tolerance on a non-straight line.
	checkSubsample(t, "-1:0, 0:0, 1:1", s1.Degree*0.0, "-1:0, 0:0, 1:1")
	// Negative tolerance should return all vertices.
	checkSubsample(t, "-1:0, 0:0, 1:1", s1.Degree*-1.0, "-1:0, 0:0, 1:1")
	// Non-zero tolerance with a straight line.
	checkSubsample(t, "0:1, 0:2, 0:3, 0:4, 0:5", s1.Degree*1.0, "0:1, 0:5")
	// And finally, verify that we still do something reasonable if the client
	// passes in an invalid polyline with two or more adjacent vertices.
	checkSubsample(t, "0:1, 0:1, 0:1, 0:2", s1.Degree*0.0, "0:1, 0:2")
}

func TestPolylineSubsampleVerticesSimpleExample(t *testing.T) {
	polyStr := "0:0, 0:1, -1:2, 0:3, 0:4, 1:4, 2:4.5, 3:4, 3.5:4, 4:4"
	checkSubsample(t, polyStr, s1.Degree*3.0, "0:0, 4:4")
	checkSubsample(t, polyStr, s1.Degree*2.0, "0:0, 2:4.5, 4:4")
	checkSubsample(t, polyStr, s1.Degree*0.9, "0:0, -1:2, 2:4.5, 4:4")
	checkSubsample(t, polyStr, s1.Degree*0.4, "0:0, 0:1, -1:2, 0:3, 0:4, 2:4.5, 4:4")
	checkSubsample(t, polyStr, s1.Degree*0, "0:0, 0:1, -1:2, 0:3, 0:4, 1:4, 2:4.5, 3:4, 3.5:4, 4:4")
}

func TestPolylineSubsampleVerticesGuarantees(t *testing.T) {
	// Check that duplicate vertices are never generated.
	checkSubsample(t, "10:10, 12:12, 10:10", s1.Degree*5.0, "10:10")
	checkSubsample(t, "0:0, 1:1, 0:0, 0:120, 0:130", s1.Degree*5.0, "0:0, 0:120, 0:130")

	// Check that points are not collapsed if they would create a line
	// segment longer than 90 degrees, and also that the code handles
	// original polyline segments longer than 90 degrees.
	checkSubsample(t, "90:0, 50:180, 20:180, -20:180, -50:180, -90:0, 30:0, 90:0",
		s1.Degree*5.0, "90:0, 20:180, -50:180, -90:0, 30:0, 90:0")

	// Check that the output polyline is parametrically equivalent and
	// not just geometrically equivalent, i.e. that backtracking is
	// preserved.  The algorithm achieves this by requiring that the
	// points must be encountered in increasing order of distance
	// along each output segment, except for points that are within
	// "tolerance" of the first vertex of each segment.

	checkSubsample(t, "10:10, 10:20, 10:30, 10:15, 10:40", s1.Degree*5.0, "10:10, 10:30, 10:15, 10:40")
	checkSubsample(t, "10:10, 10:20, 10:30, 10:10, 10:30, 10:40", s1.Degree*5.0, "10:10, 10:30, 10:10, 10:40")
	checkSubsample(t, "10:10, 12:12, 9:9, 10:20, 10:30", s1.Degree*5.0, "10:10, 10:30")
}
