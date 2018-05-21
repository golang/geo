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

import (
	"testing"
)

func TestLaxPolygonShapeEmptyPolygon(t *testing.T) {
	shape := laxPolygonFromPolygon((&Polygon{}))
	if got, want := shape.numLoops, 0; got != want {
		t.Errorf("shape.numLoops = %d, want %d", got, want)
	}
	if got, want := shape.numVertices(), 0; got != want {
		t.Errorf("shape.numVertices() = %d, want %d", got, want)
	}
	if got, want := shape.NumEdges(), 0; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), 0; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Dimension(), 2; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if !shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = false, want true")
	}
	if shape.IsFull() {
		t.Errorf("shape.IsFull() = true, want false")
	}
	if shape.ReferencePoint().Contained {
		t.Errorf("shape.ReferencePoint().Contained should be false")
	}
}

func TestLaxPolygonFull(t *testing.T) {
	shape := laxPolygonFromPolygon(PolygonFromLoops([]*Loop{makeLoop("full")}))
	if got, want := shape.numLoops, 1; got != want {
		t.Errorf("shape.numLoops = %d, want %d", got, want)
	}
	if got, want := shape.numVertices(), 0; got != want {
		t.Errorf("shape.numVertices() = %d, want %d", got, want)
	}
	if got, want := shape.NumEdges(), 0; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), 1; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Dimension(), 2; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = true, want false")
	}
	if !shape.IsFull() {
		t.Errorf("shape.IsFull() = false, want true")
	}
	if !shape.ReferencePoint().Contained {
		t.Errorf("shape.ReferencePoint().Contained = false, want true")
	}
}

func TestLaxPolygonSingleVertexPolygon(t *testing.T) {
	// Polygon doesn't support single-vertex loops, so we need to construct
	// the laxPolygon directly.
	var loops [][]Point
	loops = append(loops, parsePoints("0:0"))

	shape := laxPolygonFromPoints(loops)
	if got, want := shape.numLoops, 1; got != want {
		t.Errorf("shape.numLoops = %d, want %d", got, want)
	}
	if got, want := shape.numVertices(), 1; got != want {
		t.Errorf("shape.numVertices() = %d, want %d", got, want)
	}
	if got, want := shape.NumEdges(), 1; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), 1; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Chain(0).Start, 0; got != want {
		t.Errorf("shape.Chain(0).Start = %d, want %d", got, want)
	}
	if got, want := shape.Chain(0).Length, 1; got != want {
		t.Errorf("shape.Chain(0).Length = %d, want %d", got, want)
	}

	edge := shape.Edge(0)
	if loops[0][0] != edge.V0 {
		t.Errorf("shape.Edge(0).V0 = %v, want %v", edge.V0, loops[0][0])
	}
	if loops[0][0] != edge.V1 {
		t.Errorf("shape.Edge(0).V0 = %v, want %v", edge.V1, loops[0][0])
	}
	if edge != shape.ChainEdge(0, 0) {
		t.Errorf("shape.Edge(0) should equal shape.ChainEdge(0, 0)")
	}
	if got, want := shape.Dimension(), 2; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = true, want false")
	}
	if shape.IsFull() {
		t.Errorf("shape.IsFull() = true, want false")
	}
	if shape.ReferencePoint().Contained {
		t.Errorf("shape.ReferencePoint().Contained = true, want false")
	}
}

func TestLaxPolygonShapeSingleLoopPolygon(t *testing.T) {
	vertices := parsePoints("0:0, 0:1, 1:1, 1:0")
	lenVerts := len(vertices)
	shape := laxPolygonFromPolygon(PolygonFromLoops([]*Loop{LoopFromPoints(vertices)}))

	if got, want := shape.numLoops, 1; got != want {
		t.Errorf("shape.numLoops = %d, want %d", got, want)
	}
	if got, want := shape.numVertices(), lenVerts; got != want {
		t.Errorf("shape.numVertices() = %d, want %d", got, want)
	}
	if got, want := shape.numLoopVertices(0), lenVerts; got != want {
		t.Errorf("shape.numLoopVertices(0) = %d, want %d", got, want)
	}
	if got, want := shape.NumEdges(), lenVerts; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), 1; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Chain(0).Start, 0; got != want {
		t.Errorf("shape.Chain(0).Start = %d, want %d", got, want)
	}
	if got, want := shape.Chain(0).Length, lenVerts; got != want {
		t.Errorf("shape.Chain(0).Length = %d, want %d", got, want)
	}
	for i := 0; i < lenVerts; i++ {
		if got, want := shape.loopVertex(0, i), vertices[i]; got != want {
			t.Errorf("shape.loopVertex(%d) = %v, want %v", i, got, want)
		}

		edge := shape.Edge(i)
		if got, want := vertices[i], edge.V0; got != want {
			t.Errorf("shape.Edge(%d).V0 = %v, want %v", i, got, want)
		}
		if got, want := vertices[(i+1)%lenVerts], edge.V1; got != want {
			t.Errorf("shape.Edge(%d).V1 = %v, want %v", i, got, want)
		}
		if got, want := shape.ChainEdge(0, i).V0, edge.V0; got != want {
			t.Errorf("shape.ChainEdge(0, %d).V0 = %v, want %v", i, got, want)
		}
		if got, want := shape.ChainEdge(0, i).V1, edge.V1; got != want {
			t.Errorf("shape.ChainEdge(0, %d).V1 = %v, want %v", i, got, want)
		}
	}
	if got, want := shape.Dimension(), 2; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = true, want false")
	}
	if shape.IsFull() {
		t.Errorf("shape.IsFull() = true, want false")
	}
	if containsBruteForce(shape, OriginPoint()) {
		t.Errorf("containsBruteForce(%v, %v) = true, want false", shape, OriginPoint())
	}
}

