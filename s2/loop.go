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
	"fmt"
	"io"
	"math"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r3"
	"github.com/golang/geo/s1"
)

// Loop represents a simple spherical polygon. It consists of a sequence
// of vertices where the first vertex is implicitly connected to the
// last. All loops are defined to have a CCW orientation, i.e. the interior of
// the loop is on the left side of the edges. This implies that a clockwise
// loop enclosing a small area is interpreted to be a CCW loop enclosing a
// very large area.
//
// Loops are not allowed to have any duplicate vertices (whether adjacent or
// not), and non-adjacent edges are not allowed to intersect. Loops must have
// at least 3 vertices (except for the "empty" and "full" loops discussed
// below).
//
// There are two special loops: the "empty" loop contains no points and the
// "full" loop contains all points. These loops do not have any edges, but to
// preserve the invariant that every loop can be represented as a vertex
// chain, they are defined as having exactly one vertex each (see EmptyLoop
// and FullLoop).
type Loop struct {
	vertices []Point

	// originInside keeps a precomputed value whether this loop contains the origin
	// versus computing from the set of vertices every time.
	originInside bool

	// depth is the nesting depth of this Loop if it is contained by a Polygon
	// or other shape and is used to determine if this loop represents a hole
	// or a filled in portion.
	depth int

	// bound is a conservative bound on all points contained by this loop.
	// If l.ContainsPoint(P), then l.bound.ContainsPoint(P).
	bound Rect

	// Since bound is not exact, it is possible that a loop A contains
	// another loop B whose bounds are slightly larger. subregionBound
	// has been expanded sufficiently to account for this error, i.e.
	// if A.Contains(B), then A.subregionBound.Contains(B.bound).
	subregionBound Rect

	// index is the spatial index for this Loop.
	index *ShapeIndex
}

// LoopFromPoints constructs a loop from the given points.
func LoopFromPoints(pts []Point) *Loop {
	l := &Loop{
		vertices: pts,
	}

	l.initOriginAndBound()
	return l
}

// LoopFromCell constructs a loop corresponding to the given cell.
//
// Note that the loop and cell *do not* contain exactly the same set of
// points, because Loop and Cell have slightly different definitions of
// point containment. For example, a Cell vertex is contained by all
// four neighboring Cells, but it is contained by exactly one of four
// Loops constructed from those cells. As another example, the cell
// coverings of cell and LoopFromCell(cell) will be different, because the
// loop contains points on its boundary that actually belong to other cells
// (i.e., the covering will include a layer of neighboring cells).
func LoopFromCell(c Cell) *Loop {
	l := &Loop{
		vertices: []Point{
			c.Vertex(0),
			c.Vertex(1),
			c.Vertex(2),
			c.Vertex(3),
		},
	}

	l.initOriginAndBound()
	return l
}

// EmptyLoop returns a special "empty" loop.
func EmptyLoop() *Loop {
	return LoopFromPoints([]Point{{r3.Vector{X: 0, Y: 0, Z: 1}}})
}

// FullLoop returns a special "full" loop.
func FullLoop() *Loop {
	return LoopFromPoints([]Point{{r3.Vector{X: 0, Y: 0, Z: -1}}})
}

