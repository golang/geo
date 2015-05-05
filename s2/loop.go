package s2

import (
	"fmt"
	"math"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/s1"
)

/**
 *
 * An S2Loop represents a simple spherical polygon. It consists of a single
 * chain of vertices where the first vertex is implicitly connected to the last.
 * All loops are defined to have a CCW orientation, i.e. the interior of the
 * polygon is on the left side of the edges. This implies that a clockwise loop
 * enclosing a small area is interpreted to be a CCW loop enclosing a very large
 * area.
 *
 *  Loops are not allowed to have any duplicate vertices (whether adjacent or
 * not), and non-adjacent edges are not allowed to intersect. Loops must have at
 * least 3 vertices. Although these restrictions are not enforced in optimized
 * code, you may get unexpected results if they are violated.
 *
 *  Point containment is defined such that if the sphere is subdivided into
 * faces (loops), every point is contained by exactly one face. This implies
 * that loops do not necessarily contain all (or any) of their vertices An
 * S2LatLngRect represents a latitude-longitude rectangle. It is capable of
 * representing the empty and full rectangles as well as single points.
 *
 */
type Loop struct {
	// Edge index used for performance-critical operations. For example,
	// contains() can determine whether a point is inside a loop in nearly
	// constant time, whereas without an edge index it is forced to compare the
	// query point against every edge in the loop.
	index *EdgeIndex

	// Maps each S2Point to its order in the loop, from 1 to numVertices.
	vertexToIndex map[Point]int

	vertices []Point

	// The index (into "vertices") of the vertex that comes first in the total
	// ordering of all vertices in this loop.
	firstLogicalVertex int

	bound        Rect
	originInside bool
	depth        int
}

func LoopFromPoints(points []Point) *Loop {
	l := &Loop{
		vertices: points,
		bound:    FullRect(),
		depth:    0,
	}
	l.initOrigin()
	l.initBound()
	l.initFirstLogicalVertex()
	return l
}

func LoopFromCell(cell Cell) *Loop {
	return LoopFromCellAndRect(cell, cell.RectBound())
}

func LoopFromCellAndRect(cell Cell, bound Rect) *Loop {
	l := &Loop{
		bound:    bound,
		vertices: make([]Point, 4),
	}
	for i := 0; i < 4; i++ {
		l.vertices[i] = cell.Vertex(i)
	}
	l.initOrigin()
	l.initFirstLogicalVertex()
	return l
}

func (l *Loop) Depth() int { return l.depth }

/**
 * The depth of a loop is defined as its nesting level within its containing
 * polygon. "Outer shell" loops have depth 0, holes within those loops have
 * depth 1, shells within those holes have depth 2, etc. This field is only
 * used by the S2Polygon implementation.
 *
 * @param depth
 */
func (l *Loop) SetDepth(depth int) { l.depth = depth }

/**
 * Return true if this loop represents a hole in its containing polygon.
 */
func (l *Loop) IsHole() bool {
	return (l.depth & 1) != 0
}

/**
 * The sign of a loop is -1 if the loop represents a hole in its containing
 * polygon, and +1 otherwise.
 */
func (l *Loop) Sign() int {
	if l.IsHole() {
		return -1
	} else {
		return 1
	}
}

func (l *Loop) NumVertices() int {
	return len(l.vertices)
}

/**
 * For convenience, we make two entire copies of the vertex list available:
 * vertex(n..2*n-1) is mapped to vertex(0..n-1), where n == numVertices().
 */
func (l *Loop) Vertex(i int) Point {
	if i >= l.NumVertices() {
		i = i - l.NumVertices()
	}
	return l.vertices[i]
}

func (l *Loop) CompareTo(other *Loop) int {
	if l.NumVertices() != other.NumVertices() {
		return l.NumVertices() - other.NumVertices()
	}
	// Compare the two loops' vertices, starting with each loop's
	// firstLogicalVertex. This allows us to always catch cases where logically
	// identical loops have different vertex orderings (e.g. ABCD and BCDA).
	maxVertices := l.NumVertices()
	iThis := l.firstLogicalVertex
	iOther := other.firstLogicalVertex
	for i := 0; i < maxVertices; i++ {
		compare := l.Vertex(iThis).CompareTo(other.Vertex(iOther))
		if compare != 0 {
			return compare
		}
		iThis++
		iOther++
	}
	return 0
}

