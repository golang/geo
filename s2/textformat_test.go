/*
Copyright 2017 Google Inc. All rights reserved.

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
		writePoint(w, pt)
		if i < len(pts) {
			fmt.Fprintf(w, ", ")
		}
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
	pieces := strings.Split(s, ",")
	var lls []LatLng
	for _, piece := range pieces {
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
	if s == "empty" {
		s = ""
	}
	strs := strings.Split(s, ";")
	var loops []*Loop
	for _, str := range strs {
		if str == "" {
			continue
		}
		loop := makeLoop(strings.TrimSpace(str))
		if normalize && !loop.IsFull() {
			// TODO(roberts): Uncomment once Normalize is implemented.
			// loop.Normalize()
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

// TODO(roberts): Remaining C++ textformat related methods
// make$S2TYPE methods for missing types.
// to debug string for many types
// to/from debug for ShapeIndex
