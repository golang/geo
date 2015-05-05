package s2

import (
	"math"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/s1"
)

type EdgeCrosser struct {
	// The fields below are all constant.
	a       Point
	b       Point
	aCrossB Point

	// The fields below are updated for each vertex in the chain.

	// Previous vertex in the vertex chain.
	c Point
	// The orientation of the triangle ACB.
	acb int
}

func NewEdgeCrosser(a, b, c Point) *EdgeCrosser {
	ec := &EdgeCrosser{
		a:       a,
		b:       b,
		aCrossB: Point{a.Cross(b.Vector)},
	}
	ec.RestartAt(c)
	return ec
}

func (ec *EdgeCrosser) RestartAt(c Point) {
	ec.c = c
	ec.acb = -int(RobustCCWWithCross(ec.a, ec.b, ec.c, ec.aCrossB))
}

/**
 * This method is equivalent to calling the S2EdgeUtil.robustCrossing()
 * function (defined below) on the edges AB and CD. It returns +1 if there
 * is a crossing, -1 if there is no crossing, and 0 if two points from
 * different edges are the same. Returns 0 or -1 if either edge is
 * degenerate. As a side effect, it saves vertex D to be used as the next
 * vertex C.
 */
func (ec *EdgeCrosser) RobustCrossing(d Point) int {
	// For there to be an edge crossing, the triangles ACB, CBD, BDA, DAC must
	// all be oriented the same way (CW or CCW). We keep the orientation
	// of ACB as part of our state. When each new point D arrives, we
	// compute the orientation of BDA and check whether it matches ACB.
	// This checks whether the points C and D are on opposite sides of the
	// great circle through AB.

	// Recall that robustCCW is invariant with respect to rotating its
	// arguments, i.e. ABC has the same orientation as BDA.
	bda := int(RobustCCWWithCross(ec.a, ec.b, d, ec.aCrossB))
	var result int

	if bda == -ec.acb && bda != 0 {
		// Most common case -- triangles have opposite orientations.
		result = -1
	} else if (bda & ec.acb) == 0 {
		// At least one value is zero -- two vertices are identical.
		result = 0
	} else {
		// assert (bda == acb && bda != 0);
		result = ec.robustCrossingInternal(d) // Slow path.
	}
	// Now save the current vertex D as the next vertex C, and also save the
	// orientation of the new triangle ACB (which is opposite to the current
	// triangle BDA).
	ec.c = d
	ec.acb = -bda
	return result
}

/**
 * This method is equivalent to the S2EdgeUtil.edgeOrVertexCrossing() method
 * defined below. It is similar to robustCrossing, but handles cases where
 * two vertices are identical in a way that makes it easy to implement
 * point-in-polygon containment tests.
 */
func (ec *EdgeCrosser) EdgeOrVertexCrossing(d Point) bool {
	// We need to copy c since it is clobbered by robustCrossing().
	c2 := PointFromCoordsRaw(ec.c.X, ec.c.Y, ec.c.Z)

	crossing := ec.RobustCrossing(d)
	if crossing < 0 {
		return false
	}
	if crossing > 0 {
		return true
	}

	return VertexCrossing(ec.a, ec.b, c2, d)
}

/**
 * This function handles the "slow path" of robustCrossing().
 */
func (ec *EdgeCrosser) robustCrossingInternal(d Point) int {
	// ACB and BDA have the appropriate orientations, so now we check the
	// triangles CBD and DAC.
	cCrossD := Point{ec.c.Cross(d.Vector)}
	cbd := -int(RobustCCWWithCross(ec.c, d, ec.b, cCrossD))
	if cbd != ec.acb {
		return -1
	}

	dac := int(RobustCCWWithCross(ec.c, d, ec.a, cCrossD))
	if dac == ec.acb {
		return 1
	} else {
		return -1
	}
}

type RectBounder struct {
	a       Point
	aLatLng LatLng
	bound   Rect
}

func NewRectBounder() *RectBounder {
	return &RectBounder{bound: EmptyRect()}
}

