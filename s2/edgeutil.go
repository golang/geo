/*
Copyright 2015 Google Inc. All rights reserved.

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
	"math"

	"github.com/golang/geo/s1"
)

var (
	// edgeClipErrorUVCoord is the maximum error in a u- or v-coordinate
	// compared to the exact result, assuming that the points A and B are in
	// the rectangle [-1,1]x[1,1] or slightly outside it (by 1e-10 or less).
	edgeClipErrorUVCoord = 2.25 * dblEpsilon

	// faceClipErrorUVCoord is the maximum angle between a returned vertex
	// and the nearest point on the exact edge AB expressed as the maximum error
	// in an individual u- or v-coordinate. In other words, for each
	// returned vertex there is a point on the exact edge AB whose u- and
	// v-coordinates differ from the vertex by at most this amount.
	faceClipErrorUVCoord = 9.0 * (1.0 / math.Sqrt2) * dblEpsilon
)

// SimpleCrossing reports whether edge AB crosses CD at a point that is interior
// to both edges. Properties:
//
//  (1) SimpleCrossing(b,a,c,d) == SimpleCrossing(a,b,c,d)
//  (2) SimpleCrossing(c,d,a,b) == SimpleCrossing(a,b,c,d)
func SimpleCrossing(a, b, c, d Point) bool {
	// We compute the equivalent of SimpleCCW() for triangles ACB, CBD, BDA,
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
// (i.e. RobustCrossing(a,b,c,d) == 0), the basic rule is that a "crossing"
// occurs if AB is encountered after CD during a CCW sweep around the shared
// vertex starting from a fixed reference point.
//
// Note that according to this rule, if AB crosses CD then in general CD
// does not cross AB.  However, this leads to the correct result when
// counting polygon edge crossings.  For example, suppose that A,B,C are
// three consecutive vertices of a CCW polygon.  If we now consider the edge
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
	// if OrderedCCW() indicates that the edge AB is further CCW around the
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

// Interpolate returns the point X along the line segment AB whose distance from A
// is the given fraction "t" of the distance AB. Does NOT require that "t" be
// between 0 and 1. Note that all distances are measured on the surface of
// the sphere, so this is more complicated than just computing (1-t)*a + t*b
// and normalizing the result.
func Interpolate(t float64, a, b Point) Point {
	if t == 0 {
		return a
	}
	if t == 1 {
		return b
	}
	ab := a.Angle(b.Vector)
	return InterpolateAtDistance(s1.Angle(t)*ab, a, b)
}

// InterpolateAtDistance returns the point X along the line segment AB whose
// distance from A is the angle ax.
func InterpolateAtDistance(ax s1.Angle, a, b Point) Point {
	aRad := ax.Radians()

	// Use PointCross to compute the tangent vector at A towards B. The
	// result is always perpendicular to A, even if A=B or A=-B, but it is not
	// necessarily unit length. (We effectively normalize it below.)
	normal := a.PointCross(b)
	tangent := normal.Vector.Cross(a.Vector)

	// Now compute the appropriate linear combination of A and "tangent". With
	// infinite precision the result would always be unit length, but we
	// normalize it anyway to ensure that the error is within acceptable bounds.
	// (Otherwise errors can build up when the result of one interpolation is
	// fed into another interpolation.)
	return Point{(a.Mul(math.Cos(aRad)).Add(tangent.Mul(math.Sin(aRad) / tangent.Norm()))).Normalize()}
}
