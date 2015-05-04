package s2

import (
	"math"

	"github.com/golang/geo/r2"
)

const EPSILON float64 = 1e-14

// Number of bits in the mantissa of a double.
const EXPONENT_SHIFT uint = 52

// Mask to extract the exponent from a double.
const EXPONENT_MASK uint64 = 0x7ff0000000000000

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func exp(v float64) int {
	if v == 0 {
		return 0
	}
	bits := math.Float64bits(v)
	return (int)((EXPONENT_MASK&bits)>>EXPONENT_SHIFT) - 1022
}

/**
 * Return the angle at the vertex B in the triangle ABC. The return value is
 * always in the range [0, Pi]. The points do not need to be normalized.
 * Ensures that Angle(a,b,c) == Angle(c,b,a) for all a,b,c.
 *
 *  The angle is undefined if A or C is diametrically opposite from B, and
 * becomes numerically unstable as the length of edge AB or BC approaches 180
 * degrees.
 */
func angle(a, b, c Point) float64 {
	return a.Cross(b.Vector).Angle(c.Cross(b.Vector)).Radians()
}

func approxEqualsNumber(a, b, maxError float64) bool {
	return math.Abs(a-b) <= maxError
}

/**
 * Return the area of triangle ABC. The method used is about twice as
 * expensive as Girard's formula, but it is numerically stable for both large
 * and very small triangles. The points do not need to be normalized. The area
 * is always positive.
 *
 *  The triangle area is undefined if it contains two antipodal points, and
 * becomes numerically unstable as the length of any edge approaches 180
 * degrees.
 */
func Area(a, b, c Point) float64 {
	// This method is based on l'Huilier's theorem,
	//
	// tan(E/4) = sqrt(tan(s/2) tan((s-a)/2) tan((s-b)/2) tan((s-c)/2))
	//
	// where E is the spherical excess of the triangle (i.e. its area),
	// a, b, c, are the side lengths, and
	// s is the semiperimeter (a + b + c) / 2 .
	//
	// The only significant source of error using l'Huilier's method is the
	// cancellation error of the terms (s-a), (s-b), (s-c). This leads to a
	// *relative* error of about 1e-16 * s / min(s-a, s-b, s-c). This compares
	// to a relative error of about 1e-15 / E using Girard's formula, where E is
	// the true area of the triangle. Girard's formula can be even worse than
	// this for very small triangles, e.g. a triangle with a true area of 1e-30
	// might evaluate to 1e-5.
	//
	// So, we prefer l'Huilier's formula unless dmin < s * (0.1 * E), where
	// dmin = min(s-a, s-b, s-c). This basically includes all triangles
	// except for extremely long and skinny ones.
	//
	// Since we don't know E, we would like a conservative upper bound on
	// the triangle area in terms of s and dmin. It's possible to show that
	// E <= k1 * s * sqrt(s * dmin), where k1 = 2*sqrt(3)/Pi (about 1).
	// Using this, it's easy to show that we should always use l'Huilier's
	// method if dmin >= k2 * s^5, where k2 is about 1e-2. Furthermore,
	// if dmin < k2 * s^5, the triangle area is at most k3 * s^4, where
	// k3 is about 0.1. Since the best case error using Girard's formula
	// is about 1e-15, this means that we shouldn't even consider it unless
	// s >= 3e-4 or so.

	// We use volatile doubles to force the compiler to truncate all of these
	// quantities to 64 bits. Otherwise it may compute a value of dmin > 0
	// simply because it chose to spill one of the intermediate values to
	// memory but not one of the others.
	sa := b.Angle(c.Vector).Radians()
	sb := c.Angle(a.Vector).Radians()
	sc := a.Angle(b.Vector).Radians()
	s := 0.5 * (sa + sb + sc)
	if s >= 3e-4 {
		// Consider whether Girard's formula might be more accurate.
		s2 := s * s
		dmin := s - math.Max(sa, math.Max(sb, sc))
		if dmin < 1e-2*s*s2*s2 {
			// This triangle is skinny enough to consider Girard's formula.
			area := GirardArea(a, b, c)
			if dmin < s*(0.1*area) {
				return area
			}
		}
	}
	// Use l'Huilier's formula.
	return 4 * math.Atan(
		math.Sqrt(
			math.Max(0.0, math.Tan(0.5*s)*math.Tan(0.5*(s-sa))*math.Tan(0.5*(s-sb))*math.Tan(0.5*(s-sc)))))
}

/**
 * Return the area of the triangle computed using Girard's formula. This is
 * slightly faster than the Area() method above is not accurate for very small
 * triangles.
 */
