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

// s2textformat_test contains a collection of functions for converting geometry
// to and from a human-readable format. It is intended for testing and debugging.
// Be aware that the human-readable format is *NOT* designed to preserve the full
// precision of the original object, so it should not be used for data storage.
//
// Most functions here use the same format for inputs, a comma separated set of
// latitude-longitude coordinates in degrees. Functions that expect a different
// input document the values in the function comment.
//
// Examples of the input format:
//     ""                                 // no points
//     "-20:150"                          // one point
//     "-20:150, 10:-120, 0.123:-170.652" // three points

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/golang/geo/r3"
)

// writePoint formats the point and writes it to the given writer.
func writePoint(w io.Writer, p Point) {
	ll := LatLngFromPoint(p)
	fmt.Fprintf(w, "%.15g:%.15g", ll.Lat.Degrees(), ll.Lng.Degrees())
}

// writePoints formats the given points in debug format and writes them to the given writer.
func writePoints(w io.Writer, pts []Point) {
	for i, pt := range pts {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		writePoint(w, pt)
	}
}

// parsePoint returns a Point from the given string, or the origin if the
// string was invalid. If more than one value is given, only the first is used.
func parsePoint(s string) Point {
	p := parsePoints(s)
	if len(p) > 0 {
		return p[0]
	}

	return Point{r3.Vector{0, 0, 0}}
}

// pointToString returns a string representation suitable for reconstruction
// by the parsePoint method.
func pointToString(point Point) string {
	var buf bytes.Buffer
	writePoint(&buf, point)
	return buf.String()
}

// parsePoints returns the values in the input string as Points.
func parsePoints(s string) []Point {
	lls := parseLatLngs(s)
	if len(lls) == 0 {
		return nil
	}
	points := make([]Point, len(lls))
	for i, ll := range lls {
		points[i] = PointFromLatLng(ll)
	}
	return points
}

// pointsToString returns a string representation suitable for reconstruction
// by the parsePoints method.
func pointsToString(points []Point) string {
	var buf bytes.Buffer
	writePoints(&buf, points)
	return buf.String()
}

// parseLatLngs returns the values in the input string as LatLngs.
func parseLatLngs(s string) []LatLng {
	var lls []LatLng
	if s == "" {
		return lls
	}
	for _, piece := range strings.Split(s, ",") {
		piece = strings.TrimSpace(piece)

		// Skip empty strings.
		if piece == "" {
			continue
		}

		p := strings.Split(piece, ":")
		if len(p) != 2 {
			panic(fmt.Sprintf("invalid input string for parseLatLngs: %q", piece))
		}

		lat, err := strconv.ParseFloat(p[0], 64)
		if err != nil {
			panic(fmt.Sprintf("invalid float in parseLatLngs: %q, err: %v", p[0], err))
		}

		lng, err := strconv.ParseFloat(p[1], 64)
		if err != nil {
			panic(fmt.Sprintf("invalid float in parseLatLngs: %q, err: %v", p[1], err))
		}

		lls = append(lls, LatLngFromDegrees(lat, lng))
	}
	return lls
}

// makeRect returns the minimal bounding Rect that contains the values in the input string.
func makeRect(s string) Rect {
	var rect Rect
	lls := parseLatLngs(s)
	if len(lls) > 0 {
		rect = RectFromLatLng(lls[0])
	}

	for _, ll := range lls[1:] {
		rect = rect.AddPoint(ll)
	}

	return rect
}

// makeCellUnion returns a CellUnion from the given CellID token strings.
func makeCellUnion(tokens ...string) CellUnion {
	var cu CellUnion

	for _, t := range tokens {
		cu = append(cu, cellIDFromString(t))
	}
	return cu
}

// makeLoop constructs a Loop from the input string.
// The strings "empty" or "full" create an empty or full loop respectively.
func makeLoop(s string) *Loop {
	if s == "full" {
		return FullLoop()
	}
	if s == "empty" {
		return EmptyLoop()
	}

	return LoopFromPoints(parsePoints(s))
}

// makePolygon constructs a polygon from the sequence of loops in the input
// string. Loops are automatically normalized by inverting them if necessary
// so that they enclose at most half of the unit sphere. (Historically this was
// once a requirement of polygon loops. It also hides the problem that if the
// user thinks of the coordinates as X:Y rather than LAT:LNG, it yields a loop
// with the opposite orientation.)
//
// Loops are semicolon separated in the input string with each loop using the
// same format as above.
//
// Examples of the input format:
//     "10:20, 90:0, 20:30"                                  // one loop
//     "10:20, 90:0, 20:30; 5.5:6.5, -90:-180, -15.2:20.3"   // two loops
//     ""       // the empty polygon (consisting of no loops)
//     "empty"  // the empty polygon (consisting of no loops)
//     "full"   // the full polygon (consisting of one full loop)
func makePolygon(s string, normalize bool) *Polygon {
	var loops []*Loop
	// Avoid the case where strings.Split on empty string will still return
	// one empty value, where we want no values.
	if s == "empty" || s == "" {
		return PolygonFromLoops(loops)
	}

	for _, str := range strings.Split(s, ";") {
		// The polygon test strings mostly have a trailing semicolon
		// (to make concatenating them for tests easy). The C++
		// SplitString doesn't return empty elements where as Go does,
		// so we need to check before using it.
		if str == "" {
			continue
		}
		loop := makeLoop(strings.TrimSpace(str))
		if normalize && !loop.IsFull() {
			loop.Normalize()
		}
		loops = append(loops, loop)
	}
	return PolygonFromLoops(loops)
}

