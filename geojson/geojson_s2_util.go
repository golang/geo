//  Copyright (c) 2022 Couchbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package geojson

import (
	"strconv"
	"strings"

	index "github.com/blevesearch/bleve_index_api"
	"github.com/blevesearch/geo/s2"
	"github.com/golang/geo/s1"
)

// ------------------------------------------------------------------------

func polylineIntersectsPoint(pls []*s2.Polyline,
	point *s2.Point) bool {
	s2cell := s2.CellFromPoint(*point)

	for _, pl := range pls {
		if pl.IntersectsCell(s2cell) {
			return true
		}
	}

	return false
}

func polylineIntersectsPolygons(pls []*s2.Polyline,
	s2pgns []*s2.Polygon) bool {
	for _, pl := range pls {
		for _, s2pgn := range s2pgns {
			for i := 0; i < pl.NumEdges(); i++ {
				edge := pl.Edge(i)
				a := []float64{edge.V0.X, edge.V0.Y}
				b := []float64{edge.V1.X, edge.V1.Y}

				for i := 0; i < s2pgn.NumEdges(); i++ {
					edgeB := s2pgn.Edge(i)

					c := []float64{edgeB.V0.X, edgeB.V0.Y}
					d := []float64{edgeB.V1.X, edgeB.V1.Y}

					if doIntersect(a, b, c, d) {
						return true
					}
				}
			}
		}
	}

	return false
}

func geometryCollectionIntersectsShape(gc *GeometryCollection,
	shapeIn index.GeoJSON) bool {
	for _, shape := range gc.Members() {
		intersects, err := shapeIn.Intersects(shape)
		if err == nil && intersects {
			return true
		}
	}
	return false
}

func polygonsIntersectsLinestrings(s2pgn *s2.Polygon,
	pls []*s2.Polyline) bool {
	for _, pl := range pls {
		for i := 0; i < pl.NumEdges(); i++ {
			edge := pl.Edge(i)
			a := []float64{edge.V0.X, edge.V0.Y}
			b := []float64{edge.V1.X, edge.V1.Y}

			for j := 0; j < s2pgn.NumEdges(); j++ {
				edgeB := s2pgn.Edge(j)

				c := []float64{edgeB.V0.X, edgeB.V0.Y}
				d := []float64{edgeB.V1.X, edgeB.V1.Y}

				if doIntersect(a, b, c, d) {
					return true
				}
			}
		}
	}

	return false
}

func polygonsContainsLineStrings(s2pgns []*s2.Polygon,
	pls []*s2.Polyline) bool {
	linesWithIn := make(map[int]struct{})
	checker := s2.NewCrossingEdgeQuery(s2.NewShapeIndex())
nextLine:
	for lineIndex, pl := range pls {
		for i := 0; i < len(*pl)-1; i++ {
			start := (*pl)[i]
			end := (*pl)[i+1]

			for _, s2pgn := range s2pgns {
				containsStart := s2pgn.ContainsPoint(start)
				containsEnd := s2pgn.ContainsPoint(end)
				if containsStart && containsEnd {
					crossings := checker.Crossings(start, end, s2pgn, s2.CrossingTypeInterior)
					if len(crossings) > 0 {
						continue nextLine
					}
					linesWithIn[lineIndex] = struct{}{}
					continue nextLine
				} else {
					for _, loop := range s2pgn.Loops() {
						for i := 0; i < loop.NumVertices(); i++ {
							if !containsStart && start.ApproxEqual(loop.Vertex(i)) {
								containsStart = true
							} else if !containsEnd && end.ApproxEqual(loop.Vertex(i)) {
								containsEnd = true
							}
							if containsStart && containsEnd {
								linesWithIn[lineIndex] = struct{}{}
								continue nextLine
							}
						}
					}
				}
			}
		}
	}

	return len(pls) == len(linesWithIn)
}

func rectangleIntersectsWithPolygons(s2rect *s2.Rect,
	s2pgns []*s2.Polygon) bool {
	s2pgnFromRect := s2PolygonFromS2Rectangle(s2rect)
	for _, s2pgn := range s2pgns {
		if s2pgn.Intersects(s2pgnFromRect) {
			return true
		}
	}

	return false
}

func rectangleIntersectsWithLineStrings(s2rect *s2.Rect,
	polylines []*s2.Polyline) bool {
	for _, pl := range polylines {
		for i := 0; i < pl.NumEdges(); i++ {
			edgeA := pl.Edge(i)
			a := []float64{edgeA.V0.X, edgeA.V0.Y}
			b := []float64{edgeA.V1.X, edgeA.V1.Y}

			for j := 0; j < 4; j++ {
				v1 := s2.PointFromLatLng(s2rect.Vertex(j))
				v2 := s2.PointFromLatLng(s2rect.Vertex((j + 1) % 4))

				c := []float64{v1.X, v1.Y}
				d := []float64{v2.X, v2.Y}

				if doIntersect(a, b, c, d) {
					return true
				}
			}
		}
	}

	return false
}

func s2PolygonFromCoordinates(coordinates [][][]float64) *s2.Polygon {
	loops := make([]*s2.Loop, 0, len(coordinates))
	for _, loop := range coordinates {
		var points []s2.Point
		if loop[0][0] == loop[len(loop)-1][0] && loop[0][1] == loop[len(loop)-1][1] {
			loop = loop[:len(loop)-1]
		}
		for _, point := range loop {
			p := s2.PointFromLatLng(s2.LatLngFromDegrees(point[1], point[0]))
			points = append(points, p)
		}
		s2loop := s2.LoopFromPoints(points)
		loops = append(loops, s2loop)
	}

	rv := s2.PolygonFromOrientedLoops(loops)
	return rv
}