func GirardArea(a, b, c Point) float64 {
	// This is equivalent to the usual Girard's formula but is slightly
	// more accurate, faster to compute, and handles a == b == c without
	// a special case.

	ab := a.Cross(b.Vector)
	bc := b.Cross(c.Vector)
	ac := a.Cross(c.Vector)
	return math.Max(0.0, ab.Angle(ac).Radians()-ab.Angle(bc).Radians()+bc.Angle(ac).Radians())
}

/**
 * Like Area(), but returns a positive value for counterclockwise triangles
 * and a negative value otherwise.
 */
func SignedArea(a, b, c Point) float64 {
	return Area(a, b, c) * float64(RobustCCW(a, b, c))
}

// About centroids:
// ----------------
//
// There are several notions of the "centroid" of a triangle. First, there
// // is the planar centroid, which is simply the centroid of the ordinary
// (non-spherical) triangle defined by the three vertices. Second, there is
// the surface centroid, which is defined as the intersection of the three
// medians of the spherical triangle. It is possible to show that this
// point is simply the planar centroid projected to the surface of the
// sphere. Finally, there is the true centroid (mass centroid), which is
// defined as the area integral over the spherical triangle of (x,y,z)
// divided by the triangle area. This is the point that the triangle would
// rotate around if it was spinning in empty space.
//
// The best centroid for most purposes is the true centroid. Unlike the
// planar and surface centroids, the true centroid behaves linearly as
// regions are added or subtracted. That is, if you split a triangle into
// pieces and compute the average of their centroids (weighted by triangle
// area), the result equals the centroid of the original triangle. This is
// not true of the other centroids.
//
// Also note that the surface centroid may be nowhere near the intuitive
// "center" of a spherical triangle. For example, consider the triangle
// with vertices A=(1,eps,0), B=(0,0,1), C=(-1,eps,0) (a quarter-sphere).
// The surface centroid of this triangle is at S=(0, 2*eps, 1), which is
// within a distance of 2*eps of the vertex B. Note that the median from A
// (the segment connecting A to the midpoint of BC) passes through S, since
// this is the shortest path connecting the two endpoints. On the other
// hand, the true centroid is at M=(0, 0.5, 0.5), which when projected onto
// the surface is a much more reasonable interpretation of the "center" of
// this triangle.

/**
 * Return the centroid of the planar triangle ABC. This can be normalized to
 * unit length to obtain the "surface centroid" of the corresponding spherical
 * triangle, i.e. the intersection of the three medians. However, note that
 * for large spherical triangles the surface centroid may be nowhere near the
 * intuitive "center" (see example above).
 */
func PlanarCentroid(a, b, c Point) Point {
	return PointFromCoords((a.X+b.X+c.X)/3.0, (a.Y+b.Y+c.Y)/3.0, (a.Z+b.Z+c.Z)/3.0)
}

/**
 * Returns the true centroid of the spherical triangle ABC multiplied by the
 * signed area of spherical triangle ABC. The reasons for multiplying by the
 * signed area are (1) this is the quantity that needs to be summed to compute
 * the centroid of a union or difference of triangles, and (2) it's actually
 * easier to calculate this way.
 */
func TrueCentroid(a, b, c Point) Point {
	// I couldn't find any references for computing the true centroid of a
	// spherical triangle... I have a truly marvellous demonstration of this
	// formula which this margin is too narrow to contain :)

	// assert (isUnitLength(a) && isUnitLength(b) && isUnitLength(c));
	sina := b.Cross(c.Vector).Norm()
	sinb := c.Cross(a.Vector).Norm()
	sinc := a.Cross(b.Vector).Norm()
	ra := math.Asin(sina) / sina
	rb := math.Asin(sinb) / sinb
	rc := math.Asin(sinc) / sinc
	if sina == 0 {
		ra = 1
	}
	if sinb == 0 {
		rb = 1
	}
	if sinc == 0 {
		rc = 1
	}

	// Now compute a point M such that M.X = rX * det(ABC) / 2 for X in A,B,C.
	x := PointFromCoords(a.X, b.X, c.X)
	y := PointFromCoords(a.Y, b.Y, c.Y)
	z := PointFromCoords(a.Z, b.Z, c.Z)
	r := PointFromCoords(ra, rb, rc)
	return PointFromCoords(
		0.5*y.Cross(z.Vector).Dot(r.Vector),
		0.5*z.Cross(x.Vector).Dot(r.Vector),
		0.5*x.Cross(y.Vector).Dot(r.Vector),
	)
}

