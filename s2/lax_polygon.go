// Copyright 2023 Google Inc. All rights reserved.
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

import "slices"

// Shape interface enforcement
var _ Shape = (*LaxPolygon)(nil)

// LaxPolygon represents a region defined by a collection of zero or more
// closed loops. The interior is the region to the left of all loops. This
// is similar to Polygon except that this type supports polygons
// with degeneracies. Degeneracies are of two types: degenerate edges (from a
// vertex to itself) and sibling edge pairs (consisting of two oppositely
// oriented edges). Degeneracies can represent either "shells" or "holes"
// depending on the loop they are contained by. For example, a degenerate
// edge or sibling pair contained by a "shell" would be interpreted as a
// degenerate hole. Such edges form part of the boundary of the polygon.
//
// Loops with fewer than three vertices are interpreted as follows:
//   - A loop with two vertices defines two edges (in opposite directions).
//   - A loop with one vertex defines a single degenerate edge.
//   - A loop with no vertices is interpreted as the "full loop" containing
//     all points on the sphere. If this loop is present, then all other loops
//     must form degeneracies (i.e., degenerate edges or sibling pairs). For
//     example, two loops {} and {X} would be interpreted as the full polygon
//     with a degenerate single-point hole at X.
//
// LaxPolygon does not have any error checking, and it is perfectly fine to
// create LaxPolygon objects that do not meet the requirements below (e.g., in
// order to analyze or fix those problems). However, LaxPolygons must satisfy
// some additional conditions in order to perform certain operations:
//
//   - In order to be valid for point containment tests, the polygon must
//     satisfy the "interior is on the left" rule. This means that there must
//     not be any crossing edges, and if there are duplicate edges then all but
//     at most one of them must belong to a sibling pair (i.e., the number of
//     edges in opposite directions must differ by at most one).
//
//   - To be valid for polygon operations (BooleanOperation), degenerate
//     edges and sibling pairs cannot coincide with any other edges. For
//     example, the following situations are not allowed:
//
//     {AA, AA}     // degenerate edge coincides with another edge
//     {AA, AB}     // degenerate edge coincides with another edge
//     {AB, BA, AB} // sibling pair coincides with another edge
//
// Note that LaxPolygon is much faster to initialize and is more compact than
// Polygon, but unlike Polygon it does not have any built-in operations.
// Instead you should use ShapeIndex based operations such as BooleanOperation,
// ClosestEdgeQuery, etc.
type LaxPolygon struct {
	numLoops int
	vertices []Point

	numVerts int
	// when numLoops > 1, store a list of size (numLoops+1) where "i"
	// represents the total number of vertices in loops 0..i-1.
	loopStarts []int
}

// LaxPolygonFromLoops creates a LaxPolygon from the given set of Loops.
//
// A collection of Loops is similar, but not the same as a Polygon, so
// from this creation method, we do not need to track Loop orientation
// as hole or shell like Polygon does.
func LaxPolygonFromLoops(loops []Loop) *LaxPolygon {
	spans := make([][]Point, len(loops))
	for i, loop := range loops {
		if loop.IsFull() {
			spans[i] = []Point{} // Empty span.
		} else {
			spans[i] = make([]Point, len(loop.vertices))
			copy(spans[i], loop.vertices)
		}
	}
	return LaxPolygonFromPoints(spans)
}

// LaxPolygonFromPolygon creates a LaxPolygon from the given Polygon.
func LaxPolygonFromPolygon(p *Polygon) *LaxPolygon {
	spans := make([][]Point, len(p.loops))
	for i, loop := range p.loops {
		if loop.IsFull() {
			spans[i] = []Point{} // Empty span.
		} else {
			spans[i] = make([]Point, len(loop.vertices))
			copy(spans[i], loop.vertices)
		}
	}
	lax := LaxPolygonFromPoints(spans)

	// Polygon and LaxPolygonShape holes are oriented oppositely, so we need
	// to reverse the orientation of any loops representing holes.
	for i := 0; i < p.NumLoops(); i++ {
		if p.Loop(i).IsHole() {
			v0 := lax.loopStarts[i]
			slices.Reverse(lax.vertices[v0 : v0+lax.numLoopVertices(i)])
		}
	}
	return lax
}

