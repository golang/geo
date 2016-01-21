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

	"github.com/golang/geo/r1"
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

// RectBounder is used to compute a bounding rectangle that contains all edges
// defined by a vertex chain (v0, v1, v2, ...). All vertices must be unit length.
// Note that the bounding rectangle of an edge can be larger than the bounding
// rectangle of its endpoints, e.g. consider an edge that passes through the North Pole.
//
// The bounds are calculated conservatively to account for numerical errors
// when points are converted to LatLngs. More precisely, this function
// guarantees the following:
// Let L be a closed edge chain (Loop) such that the interior of the loop does
// not contain either pole. Now if P is any point such that L.ContainsPoint(P),
// then RectBound(L).ContainsPoint(LatLngFromPoint(P)).
type RectBounder struct {
	// The previous vertex in the chain.
	a Point
	// The previous vertex latitude longitude.
	aLL   LatLng
	bound Rect
}

// NewRectBounder returns a new instance of a RectBounder.
func NewRectBounder() *RectBounder {
	return &RectBounder{
		bound: EmptyRect(),
	}
}

// AddPoint adds the given point to the chain. The Point must be unit length.
func (r *RectBounder) AddPoint(b Point) {
	bLL := LatLngFromPoint(b)

	if r.bound.IsEmpty() {
		r.a = b
		r.aLL = bLL
		r.bound = r.bound.AddPoint(bLL)
		return
	}

	// First compute the cross product N = A x B robustly. This is the normal
	// to the great circle through A and B. We don't use RobustSign
	// since that method returns an arbitrary vector orthogonal to A if the two
	// vectors are proportional, and we want the zero vector in that case.
	n := r.a.Sub(b.Vector).Cross(r.a.Add(b.Vector)) // N = 2 * (A x B)

	// The relative error in N gets large as its norm gets very small (i.e.,
	// when the two points are nearly identical or antipodal). We handle this
	// by choosing a maximum allowable error, and if the error is greater than
	// this we fall back to a different technique. Since it turns out that
	// the other sources of error add up to at most 1.16 * dblEpsilon, and it
	// is desirable to have the total error be a multiple of dblEpsilon, we
	// have chosen the maximum error threshold here to be 3.84 * dblEpsilon.
	// It is possible to show that the error is less than this when
	//
	// n.Norm() >= 8 * sqrt(3) / (3.84 - 0.5 - sqrt(3)) * dblEpsilon
	//          = 1.91346e-15 (about 8.618 * dblEpsilon)
	nNorm := n.Norm()
	if nNorm < 1.91346e-15 {
		// A and B are either nearly identical or nearly antipodal (to within
		// 4.309 * dblEpsilon, or about 6 nanometers on the earth's surface).
		if r.a.Dot(b.Vector) < 0 {
			// The two points are nearly antipodal. The easiest solution is to
			// assume that the edge between A and B could go in any direction
			// around the sphere.
			r.bound = FullRect()
		} else {
			// The two points are nearly identical (to within 4.309 * dblEpsilon).
			// In this case we can just use the bounding rectangle of the points,
			// since after the expansion done by GetBound this Rect is
			// guaranteed to include the (lat,lng) values of all points along AB.
			r.bound = r.bound.Union(RectFromLatLng(r.aLL).AddPoint(bLL))
		}
		r.a = b
		r.aLL = bLL
		return
	}

	// Compute the longitude range spanned by AB.
	lngAB := s1.EmptyInterval().AddPoint(r.aLL.Lng.Radians()).AddPoint(bLL.Lng.Radians())
	if lngAB.Length() >= math.Pi-2*dblEpsilon {
		// The points lie on nearly opposite lines of longitude to within the
		// maximum error of the calculation. The easiest solution is to assume
		// that AB could go on either side of the pole.
		lngAB = s1.FullInterval()
	}

	// Next we compute the latitude range spanned by the edge AB. We start
	// with the range spanning the two endpoints of the edge:
	latAB := r1.IntervalFromPoint(r.aLL.Lat.Radians()).AddPoint(bLL.Lat.Radians())

	// This is the desired range unless the edge AB crosses the plane
	// through N and the Z-axis (which is where the great circle through A
	// and B attains its minimum and maximum latitudes). To test whether AB
	// crosses this plane, we compute a vector M perpendicular to this
	// plane and then project A and B onto it.
	m := n.Cross(PointFromCoords(0, 0, 1).Vector)
	mA := m.Dot(r.a.Vector)
	mB := m.Dot(b.Vector)

	// We want to test the signs of "mA" and "mB", so we need to bound
	// the error in these calculations. It is possible to show that the
	// total error is bounded by
	//
	// (1 + sqrt(3)) * dblEpsilon * nNorm + 8 * sqrt(3) * (dblEpsilon**2)
	//   = 6.06638e-16 * nNorm + 6.83174e-31

	mError := 6.06638e-16*nNorm + 6.83174e-31
	if mA*mB < 0 || math.Abs(mA) <= mError || math.Abs(mB) <= mError {
		// Minimum/maximum latitude *may* occur in the edge interior.
		//
		// The maximum latitude is 90 degrees minus the latitude of N. We
		// compute this directly using atan2 in order to get maximum accuracy
		// near the poles.
		//
		// Our goal is compute a bound that contains the computed latitudes of
		// all S2Points P that pass the point-in-polygon containment test.
		// There are three sources of error we need to consider:
		// - the directional error in N (at most 3.84 * dblEpsilon)
		// - converting N to a maximum latitude
		// - computing the latitude of the test point P
		// The latter two sources of error are at most 0.955 * dblEpsilon
		// individually, but it is possible to show by a more complex analysis
		// that together they can add up to at most 1.16 * dblEpsilon, for a
		// total error of 5 * dblEpsilon.
		//
		// We add 3 * dblEpsilon to the bound here, and GetBound() will pad
		// the bound by another 2 * dblEpsilon.
		maxLat := math.Min(
			math.Atan2(math.Sqrt(n.X*n.X+n.Y*n.Y), math.Abs(n.Z))+3*dblEpsilon,
			math.Pi/2)

		// In order to get tight bounds when the two points are close together,
		// we also bound the min/max latitude relative to the latitudes of the
		// endpoints A and B. First we compute the distance between A and B,
		// and then we compute the maximum change in latitude between any two
		// points along the great circle that are separated by this distance.
		// This gives us a latitude change "budget". Some of this budget must
		// be spent getting from A to B; the remainder bounds the round-trip
		// distance (in latitude) from A or B to the min or max latitude
		// attained along the edge AB.
		latBudget := 2 * math.Asin(0.5*(r.a.Sub(b.Vector)).Norm()*math.Sin(maxLat))
		maxDelta := 0.5*(latBudget-latAB.Length()) + dblEpsilon

		// Test whether AB passes through the point of maximum latitude or
		// minimum latitude. If the dot product(s) are small enough then the
		// result may be ambiguous.
		if mA <= mError && mB >= -mError {
			latAB.Hi = math.Min(maxLat, latAB.Hi+maxDelta)
		}
		if mB <= mError && mA >= -mError {
			latAB.Lo = math.Max(-maxLat, latAB.Lo-maxDelta)
		}
	}
	r.a = b
	r.aLL = bLL
	r.bound = r.bound.Union(Rect{latAB, lngAB})
}

// RectBound returns the bounding rectangle of the edge chain that connects the
// vertices defined so far. This bound satisfies the guarantee made
// above, i.e. if the edge chain defines a Loop, then the bound contains
// the LatLng coordinates of all Points contained by the loop.
func (r *RectBounder) RectBound() Rect {
	return r.bound.expanded(LatLng{s1.Angle(2 * dblEpsilon), 0}).PolarClosure()
}