/**
 * Calculates firstLogicalVertex, the vertex in this loop that comes first in
 * a total ordering of all vertices (by way of S2Point's compareTo function).
 */
func (l *Loop) initFirstLogicalVertex() {
	first := 0
	for i := 1; i < l.NumVertices(); i++ {
		if l.Vertex(i).CompareTo(l.Vertex(first)) < 0 {
			first = i
		}
	}
	l.firstLogicalVertex = first
}

/**
 * Return true if the loop area is at most 2*Pi.
 */
func (l *Loop) IsNormalized() bool {
	// We allow a bit of error so that exact hemispheres are
	// considered normalized.
	return l.GetArea() <= 2*math.Pi+1e-14
}

/**
 * Invert the loop if necessary so that the area enclosed by the loop is at
 * most 2*Pi.
 */
func (l *Loop) Normalize() {
	if !l.IsNormalized() {
		l.Invert()
	}
}

/**
 * Reverse the order of the loop vertices, effectively complementing the
 * region represented by the loop.
 */
func (l *Loop) Invert() {
	last := l.NumVertices() - 1
	for i := (last - 1) / 2; i >= 0; i-- {
		t := l.vertices[i]
		l.vertices[i] = l.vertices[last-i]
		l.vertices[last-i] = t
	}
	l.vertexToIndex = nil
	l.index = nil
	l.originInside = !l.originInside
	if l.bound.Lat.Lo > -math.Pi/2 && l.bound.Lat.Hi < math.Pi/2 {
		// The complement of this loop contains both poles.
		l.bound = FullRect()
	} else {
		l.initBound()
	}
	l.initFirstLogicalVertex()
}

/**
 * Helper method to get area and optionally centroid.
 */
func (l *Loop) getAreaCentroid(doCentroid bool) AreaCentroid {
	var centroid *Point
	// Don't crash even if loop is not well-defined.
	if l.NumVertices() < 3 {
		return NewAreaCentroid(0, nil)
	}

	// The triangle area calculation becomes numerically unstable as the length
	// of any edge approaches 180 degrees. However, a loop may contain vertices
	// that are 180 degrees apart and still be valid, e.g. a loop that defines
	// the northern hemisphere using four points. We handle this case by using
	// triangles centered around an origin that is slightly displaced from the
	// first vertex. The amount of displacement is enough to get plenty of
	// accuracy for antipodal points, but small enough so that we still get
	// accurate areas for very tiny triangles.
	//
	// Of course, if the loop contains a point that is exactly antipodal from
	// our slightly displaced vertex, the area will still be unstable, but we
	// expect this case to be very unlikely (i.e. a polygon with two vertices on
	// opposite sides of the Earth with one of them displaced by about 2mm in
	// exactly the right direction). Note that the approximate point resolution
	// using the E7 or S2CellId representation is only about 1cm.

	origin := l.Vertex(0)
	axis := (origin.LargestAbsComponent() + 1) % 3
	slightlyDisplaced := origin.GetAxis(axis) + math.E*1e-10
	switch axis {
	case 0:
		origin = PointFromCoordsRaw(slightlyDisplaced, origin.Y, origin.Z)
	case 1:
		origin = PointFromCoordsRaw(origin.X, slightlyDisplaced, origin.Z)
	case 2:
		origin = PointFromCoordsRaw(origin.X, origin.Y, slightlyDisplaced)
	}
	origin = Point{origin.Normalize()}

	var areaSum float64 = 0
	centroidSum := PointFromCoordsRaw(0, 0, 0)
	for i := 1; i <= l.NumVertices(); i++ {
		areaSum += SignedArea(origin, l.Vertex(i-1), l.Vertex(i))
		if doCentroid {
			// The true centroid is already premultiplied by the triangle area.
			trueCentroid := TrueCentroid(origin, l.Vertex(i-1), l.Vertex(i))
			centroidSum = Point{centroidSum.Add(trueCentroid.Vector)}
		}
	}
	// The calculated area at this point should be between -4*Pi and 4*Pi,
	// although it may be slightly larger or smaller than this due to
	// numerical errors.
	// assert (Math.abs(areaSum) <= 4 * S2.M_PI + 1e-12);

	if areaSum < 0 {
		// If the area is negative, we have computed the area to the right of the
		// loop. The area to the left is 4*Pi - (-area). Amazingly, the centroid
		// does not need to be changed, since it is the negative of the integral
		// of position over the region to the right of the loop. This is the same
		// as the integral of position over the region to the left of the loop,
		// since the integral of position over the entire sphere is (0, 0, 0).
		areaSum += 4 * math.Pi
	}
	// The loop's sign() does not affect the return result and should be taken
	// into account by the caller.
	if doCentroid {
		centroid = &centroidSum
	}
	return NewAreaCentroid(areaSum, centroid)
}