/**
 * Return true if the points A, B, C are strictly counterclockwise. Return
 * false if the points are clockwise or colinear (i.e. if they are all
 * contained on some great circle).
 *
 *  Due to numerical errors, situations may arise that are mathematically
 * impossible, e.g. ABC may be considered strictly CCW while BCA is not.
 * However, the implementation guarantees the following:
 *
 *  If SimpleCCW(a,b,c), then !SimpleCCW(c,b,a) for all a,b,c.
 *
 * In other words, ABC and CBA are guaranteed not to be both CCW
 */
func simpleCCW(a, b, c Point) bool {
	// We compute the signed volume of the parallelepiped ABC. The usual
	// formula for this is (AxB).C, but we compute it here using (CxA).B
	// in order to ensure that ABC and CBA are not both CCW. This follows
	// from the following identities (which are true numerically, not just
	// mathematically):
	//
	// (1) x.CrossProd(y) == -(y.CrossProd(x))
	// (2) (-x).DotProd(y) == -(x.DotProd(y))

	return c.Cross(a.Vector).Dot(b.Vector) > 0
}

/**
 * WARNING! This requires arbitrary precision arithmetic to be truly robust.
 * This means that for nearly colinear AB and AC, this function may return the
 * wrong answer.
 *
 * <p>
 * Like SimpleCCW(), but returns +1 if the points are counterclockwise and -1
 * if the points are clockwise. It satisfies the following conditions:
 *
 *  (1) RobustCCW(a,b,c) == 0 if and only if a == b, b == c, or c == a (2)
 * RobustCCW(b,c,a) == RobustCCW(a,b,c) for all a,b,c (3) RobustCCW(c,b,a)
 * ==-RobustCCW(a,b,c) for all a,b,c
 *
 *  In other words:
 *
 *  (1) The result is zero if and only if two points are the same. (2)
 * Rotating the order of the arguments does not affect the result. (3)
 * Exchanging any two arguments inverts the result.
 *
 *  This function is essentially like taking the sign of the determinant of
 * a,b,c, except that it has additional logic to make sure that the above
 * properties hold even when the three points are coplanar, and to deal with
 * the limitations of floating-point arithmetic.
 *
 *  Note: a, b and c are expected to be of unit length. Otherwise, the results
 * are undefined.
 */
func RobustCCW(a, b, c Point) int {
	return RobustCCWWithCross(a, b, c, Point{a.Cross(b.Vector)})
}

/**
 * A more efficient version of RobustCCW that allows the precomputed
 * cross-product of A and B to be specified.
 *
 *  Note: a, b and c are expected to be of unit length. Otherwise, the results
 * are undefined
 */
func RobustCCWWithCross(a, b, c, aCrossB Point) int {
	// assert (isUnitLength(a) && isUnitLength(b) && isUnitLength(c));

	// There are 14 multiplications and additions to compute the determinant
	// below. Since all three points are normalized, it is possible to show
	// that the average rounding error per operation does not exceed 2**-54,
	// the maximum rounding error for an operation whose result magnitude is in
	// the range [0.5,1). Therefore, if the absolute value of the determinant
	// is greater than 2*14*(2**-54), the determinant will have the same sign
	// even if the arguments are rotated (which produces a mathematically
	// equivalent result but with potentially different rounding errors).
	kMinAbsValue := 1.6e-15 // 2 * 14 * 2**-54

	det := aCrossB.Dot(c.Vector)

	// Double-check borderline cases in debug mode.
	// assert ((Math.abs(det) < kMinAbsValue) || (Math.abs(det) > 1000 * kMinAbsValue)
	//    || (det * expensiveCCW(a, b, c) > 0));

	if det > kMinAbsValue {
		return 1
	}

	if det < -kMinAbsValue {
		return -1
	}

	return ExpensiveCCW(a, b, c)
}

/**
 * A relatively expensive calculation invoked by RobustCCW() if the sign of
 * the determinant is uncertain.
 */