func (rb *RectBounder) AddPoint(b Point) {
	bLatLng := LatLngFromPoint(b)

	if rb.bound.IsEmpty() {
		rb.bound = rb.bound.AddPoint(bLatLng)
	} else {
		// We can't just call bound.addPoint(bLatLng) here, since we need to
		// ensure that all the longitudes between "a" and "b" are included.
		rb.bound = rb.bound.Union(RectFromLatLngPair(rb.aLatLng, bLatLng))

		// Check whether the min/max latitude occurs in the edge interior.
		// We find the normal to the plane containing AB, and then a vector
		// "dir" in this plane that also passes through the equator. We use
		// RobustCrossProd to ensure that the edge normal is accurate even
		// when the two points are very close together.
		aCrossB := rb.a.PointCross(b)
		dir := aCrossB.Cross(PointFromCoords(0, 0, 1).Vector)
		da := dir.Dot(rb.a.Vector)
		db := dir.Dot(b.Vector)

		if da*db < 0 {
			// Minimum/maximum latitude occurs in the edge interior. This affects
			// the latitude bounds but not the longitude bounds.
			absLat := math.Acos(math.Abs(aCrossB.Z / aCrossB.Norm()))
			lat := rb.bound.Lat
			if da < 0 {
				// It's possible that absLat < lat.lo() due to numerical errors.
				lat = r1.IntervalFromPointPair(lat.Lo, math.Max(absLat, rb.bound.Lat.Hi))
			} else {
				lat = r1.IntervalFromPointPair(math.Min(-absLat, rb.bound.Lat.Lo), lat.Hi)
			}
			rb.bound = Rect{lat, rb.bound.Lng}
		}
	}

	rb.a = b
	rb.aLatLng = bLatLng
}

func (rb *RectBounder) GetBound() Rect { return rb.bound }

/**
 * Given two edges AB and CD where at least two vertices are identical (i.e.
 * robustCrossing(a,b,c,d) == 0), this function defines whether the two edges
 * "cross" in a such a way that point-in-polygon containment tests can be
 * implemented by counting the number of edge crossings. The basic rule is
 * that a "crossing" occurs if AB is encountered after CD during a CCW sweep
 * around the shared vertex starting from a fixed reference point.
 *
 *  Note that according to this rule, if AB crosses CD then in general CD does
 * not cross AB. However, this leads to the correct result when counting
 * polygon edge crossings. For example, suppose that A,B,C are three
 * consecutive vertices of a CCW polygon. If we now consider the edge
 * crossings of a segment BP as P sweeps around B, the crossing number changes
 * parity exactly when BP crosses BA or BC.
 *
 *  Useful properties of VertexCrossing (VC):
 *
 *  (1) VC(a,a,c,d) == VC(a,b,c,c) == false (2) VC(a,b,a,b) == VC(a,b,b,a) ==
 * true (3) VC(a,b,c,d) == VC(a,b,d,c) == VC(b,a,c,d) == VC(b,a,d,c) (3) If
 * exactly one of a,b equals one of c,d, then exactly one of VC(a,b,c,d) and
 * VC(c,d,a,b) is true
 *
 * It is an error to call this method with 4 distinct vertices.
 */
func VertexCrossing(a, b, c, d Point) bool {
	// If A == B or C == D there is no intersection. We need to check this
	// case first in case 3 or more input points are identical.
	if a.Equals(b) || c.Equals(d) {
		return false
	}

	// If any other pair of vertices is equal, there is a crossing if and only
	// if orderedCCW() indicates that the edge AB is further CCW around the
	// shared vertex than the edge CD.
	if a.Equals(d) {
		return OrderedCCW(Point{a.Ortho()}, c, b, a)
	}
	if b.Equals(c) {
		return OrderedCCW(Point{b.Ortho()}, d, a, b)
	}
	if a.Equals(c) {
		return OrderedCCW(Point{a.Ortho()}, d, b, a)
	}
	if b.Equals(d) {
		return OrderedCCW(Point{b.Ortho()}, c, a, b)
	}

	// assert (false);
	return false
}

