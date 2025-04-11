// Copyright 2017 Google Inc. All rights reserved.
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
	"bytes"
	"testing"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r3"
	"github.com/golang/geo/s1"
)

func TestParseLatLng(t *testing.T) {
	tests := []struct {
		have string
		want []LatLng
	}{
		{
			have: "",
			want: []LatLng{},
		},
		{
			have: "blah",
			want: []LatLng{},
		},
		{
			have: "0:0",
			want: []LatLng{{Lat: 0, Lng: 0}},
		},
		{
			have: "0:0, 0:-90",
			want: []LatLng{
				LatLngFromDegrees(0, 0),
				LatLngFromDegrees(0, -90),
			},
		},
		{
			have: "-20:150, -20:151, -19:150",
			want: []LatLng{
				LatLngFromDegrees(-20, 150),
				LatLngFromDegrees(-20, 151),
				LatLngFromDegrees(-19, 150),
			},
		},
	}

	for _, test := range tests {
		got := parseLatLngs(test.have)

		if len(got) != len(test.want) {
			t.Errorf("parseLatLngs(%s) = %+v, got different number of results %+v", test.have, got, test.want)
			continue
		}

		for k, v := range got {
			if v != test.want[k] {
				t.Errorf("parseLatlng(%q): %d. %+v, want %+v", test.have, k, v, test.want[k])
			}
		}
	}
}

func TestTextFormatWritePoints(t *testing.T) {
	tests := []struct {
		have      []Point
		roundtrip bool
		want      string
	}{
		{
			have:      nil,
			roundtrip: false,
			want:      "",
		},
		{
			have:      []Point{},
			roundtrip: false,
			want:      "",
		},
		{
			have:      []Point{PointFromCoords(1, 0, 0)},
			roundtrip: false,
			want:      "0:0",
		},
		{
			have:      []Point{PointFromCoords(1, 0, 0), PointFromCoords(0, -1, 0)},
			roundtrip: false,
			want:      "0:0, 0:-90",
		},
		{
			// test without roundtrip precision.
			have:      []Point{PointFromCoords(1, 6.02e-23, 0)},
			roundtrip: false,
			want:      "0:3.44920592668756e-21",
		},
		{
			// test with roundtrip precision to get 2 extra digits.
			have:      []Point{PointFromCoords(1, 6.02e-23, 0)},
			roundtrip: true,
			want:      "0:3.4492059266875556e-21",
		},
	}

	for _, test := range tests {
		var buf bytes.Buffer
		writePoints(&buf, test.have, test.roundtrip)
		if got := buf.String(); got != test.want {
			t.Errorf("writePoints(%v) = %q, want %q", test.have, got, test.want)
		}
	}
}

func TestTextFormatParsePointRoundtrip(t *testing.T) {
	tests := []struct {
		have string
		want Point
	}{
		{"0:0", Point{r3.Vector{X: 1, Y: 0, Z: 0}}},
		{"90:0", Point{r3.Vector{X: 6.123233995736757e-17, Y: 0, Z: 1}}},
		{"-45:0", Point{r3.Vector{X: 0.7071067811865476, Y: 0, Z: -0.7071067811865475}}},
		{"0:0.01", Point{r3.Vector{X: 0.9999999847691292, Y: 0.00017453292431333684, Z: 0}}},
		{"0:30", Point{r3.Vector{X: 0.8660254037844387, Y: 0.49999999999999994, Z: 0}}},
		{"0:45", Point{r3.Vector{X: 0.7071067811865476, Y: 0.7071067811865475, Z: 0}}},
		{"0:90", Point{r3.Vector{X: 6.123233995736757e-17, Y: 1, Z: 0}}},
		{"30:30", Point{r3.Vector{X: 0.7500000000000001, Y: 0.4330127018922193, Z: 0.49999999999999994}}},
		{"-30:30", Point{r3.Vector{X: 0.7500000000000001, Y: 0.4330127018922193, Z: -0.49999999999999994}}},
		{"0:180", Point{r3.Vector{X: -1, Y: 6.123233995736757e-17, Z: 0}}},
		{"0:-180", Point{r3.Vector{X: -1, Y: -6.123233995736757e-17, Z: 0}}},
		{"90:-180", Point{r3.Vector{X: -6.123233995736757e-17, Y: -0, Z: 1}}},
		{"1e-20:1e-30", Point{r3.Vector{X: 1, Y: 0, Z: 0}}},
	}

	for _, test := range tests {
		pt := parsePoint(test.have)
		if !pt.ApproxEqual(test.want) {
			t.Errorf("parsePoint(%s) = %v, want %v", test.have, pt, test.want)
		}
		if got := pointToString(pt, false); got != test.have {
			t.Errorf("pointToString(parsePoint(%v), false) = %v, want %v", test.have, got, test.have)
		}
	}
}