func TestLaxPolygonShapeMultiLoopPolygon(t *testing.T) {
	// Test to make sure that the loops are oriented so that the interior of the
	// polygon is always on the left.
	loops := [][]Point{
		parsePoints("0:0, 0:3, 3:3"), // CCW
		parsePoints("1:1, 2:2, 1:2"), // CW
	}
	lenLoops := len(loops)
	shape := laxPolygonFromPoints(loops)
	if got, want := shape.numLoops, lenLoops; got != want {
		t.Errorf("shape.numLoops = %d, want %d", got, want)
	}
	if got, want := shape.NumChains(), lenLoops; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}

	numVertices := 0
	for i, loop := range loops {
		if got, want := shape.numLoopVertices(i), len(loop); got != want {
			t.Errorf("shape.numLoopVertices(%d) = %d, want %d", i, got, want)
		}
		if got, want := shape.Chain(i).Start, numVertices; got != want {
			t.Errorf("shape.Chain(%d).Start = %d, want %d", i, got, want)
		}
		if got, want := shape.Chain(i).Length, len(loop); got != want {
			t.Errorf("shape.Chain(%d).Length = %d, want %d", i, got, want)
		}
		for j, pt := range loop {
			if pt != shape.loopVertex(i, j) {
				t.Errorf("shape.loopVertex(%d, %d) = %v, want %v", i, j, shape.loopVertex(i, j), pt)
			}
			edge := shape.Edge(numVertices + j)
			if pt != edge.V0 {
				t.Errorf("shape.Edge(%d).V0 = %v, want %v", numVertices+j, edge.V0, pt)
			}
			if got, want := loop[(j+1)%len(loop)], edge.V1; got != want {
				t.Errorf("shape.Edge(%d).V1 = %v, want %v", numVertices+j, got, want)
			}
		}
		numVertices += len(loop)
	}

	if got, want := shape.numVertices(), numVertices; got != want {
		t.Errorf("shape.numVertices() = %d, want %d", got, want)
	}
	if got, want := shape.NumEdges(), numVertices; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.Dimension(), 2; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = true, want false")
	}
	if shape.IsFull() {
		t.Errorf("shape.IsFull() = true, want false")
	}
	if containsBruteForce(shape, OriginPoint()) {
		t.Errorf("containsBruteForce(%v, %v) = true, want false", shape, OriginPoint())
	}
}

func TestLaxPolygonShapeDegenerateLoops(t *testing.T) {
	loops := [][]Point{
		parsePoints("1:1, 1:2, 2:2, 1:2, 1:3, 1:2, 1:1"),
		parsePoints("0:0, 0:3, 0:6, 0:9, 0:6, 0:3, 0:0"),
		parsePoints("5:5, 6:6"),
	}

	shape := laxPolygonFromPoints(loops)
	if shape.ReferencePoint().Contained {
		t.Errorf("%v.ReferencePoint().Contained() = true, want false", shape)
	}
}

func TestLaxPolygonShapeInvertedLoops(t *testing.T) {
	loops := [][]Point{
		parsePoints("1:2, 1:1, 2:2"),
		parsePoints("3:4, 3:3, 4:4"),
	}
	shape := laxPolygonFromPoints(loops)

	if !containsBruteForce(shape, OriginPoint()) {
		t.Errorf("containsBruteForce(%v, %v) = false, want true", shape, OriginPoint())
	}
}

// TODO(roberts): TestLaxPolygonShapeCompareToLoop once fractal testing is added.
