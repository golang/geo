// Copyright 2014 Google Inc. All rights reserved.
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
	"flag"
	"io"
	"math"
	"math/rand"
	"os"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
	"github.com/golang/geo/s1"
)

// To set in testing add "--benchmark_brute_force=true" to your test command.
var benchmarkBruteForce = flag.Bool("benchmark_brute_force", false,
	"When set, use brute force algorithms in benchmarking.")

// float64Eq reports whether the two values are within the default epsilon.
func float64Eq(x, y float64) bool { return float64Near(x, y, epsilon) }

// float64Near reports whether the two values are within the given epsilon.
func float64Near(x, y, ε float64) bool {
	return math.Abs(x-y) <= ε
}

// TODO(roberts): Add in flag to allow specifying the random seed for repeatable tests.

// The Earth's mean radius in kilometers (according to NASA).
const earthRadiusKm = 6371.01

// kmToAngle converts a distance on the Earth's surface to an angle.
func kmToAngle(km float64) s1.Angle {
	return s1.Angle(km / earthRadiusKm)
}

// randomBits returns a 64-bit random unsigned integer whose lowest "num" are random, and
// whose other bits are zero.
func randomBits(num uint32) uint64 {
	// Make sure the request is for not more than 63 bits.
	if num > 63 {
		num = 63
	}
	return uint64(rand.Int63()) & ((1 << num) - 1)
}

// Return a uniformly distributed 64-bit unsigned integer.
func randomUint64() uint64 {
	return uint64(rand.Int63() | (rand.Int63() << 63))
}

// Return a uniformly distributed 32-bit unsigned integer.
func randomUint32() uint32 {
	return uint32(randomBits(32))
}

// randomFloat64 returns a uniformly distributed value in the range [0,1).
// Note that the values returned are all multiples of 2**-53, which means that
// not all possible values in this range are returned.
func randomFloat64() float64 {
	const randomFloatBits = 53
	return math.Ldexp(float64(randomBits(randomFloatBits)), -randomFloatBits)
}

// randomUniformInt returns a uniformly distributed integer in the range [0,n).
// NOTE: This is replicated here to stay in sync with how the C++ code generates
// uniform randoms. (instead of using Go's math/rand package directly).
func randomUniformInt(n int) int {
	return int(randomFloat64() * float64(n))
}

// randomUniformFloat64 returns a uniformly distributed value in the range [min, max).
func randomUniformFloat64(min, max float64) float64 {
	return min + randomFloat64()*(max-min)
}

// oneIn returns true with a probability of 1/n.
func oneIn(n int) bool {
	return randomUniformInt(n) == 0
}

// randomPoint returns a random unit-length vector.
func randomPoint() Point {
	return PointFromCoords(randomUniformFloat64(-1, 1),
		randomUniformFloat64(-1, 1), randomUniformFloat64(-1, 1))
}

// randomFrame returns a right-handed coordinate frame (three orthonormal vectors) for
// a randomly generated point.
func randomFrame() *matrix3x3 {
	return randomFrameAtPoint(randomPoint())
}

// randomFrameAtPoint returns a right-handed coordinate frame using the given
// point as the z-axis. The x- and y-axes are computed such that (x,y,z) is a
// right-handed coordinate frame (three orthonormal vectors).
func randomFrameAtPoint(z Point) *matrix3x3 {
	x := Point{z.Cross(randomPoint().Vector).Normalize()}
	y := Point{z.Cross(x.Vector).Normalize()}

	m := &matrix3x3{}
	m.setCol(0, x)
	m.setCol(1, y)
	m.setCol(2, z)
	return m
}

// randomCellIDForLevel returns a random CellID at the given level.
// The distribution is uniform over the space of cell ids, but only
// approximately uniform over the surface of the sphere.
func randomCellIDForLevel(level int) CellID {
	face := randomUniformInt(numFaces)
	pos := randomUint64() & uint64((1<<posBits)-1)
	return CellIDFromFacePosLevel(face, pos, level)
}

// randomCellID returns a random CellID at a randomly chosen
// level. The distribution is uniform over the space of cell ids,
// but only approximately uniform over the surface of the sphere.
func randomCellID() CellID {
	return randomCellIDForLevel(randomUniformInt(maxLevel + 1))
}

