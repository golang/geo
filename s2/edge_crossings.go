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
	"github.com/golang/geo/s1"
)

const (
	// intersectionError can be set somewhat arbitrarily, because the algorithm
	// uses more precision if necessary in order to achieve the specified error.
	// The only strict requirement is that intersectionError >= dblEpsilon
	// radians. However, using a larger error tolerance makes the algorithm more
	// efficient because it reduces the number of cases where exact arithmetic is
	// needed.
	intersectionError = s1.Angle(4 * dblEpsilon)

	// intersectionMergeRadius is used to ensure that intersection points that
	// are supposed to be coincident are merged back together into a single
	// vertex. This is required in order for various polygon operations (union,
	// intersection, etc) to work correctly. It is twice the intersection error
	// because two coincident intersection points might have errors in
	// opposite directions.
	intersectionMergeRadius = 2 * intersectionError
)

// A Crossing indicates how edges cross.
type Crossing int

const (
	// Cross means the edges cross.
	Cross Crossing = iota
	// MaybeCross means two vertices from different edges are the same.
	MaybeCross
	// DoNotCross means the edges do not cross.
	DoNotCross
)

// SimpleCrossing reports whether edge AB crosses CD at a point that is interior
// to both edges. Properties:
//
//  (1) SimpleCrossing(b,a,c,d) == SimpleCrossing(a,b,c,d)
//  (2) SimpleCrossing(c,d,a,b) == SimpleCrossing(a,b,c,d)
func SimpleCrossing(a, b, c, d Point) bool {
	// We compute the equivalent of Sign for triangles ACB, CBD, BDA,
	// and DAC. All of these triangles need to have the same orientation
	// (CW or CCW) for an intersection to exist.

	ab := a.Vector.Cross(b.Vector)
	acb := -(ab.Dot(c.Vector))
	bda := ab.Dot(d.Vector)
	if acb*bda <= 0 {
		return false
	}

	cd := c.Vector.Cross(d.Vector)
	cbd := -(cd.Dot(b.Vector))
	dac := cd.Dot(a.Vector)
	return (acb*cbd > 0) && (acb*dac > 0)
}

// VertexCrossing reports whether two edges "cross" in such a way that point-in-polygon
// containment tests can be implemented by counting the number of edge crossings.
//
// Given two edges AB and CD where at least two vertices are identical
// (i.e. CrossingSign(a,b,c,d) == 0), the basic rule is that a "crossing"
// occurs if AB is encountered after CD during a CCW sweep around the shared
// vertex starting from a fixed reference point.
//
// Note that according to this rule, if AB crosses CD then in general CD
// does not cross AB. However, this leads to the correct result when
// counting polygon edge crossings. For example, suppose that A,B,C are
// three consecutive vertices of a CCW polygon. If we now consider the edge
// crossings of a segment BP as P sweeps around B, the crossing number
// changes parity exactly when BP crosses BA or BC.
//
// Useful properties of VertexCrossing (VC):
//
//  (1) VC(a,a,c,d) == VC(a,b,c,c) == false
//  (2) VC(a,b,a,b) == VC(a,b,b,a) == true
//  (3) VC(a,b,c,d) == VC(a,b,d,c) == VC(b,a,c,d) == VC(b,a,d,c)
//  (3) If exactly one of a,b equals one of c,d, then exactly one of
//      VC(a,b,c,d) and VC(c,d,a,b) is true
//
// It is an error to call this method with 4 distinct vertices.
func VertexCrossing(a, b, c, d Point) bool {
	// If A == B or C == D there is no intersection. We need to check this
	// case first in case 3 or more input points are identical.
	if a.ApproxEqual(b) || c.ApproxEqual(d) {
		return false
	}

	// If any other pair of vertices is equal, there is a crossing if and only
	// if OrderedCCW indicates that the edge AB is further CCW around the
	// shared vertex O (either A or B) than the edge CD, starting from an
	// arbitrary fixed reference point.
	switch {
	case a.ApproxEqual(d):
		return OrderedCCW(Point{a.Ortho()}, c, b, a)
	case b.ApproxEqual(c):
		return OrderedCCW(Point{b.Ortho()}, d, a, b)
	case a.ApproxEqual(c):
		return OrderedCCW(Point{a.Ortho()}, d, b, a)
	case b.ApproxEqual(d):
		return OrderedCCW(Point{b.Ortho()}, c, a, b)
	}

	return false
}

// TODO(roberts): Differences from C++
// Intersection related methods