// makePolyline constructs a Polyline from the given string.
func makePolyline(s string) *Polyline {
	p := Polyline(parsePoints(s))
	return &p
}

// makeLaxPolyline constructs a laxPolyline from the given string.
func makeLaxPolyline(s string) *laxPolyline {
	return laxPolylineFromPoints(parsePoints(s))
}

// laxPolylineToString returns a string representation suitable for reconstruction
// by the makeLaxPolyline method.
func laxPolylineToString(l *laxPolyline) string {
	var buf bytes.Buffer
	writePoints(&buf, l.vertices)
	return buf.String()

}

// makeLaxPolygon creates a laxPolygon from the given debug formatted string.
// Similar to makePolygon, except that loops must be oriented so that the
// interior of the loop is always on the left, and polygons with degeneracies
// are supported. As with makePolygon, "full" denotes the full polygon and "empty"
// is not allowed (instead, simply create a laxPolygon with no loops).
func makeLaxPolygon(s string) *laxPolygon {
	var points [][]Point
	if s == "" {
		return laxPolygonFromPoints(points)
	}
	for _, l := range strings.Split(s, ";") {
		if l == "full" {
			points = append(points, []Point{})
		} else if l != "empty" {
			points = append(points, parsePoints(l))
		}
	}
	return laxPolygonFromPoints(points)
}

// makeShapeIndex builds a ShapeIndex from the given debug string containing
// the points, polylines, and loops (in the form of a single polygon)
// described by the following format:
//
//   point1|point2|... # line1|line2|... # polygon1|polygon2|...
//
// Examples:
//   1:2 | 2:3 # #                     // Two points
//   # 0:0, 1:1, 2:2 | 3:3, 4:4 #      // Two polylines
//   # # 0:0, 0:3, 3:0; 1:1, 2:1, 1:2  // Two nested loops (one polygon)
//   5:5 # 6:6, 7:7 # 0:0, 0:1, 1:0    // One of each
//   # # empty                         // One empty polygon
//   # # empty | full                  // One empty polygon, one full polygon
//
// Loops should be directed so that the region's interior is on the left.
// Loops can be degenerate (they do not need to meet Loop requirements).
//
// Note: Because whitespace is ignored, empty polygons must be specified
// as the string "empty" rather than as the empty string ("").
func makeShapeIndex(s string) *ShapeIndex {
	fields := strings.Split(s, "#")
	if len(fields) != 3 {
		panic("shapeIndex debug string must contain 2 '#' characters")
	}

	index := NewShapeIndex()

	var points []Point
	for _, p := range strings.Split(fields[0], "|") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		points = append(points, parsePoint(p))
	}
	if len(points) > 0 {
		p := PointVector(points)
		index.Add(&p)
	}

	for _, p := range strings.Split(fields[1], "|") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if polyline := makeLaxPolyline(p); polyline != nil {
			index.Add(polyline)
		}
	}

	for _, p := range strings.Split(fields[2], "|") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if polygon := makeLaxPolygon(p); polygon != nil {
			index.Add(polygon)
		}
	}

	return index
}

// shapeIndexDebugString outputs the contents of this ShapeIndex in debug
// format. The index may contain Shapes of any type. Shapes are reordered
// if necessary so that all point geometry (shapes of dimension 0) are first,
// followed by all polyline geometry, followed by all polygon geometry.
func shapeIndexDebugString(index *ShapeIndex) string {
	var buf bytes.Buffer

	for dim := 0; dim <= 2; dim++ {
		if dim > 0 {
			buf.WriteByte('#')
		}

		var count int

		// Use shapes ordered by id rather than ranging over the
		// index.shapes map to ensure that the ordering of shapes in the
		// generated string matches the C++ generated strings.
		for i := int32(0); i < index.nextID; i++ {
			shape := index.Shape(i)
			// Only use shapes that are still in the index and at the
			// current geometry level we are outputting.
			if shape == nil || shape.Dimension() != dim {
				continue
			}
			if count > 0 {
				buf.WriteString(" | ")
			} else {
				if dim > 0 {
					buf.WriteByte(' ')
				}
			}

			for c := 0; c < shape.NumChains(); c++ {
				if c > 0 {
					if dim == 2 {
						buf.WriteString("; ")
					} else {
						buf.WriteString(" | ")
					}
				}
				chain := shape.Chain(c)
				pts := []Point{shape.Edge(chain.Start).V0}
				limit := chain.Start + chain.Length
				if dim != 1 {
					limit--
				}

				for e := chain.Start; e < limit; e++ {
					pts = append(pts, shape.Edge(e).V1)
				}
				writePoints(&buf, pts)
				count++
			}
		}

		if dim == 1 || (dim == 0 && count > 0) {
			buf.WriteByte(' ')
		}
	}

	return buf.String()
}

// TODO(roberts): Remaining C++ textformat related methods
// make$S2TYPE methods for missing types.
// to debug string for many types