// randomCellUnion returns a CellUnion of the given size of randomly selected cells.
func randomCellUnion(n int) CellUnion {
	var cu CellUnion
	for i := 0; i < n; i++ {
		cu = append(cu, randomCellID())
	}
	cu.Normalize()
	return cu
}

// concentricLoopsPolygon constructs a polygon with the specified center as a
// number of concentric loops and vertices per loop.
func concentricLoopsPolygon(center Point, numLoops, verticesPerLoop int) *Polygon {
	var loops []*Loop
	for li := 0; li < numLoops; li++ {
		radius := s1.Angle(0.005 * float64(li+1) / float64(numLoops))
		loops = append(loops, RegularLoop(center, radius, verticesPerLoop))
	}
	return PolygonFromLoops(loops)
}

// skewedInt returns a number in the range [0,2^max_log-1] with bias towards smaller numbers.
func skewedInt(maxLog int) int {
	base := uint32(rand.Int31n(int32(maxLog + 1)))
	return int(randomBits(31) & ((1 << base) - 1))
}

// randomCap returns a cap with a random axis such that the log of its area is
// uniformly distributed between the logs of the two given values. The log of
// the cap angle is also approximately uniformly distributed.
func randomCap(minArea, maxArea float64) Cap {
	capArea := maxArea * math.Pow(minArea/maxArea, randomFloat64())
	return CapFromCenterArea(randomPoint(), capArea)
}

// latLngsApproxEqual reports of the two LatLngs are within the given epsilon.
func latLngsApproxEqual(a, b LatLng, epsilon float64) bool {
	return float64Near(float64(a.Lat), float64(b.Lat), epsilon) &&
		float64Near(float64(a.Lng), float64(b.Lng), epsilon)
}

// pointsApproxEqual reports whether the two points are within the given distance
// of each other. This is the same as Point.ApproxEqual but permits specifying
// the epsilon.
func pointsApproxEqual(a, b Point, epsilon float64) bool {
	return float64(a.Vector.Angle(b.Vector)) <= epsilon
}

// pointSlicesApproxEqual reports whether corresponding elements of each slice are approximately equal.
func pointSlicesApproxEqual(a, b []Point, epsilon float64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !pointsApproxEqual(a[i], b[i], epsilon) {
			return false
		}
	}
	return true
}

// r1IntervalsApproxEqual reports whether the two r1.Intervals are within the given
// epsilon of each other. This adds a test changeable value for epsilon
func r1IntervalsApproxEqual(a, b r1.Interval, epsilon float64) bool {
	if a.IsEmpty() {
		return b.Length() <= 2*epsilon
	}
	if b.IsEmpty() {
		return a.Length() <= 2*epsilon
	}
	return math.Abs(b.Lo-a.Lo) <= epsilon &&
		math.Abs(b.Hi-a.Hi) <= epsilon
}

var (
	rectErrorLat = 10 * dblEpsilon
	rectErrorLng = dblEpsilon
)

// r2PointsApproxEqual reports whether the two points are within the given epsilon.
func r2PointsApproxEqual(a, b r2.Point, epsilon float64) bool {
	return float64Near(a.X, b.X, epsilon) && float64Near(a.Y, b.Y, epsilon)
}

// r2PointSlicesApproxEqual reports whether corresponding elements of the slices are approximately equal.
func r2PointSlicesApproxEqual(a, b []r2.Point, epsilon float64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !r2PointsApproxEqual(a[i], b[i], epsilon) {
			return false
		}
	}
	return true
}

// rectsApproxEqual reports whether the two rect are within the given tolerances
// at each corner from each other. The tolerances are specific to each axis.
func rectsApproxEqual(a, b Rect, tolLat, tolLng float64) bool {
	return math.Abs(a.Lat.Lo-b.Lat.Lo) < tolLat &&
		math.Abs(a.Lat.Hi-b.Lat.Hi) < tolLat &&
		math.Abs(a.Lng.Lo-b.Lng.Lo) < tolLng &&
		math.Abs(a.Lng.Hi-b.Lng.Hi) < tolLng
}