func ExpensiveCCW(a, b, c Point) int {
	// Return zero if and only if two points are the same. This ensures (1).
	if a.Equals(b) || b.Equals(c) || c.Equals(a) {
		return 0
	}

	// Now compute the determinant in a stable way. Since all three points are
	// unit length and we know that the determinant is very close to zero, this
	// means that points are very nearly colinear. Furthermore, the most common
	// situation is where two points are nearly identical or nearly antipodal.
	// To get the best accuracy in this situation, it is important to
	// immediately reduce the magnitude of the arguments by computing either
	// A+B or A-B for each pair of points. Note that even if A and B differ
	// only in their low bits, A-B can be computed very accurately. On the
	// other hand we can't accurately represent an arbitrary linear combination
	// of two vectors as would be required for Gaussian elimination. The code
	// below chooses the vertex opposite the longest edge as the "origin" for
	// the calculation, and computes the different vectors to the other two
	// vertices. This minimizes the sum of the lengths of these vectors.
	//
	// This implementation is very stable numerically, but it still does not
	// return consistent results in all cases. For example, if three points are
	// spaced far apart from each other along a great circle, the sign of the
	// result will basically be random (although it will still satisfy the
	// conditions documented in the header file). The only way to return
	// consistent results in all cases is to compute the result using
	// arbitrary-precision arithmetic. I considered using the Gnu MP library,
	// but this would be very expensive (up to 2000 bits of precision may be
	// needed to store the intermediate results) and seems like overkill for
	// this problem. The MP library is apparently also quite particular about
	// compilers and compilation options and would be a pain to maintain.

	// We want to handle the case of nearby points and nearly antipodal points
	// accurately, so determine whether A+B or A-B is smaller in each case.
	var sab float64 = 1
	var sbc float64 = 1
	var sca float64 = 1
	if a.Dot(b.Vector) > 0 {
		sab = -1
	}
	if b.Dot(c.Vector) > 0 {
		sbc = -1
	}
	if c.Dot(a.Vector) > 0 {
		sca = -1
	}
	vab := a.Add(b.Mul(sab))
	vbc := b.Add(c.Mul(sbc))
	vca := c.Add(a.Mul(sca))
	dab := vab.Norm2()
	dbc := vbc.Norm2()
	dca := vca.Norm2()

	// Sort the difference vectors to find the longest edge, and use the
	// opposite vertex as the origin. If two difference vectors are the same
	// length, we break ties deterministically to ensure that the symmetry
	// properties guaranteed in the header file will be true.
	var sign float64
	if dca < dbc || (dca == dbc && a.LessThan(b)) {
		if dab < dbc || (dab == dbc && a.LessThan(c)) {
			// The "sab" factor converts A +/- B into B +/- A.
			sign = vab.Cross(vca).Dot(a.Vector) * sab // BC is longest
			// edge
		} else {
			sign = vca.Cross(vbc).Dot(c.Vector) * sca // AB is longest
			// edge
		}
	} else {
		if dab < dca || (dab == dca && b.LessThan(c)) {
			sign = vbc.Cross(vab).Dot(b.Vector) * sbc // CA is longest
			// edge
		} else {
			sign = vca.Cross(vbc).Dot(c.Vector) * sca // AB is longest
			// edge
		}
	}
	if sign > 0 {
		return 1
	}
	if sign < 0 {
		return -1
	}

	// The points A, B, and C are numerically indistinguishable from coplanar.
	// This may be due to roundoff error, or the points may in fact be exactly
	// coplanar. We handle this situation by perturbing all of the points by a
	// vector (eps, eps**2, eps**3) where "eps" is an infinitesmally small
	// positive number (e.g. 1 divided by a googolplex). The perturbation is
	// done symbolically, i.e. we compute what would happen if the points were
	// perturbed by this amount. It turns out that this is equivalent to
	// checking whether the points are ordered CCW around the origin first in
	// the Y-Z plane, then in the Z-X plane, and then in the X-Y plane.

	ccw := PlanarOrderedCCW(r2.Vector{a.Y, a.Z}, r2.Vector{b.Y, b.Z}, r2.Vector{c.Y, c.Z})
	if ccw == 0 {
		ccw = PlanarOrderedCCW(r2.Vector{a.Z, a.X}, r2.Vector{b.Z, b.X}, r2.Vector{c.Z, c.X})
		if ccw == 0 {
			ccw = PlanarOrderedCCW(
				r2.Vector{a.X, a.Y}, r2.Vector{b.X, b.Y}, r2.Vector{c.X, c.Y})
			// assert (ccw != 0);
		}
	}
	return ccw
}

func PlanarCCW(a, b r2.Vector) int {
	// Return +1 if the edge AB is CCW around the origin, etc.
	var sab float64 = 1
	if a.Dot(b) > 0 {
		sab = -1
	}
	vab := a.Add(b.Mul(sab))
	da := a.Norm2()
	db := b.Norm2()
	var sign float64
	if da < db || (da == db && a.LessThan(b)) {
		sign = a.Cross(vab) * sab
	} else {
		sign = vab.Cross(b)
	}
	if sign > 0 {
		return 1
	}
	if sign < 0 {
		return -1
	}
	return 0
}