/**
 * A wedge relation's test method accepts two edge chains A=(a0,a1,a2) and
 * B=(b0,b1,b2) where a1==b1, and returns either -1, 0, or 1 to indicate the
 * relationship between the region to the left of A and the region to the left
 * of B. Wedge relations are used to determine the local relationship between
 * two polygons that share a common vertex.
 *
 *  All wedge relations require that a0 != a2 and b0 != b2. Other degenerate
 * cases (such as a0 == b2) are handled as expected. The parameter "ab1"
 * denotes the common vertex a1 == b1.
 */
type WedgeRelation interface {
	Test(a0, ab1, a2, b0, b2 Point) int
}

/**
 * Given two edge chains (see WedgeRelation above), this function returns +1
 * if the region to the left of A contains the region to the left of B, and
 * 0 otherwise.
 */
type WedgeContains struct{}

func (w WedgeContains) Test(a0, ab1, a2, b0, b2 Point) int {
	// For A to contain B (where each loop interior is defined to be its left
	// side), the CCW edge order around ab1 must be a2 b2 b0 a0. We split
	// this test into two parts that test three vertices each.
	if OrderedCCW(a2, b2, b0, ab1) && OrderedCCW(b0, a0, a2, ab1) {
		return 1
	}
	return 0
}

/**
 * Given two edge chains (see WedgeRelation above), this function returns -1
 * if the region to the left of A intersects the region to the left of B,
 * and 0 otherwise. Note that regions are defined such that points along a
 * boundary are contained by one side or the other, not both. So for
 * example, if A,B,C are distinct points ordered CCW around a vertex O, then
 * the wedges BOA, AOC, and COB do not intersect.
 */
type WedgeIntersects struct{}

func (w WedgeIntersects) Test(a0, ab1, a2, b0, b2 Point) int {
	// For A not to intersect B (where each loop interior is defined to be
	// its left side), the CCW edge order around ab1 must be a0 b2 b0 a2.
	// Note that it's important to write these conditions as negatives
	// (!OrderedCCW(a,b,c,o) rather than Ordered(c,b,a,o)) to get correct
	// results when two vertices are the same.
	if OrderedCCW(a0, b2, b0, ab1) && OrderedCCW(b0, a2, a0, ab1) {
		return 0
	}
	return -1
}

/**
 * Given two edge chains (see WedgeRelation above), this function returns +1
 * if A contains B, 0 if A and B are disjoint, and -1 if A intersects but
 * does not contain B.
 */
type WedgeContainsOrIntersects struct{}

func (w WedgeContainsOrIntersects) Test(a0, ab1, a2, b0, b2 Point) int {
	// This is similar to WedgeContainsOrCrosses, except that we want to
	// distinguish cases (1) [A contains B], (3) [A and B are disjoint],
	// and (2,4,5,6) [A intersects but does not contain B].

	if OrderedCCW(a0, a2, b2, ab1) {
		// We are in case 1, 5, or 6, or case 2 if a2 == b2.
		if OrderedCCW(b2, b0, a0, ab1) {
			return 1 // Case 1
		}
		return -1 // Case 2,5,6.
	}
	// We are in cases 2, 3, or 4.
	if !OrderedCCW(a2, b0, b2, ab1) {
		return 0 // Case 3.
	}

	// We are in case 2 or 4, or case 3 if a2 == b0.
	if a2.Equals(b0) {
		return 0 // Case 3
	}
	return -1 // Case 2,4.
}

/**
 * Given two edge chains (see WedgeRelation above), this function returns +1
 * if A contains B, 0 if B contains A or the two wedges do not intersect,
 * and -1 if the edge chains A and B cross each other (i.e. if A intersects
 * both the interior and exterior of the region to the left of B). In
 * degenerate cases where more than one of these conditions is satisfied,
 * the maximum possible result is returned. For example, if A == B then the
 * result is +1.
 */
type WedgeContainsOrCrosses struct{}