// initOriginAndBound sets the origin containment for the given point and then calls
// the initialization for the bounds objects and the internal index.
func (l *Loop) initOriginAndBound() {
	if len(l.vertices) < 3 {
		// Check for the special "empty" and "full" loops (which have one vertex).
		if !l.isEmptyOrFull() {
			l.originInside = false
			return
		}

		// This is the special empty or full loop, so the origin depends on if
		// the vertex is in the southern hemisphere or not.
		l.originInside = l.vertices[0].Z < 0
	} else {
		// Point containment testing is done by counting edge crossings starting
		// at a fixed point on the sphere (OriginPoint). We need to know whether
		// the reference point (OriginPoint) is inside or outside the loop before
		// we can construct the ShapeIndex. We do this by first guessing that
		// it is outside, and then seeing whether we get the correct containment
		// result for vertex 1. If the result is incorrect, the origin must be
		// inside the loop.
		//
		// A loop with consecutive vertices A,B,C contains vertex B if and only if
		// the fixed vector R = B.Ortho is contained by the wedge ABC. The
		// wedge is closed at A and open at C, i.e. the point B is inside the loop
		// if A = R but not if C = R. This convention is required for compatibility
		// with VertexCrossing. (Note that we can't use OriginPoint
		// as the fixed vector because of the possibility that B == OriginPoint.)
		l.originInside = false
		v1Inside := OrderedCCW(Point{l.vertices[1].Ortho()}, l.vertices[0], l.vertices[2], l.vertices[1])
		if v1Inside != l.ContainsPoint(l.vertices[1]) {
			l.originInside = true
		}
	}

	// We *must* call initBound before initializing the index, because
	// initBound calls ContainsPoint which does a bounds check before using
	// the index.
	l.initBound()

	// Create a new index and add us to it.
	l.index = NewShapeIndex()
	l.index.Add(l)
}

// initBound sets up the approximate bounding Rects for this loop.
func (l *Loop) initBound() {
	// Check for the special "empty" and "full" loops.
	if l.isEmptyOrFull() {
		if l.IsEmpty() {
			l.bound = EmptyRect()
		} else {
			l.bound = FullRect()
		}
		l.subregionBound = l.bound
		return
	}

	// The bounding rectangle of a loop is not necessarily the same as the
	// bounding rectangle of its vertices. First, the maximal latitude may be
	// attained along the interior of an edge. Second, the loop may wrap
	// entirely around the sphere (e.g. a loop that defines two revolutions of a
	// candy-cane stripe). Third, the loop may include one or both poles.
	// Note that a small clockwise loop near the equator contains both poles.
	bounder := NewRectBounder()
	for i := 0; i <= len(l.vertices); i++ { // add vertex 0 twice
		bounder.AddPoint(l.Vertex(i))
	}
	b := bounder.RectBound()

	if l.ContainsPoint(Point{r3.Vector{0, 0, 1}}) {
		b = Rect{r1.Interval{b.Lat.Lo, math.Pi / 2}, s1.FullInterval()}
	}
	// If a loop contains the south pole, then either it wraps entirely
	// around the sphere (full longitude range), or it also contains the
	// north pole in which case b.Lng.IsFull() due to the test above.
	// Either way, we only need to do the south pole containment test if
	// b.Lng.IsFull().
	if b.Lng.IsFull() && l.ContainsPoint(Point{r3.Vector{0, 0, -1}}) {
		b.Lat.Lo = -math.Pi / 2
	}
	l.bound = b
	l.subregionBound = ExpandForSubregions(l.bound)
}

// ContainsOrigin reports true if this loop contains s2.OriginPoint().
func (l *Loop) ContainsOrigin() bool {
	return l.originInside
}

// HasInterior returns true because all loops have an interior.
func (l *Loop) HasInterior() bool {
	return true
}

// NumEdges returns the number of edges in this shape.
func (l *Loop) NumEdges() int {
	if l.isEmptyOrFull() {
		return 0
	}
	return len(l.vertices)
}

// Edge returns the endpoints for the given edge index.
func (l *Loop) Edge(i int) Edge {
	return Edge{l.Vertex(i), l.Vertex(i + 1)}
}

// NumChains reports the number of contiguous edge chains in the Loop.
func (l *Loop) NumChains() int {
	if l.isEmptyOrFull() {
		return 0
	}
	return 1
}

// Chain returns the i-th edge chain in the Shape.
func (l *Loop) Chain(chainID int) Chain {
	return Chain{0, l.NumEdges()}
}

// ChainEdge returns the j-th edge of the i-th edge chain.
func (l *Loop) ChainEdge(chainID, offset int) Edge {
	return Edge{l.Vertex(offset), l.Vertex(offset + 1)}
}

// ChainPosition returns a ChainPosition pair (i, j) such that edgeID is the
// j-th edge of the Loop.
func (l *Loop) ChainPosition(edgeID int) ChainPosition {
	return ChainPosition{0, edgeID}
}