// matricesApproxEqual reports whether all cells in both matrices are equal within
// the default floating point epsilon.
func matricesApproxEqual(m1, m2 *matrix3x3) bool {
	return float64Eq(m1[0][0], m2[0][0]) &&
		float64Eq(m1[0][1], m2[0][1]) &&
		float64Eq(m1[0][2], m2[0][2]) &&

		float64Eq(m1[1][0], m2[1][0]) &&
		float64Eq(m1[1][1], m2[1][1]) &&
		float64Eq(m1[1][2], m2[1][2]) &&

		float64Eq(m1[2][0], m2[2][0]) &&
		float64Eq(m1[2][1], m2[2][1]) &&
		float64Eq(m1[2][2], m2[2][2])
}

// samplePointFromRect returns a point chosen uniformly at random (with respect
// to area on the sphere) from the given rectangle.
func samplePointFromRect(rect Rect) Point {
	// First choose a latitude uniformly with respect to area on the sphere.
	sinLo := math.Sin(rect.Lat.Lo)
	sinHi := math.Sin(rect.Lat.Hi)
	lat := math.Asin(randomUniformFloat64(sinLo, sinHi))

	// Now choose longitude uniformly within the given range.
	lng := rect.Lng.Lo + randomFloat64()*rect.Lng.Length()

	return PointFromLatLng(LatLng{s1.Angle(lat), s1.Angle(lng)}.Normalized())
}

// samplePointFromCap returns a point chosen uniformly at random (with respect
// to area) from the given cap.
func samplePointFromCap(c Cap) Point {
	// We consider the cap axis to be the "z" axis. We choose two other axes to
	// complete the coordinate frame.
	m := getFrame(c.Center())

	// The surface area of a spherical cap is directly proportional to its
	// height. First we choose a random height, and then we choose a random
	// point along the circle at that height.
	h := randomFloat64() * c.Height()
	theta := 2 * math.Pi * randomFloat64()
	r := math.Sqrt(h * (2 - h))

	// The result should already be very close to unit-length, but we might as
	// well make it accurate as possible.
	return Point{fromFrame(m, PointFromCoords(math.Cos(theta)*r, math.Sin(theta)*r, 1-h)).Normalize()}
}

// perturbATowardsB returns a point that has been shifted some distance towards the
// second point based on a random number.
func perturbATowardsB(a, b Point) Point {
	choice := randomFloat64()
	if choice < 0.1 {
		return a
	}
	if choice < 0.3 {
		// Return a point that is exactly proportional to A and that still
		// satisfies IsUnitLength().
		for {
			b := Point{a.Mul(2 - a.Norm() + 5*(randomFloat64()-0.5)*dblEpsilon)}
			if !b.ApproxEqual(a) && b.IsUnit() {
				return b
			}
		}
	}
	if choice < 0.5 {
		// Return a point such that the distance squared to A will underflow.
		return InterpolateAtDistance(1e-300, a, b)
	}
	// Otherwise return a point whose distance from A is near dblEpsilon such
	// that the log of the pdf is uniformly distributed.
	distance := dblEpsilon * 1e-5 * math.Pow(1e6, randomFloat64())
	return InterpolateAtDistance(s1.Angle(distance), a, b)
}

// perturbedCornerOrMidpoint returns a Point from a line segment whose endpoints are
// difficult to handle correctly. Given two adjacent cube vertices P and Q,
// it returns either an edge midpoint, face midpoint, or corner vertex that is
// in the plane of PQ and that has been perturbed slightly. It also sometimes
// returns a random point from anywhere on the sphere.
func perturbedCornerOrMidpoint(p, q Point) Point {
	a := p.Mul(float64(randomUniformInt(3) - 1)).Add(q.Mul(float64(randomUniformInt(3) - 1)))
	if oneIn(10) {
		// This perturbation often has no effect except on coordinates that are
		// zero, in which case the perturbed value is so small that operations on
		// it often result in underflow.
		a = a.Add(randomPoint().Mul(math.Pow(1e-300, randomFloat64())))
	} else if oneIn(2) {
		// For coordinates near 1 (say > 0.5), this perturbation yields values
		// that are only a few representable values away from the initial value.
		a = a.Add(randomPoint().Mul(4 * dblEpsilon))
	} else {
		// A perturbation whose magnitude is in the range [1e-25, 1e-10].
		a = a.Add(randomPoint().Mul(1e-10 * math.Pow(1e-15, randomFloat64())))
	}

	if a.Norm2() < math.SmallestNonzeroFloat64 {
		// If a.Norm2() is denormalized, Normalize() loses too much precision.
		return perturbedCornerOrMidpoint(p, q)
	}
	return Point{a}
}

