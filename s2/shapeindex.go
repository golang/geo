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

// Shape defines an interface for any s2 type that needs to be indexable.
type Shape interface {
	// NumEdges returns the number of edges in this shape.
	NumEdges() int

	// Edge returns endpoints for the given edge index.
	Edge(i int) (a, b Point)

	// HasInterior returns true if this shape has an interior.
	// i.e. the Shape consists of one or more closed non-intersecting loops.
	HasInterior() bool

	// ContainsOrigin returns true if this shape contains s2.Origin.
	// Shapes that do not have an interior will return false.
	ContainsOrigin() bool
}

// A minimal check for types that should satisfy the Shape interface.
var (
	_ Shape = Loop{}
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

var (
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
)

// ShapeIndex indexes a set of Shapes, where a Shape is some collection of
// edges. A shape can be as simple as a single edge, or as complex as a set of loops.
// For Shapes that have interiors, the index makes it very fast to determine which
// Shape(s) contain a given point or region.
type ShapeIndex struct {
	// shapes maps all shapes to their index.
	shapes map[Shape]int32

	maxEdgesPerCell int

	// nextID tracks the next ID to hand out. IDs are not reused when shapes
	// are removed from the index.
	nextID int32
}

// NewShapeIndex creates a new ShapeIndex.
func NewShapeIndex() *ShapeIndex {
	return &ShapeIndex{
		maxEdgesPerCell: 10,
		shapes:          make(map[Shape]int32),
	}
}

// Add adds the given shape to the index and assign an ID to it.
func (s *ShapeIndex) Add(shape Shape) {
	s.shapes[shape] = s.nextID
	s.nextID++
}

// Remove removes the given shape from the index.
func (s *ShapeIndex) Remove(shape Shape) {
	delete(s.shapes, shape)
}

// Len reports the number of Shapes in this index.
func (s *ShapeIndex) Len() int {
	return len(s.shapes)
}

// Reset clears the contents of the index and resets it to its original state.
func (s *ShapeIndex) Reset() {
	s.shapes = make(map[Shape]int32)
	s.nextID = 0
}