// LaxPolygonFromPoints creates a LaxPolygon from the given points.
func LaxPolygonFromPoints(loops [][]Point) *LaxPolygon {
	p := &LaxPolygon{}
	p.numLoops = len(loops)
	switch p.numLoops {
	case 0:
		p.numVerts = 0
		p.vertices = nil
	case 1:
		p.numVerts = len(loops[0])
		p.vertices = make([]Point, p.numVerts)
		p.loopStarts = []int{0, 0}
		copy(p.vertices, loops[0])
	default:
		p.numVerts = 0
		p.loopStarts = make([]int, p.numLoops+1)
		for i, loop := range loops {
			p.loopStarts[i] = p.numVerts
			p.numVerts += len(loop)
			p.vertices = append(p.vertices, loop...)
		}

		p.loopStarts[p.numLoops] = p.numVerts
	}
	return p
}

// numVertices reports the total number of vertices in all loops.
func (p *LaxPolygon) numVertices() int {
	if p.numLoops <= 1 {
		return p.numVerts
	}
	return p.loopStarts[p.numLoops]
}

// numLoopVertices reports the total number of vertices in the given loop.
func (p *LaxPolygon) numLoopVertices(i int) int {
	if p.numLoops == 1 {
		return p.numVerts
	}
	return p.loopStarts[i+1] - p.loopStarts[i]
}

// loopVertex returns the vertex from loop i at index j.
//
// This requires:
//
//	0 <= i < len(loops)
//	0 <= j < numLoopVertices(i)
func (p *LaxPolygon) loopVertex(i, j int) Point {
	if p.numLoops == 1 {
		return p.vertices[j]
	}

	return p.vertices[p.loopStarts[i]+j]
}

func (p *LaxPolygon) NumEdges() int { return p.numVertices() }

func (p *LaxPolygon) Edge(e int) Edge {
	pos := p.ChainPosition(e)
	return p.ChainEdge(pos.ChainID, pos.Offset)
}

func (p *LaxPolygon) Dimension() int                 { return 2 }
func (p *LaxPolygon) typeTag() typeTag               { return typeTagLaxPolygon }
func (p *LaxPolygon) privateInterface()              {}
func (p *LaxPolygon) IsEmpty() bool                  { return defaultShapeIsEmpty(p) }
func (p *LaxPolygon) IsFull() bool                   { return defaultShapeIsFull(p) }
func (p *LaxPolygon) ReferencePoint() ReferencePoint { return referencePointForShape(p) }
func (p *LaxPolygon) NumChains() int                 { return p.numLoops }
func (p *LaxPolygon) Chain(i int) Chain {
	if p.numLoops == 1 {
		return Chain{Start: 0, Length: p.numVertices()}
	}

	start := p.loopStarts[i]
	return Chain{Start: start, Length: p.loopStarts[i+1] - start}
}

func (p *LaxPolygon) ChainEdge(i, j int) Edge {
	n := p.numLoopVertices(i)
	k := 0
	if j+1 != n {
		k = j + 1
	}

	if p.numLoops == 1 {
		return Edge{V0: p.vertices[j], V1: p.vertices[k]}
	}

	start := p.loopStarts[i]
	return Edge{V0: p.vertices[start+j], V1: p.vertices[start+k]}
}

func (p *LaxPolygon) ChainPosition(edgeID int) ChainPosition {

	if p.numLoops == 1 {
		return ChainPosition{ChainID: 0, Offset: edgeID}
	}

	// We need the loopStart that is less than or equal to the edgeID.
	nextLoop := 0
	for p.loopStarts[nextLoop] <= edgeID {
		nextLoop++
	}

	return ChainPosition{
		ChainID: nextLoop - 1,
		Offset:  edgeID - p.loopStarts[nextLoop-1],
	}
}

// TODO(roberts): Remaining to port from C++:
// encode/decode
// Support for EncodedLaxPolygon
