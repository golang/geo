/*
Copyright 2017 Google Inc. All rights reserved.

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

// Shape interface enforcement
var _ Shape = (*laxLoop)(nil)

// laxLoop represents a closed loop of edges surrounding an interior
// region. It is similar to Loop except that this class allows
// duplicate vertices and edges. Loops may have any number of vertices,
// including 0, 1, or 2. (A one-vertex loop defines a degenerate edge
// consisting of a single point.)
//
// Note that laxLoop is faster to initialize and more compact than
// Loop, but does not support the same operations as Loop.
type laxLoop struct {
	numVertices int
	vertices    []Point
}

func laxLoopFromPoints(vertices []Point) *laxLoop {
	l := &laxLoop{
		numVertices: len(vertices),
		vertices:    make([]Point, len(vertices)),
	}
	copy(l.vertices, vertices)
	return l
}

func laxLoopFromLoop(loop *Loop) *laxLoop {
	if loop.IsFull() {
		panic("FullLoops are not yet supported")
	}
	if loop.IsEmpty() {
		return &laxLoop{}
	}

	l := &laxLoop{
		numVertices: len(loop.vertices),
		vertices:    make([]Point, len(loop.vertices)),
	}
	copy(l.vertices, loop.vertices)
	return l
}

func (l *laxLoop) vertex(i int) Point { return l.vertices[i] }
func (l *laxLoop) NumEdges() int      { return l.numVertices }
func (l *laxLoop) Edge(e int) Edge {
	e1 := e + 1
	if e1 == l.numVertices {
		e1 = 0
	}
	return Edge{l.vertices[e], l.vertices[e1]}

}
func (l *laxLoop) dimension() dimension { return polygonGeometry }
func (l *laxLoop) HasInterior() bool    { return l.dimension() == polygonGeometry }
func (l *laxLoop) ReferencePoint() ReferencePoint {
	// ReferencePoint interprets a loop with no vertices as full.
	if l.numVertices == 0 {
		return OriginReferencePoint(false)
	}
	return referencePointForShape(l)
}
func (l *laxLoop) NumChains() int    { return minInt(1, l.numVertices) }
func (l *laxLoop) Chain(i int) Chain { return Chain{0, l.numVertices} }
func (l *laxLoop) ChainEdge(i, j int) Edge {
	var k int
	if j+1 == l.numVertices {
		k = j + 1
	}
	return Edge{l.vertices[j], l.vertices[k]}
}
func (l *laxLoop) ChainPosition(e int) ChainPosition {
	return ChainPosition{0, e}
}