func TestTextFormatParsePointRoundtripEdgecases(t *testing.T) {
	tests := []struct {
		have    string
		wantPt  Point
		wantStr string
	}{
		// just past pole, lng should shift by 180
		{
			have:    "91:0",
			wantPt:  Point{r3.Vector{X: -0.017452406437283473, Y: 0, Z: 0.9998476951563913}},
			wantStr: "89:-180",
		},
		{
			have:    "-91:0",
			wantPt:  Point{r3.Vector{X: -0.017452406437283473, Y: -0, Z: -0.9998476951563913}},
			wantStr: "-89:-180",
		},

		// The conversion from degrees to radians and back leads to little
		// bits of floating point noise, so we end up with things like
		// 7.01e-15 instead of 0.

		// values wrap around past the North pole back past the equator on the
		// other side of the earth.
		{
			have:    "179.99:0",
			wantPt:  Point{r3.Vector{X: -0.9999999847691292, Y: -0, Z: 0.00017453292431344843}},
			wantStr: "0.0100000000000064:-180",
		},
		/*
			// TODO(roberts): This test output differs between gccgo and 6g in the last significant digit.
			{
				have:    "180:0",
				wantPt:  Point{r3.Vector{X: -1, Y: -0, Z: 1.2246467991473515e-16}},
				wantStr: "7.01670929853487e-15:-180",
			},
		*/
		{
			have:    "181.0:0",
			wantPt:  Point{r3.Vector{X: -0.9998476951563913, Y: -0, Z: -0.017452406437283637}},
			wantStr: "-1.00000000000001:-180",
		},
		/*
			// past pole to equator, lng should shift by 180 as well.
			// TODO(roberts): This test output differs between gccgo and 6g in the last significant digit.
			{
				have:    "-180:90",
				wantPt:  Point{r3.Vector{X: -6.123233995736757e-17, Y: -1, Z: 1.2246467991473515e-16}},
				wantStr: "-7.01670929853487e-15:-90",
			},
		*/

		// string contains more than one value, only first is used in making a point.
		{
			have:    "37.4210:-122.0866, 37.4231:-122.0819",
			wantPt:  Point{r3.Vector{X: -0.4218751185559026, Y: -0.6728760966593905, Z: 0.6076669670863027}},
			wantStr: "37.421:-122.0866",
		},
	}

	for _, test := range tests {
		pt := parsePoint(test.have)
		if !pt.ApproxEqual(test.wantPt) {
			t.Errorf("parsePoint(%s) = %v, want %v", test.have, pt, test.wantPt)
		}
		if got := pointToString(pt, false); got != test.wantStr {
			t.Errorf("pointToString(parsePoint(%v), false) = %v, want %v", test.have, got, test.wantStr)
		}
	}
}