/**
 * Return the area of the loop interior, i.e. the region on the left side of
 * the loop. The return value is between 0 and 4*Pi and the true centroid of
 * the loop multiplied by the area of the loop (see S2.java for details on
 * centroids). Note that the centroid may not be contained by the loop.
 */
func (l *Loop) GetAreaAndCentroid() AreaCentroid {
	return l.getAreaCentroid(true)
}

/**
 * Return the area of the polygon interior, i.e. the region on the left side
 * of an odd number of loops. The return value is between 0 and 4*Pi.
 */
func (l *Loop) GetArea() float64 {
	return l.getAreaCentroid(false).GetArea()
}

/**
 * Return the true centroid of the polygon multiplied by the area of the
 * polygon (see {@link S2} for details on centroids). Note that the centroid
 * may not be contained by the polygon.
 */
func (l *Loop) GetCentroid() Point {
	return l.getAreaCentroid(true).GetCentroid()
}

func (l *Loop) ContainsLoop(b *Loop) bool {
	// For this loop A to contains the given loop B, all of the following must
	// be true:
	//
	// (1) There are no edge crossings between A and B except at vertices.
	//
	// (2) At every vertex that is shared between A and B, the local edge
	// ordering implies that A contains B.
	//
	// (3) If there are no shared vertices, then A must contain a vertex of B
	// and B must not contain a vertex of A. (An arbitrary vertex may be
	// chosen in each case.)
	//
	// The second part of (3) is necessary to detect the case of two loops whose
	// union is the entire sphere, i.e. two loops that contains each other's
	// boundaries but not each other's interiors.

	if !l.bound.ContainsRect(b.RectBound()) {
		return false
	}

	// Unless there are shared vertices, we need to check whether A contains a
	// vertex of B. Since shared vertices are rare, it is more efficient to do
	// this test up front as a quick rejection test.
	if !l.ContainsPoint(b.Vertex(0)) && l.findVertex(b.Vertex(0)) < 0 {
		return false
	}

	// Now check whether there are any edge crossings, and also check the loop
	// relationship at any shared vertices.
	if l.checkEdgeCrossings(b, WedgeContains{}) <= 0 {
		return false
	}

	// At this point we know that the boundaries of A and B do not intersect,
	// and that A contains a vertex of B. However we still need to check for
	// the case mentioned above, where (A union B) is the entire sphere.
	// Normally this check is very cheap due to the bounding box precondition.
	if l.bound.Union(b.RectBound()).IsFull() {
		if b.ContainsPoint(l.Vertex(0)) && b.findVertex(l.Vertex(0)) < 0 {
			return false
		}
	}
	return true
}

/**
 * Return true if the region contained by this loop intersects the region
 * contained by the given other loop.
 */
func (l *Loop) IntersectsLoop(b *Loop) bool {
	// a->Intersects(b) if and only if !a->Complement()->Contains(b).
	// This code is similar to Contains(), but is optimized for the case
	// where both loops enclose less than half of the sphere.

	if !l.bound.IntersectsRect(b.RectBound()) {
		return false
	}

	// Normalize the arguments so that B has a smaller longitude span than A.
	// This makes intersection tests much more efficient in the case where
	// longitude pruning is used (see CheckEdgeCrossings).
	if b.RectBound().Lng.Length() > l.bound.Lng.Length() {
		return b.IntersectsLoop(l)
	}

	// Unless there are shared vertices, we need to check whether A contains a
	// vertex of B. Since shared vertices are rare, it is more efficient to do
	// this test up front as a quick acceptance test.
	if l.ContainsPoint(b.Vertex(0)) && l.findVertex(b.Vertex(0)) < 0 {
		return true
	}

	// Now check whether there are any edge crossings, and also check the loop
	// relationship at any shared vertices.
	if l.checkEdgeCrossings(b, WedgeIntersects{}) < 0 {
		return true
	}

	// We know that A does not contain a vertex of B, and that there are no edge
	// crossings. Therefore the only way that A can intersect B is if B
	// entirely contains A. We can check this by testing whether B contains an
	// arbitrary non-shared vertex of A. Note that this check is cheap because
	// of the bounding box precondition and the fact that we normalized the
	// arguments so that A's longitude span is at least as long as B's.
	if b.RectBound().ContainsRect(l.bound) {
		if b.ContainsPoint(l.Vertex(0)) && b.findVertex(l.Vertex(0)) < 0 {
			return true
		}
	}

	return false
}