// readLoops returns a slice of Loops read from a file encoded using Loops Encode.
func readLoops(filename string) ([]*Loop, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var loops []*Loop

	// Test loop files are expected to be a direct concatenation of encoded loops with
	// no separator tokens. Because there is no way of knowing a priori how many items
	// or how many bytes ahead of time, the only way to end the loop is when we hit EOF.
	for {
		l := &Loop{}
		err := l.Decode(f)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		loops = append(loops, l)
	}

	return loops, nil
}

// fractal is a simple type that generates "Koch snowflake" fractals (see Wikipedia
// for an introduction). There is an option to control the fractal dimension
// (between 1.0 and 2.0); values between 1.02 and 1.50 are reasonable simulations
// of various coastlines. The default dimension (about 1.26) corresponds to the
// standard Koch snowflake. (The west coast of Britain has a fractal dimension
// of approximately 1.25.)
//
// The fractal is obtained by starting with an equilateral triangle and
// recursively subdividing each edge into four segments of equal length.
// Therefore the shape at level "n" consists of 3*(4**n) edges. Multi-level
// fractals are also supported: if you set MinLevel to a non-negative
// value, then the recursive subdivision has an equal probability of
// stopping at any of the levels between the given min and max (inclusive).
// This yields a fractal where the perimeter of the original triangle is
// approximately equally divided between fractals at the various possible
// levels. If there are k distinct levels {min,..,max}, the expected number
// of edges at each level "i" is approximately 3*(4**i)/k.
type fractal struct {
	maxLevel int
	// minLevel defines the minimum subdivision level of the fractal
	// A min level of 0 should be avoided since this creates a
	// significant chance that none of the edges will end up subdivided.
	minLevel int

	// dimension of the fractal. A value of approximately 1.26 corresponds
	// to the stardard Koch curve. The value must lie in the range [1.0, 2.0).
	dimension float64

	// The ratio of the sub-edge length to the original edge length at each
	// subdivision step.
	edgeFraction float64

	// The distance from the original edge to the middle vertex at each
	// subdivision step, as a fraction of the original edge length.
	offsetFraction float64
}

// newFractal returns a new instance of the fractal type with appropriate defaults.
func newFractal() *fractal {
	return &fractal{
		maxLevel:       -1,
		minLevel:       -1,
		dimension:      math.Log(4) / math.Log(3), // standard Koch curve
		edgeFraction:   0,
		offsetFraction: 0,
	}
}

// setLevelForApproxMinEdges sets the min level to produce approximately the
// given number of edges. The values are rounded to a nearby value of 3*(4**n).
func (f *fractal) setLevelForApproxMinEdges(minEdges int) {
	// Map values in the range [3*(4**n)/2, 3*(4**n)*2) to level n.
	f.minLevel = int(math.Round(0.5 * math.Log2(float64(minEdges)/3)))
}

// setLevelForApproxMaxEdges sets the max level to produce approximately the
// given number of edges. The values are rounded to a nearby value of 3*(4**n).
func (f *fractal) setLevelForApproxMaxEdges(maxEdges int) {
	// Map values in the range [3*(4**n)/2, 3*(4**n)*2) to level n.
	f.maxLevel = int(math.Round(0.5 * math.Log2(float64(maxEdges)/3)))
}

// minRadiusFactor returns a lower bound on the ratio (Rmin / R), where "R" is the
// radius passed to makeLoop and Rmin is the minimum distance from the
// fractal boundary to its center, where all distances are measured in the
// tangent plane at the fractal's center. This can be used to inscribe
// another geometric figure within the fractal without intersection.
func (f *fractal) minRadiusFactor() float64 {
	// The minimum radius is attained at one of the vertices created by the
	// first subdivision step as long as the dimension is not too small (at
	// least minDimensionForMinRadiusAtLevel1, see below). Otherwise we fall
	// back on the incircle radius of the original triangle, which is always a
	// lower bound (and is attained when dimension = 1).
	//
	// The value below was obtained by letting AE be an original triangle edge,
	// letting ABCDE be the corresponding polyline after one subdivision step,
	// and then letting BC be tangent to the inscribed circle at the center of
	// the fractal O. This gives rise to a pair of similar triangles whose edge
	// length ratios can be used to solve for the corresponding "edge fraction".
	// This method is slightly conservative because it is computed using planar
	// rather than spherical geometry. The value below is equal to
	// -math.Log(4)/math.Log((2 + math.Cbrt(2) - math.Cbrt(4))/6).
	const minDimensionForMinRadiusAtLevel1 = 1.0852230903040407
	if f.dimension >= minDimensionForMinRadiusAtLevel1 {
		return math.Sqrt(1 + 3*f.edgeFraction*(f.edgeFraction-1))
	}
	return 0.5
}

