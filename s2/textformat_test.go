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

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/geo/r3"
)

// parsePoint returns an Point from the latitude-longitude coordinate in degrees
// in the given string, or the origin if the string was invalid.
// e.g., "-20:150"
func parsePoint(s string) Point {
	p := parsePoints(s)
	if len(p) > 0 {
		return p[0]
	}

	return Point{r3.Vector{0, 0, 0}}
}

// parsePoints takes a string of lat:lng points and returns the set of Points it defines.
func parsePoints(s string) []Point {
	lls := parseLatLngs(s)
	points := make([]Point, len(lls))
	for i, ll := range lls {
		points[i] = PointFromLatLng(ll)
	}
	return points
}

// parseLatLngs splits up a string of lat:lng points and returns the list of parsed
// entries.
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

// parseRect returns the minimal bounding Rect that contains the one or more
// latitude-longitude coordinates in degrees in the given string.
// Examples of input:
//   "-20:150"                     // one point
//   "-20:150, -20:151, -19:150"   // three points
//
// TODO(roberts): Rename this to makeRect for consistency with other methods.
func parseRect(s string) Rect {
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

// makeLoop constructs a loop from a comma separated string of lat:lng
// coordinates in degrees. Example of the input format:
//   "-20:150, 10:-120, 0.123:-170.652"
// The special strings "empty" or "full" create an empty or full loop respectively.
func makeLoop(s string) *Loop {
	if s == "full" {
		return FullLoop()
	}
	if s == "empty" {
		return EmptyLoop()
	}

	return LoopFromPoints(parsePoints(s))
}

// makePolygon constructs a polygon from the set of semicolon separated CSV
// strings of lat:lng points defining each loop in the polygon. If the normalize
// flag is set to true, loops are normalized by inverting them
// if necessary so that they enclose at most half of the unit sphere.
//
// Examples of the input format:
//     "10:20, 90:0, 20:30"                                  // one loop
//     "10:20, 90:0, 20:30; 5.5:6.5, -90:-180, -15.2:20.3"   // two loops
//     ""       // the empty polygon (consisting of no loops)
//     "full"   // the full polygon (consisting of one full loop)
//     "empty"  // **INVALID** (a polygon consisting of one empty loop)
func makePolygon(s string, normalize bool) *Polygon {
	strs := strings.Split(s, ";")
	var loops []*Loop
	for _, str := range strs {
		if str == "" {
			continue
		}
		loop := makeLoop(strings.TrimSpace(str))
		if normalize {
			// TODO(roberts): Uncomment once Normalize is implemented.
			// loop.Normalize()
		}
		loops = append(loops, loop)
	}
	return PolygonFromLoops(loops)
}

// makePolyline constructs a Polyline from the given string of lat:lng values.
func makePolyline(s string) *Polyline {
	p := Polyline(parsePoints(s))
	return &p
}

// TODO(roberts): Remaining C++ textformat related methods
// to debug string methods for many types.
// make type methods for remaining types.
// to/from debug for ShapeIndex
