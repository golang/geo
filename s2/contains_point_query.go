// Copyright 2018 Google Inc. All rights reserved.
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

// VertexModel defines whether shapes are considered to contain their vertices.
// Note that these definitions differ from the ones used by BooleanOperation.
//
// Note that points other than vertices are never contained by polylines.
// If you want need this behavior, use ClosestEdgeQuery's IsDistanceLess
// with a suitable distance threshold instead.
type VertexModel int

const (
	// VertexModelOpen means no shapes contain their vertices (not even
	// points). Therefore Contains(Point) returns true if and only if the
	// point is in the interior of some polygon.
	VertexModelOpen VertexModel = iota

	// VertexModelSemiOpen means that polygon point containment is defined
	// such that if several polygons tile the region around a vertex, then
	// exactly one of those polygons contains that vertex. Points and
	// polylines still do not contain any vertices.
	VertexModelSemiOpen

	// VertexModelClosed means all shapes contain their vertices (including
	// points and polylines).
	VertexModelClosed
)

// ContainsPointQuery determines whether one or more shapes in a ShapeIndex
// contain a given Point. The ShapeIndex may contain any number of points,
// polylines, and/or polygons (possibly overlapping). Shape boundaries may be
// modeled as Open, SemiOpen, or Closed (this affects whether or not shapes are
// considered to contain their vertices).
//
// Note that if you need to do a large number of point containment
// tests, it is more efficient to re-use the query rather than creating a new
// one each time.
type ContainsPointQuery struct {
	model VertexModel
	index *ShapeIndex
	iter  *ShapeIndexIterator
}

// NewContainsPointQuery creates a new instance of the ContainsPointQuery for the index
// and given vertex model choice.
func NewContainsPointQuery(index *ShapeIndex, model VertexModel) *ContainsPointQuery {
	return &ContainsPointQuery{
		index: index,
		model: model,
		iter:  index.Iterator(),
	}
}

// Contains reports whether any shape in the queries index contains the point p
// under the queries vertex model (Open, SemiOpen, or Closed).
func (q *ContainsPointQuery) Contains(p Point) bool {
	if !q.iter.LocatePoint(p) {
		return false
	}

	cell := q.iter.IndexCell()
	for _, clipped := range cell.shapes {
		if q.shapeContains(clipped, q.iter.Center(), p) {
			return true
		}
	}
	return false
}

// shapeContains reports whether the clippedShape from the iterator's center position contains
// the given point.
func (q *ContainsPointQuery) shapeContains(clipped *clippedShape, center, p Point) bool {
	inside := clipped.containsCenter
	numEdges := clipped.numEdges()
	if numEdges <= 0 {
		return inside
	}

	shape := q.index.Shape(clipped.shapeID)
	if shape.Dimension() != 2 {
		// Points and polylines can be ignored unless the vertex model is Closed.
		if q.model != VertexModelClosed {
			return false
		}

		// Otherwise, the point is contained if and only if it matches a vertex.
		for _, edgeID := range clipped.edges {
			edge := shape.Edge(edgeID)
			if edge.V0 == p || edge.V1 == p {
				return true
			}
		}
		return false
	}

	// Test containment by drawing a line segment from the cell center to the
	// given point and counting edge crossings.
	crosser := NewEdgeCrosser(center, p)
	for _, edgeID := range clipped.edges {
		edge := shape.Edge(edgeID)
		sign := crosser.CrossingSign(edge.V0, edge.V1)
		if sign == DoNotCross {
			continue
		}
		if sign == MaybeCross {
			// For the Open and Closed models, check whether p is a vertex.
			if q.model != VertexModelSemiOpen && (edge.V0 == p || edge.V1 == p) {
				return (q.model == VertexModelClosed)
			}
			// C++ plays fast and loose with the int <-> bool conversions here.
			if VertexCrossing(crosser.a, crosser.b, edge.V0, edge.V1) {
				sign = Cross
			} else {
				sign = DoNotCross
			}
		}
		inside = inside != (sign == Cross)
	}

	return inside
}

// ShapeContains reports whether the given shape contains the point under this
// queries vertex model (Open, SemiOpen, or Closed).
//
// This requires the shape belongs to this queries index.
func (q *ContainsPointQuery) ShapeContains(shape Shape, p Point) bool {
	if !q.iter.LocatePoint(p) {
		return false
	}

	clipped := q.iter.IndexCell().findByShapeID(q.index.idForShape(shape))
	if clipped == nil {
		return false
	}
	return q.shapeContains(clipped, q.iter.Center(), p)
}

// TODO(roberts): Remaining methods from C++
// func (q *ContainsPointQuery) ContainingShapes(p Point) []Shape
// type shapeVisitorFunc func(shape Shape) bool
// func (q *ContainsPointQuery) VisitContainingShapes(p Point, v shapeVisitorFunc) bool
// type edgeVisitorFunc func(shape ShapeEdge) bool
// func (q *ContainsPointQuery) VisitIncidentEdges(p Point, v edgeVisitorFunc) bool