// dimension returns the dimension of the geometry represented by this Loop.
func (l *Loop) dimension() dimension { return polygonGeometry }

// IsEmpty reports true if this is the special "empty" loop that contains no points.
func (l *Loop) IsEmpty() bool {
	return l.isEmptyOrFull() && !l.ContainsOrigin()
}

// IsFull reports true if this is the special "full" loop that contains all points.
func (l *Loop) IsFull() bool {
	return l.isEmptyOrFull() && l.ContainsOrigin()
}

// isEmptyOrFull reports true if this loop is either the "empty" or "full" special loops.
func (l *Loop) isEmptyOrFull() bool {
	return len(l.vertices) == 1
}

// Vertices returns the vertices in the loop.
func (l *Loop) Vertices() []Point {
	return l.vertices
}

// RectBound returns a tight bounding rectangle. If the loop contains the point,
// the bound also contains it.
func (l *Loop) RectBound() Rect {
	return l.bound
}

// CapBound returns a bounding cap that may have more padding than the corresponding
// RectBound. The bound is conservative such that if the loop contains a point P,
// the bound also contains it.
func (l *Loop) CapBound() Cap {
	return l.bound.CapBound()
}

// Vertex returns the vertex for the given index. For convenience, the vertex indices
// wrap automatically for methods that do index math such as Edge.
// i.e., Vertex(NumEdges() + n) is the same as Vertex(n).
func (l *Loop) Vertex(i int) Point {
	return l.vertices[i%len(l.vertices)]
}

// NumVertices returns the number of vertices in this loop.
func (l *Loop) NumVertices() int {
	return len(l.vertices)
}

// bruteForceContainsPoint reports if the given point is contained by this loop.
// This method does not use the ShapeIndex, so it is only preferable below a certain
// size of loop.
func (l *Loop) bruteForceContainsPoint(p Point) bool {
	origin := OriginPoint()
	inside := l.originInside
	crosser := NewChainEdgeCrosser(origin, p, l.Vertex(0))
	for i := 1; i <= len(l.vertices); i++ { // add vertex 0 twice
		inside = inside != crosser.EdgeOrVertexChainCrossing(l.Vertex(i))
	}
	return inside
}

// ContainsPoint returns true if the loop contains the point.
func (l *Loop) ContainsPoint(p Point) bool {
	// Empty and full loops don't need a special case, but invalid loops with
	// zero vertices do, so we might as well handle them all at once.
	if len(l.vertices) < 3 {
		return l.originInside
	}

	// For small loops, and during initial construction, it is faster to just
	// check all the crossing.
	const maxBruteForceVertices = 32
	if len(l.vertices) < maxBruteForceVertices || l.index == nil {
		return l.bruteForceContainsPoint(p)
	}

	// Otherwise, look up the point in the index.
	it := l.index.Iterator()
	if !it.LocatePoint(p) {
		return false
	}
	return l.iteratorContainsPoint(it, p)
}

// ContainsCell reports whether the given Cell is contained by this Loop.
func (l *Loop) ContainsCell(target Cell) bool {
	it := l.index.Iterator()
	relation := it.LocateCellID(target.ID())

	// If "target" is disjoint from all index cells, it is not contained.
	// Similarly, if "target" is subdivided into one or more index cells then it
	// is not contained, since index cells are subdivided only if they (nearly)
	// intersect a sufficient number of edges.  (But note that if "target" itself
	// is an index cell then it may be contained, since it could be a cell with
	// no edges in the loop interior.)
	if relation != Indexed {
		return false
	}

	// Otherwise check if any edges intersect "target".
	if l.boundaryApproxIntersects(it, target) {
		return false
	}

	// Otherwise check if the loop contains the center of "target".
	return l.iteratorContainsPoint(it, target.Center())
}

