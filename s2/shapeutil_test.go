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

import (
	"testing"

	"github.com/golang/geo/s1"
)

// This file will contain a number of Shape utility types used in different
// parts of testing.
//
//  - edgeVectorShape: represents an arbitrary collection of edges.
//  - laxLoop:         like Loop but allows duplicate vertices & edges,
//                     more compact representation, faster to initialize.
//
//  TODO(roberts): Add remaining testing types here.

// Shape interface enforcement
var (
	_ Shape = (*edgeVectorShape)(nil)
	_ Shape = (*laxLoop)(nil)
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

func TestShapeutilEdgeVectorShapeEdgeAccess(t *testing.T) {
	shape := &edgeVectorShape{}

	const numEdges = 100
	for i := 0; i < numEdges; i++ {
		a := randomPoint()
		shape.Add(a, randomPoint())
	}
	if got, want := shape.NumEdges(), numEdges; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), numEdges; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.dimension(), polylineGeometry; got != want {
		t.Errorf("shape.dimension() = %v, want %v", got, want)
	}

	for i := 0; i < numEdges; i++ {
		if got, want := shape.Chain(i).Start, i; got != want {
			t.Errorf("shape.Chain(i).Start = %v, want %v", got, want)
		}
		if got, want := shape.Chain(i).Length, 1; got != want {
			t.Errorf("shape.Chain(i).Length = %v, want %v", got, want)
		}
	}
}

func TestEdgeVectorShapeSingletonConstructor(t *testing.T) {
	a := PointFromCoords(1, 0, 0)
	b := PointFromCoords(0, 1, 0)

	var shape Shape = edgeVectorShapeFromPoints(a, b)
	if shape.NumEdges() != 1 {
		t.Errorf("shape created from one edge should only have one edge, got %v", shape.NumEdges())
	}
	if shape.NumChains() != 1 {
		t.Errorf("should only have one edge got %v", shape.NumChains())
	}
	edge := shape.Edge(0)

	if edge.V0 != a {
		t.Errorf("vertex 0 of the edge should be the same as was used to create it. got %v, want %v", edge.V0, a)
	}
	if edge.V1 != b {
		t.Errorf("vertex 1 of the edge should be the same as was used to create it. got %v, want %v", edge.V1, b)
	}
}

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
func (l *laxLoop) NumChains() int    { return min(1, l.numVertices) }
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

func TestShapeutilLaxLoopEmptyLoop(t *testing.T) {
	shape := Shape(laxLoopFromLoop(EmptyLoop()))

	if shape.NumEdges() != 0 {
		t.Errorf("empty laxLoop.NumEdges() = %v, want 0", shape.NumEdges())
	}
	if shape.NumChains() != 0 {
		t.Errorf("empty laxLoop.NumChains() = %v, want 0", shape.NumChains())
	}
	if shape.dimension() != polygonGeometry {
		t.Errorf("laxLoop.dimension() = %v, want %v", shape.dimension(), polygonGeometry)
	}
	if shape.ReferencePoint().Contained {
		t.Errorf("empty laxLoop.ReferencePoint().Contained should be false")
	}
}

func TestShapeutilLaxLoopNonEmptyLoop(t *testing.T) {
	vertices := parsePoints("0:0, 0:1, 1:1, 1:0")
	shape := Shape(laxLoopFromPoints(vertices))
	if got, want := len(shape.(*laxLoop).vertices), len(vertices); got != want {
		t.Errorf("laxLoop.numVertices = %v, want %v", got, want)
	}
	if got, want := shape.NumEdges(), len(vertices); got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), 1; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Chain(0).Start, 0; got != want {
		t.Errorf("shape.Chain(0).Start = %v, want %v", got, want)
	}
	if got, want := shape.Chain(0).Length, len(vertices); got != want {
		t.Errorf("shape.Chain(0).Length = %v, want %v", got, want)
	}
	for i := 0; i < len(vertices); i++ {
		if got, want := shape.(*laxLoop).vertex(i), vertices[i]; got != want {
			t.Errorf("%d. vertex(%d) = %v, want %v", i, i, got, want)
		}
		edge := shape.Edge(i)
		if vertices[i] != edge.V0 {
			t.Errorf("%d. edge.V0 = %v, want %v", i, edge.V0, vertices[i])
		}
		if got, want := edge.V1, vertices[(i+1)%len(vertices)]; got != want {
			t.Errorf("%d. edge.V1 = %v, want %v", i, got, want)
		}
	}
	if got, want := shape.dimension(), polygonGeometry; got != want {
		t.Errorf("shape.dimension() = %v, want %v", got, want)
	}
	if !shape.HasInterior() {
		t.Errorf("shape.HasInterior() = false, want true")
	}
	if shape.ReferencePoint().Contained {
		t.Errorf("shape.ReferencePoint().Contained = true, want false")
	}
}

// TODO(roberts): Uncomment once LaxPolyline and LaxPolygon are added to here.
/*
func TestShapeutilContainsBruteForceNoInterior(t *testing.T) {
	// Defines a polyline that almost entirely encloses the point 0:0.
	polyline := makeLaxPolyline("0:0, 0:1, 1:-1, -1:-1, -1e9:1")
	if containsBruteForce(polyline, parsePoint("0:0")) {
		t.Errorf("containsBruteForce(%v, %v) = true, want false")
	}
}

func TestShapeutilContainsBruteForceContainsReferencePoint(t *testing.T) {
	// Checks that containsBruteForce agrees with ReferencePoint.
	polygon := makeLaxPolygon("0:0, 0:1, 1:-1, -1:-1, -1e9:1")
	ref, _ := polygon.ReferencePoint()
	if got := containsBruteForce(polygon, ref.point); got != ref.contained {
		t.Errorf("containsBruteForce(%v, %v) = %v, want %v", polygon, ref.Point, got, ref.contained)
	}
}
*/

func TestShapeutilContainsBruteForceConsistentWithLoop(t *testing.T) {
	// Checks that containsBruteForce agrees with Loop Contains.
	loop := RegularLoop(parsePoint("89:-179"), s1.Angle(10)*s1.Degree, 100)
	for i := 0; i < loop.NumVertices(); i++ {
		if got, want := loop.ContainsPoint(loop.Vertex(i)),
			containsBruteForce(loop, loop.Vertex(i)); got != want {
			t.Errorf("loop.ContainsPoint(%v) = %v, containsBruteForce(shape, %v) = %v, should be the same", loop.Vertex(i), got, loop.Vertex(i), want)
		}
	}
}
