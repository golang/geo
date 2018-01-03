// Copyright 2017 Google Inc. All rights reserved.
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
var (
	_ Shape = (*edgeVectorShape)(nil)
)

// edgeVectorShape is a Shape representing an arbitrary set of edges. It
// is used for testing, but it can also be useful if you have, say, a
// collection of polylines and don't care about memory efficiency (since
// this type would store most of the vertices twice).
type edgeVectorShape struct {
	edges []Edge
}

// edgeVectorShapeFromPoints returns an edgeVectorShape of length 1 from the given points.
func edgeVectorShapeFromPoints(a, b Point) *edgeVectorShape {
	e := &edgeVectorShape{
		edges: []Edge{
			Edge{a, b},
		},
	}
	return e
}

// Add adds the given edge to the shape.
func (e *edgeVectorShape) Add(a, b Point) {
	e.edges = append(e.edges, Edge{a, b})
}
func (e *edgeVectorShape) NumEdges() int                          { return len(e.edges) }
func (e *edgeVectorShape) Edge(id int) Edge                       { return e.edges[id] }
func (e *edgeVectorShape) HasInterior() bool                      { return false }
func (e *edgeVectorShape) ReferencePoint() ReferencePoint         { return OriginReferencePoint(false) }
func (e *edgeVectorShape) NumChains() int                         { return len(e.edges) }
func (e *edgeVectorShape) Chain(chainID int) Chain                { return Chain{chainID, 1} }
func (e *edgeVectorShape) ChainEdge(chainID, offset int) Edge     { return e.edges[chainID] }
func (e *edgeVectorShape) ChainPosition(edgeID int) ChainPosition { return ChainPosition{edgeID, 0} }
func (e *edgeVectorShape) dimension() dimension                   { return polylineGeometry }