// IntersectsCell reports whether this Loop intersects the given cell.
func (l *Loop) IntersectsCell(target Cell) bool {
	it := l.index.Iterator()
	relation := it.LocateCellID(target.ID())

	// If target does not overlap any index cell, there is no intersection.
	if relation == Disjoint {
		return false
	}
	// If target is subdivided into one or more index cells, there is an
	// intersection to within the S2ShapeIndex error bound (see Contains).
	if relation == Subdivided {
		return true
	}
	// If target is an index cell, there is an intersection because index cells
	// are created only if they have at least one edge or they are entirely
	// contained by the loop.
	if it.CellID() == target.id {
		return true
	}
	// Otherwise check if any edges intersect target.
	if l.boundaryApproxIntersects(it, target) {
		return true
	}
	// Otherwise check if the loop contains the center of target.
	return l.iteratorContainsPoint(it, target.Center())
}

// boundaryApproxIntersects reports if the loop's boundary intersects target.
// It may also return true when the loop boundary does not intersect target but
// some edge comes within the worst-case error tolerance.
//
// This requires that it.Locate(target) returned Indexed.
func (l *Loop) boundaryApproxIntersects(it *ShapeIndexIterator, target Cell) bool {
	aClipped := it.IndexCell().findByShapeID(0)

	// If there are no edges, there is no intersection.
	if len(aClipped.edges) == 0 {
		return false
	}

	// We can save some work if target is the index cell itself.
	if it.CellID() == target.ID() {
		return true
	}

	// Otherwise check whether any of the edges intersect target.
	maxError := (faceClipErrorUVCoord + intersectsRectErrorUVDist)
	bound := target.BoundUV().ExpandedByMargin(maxError)
	for _, ai := range aClipped.edges {
		v0, v1, ok := ClipToPaddedFace(l.Vertex(ai), l.Vertex(ai+1), target.Face(), maxError)
		if ok && edgeIntersectsRect(v0, v1, bound) {
			return true
		}
	}
	return false
}

// iteratorContainsPoint reports if the iterator that is positioned at the ShapeIndexCell
// that may contain p, contains the point p.
func (l *Loop) iteratorContainsPoint(it *ShapeIndexIterator, p Point) bool {
	// Test containment by drawing a line segment from the cell center to the
	// given point and counting edge crossings.
	aClipped := it.IndexCell().findByShapeID(0)
	inside := aClipped.containsCenter
	if len(aClipped.edges) > 0 {
		center := it.Center()
		crosser := NewEdgeCrosser(center, p)
		aiPrev := -2
		for _, ai := range aClipped.edges {
			if ai != aiPrev+1 {
				crosser.RestartAt(l.Vertex(ai))
			}
			aiPrev = ai
			inside = inside != crosser.EdgeOrVertexChainCrossing(l.Vertex(ai+1))
		}
	}
	return inside
}

// RegularLoop creates a loop with the given number of vertices, all
// located on a circle of the specified radius around the given center.
func RegularLoop(center Point, radius s1.Angle, numVertices int) *Loop {
	return RegularLoopForFrame(getFrame(center), radius, numVertices)
}

// RegularLoopForFrame creates a loop centered around the z-axis of the given
// coordinate frame, with the first vertex in the direction of the positive x-axis.
func RegularLoopForFrame(frame matrix3x3, radius s1.Angle, numVertices int) *Loop {
	return LoopFromPoints(regularPointsForFrame(frame, radius, numVertices))
}

// CanonicalFirstVertex returns a first index and a direction (either +1 or -1)
// such that the vertex sequence (first, first+dir, ..., first+(n-1)*dir) does
// not change when the loop vertex order is rotated or inverted. This allows the
// loop vertices to be traversed in a canonical order. The return values are
// chosen such that (first, ..., first+n*dir) are in the range [0, 2*n-1] as
// expected by the Vertex method.
func (l *Loop) CanonicalFirstVertex() (firstIdx, direction int) {
	firstIdx = 0
	n := len(l.vertices)
	for i := 1; i < n; i++ {
		if l.Vertex(i).Cmp(l.Vertex(firstIdx).Vector) == -1 {
			firstIdx = i
		}
	}

	// 0 <= firstIdx <= n-1, so (firstIdx+n*dir) <= 2*n-1.
	if l.Vertex(firstIdx+1).Cmp(l.Vertex(firstIdx+n-1).Vector) == -1 {
		return firstIdx, 1
	}

	// n <= firstIdx <= 2*n-1, so (firstIdx+n*dir) >= 0.
	firstIdx += n
	return firstIdx, -1
}