/**
 * Given two loops of a polygon, return true if A contains B. This version of
 * contains() is much cheaper since it does not need to check whether the
 * boundaries of the two loops cross.
 */
func (l *Loop) ContainsNested(b *Loop) bool {
	if !l.bound.ContainsRect(b.RectBound()) {
		return false
	}

	// We are given that A and B do not share any edges, and that either one
	// loop contains the other or they do not intersect.
	m := l.findVertex(b.Vertex(1))
	if m < 0 {
		// Since b->vertex(1) is not shared, we can check whether A contains it.
		return l.ContainsPoint(b.Vertex(1))
	}
	// Check whether the edge order around b->vertex(1) is compatible with
	// A containin B.
	return (WedgeContains{}).Test(l.Vertex(m-1), l.Vertex(m), l.Vertex(m+1), b.Vertex(0), b.Vertex(2)) > 0
}

/**
 * Return +1 if A contains B (i.e. the interior of B is a subset of the
 * interior of A), -1 if the boundaries of A and B cross, and 0 otherwise.
 * Requires that A does not properly contain the complement of B, i.e. A and B
 * do not contain each other's boundaries. This method is used for testing
 * whether multi-loop polygons contain each other.
 */
func (l *Loop) ContainsOrCrosses(b *Loop) int {
	// There can be containment or crossing only if the bounds intersect.
	if !l.bound.IntersectsRect(b.RectBound()) {
		return 0
	}

	// Now check whether there are any edge crossings, and also check the loop
	// relationship at any shared vertices. Note that unlike Contains() or
	// Intersects(), we can't do a point containment test as a shortcut because
	// we need to detect whether there are any edge crossings.
	result := l.checkEdgeCrossings(b, WedgeContainsOrCrosses{})

	// If there was an edge crossing or a shared vertex, we know the result
	// already. (This is true even if the result is 1, but since we don't
	// bother keeping track of whether a shared vertex was seen, we handle this
	// case below.)
	if result <= 0 {
		return result
	}

	// At this point we know that the boundaries do not intersect, and we are
	// given that (A union B) is a proper subset of the sphere. Furthermore
	// either A contains B, or there are no shared vertices (due to the check
	// above). So now we just need to distinguish the case where A contains B
	// from the case where B contains A or the two loops are disjoint.
	if !l.bound.ContainsRect(b.RectBound()) {
		return 0
	}
	if !l.ContainsPoint(b.Vertex(0)) && l.findVertex(b.Vertex(0)) < 0 {
		return 0
	}

	return 1
}

/**
 * Returns true if two loops have the same boundary except for vertex
 * perturbations. More precisely, the vertices in the two loops must be in the
 * same cyclic order, and corresponding vertex pairs must be separated by no
 * more than maxError. Note: This method mostly useful only for testing
 * purposes.
 */
func (l *Loop) BoundaryApproxEquals(b *Loop, maxError float64) bool {
	if l.NumVertices() != b.NumVertices() {
		return false
	}
	maxVertices := l.NumVertices()
	iThis := l.firstLogicalVertex
	iOther := b.firstLogicalVertex
	for i := 0; i < maxVertices; i++ {
		if !l.Vertex(iThis).ApproxEquals(b.Vertex(iOther), maxError) {
			return false
		}
		iThis++
		iOther++
	}
	return true
}

// CapBound returns a bounding spherical cap. This is not guaranteed to be exact.
func (l *Loop) CapBound() Cap {
	return l.bound.CapBound()
}

// RectBound returns a bounding latitude-longitude rectangle that contains
// the region. The bounds are not guaranteed to be tight.
func (l *Loop) RectBound() Rect {
	return l.bound
}

/**
 * If this method returns true, the region completely contains the given cell.
 * Otherwise, either the region does not contain the cell or the containment
 * relationship could not be determined.
 */