// maxRadiusFactor returns the ratio (Rmax / R), where "R" is the radius passed
// to makeLoop and Rmax is the maximum distance from the fractal boundary
// to its center, where all distances are measured in the tangent plane at
// the fractal's center. This can be used to inscribe the fractal within
// some other geometric figure without intersection.
func (f *fractal) maxRadiusFactor() float64 {
	// The maximum radius is always attained at either an original triangle
	// vertex or at a middle vertex from the first subdivision step.
	return math.Max(1.0, f.offsetFraction*math.Sqrt(3)+0.5)
}

// r2VerticesHelper recursively subdivides the edge to the desired level given
// the two endpoints (v0,v4) of an edge, and returns all vertices of the resulting
// curve up to but not including the endpoint v4.
func (f *fractal) r2VerticesHelper(v0, v4 r2.Point, level int) []r2.Point {
	if level >= f.minLevel && oneIn(f.maxLevel-level+1) {
		// stop at this level
		return []r2.Point{v0}
	}

	var vertices []r2.Point

	// Otherwise compute the intermediate vertices v1, v2, and v3.
	dir := v4.Sub(v0)
	v1 := v0.Add(dir.Mul(f.edgeFraction))
	v2 := v0.Add(v4).Mul(0.5).Sub(dir.Ortho().Mul(f.offsetFraction))
	v3 := v4.Sub(dir.Mul(f.edgeFraction))

	// And recurse on the four sub-edges.
	vertices = append(vertices, f.r2VerticesHelper(v0, v1, level+1)...)
	vertices = append(vertices, f.r2VerticesHelper(v1, v2, level+1)...)
	vertices = append(vertices, f.r2VerticesHelper(v2, v3, level+1)...)
	vertices = append(vertices, f.r2VerticesHelper(v3, v4, level+1)...)

	return vertices
}

// generateR2Vertices returns the set of r2 plane vertices for the fractal
// based on its current settings.
func (f *fractal) generateR2Vertices() []r2.Point {
	var vertices []r2.Point

	// The Koch "snowflake" consists of three Koch curves whose initial edges
	// form an equilateral triangle.
	v0 := r2.Point{1.0, 0.0}
	v1 := r2.Point{-0.5, math.Sqrt(3) / 2}
	v2 := r2.Point{-0.5, -math.Sqrt(3) / 2}
	vertices = append(vertices, f.r2VerticesHelper(v0, v1, 0)...)
	vertices = append(vertices, f.r2VerticesHelper(v1, v2, 0)...)
	vertices = append(vertices, f.r2VerticesHelper(v2, v0, 0)...)

	return vertices
}

// makeLoop returns a fractal loop centered around the z-axis of the given
// coordinate frame, with the first vertex in the direction of the
// positive x-axis. In order to avoid self-intersections, the fractal is
// generated by first drawing it in a 2D tangent plane to the unit sphere
// (touching at the fractal's center point) and then projecting the edges
// onto the sphere. This has the side effect of shrinking the fractal
// slightly compared to its nominal radius.
func (f *fractal) makeLoop(frame *matrix3x3, nominalRadius s1.Angle) *Loop {
	// update dependent values before making the loop.
	if f.minLevel < 0 || f.minLevel > f.maxLevel {
		f.minLevel = f.maxLevel
	}

	f.edgeFraction = math.Pow(4.0, -1.0/f.dimension)
	f.offsetFraction = math.Sqrt(f.edgeFraction - 0.25)

	r2pts := f.generateR2Vertices()
	verts := make([]Point, 0, len(r2pts))
	r := nominalRadius.Radians()

	for _, pt := range r2pts {
		verts = append(verts, fromFrame(*frame, PointFromCoords(pt.X*r, pt.Y*r, 1)))
	}
	return LoopFromPoints(verts)
}

// TODO(roberts): Items remaining to port:
// CheckCovering
// CheckResultSet
// CheckDistanceResults