// TurningAngle returns the sum of the turning angles at each vertex. The return
// value is positive if the loop is counter-clockwise, negative if the loop is
// clockwise, and zero if the loop is a great circle. Degenerate and
// nearly-degenerate loops are handled consistently with Sign. So for example,
// if a loop has zero area (i.e., it is a very small CCW loop) then the turning
// angle will always be negative.
//
// This quantity is also called the "geodesic curvature" of the loop.
func (l *Loop) TurningAngle() float64 {
	// For empty and full loops, we return the limit value as the loop area
	// approaches 0 or 4*Pi respectively.
	if l.isEmptyOrFull() {
		if l.ContainsOrigin() {
			return -2 * math.Pi
		}
		return 2 * math.Pi
	}

	// Don't crash even if the loop is not well-defined.
	if len(l.vertices) < 3 {
		return 0
	}

	// To ensure that we get the same result when the vertex order is rotated,
	// and that the result is negated when the vertex order is reversed, we need
	// to add up the individual turn angles in a consistent order. (In general,
	// adding up a set of numbers in a different order can change the sum due to
	// rounding errors.)
	//
	// Furthermore, if we just accumulate an ordinary sum then the worst-case
	// error is quadratic in the number of vertices. (This can happen with
	// spiral shapes, where the partial sum of the turning angles can be linear
	// in the number of vertices.) To avoid this we use the Kahan summation
	// algorithm (http://en.wikipedia.org/wiki/Kahan_summation_algorithm).
	n := len(l.vertices)
	i, dir := l.CanonicalFirstVertex()
	sum := TurnAngle(l.Vertex((i+n-dir)%n), l.Vertex(i), l.Vertex((i+dir)%n))

	compensation := s1.Angle(0)
	for n-1 > 0 {
		i += dir
		angle := TurnAngle(l.Vertex(i-dir), l.Vertex(i), l.Vertex(i+dir))
		oldSum := sum
		angle += compensation
		sum += angle
		compensation = (oldSum - sum) + angle
		n--
	}
	return float64(dir) * float64(sum+compensation)
}

// turningAngleMaxError return the maximum error in TurningAngle. The value is not
// constant; it depends on the loop.
func (l *Loop) turningAngleMaxError() float64 {
	// The maximum error can be bounded as follows:
	//   2.24 * dblEpsilon    for RobustCrossProd(b, a)
	//   2.24 * dblEpsilon    for RobustCrossProd(c, b)
	//   3.25 * dblEpsilon    for Angle()
	//   2.00 * dblEpsilon    for each addition in the Kahan summation
	//   ------------------
	//   9.73 * dblEpsilon
	maxErrorPerVertex := 9.73 * dblEpsilon
	return maxErrorPerVertex * float64(len(l.vertices))
}

// Encode encodes the Loop.
func (l Loop) Encode(w io.Writer) error {
	e := &encoder{w: w}
	l.encode(e)
	return e.err
}

func (l Loop) encode(e *encoder) {
	e.writeInt8(encodingVersion)
	e.writeUint32(uint32(len(l.vertices)))
	for _, v := range l.vertices {
		e.writeFloat64(v.X)
		e.writeFloat64(v.Y)
		e.writeFloat64(v.Z)
	}

	e.writeBool(l.originInside)
	// The depth of this loop within a polygon. Go does not currently track this value.
	e.writeInt32(0)

	// Encode the bound.
	l.bound.encode(e)
}

// Decode decodes a loop.
func (l *Loop) Decode(r io.Reader) error {
	*l = Loop{}
	d := &decoder{r: asByteReader(r)}
	version := int8(d.readUint8())
	if version != encodingVersion {
		return fmt.Errorf("cannot decode version %d, only %d", version, encodingVersion)
	}
	l.decode(d)
	return d.err
}