func (l *Loop) ContainsCell(cell Cell) bool {
	// It is faster to construct a bounding rectangle for an S2Cell than for
	// a general polygon. A future optimization could also take advantage of
	// the fact than an S2Cell is convex.

	cellBound := cell.RectBound()
	if !l.bound.ContainsRect(cellBound) {
		return false
	}
	cellLoop := LoopFromCellAndRect(cell, cellBound)
	return l.ContainsLoop(cellLoop)
}

/**
 * If this method returns false, the region does not intersect the given cell.
 * Otherwise, either region intersects the cell, or the intersection
 * relationship could not be determined.
 */
func (l *Loop) IntersectsCell(cell Cell) bool {
	// It is faster to construct a bounding rectangle for an S2Cell than for
	// a general polygon. A future optimization could also take advantage of
	// the fact than an S2Cell is convex.

	cellBound := cell.RectBound()
	if !l.bound.IntersectsRect(cellBound) {
		return false
	}
	return LoopFromCellAndRect(cell, cellBound).IntersectsLoop(l)
}

/**
 * The point 'p' does not need to be normalized.
 */
func (l *Loop) ContainsPoint(p Point) bool {
	if !l.bound.ContainsLatLng(LatLngFromPoint(p)) {
		return false
	}

	inside := l.originInside
	origin := PointFromCoordsRaw(0, 1, 0)
	crosser := NewEdgeCrosser(origin, p, l.Vertex(l.NumVertices()-1))

	// The s2edgeindex library is not optimized yet for long edges,
	// so the tradeoff to using it comes with larger loops.
	if l.NumVertices() < 2000 {
		for i := 0; i < l.NumVertices(); i++ {
			inside = inside != crosser.EdgeOrVertexCrossing(l.Vertex(i))
		}
	} else {
		it := l.getEdgeIterator(l.NumVertices())
		previousIndex := -2
		for it.GetCandidates(origin, p); it.HasNext(); it.Next() {
			ai := it.Index()
			if previousIndex != ai-1 {
				crosser.RestartAt(l.Vertex(ai))
			}
			previousIndex = ai
			inside = inside != crosser.EdgeOrVertexCrossing(l.Vertex(ai+1))
		}
	}

	return inside
}

/**
 * Returns the shortest distance from a point P to this loop, given as the
 * angle formed between P, the origin and the nearest point on the loop to P.
 * This angle in radians is equivalent to the arclength along the unit sphere.
 */
func (l *Loop) GetDistance(p Point) s1.Angle {
	normalized := Point{p.Normalize()}

	// The furthest point from p on the sphere is its antipode, which is an
	// angle of PI radians. This is an upper bound on the angle.
	minDistance := math.Pi
	for i := 0; i < l.NumVertices(); i++ {
		minDistance = math.Min(minDistance, getDistance(normalized, l.Vertex(i), l.Vertex(i+1)).Radians())
	}
	return s1.Angle(minDistance)
}

/**
 * Creates an edge index over the vertices, which by itself takes no time.
 * Then the expected number of queries is used to determine whether brute
 * force lookups are likely to be slower than really creating an index, and if
 * so, we do so. Finally an iterator is returned that can be used to perform
 * edge lookups.
 */
func (l *Loop) getEdgeIterator(expectedQueries int) *DataEdgeIterator {
	if l.index == nil {
		l.index = NewEdgeIndex(
			func() int { return l.NumVertices() },
			func(i int) Point { return l.Vertex(i) },
			func(i int) Point { return l.Vertex(i + 1) },
		)
	}
	l.index.PredictAdditionalCalls(expectedQueries)
	return NewDataEdgeIterator(l.index)
}