func (w WedgeContainsOrCrosses) Test(a0, ab1, a2, b0, b2 Point) int {
	// There are 6 possible edge orderings at a shared vertex (all
	// of these orderings are circular, i.e. abcd == bcda):
	//
	// (1) a2 b2 b0 a0: A contains B
	// (2) a2 a0 b0 b2: B contains A
	// (3) a2 a0 b2 b0: A and B are disjoint
	// (4) a2 b0 a0 b2: A and B intersect in one wedge
	// (5) a2 b2 a0 b0: A and B intersect in one wedge
	// (6) a2 b0 b2 a0: A and B intersect in two wedges
	//
	// In cases (4-6), the boundaries of A and B cross (i.e. the boundary
	// of A intersects the interior and exterior of B and vice versa).
	// Thus we want to distinguish cases (1), (2-3), and (4-6).
	//
	// Note that the vertices may satisfy more than one of the edge
	// orderings above if two or more vertices are the same. The tests
	// below are written so that we take the most favorable
	// interpretation, i.e. preferring (1) over (2-3) over (4-6). In
	// particular note that if orderedCCW(a,b,c,o) returns true, it may be
	// possible that orderedCCW(c,b,a,o) is also true (if a == b or b == c).

	if OrderedCCW(a0, a2, b2, ab1) {
		// The cases with this vertex ordering are 1, 5, and 6,
		// although case 2 is also possible if a2 == b2.
		if OrderedCCW(b2, b0, a0, ab1) {
			return 1 // Case 1 (A contains B)
		}

		// We are in case 5 or 6, or case 2 if a2 == b2.
		if a2.Equals(b2) {
			return 0 // Case 2
		}
		return -1 // Case 5,6.
	}
	// We are in case 2, 3, or 4.
	if OrderedCCW(a0, b0, a2, ab1) {
		return 0 // Case 2,3
	}
	return -1 // Case 4.
}

/**
 * Given a point X and an edge AB, return the distance ratio AX / (AX + BX).
 * If X happens to be on the line segment AB, this is the fraction "t" such
 * that X == Interpolate(A, B, t). Requires that A and B are distinct.
 */
func getDistanceFraction(x, a0, a1 Point) float64 {
	if a0.Equals(a1) {
		panic("a0 and a1 are equal")
	}
	d0 := x.Distance(a0).Radians()
	d1 := x.Distance(a1).Radians()
	return d0 / (d0 + d1)
}

/**
 * Return the minimum distance from X to any point on the edge AB. The result
 * is very accurate for small distances but may have some numerical error if
 * the distance is large (approximately Pi/2 or greater). The case A == B is
 * handled correctly. Note: x, a and b must be of unit length. Throws
 * IllegalArgumentException if this is not the case.
 */
func getDistance(x, a, b Point) s1.Angle {
	return getDistanceWithCross(x, a, b, a.PointCross(b))
}

/**
 * A slightly more efficient version of getDistance() where the cross product
 * of the two endpoints has been precomputed. The cross product does not need
 * to be normalized, but should be computed using S2.robustCrossProd() for the
 * most accurate results.
 */
func getDistanceWithCross(x, a, b, aCrossB Point) s1.Angle {
	if !x.IsUnit() || !a.IsUnit() || !b.IsUnit() {
		panic("x, a and b need to be unit length")
	}

	// There are three cases. If X is located in the spherical wedge defined by
	// A, B, and the axis A x B, then the closest point is on the segment AB.
	// Otherwise the closest point is either A or B; the dividing line between
	// these two cases is the great circle passing through (A x B) and the
	// midpoint of AB.

	if simpleCCW(aCrossB, a, x) && simpleCCW(x, b, aCrossB) {
		// The closest point to X lies on the segment AB. We compute the distance
		// to the corresponding great circle. The result is accurate for small
		// distances but not necessarily for large distances (approaching Pi/2).

		sinDist := math.Abs(x.Dot(aCrossB.Vector)) / aCrossB.Norm()
		return s1.Angle(math.Asin(math.Min(1.0, sinDist)))
	}

	// Otherwise, the closest point is either A or B. The cheapest method is
	// just to compute the minimum of the two linear (as opposed to spherical)
	// distances and convert the result to an angle. Again, this method is
	// accurate for small but not large distances (approaching Pi).

	linearDist2 := math.Min(x.Sub(a.Vector).Norm2(), x.Sub(b.Vector).Norm2())
	return s1.Angle(2 * math.Asin(math.Min(1.0, 0.5*math.Sqrt(linearDist2))))
}
