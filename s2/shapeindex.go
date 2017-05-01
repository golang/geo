/*
Copyright 2016 Google Inc. All rights reserved.

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
	"github.com/golang/geo/r2"
)

// dimension defines the types of geometry dimensions that a Shape supports.
type dimension int

const (
	pointGeometry dimension = iota
	polylineGeometry
	polygonGeometry
)

// Shape defines an interface for any S2 type that needs to be indexable. A shape
// is a collection of edges that optionally defines an interior. It can be used to
// represent a set of points, a set of polylines, or a set of polygons.
type Shape interface {
	// NumEdges returns the number of edges in this shape.
	NumEdges() int

	// Edge returns endpoints for the given edge index.
	// Zero-length edges are allowed, and can be used to represent points.
	Edge(i int) (a, b Point)

	// numChains reports the number of contiguous edge chains in the shape.
	// For example, a shape whose edges are [AB, BC, CD, AE, EF] would consist
	// of two chains (AB,BC,CD and AE,EF). This method allows some algorithms
	// to be optimized by skipping over edge chains that do not affect the output.
	//
	// Note that it is always acceptable to implement this method by returning
	// NumEdges, i.e. every chain consists of a single edge.
	numChains() int

	// chainStart returns the id of the first edge in the i-th edge chain,
	// and returns NumEdges when i == numChains. For example, if there are
	// two chains AB,BC,CD and AE,EF, the chain starts would be [0, 3, 5].
	//
	// This requires the following:
	// 0 <= i <= numChains()
	// chainStart(0) == 0
	// chainStart(i) < chainStart(i+1)
	// chainStart(numChains()) == NumEdges()
	chainStart(i int) int

	// dimension returns the dimension of the geometry represented by this shape.
	//
	// Note that this method allows degenerate geometry of different dimensions
	// to be distinguished, e.g. it allows a point to be distinguished from a
	// polyline or polygon that has been simplified to a single point.
	dimension() dimension

	// HasInterior reports whether this shape has an interior. If so, it must be possible
	// to assemble the edges into a collection of non-crossing loops.  Edges may
	// be returned in any order, and edges may be oriented arbitrarily with
	// respect to the shape interior.  (However, note that some Shape types
	// may have stronger requirements.)
	HasInterior() bool

	// ContainsOrigin returns true if this shape contains s2.Origin.
	// Shapes that do not have an interior will return false.
	ContainsOrigin() bool
}

// A minimal check for types that should satisfy the Shape interface.
var (
	_ Shape = &Loop{}
	_ Shape = &Polygon{}
	_ Shape = &Polyline{}
)

// CellRelation describes the possible relationships between a target cell
// and the cells of the ShapeIndex. If the target is an index cell or is
// contained by an index cell, it is Indexed. If the target is subdivided
// into one or more index cells, it is Subdivided. Otherwise it is Disjoint.
type CellRelation int

// The possible CellRelations for a ShapeIndex.
const (
	Indexed CellRelation = iota
	Subdivided
	Disjoint
)

const (
	// cellPadding defines the total error when clipping an edge which comes
	// from two sources:
	// (1) Clipping the original spherical edge to a cube face (the face edge).
	//     The maximum error in this step is faceClipErrorUVCoord.
	// (2) Clipping the face edge to the u- or v-coordinate of a cell boundary.
	//     The maximum error in this step is edgeClipErrorUVCoord.
	// Finally, since we encounter the same errors when clipping query edges, we
	// double the total error so that we only need to pad edges during indexing
	// and not at query time.
	cellPadding = 2.0 * (faceClipErrorUVCoord + edgeClipErrorUVCoord)

	// cellSizeToLongEdgeRatio defines the cell size relative to the length of an
	// edge at which it is first considered to be long. Long edges do not
	// contribute toward the decision to subdivide a cell further. For example,
	// a value of 2.0 means that the cell must be at least twice the size of the
	// edge in order for that edge to be counted. There are two reasons for not
	// counting long edges: (1) such edges typically need to be propagated to
	// several children, which increases time and memory costs without much benefit,
	// and (2) in pathological cases, many long edges close together could force
	// subdivision to continue all the way to the leaf cell level.
	cellSizeToLongEdgeRatio = 1.0
)

// clippedShape represents the part of a shape that intersects a Cell.
// It consists of the set of edge ids that intersect that cell and a boolean
// indicating whether the center of the cell is inside the shape (for shapes
// that have an interior).
//
// Note that the edges themselves are not clipped; we always use the original
// edges for intersection tests so that the results will be the same as the
// original shape.
type clippedShape struct {
	// shapeID is the index of the shape this clipped shape is a part of.
	shapeID int32

	// containsCenter indicates if the center of the CellID this shape has been
	// clipped to falls inside this shape. This is false for shapes that do not
	// have an interior.
	containsCenter bool

	// edges is the ordered set of ShapeIndex original edge ids. Edges
	// are stored in increasing order of edge id.
	edges []int
}

// newClippedShape returns a new clipped shape for the given shapeID and number of expected edges.
func newClippedShape(id int32, numEdges int) *clippedShape {
	return &clippedShape{
		shapeID: id,
		edges:   make([]int, numEdges),
	}
}

// numEdges returns the number of edges that intersect the CellID of the Cell this was clipped to.
func (c *clippedShape) numEdges() int {
	return len(c.edges)
}

// containsEdge reports if this clipped shape contains the given edge id.
func (c *clippedShape) containsEdge(id int) bool {
	// Linear search is fast because the number of edges per shape is typically
	// very small (less than 10).
	for _, e := range c.edges {
		if e == id {
			return true
		}
	}
	return false
}

// ShapeIndexCell stores the index contents for a particular CellID.
type ShapeIndexCell struct {
	shapes []*clippedShape
}

// add adds the given clipped shape to this index cell.
func (s *ShapeIndexCell) add(c *clippedShape) {
	s.shapes = append(s.shapes, c)
}

// findByShapeID returns the clipped shape that contains the given shapeID,
// or nil if none of the clipped shapes contain it.
func (s *ShapeIndexCell) findByShapeID(shapeID int32) *clippedShape {
	// Linear search is fine because the number of shapes per cell is typically
	// very small (most often 1), and is large only for pathological inputs
	// (e.g. very deeply nested loops).
	for _, clipped := range s.shapes {
		if clipped.shapeID == shapeID {
			return clipped
		}
	}
	return nil
}

// faceEdge and clippedEdge store temporary edge data while the index is being
// updated.
//
// While it would be possible to combine all the edge information into one
// structure, there are two good reasons for separating it:
//
//  - Memory usage. Separating the two means that we only need to
//    store one copy of the per-face data no matter how many times an edge is
//    subdivided, and it also lets us delay computing bounding boxes until
//    they are needed for processing each face (when the dataset spans
//    multiple faces).
//
//  - Performance. UpdateEdges is significantly faster on large polygons when
//    the data is separated, because it often only needs to access the data in
//    clippedEdge and this data is cached more successfully.

// faceEdge represents an edge that has been projected onto a given face,
type faceEdge struct {
	shapeID     int32    // The ID of shape that this edge belongs to
	edgeID      int      // Edge ID within that shape
	maxLevel    int      // Not desirable to subdivide this edge beyond this level
	hasInterior bool     // Belongs to a shape that has an interior
	a, b        r2.Point // The edge endpoints, clipped to a given face
	va, vb      Point    // The original vertices of this edge.
}

// clippedEdge represents the portion of that edge that has been clipped to a given Cell.
type clippedEdge struct {
	faceEdge faceEdge // The original unclipped edge
	bound    r2.Rect  // Bounding box for the clipped portion
}

// ShapeIndexIterator is an iterator that provides low-level access to
// the cells of the index. Cells are returned in increasing order of CellID.
//
//   for it := index.Iterator(); !it.Done(); it.Next() {
//     fmt.Print(it.CellID())
//   }
//
type ShapeIndexIterator struct {
	index    *ShapeIndex
	position int
}

// CellID returns the CellID of the cell at the current position of the iterator.
func (s *ShapeIndexIterator) CellID() CellID {
	if s.position >= len(s.index.cells) {
		return 0
	}
	return s.index.cells[s.position]
}

// IndexCell returns the ShapeIndexCell at the current position of the iterator.
func (s *ShapeIndexIterator) IndexCell() *ShapeIndexCell {
	return s.index.cellMap[s.CellID()]
}

// Center returns the Point at the center of the current position of the iterator.
func (s *ShapeIndexIterator) Center() Point {
	return s.CellID().Point()
}

// Reset the iterator to be positioned at the first cell in the index.
func (s *ShapeIndexIterator) Reset() {
	if !s.index.IsFresh() {
		// TODO: handle the case when the index needs updating first.
	}
	s.position = 0
}

// AtBegin reports if the iterator is positioned at the first index cell.
func (s *ShapeIndexIterator) AtBegin() bool {
	return s.position == 0
}

// Next advances the iterator to the next cell in the index.
func (s *ShapeIndexIterator) Next() {
	s.position++
}

// Prev advances the iterator to the previous cell in the index.
// If the iterator is at the first cell the call does nothing.
func (s *ShapeIndexIterator) Prev() {
	if s.position > 0 {
		s.position--
	}
}

// Done reports if the iterator is positioned at or after the last index cell.
func (s *ShapeIndexIterator) Done() bool {
	return s.position >= len(s.index.cells)
}

// seek positions the iterator at the first cell whose ID >= target starting from the
// current position of the iterator, or at the end of the index if no such cell exists.
// If the iterator is currently at the end, nothing is done.
func (s *ShapeIndexIterator) seek(target CellID) {
	// In C++, this relies on the lower_bound method of the underlying btree_map.
	// TODO(roberts): Convert this to a binary search since the list of cells is ordered.
	for k, v := range s.index.cells {
		// We've passed the cell that is after us, so we are done.
		if v >= target {
			s.position = k
			break
		}
		// Otherwise, advance the position.
		s.position++
	}
}

// seekForward advances the iterator to the next cell with cellID >= target if the
// iterator is not Done or already satisfies the condition.
func (s *ShapeIndexIterator) seekForward(target CellID) {
	if !s.Done() && s.CellID() < target {
		s.seek(target)
	}
}

// indexStatus is an enumeration of states the index can be in.
type indexStatus int

const (
	stale    indexStatus = iota // There are pending updates.
	updating                    // Updates are currently being applied.
	fresh                       // There are no pending updates.
)

// ShapeIndex indexes a set of Shapes, where a Shape is some collection of edges
// that optionally defines an interior. It can be used to represent a set of
// points, a set of polylines, or a set of polygons. For Shapes that have
// interiors, the index makes it very fast to determine which Shape(s) contain
// a given point or region.
type ShapeIndex struct {
	// shapes is a map of shape ID to shape.
	shapes map[int]Shape

	// The maximum number of edges per cell.
	// TODO(roberts): Update the comments when the usage of this is implemented.
	maxEdgesPerCell int

	// nextID tracks the next ID to hand out. IDs are not reused when shapes
	// are removed from the index.
	nextID int

	// cellMap is a map from CellID to the set of clipped shapes that intersect that
	// cell. The cell ids cover a set of non-overlapping regions on the sphere.
	// In C++, this is a BTree, so the cells are ordered naturally by the data structure.
	cellMap map[CellID]*ShapeIndexCell
	// Track the ordered list of cell ids.
	cells []CellID

	// The current status of the index.
	status indexStatus
}

// NewShapeIndex creates a new ShapeIndex.
func NewShapeIndex() *ShapeIndex {
	return &ShapeIndex{
		maxEdgesPerCell: 10,
		shapes:          make(map[int]Shape),
		cellMap:         make(map[CellID]*ShapeIndexCell),
		cells:           nil,
		status:          fresh,
	}
}

// Iterator returns an iterator for this index.
func (s *ShapeIndex) Iterator() *ShapeIndexIterator {
	return &ShapeIndexIterator{index: s}
}

// Begin positions the iterator at the first cell in the index.
func (s *ShapeIndex) Begin() *ShapeIndexIterator {
	return &ShapeIndexIterator{index: s}
}

// End positions the iterator at the last cell in the index.
func (s *ShapeIndex) End() *ShapeIndexIterator {
	// TODO(roberts): It's possible that updates could happen to the index between
	// the time this is called and the time the iterators position is used and this
	// will be invalid or not the end. For now, things will be undefined if this
	// happens. See about referencing the IsFresh to guard for this in the future.
	return &ShapeIndexIterator{
		index:    s,
		position: len(s.cells),
	}
}

// Add adds the given shape to the index and assign an ID to it.
func (s *ShapeIndex) Add(shape Shape) {
	s.shapes[s.nextID] = shape
	s.nextID++
	s.status = stale
}

// Remove removes the given shape from the index.
func (s *ShapeIndex) Remove(shape Shape) {
	for k, v := range s.shapes {
		if v == shape {
			delete(s.shapes, k)
			s.status = stale
			return
		}
	}
}

// Len reports the number of Shapes in this index.
func (s *ShapeIndex) Len() int {
	return len(s.shapes)
}

// Reset resets the index to its original state.
func (s *ShapeIndex) Reset() {
	s.shapes = make(map[int]Shape)
	s.nextID = 0
	s.cellMap = make(map[CellID]*ShapeIndexCell)
	s.cells = nil
	s.status = fresh
}

// NumEdges returns the number of edges in this index.
func (s *ShapeIndex) NumEdges() int {
	numEdges := 0
	for _, shape := range s.shapes {
		numEdges += shape.NumEdges()
	}
	return numEdges
}

// IsFresh reports if there are no pending updates that need to be applied.
// This can be useful to avoid building the index unnecessarily, or for
// choosing between two different algorithms depending on whether the index
// is available.
//
// The returned index status may be slightly out of date if the index was
// built in a different thread. This is fine for the intended use (as an
// efficiency hint), but it should not be used by internal methods.
func (s *ShapeIndex) IsFresh() bool {
	return s.status == fresh
}