/** Return true if this loop is valid. */
func (l *Loop) IsValid() bool {
	if l.NumVertices() < 3 {
		fmt.Println("Degenerate loop")
		return false
	}

	// All vertices must be unit length.
	for i := 0; i < l.NumVertices(); i++ {
		if !l.Vertex(i).IsUnit() {
			fmt.Printf("Vertex %d is not unit length\n", i)
			return false
		}
	}

	// Loops are not allowed to have any duplicate vertices.
	vmap := make(map[Point]int)
	for i := 0; i < l.NumVertices(); i++ {
		if previousVertexIndex, ok := vmap[l.Vertex(i)]; ok {
			fmt.Printf("Duplicate vertices: %d and %d\n", previousVertexIndex, i)
			return false
		}
		vmap[l.Vertex(i)] = i
	}

	// Non-adjacent edges are not allowed to intersect.
	MAX_INTERSECTION_ERROR := 1e-15
	crosses := false
	it := l.getEdgeIterator(l.NumVertices())
	for a1 := 0; a1 < l.NumVertices(); a1++ {
		a2 := (a1 + 1) % l.NumVertices()
		crosser := NewEdgeCrosser(l.Vertex(a1), l.Vertex(a2), l.Vertex(0))
		previousIndex := -2
		for it.GetCandidates(l.Vertex(a1), l.Vertex(a2)); it.HasNext(); it.Next() {
			b1 := it.Index()
			b2 := (b1 + 1) % l.NumVertices()
			// If either 'a' index equals either 'b' index, then these two edges
			// share a vertex. If a1==b1 then it must be the case that a2==b2, e.g.
			// the two edges are the same. In that case, we skip the test, since we
			// don't want to test an edge against itself. If a1==b2 or b1==a2 then
			// we have one edge ending at the start of the other, or in other words,
			// the edges share a vertex -- and in S2 space, where edges are always
			// great circle segments on a sphere, edges can only intersect at most
			// once, so we don't need to do further checks in that case either.
			if a1 != b2 && a2 != b1 && a1 != b1 {
				// WORKAROUND(shakusa, ericv): S2.robustCCW() currently
				// requires arbitrary-precision arithmetic to be truly robust. That
				// means it can give the wrong answers in cases where we are trying
				// to determine edge intersections. The workaround is to ignore
				// intersections between edge pairs where all four points are
				// nearly colinear.
				abc := angle(l.Vertex(a1), l.Vertex(a2), l.Vertex(b1))
				abcNearlyLinear := approxEqualsNumber(abc, 0, MAX_INTERSECTION_ERROR) || approxEqualsNumber(abc, math.Pi, MAX_INTERSECTION_ERROR)
				abd := angle(l.Vertex(a1), l.Vertex(a2), l.Vertex(b2))
				abdNearlyLinear := approxEqualsNumber(abd, 0, MAX_INTERSECTION_ERROR) || approxEqualsNumber(abd, math.Pi, MAX_INTERSECTION_ERROR)
				if abcNearlyLinear && abdNearlyLinear {
					continue
				}

				if previousIndex != b1 {
					crosser.RestartAt(l.Vertex(b1))
				}

				// Beware, this may return the loop is valid if there is a
				// "vertex crossing".
				// TODO(user): Fix that.
				crosses = crosser.RobustCrossing(l.Vertex(b2)) > 0
				previousIndex = b2
				if crosses {
					fmt.Printf("Edges %d and %d cross\n", a1, b1)
					fmt.Printf("Edge locations in degrees: %s-%s and %s-%s\n",
						LatLngFromPoint(l.Vertex(a1)).StringDegrees(),
						LatLngFromPoint(l.Vertex(a2)).StringDegrees(),
						LatLngFromPoint(l.Vertex(b1)).StringDegrees(),
						LatLngFromPoint(l.Vertex(b2)).StringDegrees())
					return false
				}
			}
		}
	}

	return true
}

/**
 * Static version of isValid(), to be used only when an S2Loop instance is not
 * available, but validity of the points must be checked.
 *
 * @return true if the given loop is valid. Creates an instance of S2Loop and
 *         defers this call to {@link #isValid()}.
 */
func LoopIsValid(vertices []Point) bool {
	return LoopFromPoints(vertices).IsValid()
}

func (l *Loop) ToString() string {
	result := fmt.Sprintf("Loop, %d points. [", l.NumVertices())
	for _, v := range l.vertices {
		result = result + v.String() + " "
	}
	return result + "]"
}