func (l *Loop) decode(d *decoder) {
	// Empty loops are explicitly allowed here: a newly created loop has zero vertices
	// and such loops encode and decode properly.
	nvertices := d.readUint32()
	if nvertices > maxEncodedVertices {
		if d.err == nil {
			d.err = fmt.Errorf("too many vertices (%d; max is %d)", nvertices, maxEncodedVertices)

		}
		return
	}
	l.vertices = make([]Point, nvertices)
	for i := range l.vertices {
		l.vertices[i].X = d.readFloat64()
		l.vertices[i].Y = d.readFloat64()
		l.vertices[i].Z = d.readFloat64()
	}
	l.originInside = d.readBool()
	l.depth = int(d.readUint32())
	l.bound.decode(d)
	l.subregionBound = ExpandForSubregions(l.bound)

	l.index = NewShapeIndex()
	l.index.Add(l)
}

// Bitmasks to read from properties.
const (
	originInside = 1 << iota
	boundEncoded
)

func (l *Loop) xyzFaceSiTiVertices() []xyzFaceSiTi {
	ret := make([]xyzFaceSiTi, len(l.vertices))
	for i, v := range l.vertices {
		ret[i].xyz = v
		ret[i].face, ret[i].si, ret[i].ti, ret[i].level = xyzToFaceSiTi(v)
	}
	return ret
}

func (l *Loop) encodeCompressed(e *encoder, snapLevel int) {
	vertices := l.xyzFaceSiTiVertices()
	if len(vertices) > maxEncodedVertices {
		if e.err == nil {
			e.err = fmt.Errorf("too many vertices (%d; max is %d)", len(vertices), maxEncodedVertices)

		}
		return
	}
	e.writeUvarint(uint64(len(vertices)))
	encodePointsCompressed(e, vertices, snapLevel)

	props := l.compressedEncodingProperties()
	e.writeUvarint(props)
	e.writeUvarint(uint64(l.depth))
	if props&boundEncoded != 0 {
		l.bound.encode(e)
	}
}

func (l *Loop) compressedEncodingProperties() uint64 {
	var properties uint64
	if l.originInside {
		properties |= originInside
	}

	// Write whether there is a bound so we can change the threshold later.
	// Recomputing the bound multiplies the decode time taken per vertex
	// by a factor of about 3.5.  Without recomputing the bound, decode
	// takes approximately 125 ns / vertex.  A loop with 63 vertices
	// encoded without the bound will take ~30us to decode, which is
	// acceptable.  At ~3.5 bytes / vertex without the bound, adding
	// the bound will increase the size by <15%, which is also acceptable.
	const minVerticesForBound = 64
	if len(l.vertices) >= minVerticesForBound {
		properties |= boundEncoded
	}

	return properties
}

func (l *Loop) decodeCompressed(d *decoder, snapLevel int) {
	nvertices := d.readUvarint()
	if d.err != nil {
		return
	}
	if nvertices > maxEncodedVertices {
		d.err = fmt.Errorf("too many vertices (%d; max is %d)", nvertices, maxEncodedVertices)
		return
	}
	l.vertices = make([]Point, nvertices)
	decodePointsCompressed(d, snapLevel, l.vertices)
	properties := d.readUvarint()

	// Make sure values are valid before using.
	if d.err != nil {
		return
	}

	l.originInside = (properties & originInside) != 0

	l.depth = int(d.readUvarint())

	if masked := properties & (1 << boundEncoded); masked != 0 {
		l.bound.decode(d)
		if d.err != nil {
			return
		}
		l.subregionBound = ExpandForSubregions(l.bound)
	} else {
		l.initBound()
	}

	l.index = NewShapeIndex()
	l.index.Add(l)
}

// TODO(roberts): Differences from the C++ version:
// IsNormalized
// Normalize
// Invert
// Area
// Centroid
// DistanceToPoint
// DistanceToBoundary
// Project
// ProjectToBoundary
// ContainsLoop
// IntersectsLoop
// EqualsLoop
// LoopRelations
// FindVertex
// ContainsNested
// BoundaryEquals
// BoundaryApproxEquals
// BoundaryNear
// SurfaceIntegral
// CompareBoundary
// ContainsNonCrossingBoundary
