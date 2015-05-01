package s2

import (
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
	// index EdgeIndex

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

func LoopFromCell(cell Cell, bound Rect) *Loop {
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
	if l.ContainsPoint(PointFromCoords(0, 0, 1)) {
		b = Rect{r1.IntervalFromPointPair(b.Lat.Lo, math.Pi/2), s1.FullInterval()}
	}
	// If a loop contains the south pole, then either it wraps entirely
	// around the sphere (full longitude range), or it also contains the
	// north pole in which case b.lng().isFull() due to the test above.

	if b.Lng.IsFull() && l.ContainsPoint(PointFromCoords(0, 0, -1)) {
		b = Rect{r1.IntervalFromPointPair(-math.Pi/2, b.Lat.Hi), b.Lng}
	}
	l.bound = b
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

// CapBound returns a bounding spherical cap. This is not guaranteed to be exact.
func (l *Loop) CapBound() Cap {
	// TODO: Implement
	return EmptyCap()
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
	cellLoop := LoopFromCell(cell, cellBound)
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
	return LoopFromCell(cell, cellBound).IntersectsLoop(l)
}

/**
 * The point 'p' does not need to be normalized.
 */
func (l *Loop) ContainsPoint(p Point) bool {
	if !l.bound.ContainsLatLng(LatLngFromPoint(p)) {
		return false
	}

	inside := l.originInside
	origin := PointFromCoords(0, 1, 0)
	crosser := NewEdgeCrosser(origin, p, l.Vertex(l.NumVertices()-1))

	// The s2edgeindex library is not optimized yet for long edges,
	// so the tradeoff to using it comes with larger loops.
	if l.NumVertices() < 2000 {
		for i := 0; i < l.NumVertices(); i++ {
			inside = inside != crosser.EdgeOrVertexCrossing(l.Vertex(i))
		}
	} else {
		panic("TODO")
		// DataEdgeIterator it = getEdgeIterator(numVertices)
		// int previousIndex = -2
		// for (it.getCandidates(origin, p); it.hasNext(); it.next()) {
		//   int ai = it.index()
		//   if (previousIndex != ai - 1) {
		//     crosser.restartAt(vertices[ai])
		//   }
		//   previousIndex = ai
		//   inside ^= crosser.EdgeOrVertexCrossing(vertex(ai + 1))
		// }
	}

	return inside
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
	if (l.checkEdgeCrossings(b, WedgeIntersects{}) < 0) {
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
	// DataEdgeIterator it = getEdgeIterator(b.numVertices);
	result := 1
	// since 'this' usually has many more vertices than 'b', use the index on
	// 'this' and loop over 'b'
	// for (int j = 0; j < b.numVertices(); ++j) {
	//   S2EdgeUtil.EdgeCrosser crosser =
	//     new S2EdgeUtil.EdgeCrosser(b.vertex(j), b.vertex(j + 1), vertex(0));
	//   int previousIndex = -2;
	//   for (it.getCandidates(b.vertex(j), b.vertex(j + 1)); it.hasNext(); it.next()) {
	//     int i = it.index();
	//     if (previousIndex != i - 1) {
	//       crosser.restartAt(vertex(i));
	//     }
	//     previousIndex = i;
	//     int crossing = crosser.robustCrossing(vertex(i + 1));
	//     if (crossing < 0) {
	//       continue;
	//     }
	//     if (crossing > 0) {
	//       return -1; // There is a proper edge crossing.
	//     }
	//     if (vertex(i + 1).equals(b.vertex(j + 1))) {
	//       result = Math.min(result, relation.test(
	//           vertex(i), vertex(i + 1), vertex(i + 2), b.vertex(j), b.vertex(j + 2)));
	//       if (result < 0) {
	//         return result;
	//       }
	//     }
	//   }
	// }
	return result
}