func (l *Loop) initOrigin() {
	// The bounding box does not need to be correct before calling this
	// function, but it must at least contain vertex(1) since we need to
	// do a Contains() test on this point below.
	if !l.bound.ContainsLatLng(LatLngFromPoint(l.Vertex(1))) {
		panic("Bounds needs to at least contain Vertex(1)")
	}

	// To ensure that every point is contained in exactly one face of a
	// subdivision of the sphere, all containment tests are done by counting the
	// edge crossings starting at a fixed point on the sphere (S2::Origin()).
	// We need to know whether this point is inside or outside of the loop.
	// We do this by first guessing that it is outside, and then seeing whether
	// we get the correct containment result for vertex 1. If the result is
	// incorrect, the origin must be inside the loop.
	//
	// A loop with consecutive vertices A,B,C contains vertex B if and only if
	// the fixed vector R = S2::Ortho(B) is on the left side of the wedge ABC.
	// The test below is written so that B is inside if C=R but not if A=R.

	l.originInside = false // Initialize before calling Contains().
	v1Inside := OrderedCCW(Point{l.Vertex(1).Ortho()}, l.Vertex(0), l.Vertex(2), l.Vertex(1))
	if v1Inside != l.ContainsPoint(l.Vertex(1)) {
		l.originInside = true
	}
}

func (l *Loop) initBound() {
	// The bounding rectangle of a loop is not necessarily the same as the
	// bounding rectangle of its vertices. First, the loop may wrap entirely
	// around the sphere (e.g. a loop that defines two revolutions of a
	// candy-cane stripe). Second, the loop may include one or both poles.
	// Note that a small clockwise loop near the equator contains both poles.

	bounder := NewRectBounder()
	for i := 0; i <= l.NumVertices(); i++ {
		bounder.AddPoint(l.Vertex(i))
	}
	b := bounder.GetBound()
	// Note that we need to initialize bound with a temporary value since
	// contains() does a bounding rectangle check before doing anything else.
	l.bound = FullRect()
	if l.ContainsPoint(PointFromCoordsRaw(0, 0, 1)) {
		b = Rect{r1.IntervalFromPointPair(b.Lat.Lo, math.Pi/2), s1.FullInterval()}
	}
	// If a loop contains the south pole, then either it wraps entirely
	// around the sphere (full longitude range), or it also contains the
	// north pole in which case b.lng().isFull() due to the test above.

	if b.Lng.IsFull() && l.ContainsPoint(PointFromCoordsRaw(0, 0, -1)) {
		b = Rect{r1.IntervalFromPointPair(-math.Pi/2, b.Lat.Hi), b.Lng}
	}
	l.bound = b
}

/**
 * Return the index of a vertex at point "p", or -1 if not found. The return
 * value is in the range 1..num_vertices_ if found.
 */
func (l *Loop) findVertex(p Point) int {
	if l.vertexToIndex == nil {
		l.vertexToIndex = make(map[Point]int)
		for i := 1; i <= l.NumVertices(); i++ {
			l.vertexToIndex[l.Vertex(i)] = i
		}
	}
	if index, ok := l.vertexToIndex[p]; ok {
		return index
	}
	return -1
}

/**
 * This method encapsulates the common code for loop containment and
 * intersection tests. It is used in three slightly different variations to
 * implement contains(), intersects(), and containsOrCrosses().
 *
 *  In a nutshell, this method checks all the edges of this loop (A) for
 * intersection with all the edges of B. It returns -1 immediately if any edge
 * intersections are found. Otherwise, if there are any shared vertices, it
 * returns the minimum value of the given WedgeRelation for all such vertices
 * (returning immediately if any wedge returns -1). Returns +1 if there are no
 * intersections and no shared vertices.
 */
func (l *Loop) checkEdgeCrossings(b *Loop, relation WedgeRelation) int {
	it := l.getEdgeIterator(b.NumVertices())
	result := 1
	// since 'this' usually has many more vertices than 'b', use the index on
	// 'this' and loop over 'b'
	for j := 0; j < b.NumVertices(); j++ {
		crosser := NewEdgeCrosser(b.Vertex(j), b.Vertex(j+1), l.Vertex(0))
		previousIndex := -2
		for it.GetCandidates(b.Vertex(j), b.Vertex(j+1)); it.HasNext(); it.Next() {
			i := it.Index()
			if previousIndex != i-1 {
				crosser.RestartAt(l.Vertex(i))
			}
			previousIndex = i
			crossing := crosser.RobustCrossing(l.Vertex(i + 1))
			if crossing < 0 {
				continue
			}
			if crossing > 0 {
				return -1 // There is a proper edge crossing.
			}
			if l.Vertex(i + 1).Equals(b.Vertex(j + 1)) {
				result = min(result, relation.Test(
					l.Vertex(i),
					l.Vertex(i+1),
					l.Vertex(i+2),
					b.Vertex(j),
					b.Vertex(j+2),
				))
				if result < 0 {
					return result
				}
			}
		}
	}
	return result
}