func TestTextFormatParsePointsLatLngs(t *testing.T) {
	tests := []struct {
		have    string
		wantPts []Point
		wantLLs []LatLng
	}{
		{
			have:    "0:0",
			wantPts: []Point{{r3.Vector{X: 1, Y: 0, Z: 0}}},
			wantLLs: []LatLng{{Lat: 0, Lng: 0}},
		},
		{
			have:    "      0:0,    ",
			wantPts: []Point{{r3.Vector{X: 1, Y: 0, Z: 0}}},
			wantLLs: []LatLng{{Lat: 0, Lng: 0}},
		},
		{
			have: "90:0,-90:0",
			wantPts: []Point{
				{r3.Vector{X: 6.123233995736757e-17, Y: 0, Z: 1}},
				{r3.Vector{X: 6.123233995736757e-17, Y: 0, Z: -1}},
			},
			wantLLs: []LatLng{
				{Lat: 90 * s1.Degree, Lng: 0},
				{Lat: -90 * s1.Degree, Lng: 0},
			},
		},
		{
			have: "90:0, 0:90, -90:0, 0:-90",
			wantPts: []Point{
				{r3.Vector{X: 6.123233995736757e-17, Y: 0, Z: 1}},
				{r3.Vector{X: 6.123233995736757e-17, Y: 1, Z: 0}},
				{r3.Vector{X: 6.123233995736757e-17, Y: 0, Z: -1}},
				{r3.Vector{X: 6.123233995736757e-17, Y: -1, Z: 0}},
			},
			wantLLs: []LatLng{
				{Lat: 90 * s1.Degree, Lng: 0},
				{Lat: 0, Lng: 90 * s1.Degree},
				{Lat: -90 * s1.Degree, Lng: 0},
				{Lat: 0, Lng: -90 * s1.Degree},
			},
		},
		{
			have: "37.4210:-122.0866, 37.4231:-122.0819",
			wantPts: []Point{
				{r3.Vector{X: -0.421875118555903, Y: -0.672876096659391, Z: 0.607666967086303}},
				{r3.Vector{X: -0.421808091075447, Y: -0.672891829588934, Z: 0.607696075333505}},
			},
			wantLLs: []LatLng{
				{Lat: s1.Degree * 37.4210, Lng: s1.Degree * -122.0866},
				{Lat: s1.Degree * 37.4231, Lng: s1.Degree * -122.0819},
			},
		},
		{
			// empty string, should get back nothing.
			have: "",
		},
		{
			// two empty elements, both should be skipped.
			have: ",",
		},
		{
			// Oversized values should come through as expected.
			have: "9000:1234.56",
			wantPts: []Point{
				{r3.Vector{X: -0.903035619536086, Y: 0.429565675827430, Z: 9.82193362e-16}},
			},

			wantLLs: []LatLng{
				{Lat: 9000 * s1.Degree, Lng: 1234.56 * s1.Degree},
			},
		},
	}

	for _, test := range tests {
		for i, pt := range parsePoints(test.have) {
			if !pt.ApproxEqual(test.wantPts[i]) {
				t.Errorf("parsePoints(%s): [%d]: got %v, want %v", test.have, i, pt, test.wantPts[i])
			}
		}

		// TODO(roberts): Test the roundtrip back to pointsToString()

		for i, ll := range parseLatLngs(test.have) {
			if ll != test.wantLLs[i] {
				t.Errorf("parseLatLngs(%s): [%d]: got %v, want %v", test.have, i, ll, test.wantLLs[i])
			}
		}

		// TODO(roberts): Test the roundtrip back to latlngssToString()
	}
}

func TestTextFormatParseRect(t *testing.T) {
	tests := []struct {
		have string
		want Rect
	}{
		{"0:0", Rect{}},
		{
			"1:1",
			Rect{
				r1.Interval{Lo: float64(s1.Degree), Hi: float64(s1.Degree)},
				s1.Interval{Lo: float64(s1.Degree), Hi: float64(s1.Degree)},
			},
		},
		{
			"1:1, 2:2, 3:3",
			Rect{
				r1.Interval{Lo: float64(s1.Degree), Hi: 3 * float64(s1.Degree)},
				s1.Interval{Lo: float64(s1.Degree), Hi: 3 * float64(s1.Degree)},
			},
		},
		{
			"-90:-180, 90:180",
			Rect{
				r1.Interval{Lo: -90 * float64(s1.Degree), Hi: 90 * float64(s1.Degree)},
				s1.Interval{Lo: 180 * float64(s1.Degree), Hi: -180 * float64(s1.Degree)},
			},
		},
		{
			"-89.99:0, 89.99:179.99",
			Rect{
				r1.Interval{Lo: -89.99 * float64(s1.Degree), Hi: 89.99 * float64(s1.Degree)},
				s1.Interval{Lo: 0, Hi: 179.99 * float64(s1.Degree)},
			},
		},
		{
			"-89.99:-179.99, 89.99:179.99",
			Rect{
				r1.Interval{Lo: -89.99 * float64(s1.Degree), Hi: 89.99 * float64(s1.Degree)},
				s1.Interval{Lo: 179.99 * float64(s1.Degree), Hi: -179.99 * float64(s1.Degree)},
			},
		},
		{
			"37.4210:-122.0866, 37.4231:-122.0819",
			Rect{
				r1.Interval{Lo: float64(s1.Degree * 37.4210), Hi: float64(s1.Degree * 37.4231)},
				s1.Interval{Lo: float64(s1.Degree * -122.0866), Hi: float64(s1.Degree * -122.0819)},
			},
		},
		{
			"-876.54:-654.43, 963.84:2468.35",
			Rect{
				r1.Interval{Lo: -876.54 * float64(s1.Degree), Hi: -876.54 * float64(s1.Degree)},
				s1.Interval{Lo: -654.43 * float64(s1.Degree), Hi: -654.43 * float64(s1.Degree)},
			},
		},
	}
	for _, test := range tests {
		if got := makeRect(test.have); got != test.want {
			t.Errorf("makeRect(%s) = %v, want %v", test.have, got, test.want)
		}
	}
}

