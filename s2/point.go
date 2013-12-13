package s2

import (
	"code.google.com/p/gos2/r3"
	"code.google.com/p/gos2/s1"
)

// Point represents a point on the unit sphere as a normalized 3D vector.
//
// Points are guaranteed to be close to normal in the sense that the norm of any points will be very close to 1.
//
// Fields should be treated as read-only. Use one of the factory methods for creation.
type Point struct {
	r3.Vector
}

// PointFromCoords creates a new normalized point from coordinates.
//
// This always returns a valid point. If the given coordinates can not be normalized the origin point will be returned.
func PointFromCoords(x, y, z float64) Point {
	if x == 0 && y == 0 && z == 0 {
		return OriginPoint()
	}
	return Point{r3.Vector{x, y, z}.Normalize()}
}

// OriginPoint returns a unique "origin" on the sphere for operations that need a fixed
// reference point. In particular, this is the "point at infinity" used for
// point-in-polygon testing (by counting the number of edge crossings).
//
// It should *not* be a point that is commonly used in edge tests in order
// to avoid triggering code to handle degenerate cases (this rules out the
// north and south poles). It should also not be on the boundary of any
// low-level S2Cell for the same reason.
func OriginPoint() Point {
	return Point{r3.Vector{0.00456762077230, 0.99947476613078, 0.03208315302933}}
}

// PointCross returns a Point that is orthogonal to both p and op. This is similar to
// p.Cross(op) (the true cross product) except that it does a better job of
// ensuring orthogonality when the Point is nearly parallel to op, it returns
// a non-zero result even when p == op or p == -op and the result is a Point,
// so it will have norm 1.
//
// It satisfies the following properties (f == PointCross):
//
//   (1) f(p, op) != 0 for all p, op
//   (2) f(op,p) == -f(p,op) unless p == op or p == -op
//   (3) f(-p,op) == -f(p,op) unless p == op or p == -op
//   (4) f(p,-op) == -f(p,op) unless p == op or p == -op
func (p Point) PointCross(op Point) Point {
	// NOTE(dnadasi): In the C++ API the equivalent method here was known as "RobustCrossProd",
	// but PointCross more accurately describes how this method is used.
	x := p.Add(op.Vector).Cross(op.Sub(p.Vector))

	if x.ApproxEqual(r3.Vector{0, 0, 0}) {
		// The only result that makes sense mathematically is to return zero, but
		// we find it more convenient to return an arbitrary orthogonal vector.
		return Point{p.Ortho()}
	}

	return Point{x.Normalize()}
}

// CCW returns true if the points A, B, C are strictly counterclockwise,
// and returns false if the points are clockwise or collinear (i.e. if they are all
// contained on some great circle).
//
// Due to numerical errors, situations may arise that are mathematically
// impossible, e.g. ABC may be considered strictly CCW while BCA is not.
// However, the implementation guarantees the following:
//
//   If CCW(a,b,c), then !CCW(c,b,a) for all a,b,c.
func CCW(a, b, c Point) bool {
	// NOTE(dnadasi): In the C++ API the equivalent method here was known as "SimpleCCW",
	// but CCW seems like a fine name at least until the need for a RobustCCW is demonstrated.

	// We compute the signed volume of the parallelepiped ABC. The usual
	// formula for this is (A ⨯ B) · C, but we compute it here using (C ⨯ A) · B
	// in order to ensure that ABC and CBA are not both CCW. This follows
	// from the following identities (which are true numerically, not just
	// mathematically):
	//
	//     (1) x ⨯ y == -(y ⨯ x)
	//     (2) -x · y == -(x · y)
	return c.Cross(a.Vector).Dot(b.Vector) > 0
}

// Distance returns the angle between two points.
func (a Point) Distance(b Point) s1.Angle {
	return a.Vector.Angle(b.Vector)
}

// ApproxEqual reports if the two points are similar enough to be equal.
func (p Point) ApproxEqual(other Point) bool {
	const epsilon = 1e-14
	return p.Vector.Angle(other.Vector) <= epsilon
}

// TODO(dnadasi):
//   - Other CCW methods
//   - Area methods
//   - Centroid methods