func PlanarOrderedCCW(a, b, c r2.Vector) int {
	sum := 0
	sum += PlanarCCW(a, b)
	sum += PlanarCCW(b, c)
	sum += PlanarCCW(c, a)
	if sum > 0 {
		return 1
	}
	if sum < 0 {
		return -1
	}
	return 0
}

/**
 * Return true if the edges OA, OB, and OC are encountered in that order while
 * sweeping CCW around the point O. You can think of this as testing whether
 * A <= B <= C with respect to a continuous CCW ordering around O.
 *
 * Properties:
 * <ol>
 *   <li>If orderedCCW(a,b,c,o) && orderedCCW(b,a,c,o), then a == b</li>
 *   <li>If orderedCCW(a,b,c,o) && orderedCCW(a,c,b,o), then b == c</li>
 *   <li>If orderedCCW(a,b,c,o) && orderedCCW(c,b,a,o), then a == b == c</li>
 *   <li>If a == b or b == c, then orderedCCW(a,b,c,o) is true</li>
 *   <li>Otherwise if a == c, then orderedCCW(a,b,c,o) is false</li>
 * </ol>
 */
func OrderedCCW(a, b, c, o Point) bool {
	// The last inequality below is ">" rather than ">=" so that we return true
	// if A == B or B == C, and otherwise false if A == C. Recall that
	// RobustCCW(x,y,z) == -RobustCCW(z,y,x) for all x,y,z.

	sum := 0
	if RobustCCW(b, o, a) >= 0 {
		sum++
	}
	if RobustCCW(c, o, b) >= 0 {
		sum++
	}
	if RobustCCW(a, o, c) > 0 {
		sum++
	}
	return sum >= 2
}

// Defines an area or a length cell metric.
type Metric struct {
	deriv float64
	dim   uint
}

// Defines a cell metric of the given dimension (1 == length, 2 == area).
func NewMetric(dim uint, deriv float64) Metric {
	return Metric{deriv, dim}
}

// The "deriv" value of a metric is a derivative, and must be multiplied by
// a length or area in (s,t)-space to get a useful value.
func (m Metric) Deriv() float64 { return m.deriv }

// Return the value of a metric for cells at the given level.
func (m Metric) GetValue(level int) float64 {
	return math.Pow(m.deriv, float64(int(m.dim)*(1-level)))
}

/**
 * Return the level at which the metric has approximately the given value.
 * For example, S2::kAvgEdge.GetClosestLevel(0.1) returns the level at which
 * the average cell edge length is approximately 0.1. The return value is
 * always a valid level.
 */
func (m Metric) getClosestLevel(value float64) int {
	if m.dim == 1 {
		return m.getMinLevel(math.Sqrt2 * value)
	}
	return m.getMinLevel(2 * value)
}

/**
 * Return the minimum level such that the metric is at most the given value,
 * or S2CellId::kMaxLevel if there is no such level. For example,
 * S2::kMaxDiag.GetMinLevel(0.1) returns the minimum level such that all
 * cell diagonal lengths are 0.1 or smaller. The return value is always a
 * valid level.
 */
func (m Metric) getMinLevel(value float64) int {
	if value <= 0 {
		return MAX_LEVEL
	}

	// This code is equivalent to computing a floating-point "level"
	// value and rounding up.
	exponent := exp(value / (float64(int(1)<<m.dim) * m.deriv))
	level := max(0, min(MAX_LEVEL, -((exponent-1)>>(m.dim-1))))
	// assert (level == S2CellId.MAX_LEVEL || getValue(level) <= value);
	// assert (level == 0 || getValue(level - 1) > value);
	return level
}

/**
 * Return the maximum level such that the metric is at least the given
 * value, or zero if there is no such level. For example,
 * S2.kMinWidth.GetMaxLevel(0.1) returns the maximum level such that all
 * cells have a minimum width of 0.1 or larger. The return value is always a
 * valid level.
 */
func (m Metric) getMaxLevel(value float64) int {
	if value <= 0 {
		return MAX_LEVEL
	}

	// This code is equivalent to computing a floating-point "level"
	// value and rounding down.
	exponent := exp(float64(int(1)<<m.dim) * m.deriv / value)
	level := max(0, min(MAX_LEVEL, ((exponent-1)>>(m.dim-1))))
	// assert (level == 0 || getValue(level) >= value);
	// assert (level == S2CellId.MAX_LEVEL || getValue(level + 1) < value);
	return level
}