func TestTextFormatMakeCellUnion(t *testing.T) {
	tests := []struct {
		have []string
		want CellUnion
	}{
		{
			have: nil,
			want: CellUnion{},
		},
		{
			have: []string{},
			want: CellUnion{},
		},
		{
			have: []string{"google"},
			want: CellUnion{0},
		},
		{
			have: []string{"0/"},
			want: CellUnion{CellIDFromFace(0)},
		},
		{
			have: []string{"2/010", "2/011", "2/02"},
			want: CellUnion{0x4240000000000000, 0x42c0000000000000, 0x4500000000000000},
		},
	}

	for _, test := range tests {
		got := makeCellUnion(test.have...)
		got.Normalize()
		if !got.Equal(test.want) {
			t.Errorf("makeCellUnion(%v) = %v, want %v", test.have, got, test.want)
		}
	}
}

func TestTextFormatMakeLaxPolyline(t *testing.T) {
	l := makeLaxPolyline("-20:150, -20:151, -19:150")

	// No easy equality check for laxPolylines; check vertices instead.
	if len(l.vertices) != 3 {
		t.Errorf("len(l.vertices) = %d, want 3", len(l.vertices))
	}
	if got, want := LatLngFromPoint(l.vertices[0]), LatLngFromDegrees(-20, 150); !latLngsApproxEqual(got, want, epsilon) {
		t.Errorf("vertex(0) = %v, want %v", got, want)
	}
	if got, want := LatLngFromPoint(l.vertices[1]), LatLngFromDegrees(-20, 151); !latLngsApproxEqual(got, want, epsilon) {
		t.Errorf("vertex(1) = %v, want %v", got, want)
	}
	if got, want := LatLngFromPoint(l.vertices[2]), LatLngFromDegrees(-19, 150); !latLngsApproxEqual(got, want, epsilon) {
		t.Errorf("vertex(2) = %v, want %v", got, want)
	}

	// TODO(roberts): test out an invalid value
	// makeLaxPolyline("blah")
}

func TestTextFormatMakeLaxPolygonEmpty(t *testing.T) {
	// Verify that "" and "empty" both create empty polygons.
	shape := makeLaxPolygon("")
	if got, want := shape.numLoops, 0; got != want {
		t.Errorf("laxPolygon.numLoops = %d, want %d", got, want)
	}
	shape = makeLaxPolygon("empty")
	if got, want := shape.numLoops, 0; got != want {
		t.Errorf("laxPolygon.numLoops = %d, want %d", got, want)
	}
}

func TestTextFormatMakeLaxPolygonFull(t *testing.T) {
	shape := makeLaxPolygon("full")
	if got, want := shape.numLoops, 1; got != want {
		t.Errorf("laxPolygon.numLoops = %d, want %d", got, want)
	}
	if got, want := shape.numLoopVertices(0), 0; got != want {
		t.Errorf("laxPolygon.numLoopVertices(%d) = %d, want %d", 0, got, want)
	}
}

func TestTextFormatMakeLaxPolygonFullWithHole(t *testing.T) {
	shape := makeLaxPolygon("full; 0:0")
	if got, want := shape.numLoops, 2; got != want {
		t.Errorf("laxPolygon.numLoops = %d, want %d", got, want)
	}
	if got, want := shape.numLoopVertices(0), 0; got != want {
		t.Errorf("laxPolygon.numLoopVertices(%d) = %d, want %d", 0, got, want)
	}
	if got, want := shape.numLoopVertices(1), 1; got != want {
		t.Errorf("laxPolygon.numLoopVertices(%d) = %d, want %d", 1, got, want)

	}
	if got, want := shape.NumEdges(), 1; got != want {
		t.Errorf("laxPolygon.NumEdges() = %d, want %d", got, want)
	}
}

func TestTextFormatShapeIndexDebugStringRoundTrip(t *testing.T) {
	// TODO(rsned): Incorporate the roundtripPrecision parameter to the tests.
	tests := []string{
		"# #",
		"0:0 # #",
		"0:0 | 1:0 # #",
		"0:0 | 1:0 # #",
		"# 0:0, 0:0 #",
		"# 0:0, 0:0 | 1:0, 2:0 #",
		"# # 0:0",
		"# # 0:0, 0:1",
		"# # 0:0, 0:1, 1:0",
		"# # 0:0, 0:1, 1:0, 2:2",
		"# # full",
	}

	for _, want := range tests {
		if got := shapeIndexDebugString(makeShapeIndex(want), false); got != want {
			t.Errorf("ShapeIndex failed roundtrip to string. got %q, want %q", got, want)
		}
	}
}

// TODO(roberts): Remaining tests
// to debug string tests for
//   SpecialCases, EmptyLoop, EmptyPolyline, Empty Othertypes
//
// make type tests for ValidInput and InvalidInput for
//   Points, Rect, Loop, Polyline, Polygon,
//   LaxPolygon, ShapeIndex
