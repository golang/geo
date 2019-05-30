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

// Shape interface enforcement
var _ Shape = (*laxPolygon)(nil)

// laxPolygon represents a region defined by a collection of zero or more
// closed loops. The interior is the region to the left of all loops. This
// is similar to Polygon except that this class supports polygons
// with degeneracies. Degeneracies are of two types: degenerate edges (from a
// vertex to itself) and sibling edge pairs (consisting of two oppositely
// oriented edges). Degeneracies can represent either "shells" or "holes"
// depending on the loop they are contained by. For example, a degenerate
// edge or sibling pair contained by a "shell" would be interpreted as a
// degenerate hole. Such edges form part of the boundary of the polygon.
//
// Loops with fewer than three vertices are interpreted as follows:
// - A loop with two vertices defines two edges (in opposite directions).
// - A loop with one vertex defines a single degenerate edge.
// - A loop with no vertices is interpreted as the "full loop" containing
//   all points on the sphere. If this loop is present, then all other loops
//   must form degeneracies (i.e., degenerate edges or sibling pairs). For
//   example, two loops {} and {X} would be interpreted as the full polygon
//   with a degenerate single-point hole at X.
//
// laxPolygon does not have any error checking, and it is perfectly fine to
// create laxPolygon objects that do not meet the requirements below (e.g., in
// order to analyze or fix those problems). However, laxPolygons must satisfy
// some additional conditions in order to perform certain operations:
//
// - In order to be valid for point containment tests, the polygon must
//   satisfy the "interior is on the left" rule. This means that there must
//   not be any crossing edges, and if there are duplicate edges then all but
//   at most one of thm must belong to a sibling pair (i.e., the number of
//   edges in opposite directions must differ by at most one).
//
// - To be valid for polygon operations (BoundaryOperation), degenerate
//   edges and sibling pairs cannot coincide with any other edges. For
//   example, the following situations are not allowed:
//
//    {AA, AA}     // degenerate edge coincides with another edge
//    {AA, AB}     // degenerate edge coincides with another edge
//    {AB, BA, AB} // sibling pair coincides with another edge
//
// Note that laxPolygon is much faster to initialize and is more compact than
// Polygon, but unlike Polygon it does not have any built-in operations.
// Instead you should use ShapeIndex based operations such as BoundaryOperation,
// ClosestEdgeQuery, etc.
type laxPolygon struct {
	numLoops int
	vertices []Point

	numVerts           int
	cumulativeVertices []int
}

// laxPolygonFromPolygon creates a laxPolygon from the given Polygon.
func laxPolygonFromPolygon(p *Polygon) *laxPolygon {
	spans := make([][]Point, len(p.loops))
	for i, loop := range p.loops {
		if loop.IsFull() {
			spans[i] = []Point{} // Empty span.
		} else {
			spans[i] = make([]Point, len(loop.vertices))
			copy(spans[i], loop.vertices)
		}
	}
	return laxPolygonFromPoints(spans)
}

// laxPolygonFromPoints creates a laxPolygon from the given points.
func laxPolygonFromPoints(loops [][]Point) *laxPolygon {
	p := &laxPolygon{}
	p.numLoops = len(loops)
	if p.numLoops == 0 {
		p.numVerts = 0
		p.vertices = nil
	} else if p.numLoops == 1 {
		p.numVerts = len(loops[0])
		p.vertices = make([]Point, p.numVerts)
		copy(p.vertices, loops[0])
	} else {
		p.cumulativeVertices = make([]int, p.numLoops+1)
		numVertices := 0
		for i, loop := range loops {
			p.cumulativeVertices[i] = numVertices
			numVertices += len(loop)
		}

		p.cumulativeVertices[p.numLoops] = numVertices
		for _, points := range loops {
			p.vertices = append(p.vertices, points...)
		}
	}
	return p
}

// numVertices reports the total number of vertices in all loops.
func (p *laxPolygon) numVertices() int {
	if p.numLoops <= 1 {
		return p.numVerts
	}
	return p.cumulativeVertices[p.numLoops]
}

// numLoopVertices reports the total number of vertices in the given loop.
func (p *laxPolygon) numLoopVertices(i int) int {
	if p.numLoops == 1 {
		return p.numVerts
	}
	return p.cumulativeVertices[i+1] - p.cumulativeVertices[i]
}

// loopVertex returns the vertex from loop i at index j.
//
// This requires:
//     0 <= i < len(loops)
//     0 <= j < len(loop[i].vertices)
func (p *laxPolygon) loopVertex(i, j int) Point {
	if p.numLoops == 1 {
		return p.vertices[j]
	}

	return p.vertices[p.cumulativeVertices[i]+j]
}

func (p *laxPolygon) NumEdges() int { return p.numVertices() }

func (p *laxPolygon) Edge(e int) Edge {
	e1 := e + 1
	if p.numLoops == 1 {
		// wrap the end vertex if this is the last edge.
		if e1 == p.numVerts {
			e1 = 0
		}
		return Edge{p.vertices[e], p.vertices[e1]}
	}

	// TODO(roberts): If this turns out to be performance critical in tests
	// incorporate the maxLinearSearchLoops like in C++.

	// Check if e1 would cross a loop boundary in the set of all vertices.
	nextLoop := 0
	for p.cumulativeVertices[nextLoop] <= e {
		nextLoop++
	}

	// If so, wrap around to the first vertex of the loop.
	if e1 == p.cumulativeVertices[nextLoop] {
		e1 = p.cumulativeVertices[nextLoop-1]
	}

	return Edge{p.vertices[e], p.vertices[e1]}
}

func (p *laxPolygon) Dimension() int                 { return 2 }
func (p *laxPolygon) typeTag() typeTag               { return typeTagLaxPolygon }
func (p *laxPolygon) privateInterface()              {}
func (p *laxPolygon) IsEmpty() bool                  { return defaultShapeIsEmpty(p) }
func (p *laxPolygon) IsFull() bool                   { return defaultShapeIsFull(p) }
func (p *laxPolygon) ReferencePoint() ReferencePoint { return referencePointForShape(p) }
func (p *laxPolygon) NumChains() int                 { return p.numLoops }
func (p *laxPolygon) Chain(i int) Chain {
	if p.numLoops == 1 {
		return Chain{0, p.numVertices()}
	}
	start := p.cumulativeVertices[i]
	return Chain{start, p.cumulativeVertices[i+1] - start}
}

func (p *laxPolygon) ChainEdge(i, j int) Edge {
	n := p.numLoopVertices(i)
	k := 0
	if j+1 != n {
		k = j + 1
	}
	if p.numLoops == 1 {
		return Edge{p.vertices[j], p.vertices[k]}
	}
	base := p.cumulativeVertices[i]
	return Edge{p.vertices[base+j], p.vertices[base+k]}
}

func (p *laxPolygon) ChainPosition(e int) ChainPosition {
	if p.numLoops == 1 {
		return ChainPosition{0, e}
	}

	// TODO(roberts): If this turns out to be performance critical in tests
	// incorporate the maxLinearSearchLoops like in C++.

	// Find the index of the first vertex of the loop following this one.
	nextLoop := 1
	for p.cumulativeVertices[nextLoop] <= e {
		nextLoop++
	}

	return ChainPosition{p.cumulativeVertices[nextLoop] - p.cumulativeVertices[1], e - p.cumulativeVertices[nextLoop-1]}
}
