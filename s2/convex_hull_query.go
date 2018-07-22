// Copyright 2018 Google Inc. All rights reserved.
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
	"sort"

	"github.com/golang/geo/r3"
)

// ConvexHullQuery builds the convex hull of any collection of points,
// polylines, loops, and polygons.  It returns a single convex loop.
//
// The convex hull is defined as the smallest convex region on the sphere that
// contains all of your input geometry.  Recall that a region is "convex" if
// for every pair of points inside the region, the straight edge between them
// is also inside the region.  In our case, a "straight" edge is a geodesic,
// i.e. the shortest path on the sphere between two points.
//
// Containment of input geometry is defined as follows:
//
//  - Each input loop and polygon is contained by the convex hull exactly
//    (i.e., according to S2Polygon::Contains(S2Polygon)).
//
//  - Each input point is either contained by the convex hull or is a vertex
//    of the convex hull. (Recall that S2Loops do not necessarily contain their
//    vertices.)
//
//  - For each input polyline, the convex hull contains all of its vertices
//    according to the rule for points above.  (The definition of convexity
//    then ensures that the convex hull also contains the polyline edges.)
//
// To use this class, call the Add*() methods to add your input geometry, and
// then call ConvexHull().  Note that ConvexHull() does *not* reset the
// state; you can continue adding geometry if desired and compute the convex
// hull again.  If you want to start from scratch, simply declare a new
// S2ConvexHullQuery object (they are cheap to create).
type ConvexHullQuery struct {
	points []Point
	bound  Rect
}

// NewConvexHullQuery returns a new instance of a ConvexHullQuery.
func NewConvexHullQuery() *ConvexHullQuery {
	return &ConvexHullQuery{
		bound: EmptyRect(),
	}
}

// AddPoint adds one or more points to the input geometry.
func (q *ConvexHullQuery) AddPoint(point Point) {
	q.bound.AddPoint(LatLngFromPoint(point))
	q.points = append(q.points, point)
}

// AddPolyline add a polyline to the input geometry.
func (q *ConvexHullQuery) AddPolyline(polyline Polyline) {
	q.bound = q.bound.Union(polyline.RectBound())
	q.points = append(q.points, polyline...)
}

// AddLoop adds one or more loops to the input geometry.
func (q *ConvexHullQuery) AddLoop(loop *Loop) {
	q.bound = q.bound.Union(loop.bound)
	if loop.isEmptyOrFull() {
		return
	}

	q.points = append(q.points, loop.vertices...)
}

// AddPolygon adds one or more polygons to the input geometry.
func (q *ConvexHullQuery) AddPolygon(polygon *Polygon) {
	for i := range polygon.loops {
		q.AddLoop(polygon.loops[i])
	}
}

// CapBound computes a bounding cap for the input geometry provided.
//
// Note that this method does not clear the geometry; you can continue
// adding to it and call this method again if desired.
func (q *ConvexHullQuery) CapBound() Cap {
	return q.bound.CapBound()
}

// ConvexHull computes the convex hull of the input geometry provided.
//
// If there is no geometry, this method returns an empty loop containing no
// points (see S2Loop::is_empty()).
//
// If the geometry spans more than half of the sphere, this method returns a
// full loop containing the entire sphere (see S2Loop::is_full()).
//
// If the geometry contains 1 or 2 points, or a single edge, this method
// returns a very small loop consisting of three vertices (which are a
// superset of the input vertices).
//
// Note that this method does not clear the geometry; you can continue
// adding to it and call this method again if desired.
func (q *ConvexHullQuery) ConvexHull() *Loop {
	cap := q.CapBound()
	if cap.Height() >= 1 {
		// The bounding cap is not convex.  The current bounding cap
		// implementation is not optimal, but nevertheless it is likely that the
		// input geometry itself is not contained by any convex polygon.  In any
		// case, we need a convex bounding cap to proceed with the algorithm below
		// (in order to construct a point "origin" that is definitely outside the
		// convex hull).
		return FullLoop()
	}

	// This code implements Andrew's monotone chain algorithm, which is a simple
	// variant of the Graham scan.  Rather than sorting by x-coordinate, instead
	// we sort the points in CCW order around an origin O such that all points
	// are guaranteed to be on one side of some geodesic through O.  This
	// ensures that as we scan through the points, each new point can only
	// belong at the end of the chain (i.e., the chain is monotone in terms of
	// the angle around O from the starting point).
	origin := cap.Center().Ortho()
	sortPointsCcwAround(q.points, origin)

	// Remove duplicates.  We need to do this before checking whether there are
	// fewer than 3 points.
	q.points = uniquePoints(q.points)

	switch len(q.points) {
	case 0:
		return EmptyLoop()
	case 1:
		return singlePointLoop(q.points[0])
	case 2:
		return singleEdgeLoop(q.points[0], q.points[1])
	}

	var lower, upper []Point
	q.getMonotoneChain(&lower)
	reversePoints(q.points)
	q.getMonotoneChain(&upper)

	// Remove duplicates
	lower = lower[1 : len(lower)-1]

	// Combine chains and return loop
	return LoopFromPoints(append(lower, upper...))
}

// pointsCcwAroundSorter implements the Sort interface for slices of Point
// with a comparator for sorting points in CCW around a central point "center".
type pointsCcwAroundSorter struct {
	center Point
	points []Point
}

func (s pointsCcwAroundSorter) Len() int           { return len(s.points) }
func (s pointsCcwAroundSorter) Swap(i, j int)      { s.points[i], s.points[j] = s.points[j], s.points[i] }
func (s pointsCcwAroundSorter) Less(i, j int) bool { return Sign(s.center, s.points[i], s.points[j]) }

func sortPointsCcwAround(points []Point, origin r3.Vector) {
	sorter := pointsCcwAroundSorter{
		center: Point{origin},
		points: points,
	}
	sort.Sort(sorter)
}

func uniquePoints(points []Point) []Point {
	seen := make(map[Point]struct{}, len(points))
	i := 0
	for _, v := range points {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		points[i] = v
		i++
	}
	return points[:i]
}

func reversePoints(points []Point) {
	for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
		points[left], points[right] = points[right], points[left]
	}
}

func singlePointLoop(p Point) *Loop {
	const offset = 1e-15
	d0 := p.Ortho()
	d1 := p.Cross(d0)
	vertices := make([]Point, 3)
	vertices[0] = p
	vertices[1] = Point{p.Add(d0.Mul(offset)).Normalize()}
	vertices[2] = Point{p.Add(d1.Mul(offset)).Normalize()}
	return LoopFromPoints(vertices)
}

func singleEdgeLoop(a, b Point) *Loop {
	vertices := make([]Point, 3)
	vertices[0] = a
	vertices[1] = b
	vertices[2] = Point{a.Add(b.Vector).Normalize()}
	return LoopFromPoints(vertices)
}

func (q *ConvexHullQuery) getMonotoneChain(output *[]Point) {
	for i := range q.points {
		for len(*output) >= 2 && Sign((*output)[len(*output)-2], (*output)[len(*output)-1], q.points[i]) {
			*output = (*output)[:len(*output)-1]
		}
		*output = append(*output, q.points[i])
	}
}
