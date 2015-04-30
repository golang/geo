package s2

import (
	"fmt"
	"math"

	"github.com/golang/geo/r1"
)

type EdgeCrosser struct {
	// The fields below are all constant.
	a Point
	b Point

	// The fields below are updated for each vertex in the chain.

	// Previous vertex in the vertex chain.
	c Point
	// The orientation of the triangle ACB.
	acb int
}

func NewEdgeCrosser(a, b, c Point) EdgeCrosser {
	ec := EdgeCrosser{
		a: a,
		b: b,
	}
	ec.RestartAt(c)
	return ec
}

func (ec EdgeCrosser) RestartAt(c Point) {
	ec.c = c
	ec.acb = -int(RobustSign(ec.a, ec.b, c))
}

/**
 * This method is equivalent to calling the S2EdgeUtil.robustCrossing()
 * function (defined below) on the edges AB and CD. It returns +1 if there
 * is a crossing, -1 if there is no crossing, and 0 if two points from
 * different edges are the same. Returns 0 or -1 if either edge is
 * degenerate. As a side effect, it saves vertex D to be used as the next
 * vertex C.
 */
func (ec EdgeCrosser) RobustCrossing(d Point) int {
	// For there to be an edge crossing, the triangles ACB, CBD, BDA, DAC must
	// all be oriented the same way (CW or CCW). We keep the orientation
	// of ACB as part of our state. When each new point D arrives, we
	// compute the orientation of BDA and check whether it matches ACB.
	// This checks whether the points C and D are on opposite sides of the
	// great circle through AB.

	// Recall that robustCCW is invariant with respect to rotating its
	// arguments, i.e. ABC has the same orientation as BDA.
	bda := int(RobustSign(ec.a, ec.b, d))
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
 * This function handles the "slow path" of robustCrossing().
 */
func (ec EdgeCrosser) robustCrossingInternal(d Point) int {
	// ACB and BDA have the appropriate orientations, so now we check the
	// triangles CBD and DAC.
	cbd := -int(RobustSign(ec.c, d, ec.b))
	if cbd != ec.acb {
		return -1
	}

	dac := int(RobustSign(ec.c, d, ec.a))
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

func NewRectBounder() RectBounder {
	return RectBounder{bound: EmptyRect()}
}

func (rb RectBounder) AddPoint(b Point) {
	bLatLng := LatLngFromPoint(b)

	if rb.bound.IsEmpty() {
		fmt.Printf("AddPoint %s\n", rb.bound.String())
		rb.bound = rb.bound.AddPoint(bLatLng)
		fmt.Printf("AddPoint %s\n", rb.bound.String())
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

func (rb RectBounder) GetBound() Rect { return rb.bound }