func s2PolygonFromS2Rectangle(s2rect *s2.Rect) *s2.Polygon {
	loops := make([]*s2.Loop, 0, 1)
	var points []s2.Point
	for j := 0; j < 4; j++ {
		points = append(points, s2.PointFromLatLng(s2rect.Vertex(j%4)))
	}

	loops = append(loops, s2.LoopFromPoints(points))
	return s2.PolygonFromLoops(loops)
}

func DeduplicateTerms(terms []string) []string {
	var rv []string
	hash := make(map[string]struct{}, len(terms))
	for _, term := range terms {
		if _, exists := hash[term]; !exists {
			rv = append(rv, term)
			hash[term] = struct{}{}
		}
	}

	return rv
}

//----------------------------------------------------------------------

var earthRadiusInMeter = 6378137.0

func radiusInMetersToS1Angle(radius float64) s1.Angle {
	return s1.Angle(radius / earthRadiusInMeter)
}

func s2PolylinesFromCoordinates(coordinates [][][]float64) []*s2.Polyline {
	var polylines []*s2.Polyline
	for _, lines := range coordinates {
		var latlngs []s2.LatLng
		for _, line := range lines {
			v := s2.LatLngFromDegrees(line[1], line[0])
			latlngs = append(latlngs, v)
		}
		polylines = append(polylines, s2.PolylineFromLatLngs(latlngs))
	}
	return polylines
}

func s2RectFromBounds(topLeft, bottomRight []float64) *s2.Rect {
	rect := s2.EmptyRect()
	rect = rect.AddPoint(s2.LatLngFromDegrees(topLeft[1], topLeft[0]))
	rect = rect.AddPoint(s2.LatLngFromDegrees(bottomRight[1], bottomRight[0]))
	return &rect
}

func s2Cap(vertices []float64, radiusInMeter float64) *s2.Cap {
	cp := s2.PointFromLatLng(s2.LatLngFromDegrees(vertices[1], vertices[0]))
	angle := radiusInMetersToS1Angle(float64(radiusInMeter))
	cap := s2.CapFromCenterAngle(cp, angle)
	return &cap
}

func max(a, b float64) float64 {
	if a >= b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a >= b {
		return b
	}
	return a
}

func onsegment(p, q, r []float64) bool {
	if q[0] <= max(p[0], r[0]) && q[0] >= min(p[0], r[0]) &&
		q[1] <= max(p[1], r[1]) && q[1] >= min(p[1], r[1]) {
		return true
	}

	return false
}

func doIntersect(p1, q1, p2, q2 []float64) bool {
	o1 := orientation(p1, q1, p2)
	o2 := orientation(p1, q1, q2)
	o3 := orientation(p2, q2, p1)
	o4 := orientation(p2, q2, q1)

	if o1 != o2 && o3 != o4 {
		return true
	}

	if o1 == 0 && onsegment(p1, p2, q1) {
		return true
	}

	if o2 == 0 && onsegment(p1, q2, q1) {
		return true
	}

	if o3 == 0 && onsegment(p2, p1, q2) {
		return true
	}

	if o4 == 0 && onsegment(p2, q1, q2) {
		return true
	}

	return false
}

func orientation(p, q, r []float64) int {
	val := (q[1]-p[1])*(r[0]-q[0]) - (q[0]-p[0])*(r[1]-q[1])
	if val == 0 {
		return 0
	}
	if val > 0 {
		return 1
	}
	return 2
}

func StripCoveringTerms(terms []string) []string {
	rv := make([]string, 0, len(terms))
	for _, term := range terms {
		if strings.HasPrefix(term, "$") {
			rv = append(rv, term[1:])
			continue
		}
		rv = append(rv, term)
	}
	return DeduplicateTerms(rv)
}

type distanceUnit struct {
	conv     float64
	suffixes []string
}

var inch = distanceUnit{0.0254, []string{"in", "inch"}}
var yard = distanceUnit{0.9144, []string{"yd", "yards"}}
var feet = distanceUnit{0.3048, []string{"ft", "feet"}}
var kilom = distanceUnit{1000, []string{"km", "kilometers"}}
var nauticalm = distanceUnit{1852.0, []string{"nm", "nauticalmiles"}}
var millim = distanceUnit{0.001, []string{"mm", "millimeters"}}
var centim = distanceUnit{0.01, []string{"cm", "centimeters"}}
var miles = distanceUnit{1609.344, []string{"mi", "miles"}}
var meters = distanceUnit{1, []string{"m", "meters"}}

var distanceUnits = []*distanceUnit{
	&inch, &yard, &feet, &kilom, &nauticalm, &millim, &centim, &miles, &meters,
}

// ParseDistance attempts to parse a distance string and return distance in
// meters.  Example formats supported:
// "5in" "5inch" "7yd" "7yards" "9ft" "9feet" "11km" "11kilometers"
// "3nm" "3nauticalmiles" "13mm" "13millimeters" "15cm" "15centimeters"
// "17mi" "17miles" "19m" "19meters"
// If the unit cannot be determined, the entire string is parsed and the
// unit of meters is assumed.
// If the number portion cannot be parsed, 0 and the parse error are returned.
func ParseDistance(d string) (float64, error) {
	for _, unit := range distanceUnits {
		for _, unitSuffix := range unit.suffixes {
			if strings.HasSuffix(d, unitSuffix) {
				parsedNum, err := strconv.ParseFloat(d[0:len(d)-len(unitSuffix)], 64)
				if err != nil {
					return 0, err
				}
				return parsedNum * unit.conv, nil
			}
		}
	}
	// no unit matched, try assuming meters?
	parsedNum, err := strconv.ParseFloat(d, 64)
	if err != nil {
		return 0, err
	}
	return parsedNum, nil
}
