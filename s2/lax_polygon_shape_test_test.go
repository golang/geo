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
	if shape.numLoops != 0 {
		t.Errorf("empty laxPolygon should have no loops")
	}

	if shape.numVertices() != 0 {
		t.Errorf("empty laxPolygon should have no vertices")
	}

	if shape.NumEdges() != 0 {
		t.Errorf("empty laxPolygon should have no edges")
	}

	if shape.NumChains() != 0 {
		t.Errorf("empty laxPolygon should have no chains")
	}

	if shape.dimension() != polygonGeometry {
		t.Errorf("laxPolygons dimension = %v, want polygonGeometry", shape.dimension())
	}

	if !shape.HasInterior() {
		t.Errorf("laxPolygons should have interiors")
	}
	if shape.ReferencePoint().Contained {
		t.Errorf("empty laxPolygon should not contain the reference point")
	}
}

func TestLaxPolygonFull(t *testing.T) {
	shape := laxPolygonFromPolygon(PolygonFromLoops([]*Loop{makeLoop("full")}))

	if shape.numLoops != 1 {
		t.Errorf("full laxPolygon should have 1 loop")
	}
	if shape.numVertices() != 0 {
		t.Errorf("full laxPolygon should have no vertices")
	}
	if shape.NumEdges() != 0 {
		t.Errorf("full laxPolygon should have no edges")
	}
	if shape.NumChains() != 1 {
		t.Errorf("full laxPolygon should have 1 chain")
	}
	if shape.dimension() != polygonGeometry {
		t.Errorf("laxPolygons dimension = %v, want polygonGeometry", shape.dimension())
	}
	if !shape.HasInterior() {
		t.Errorf("laxPolygons should have interiors")
	}
	if !shape.ReferencePoint().Contained {
		t.Errorf("full laxPolygon should contain the reference point")
	}
}

func TestLaxPolygonSingleVertexPolygon(t *testing.T) {
	// Polygon doesn't support single-vertex loops, so we need to construct
	// the laxPolygon directly.
	var loops [][]Point
	loops = append(loops, parsePoints("0:0"))

	shape := laxPolygonFromPoints(loops)

	if shape.numLoops != 1 {
		t.Errorf("laxPolygon with 1 point should have 1 loop")
	}
	if shape.numVertices() != 1 {
		t.Errorf("laxPolygon with 1 point should have 1 vertex")
	}
	if shape.NumEdges() != 1 {
		t.Errorf("laxPolygon with 1 point should have 1 edge")
	}
	if shape.NumChains() != 1 {
		t.Errorf("laxPolygon with 1 point should have 1 chain")
	}
	if shape.Chain(0).Start != 0 {
		t.Errorf("laxPolygon with 1 point should have chain that starts at vertex 0")
	}
	if shape.Chain(0).Length != 1 {
		t.Errorf("laxPolygon with 1 point should have chain of length 1")
	}

	edge := shape.Edge(0)
	if loops[0][0] != edge.V0 {
		t.Errorf("laxPolygon.Edge(0).V0 = %v, want %v", edge.V0, loops[0][0])
	}
	if loops[0][0] != edge.V1 {
		t.Errorf("laxPolygon.Edge(0).V0 = %v, want %v", edge.V1, loops[0][0])
	}
	if edge != shape.ChainEdge(0, 0) {
		t.Errorf("laxPolygon.Edge(0) should equal laxPolygon.ChainEdge(0, 0)")
	}
	if polygonGeometry != shape.dimension() {
		t.Errorf("laxPolygons dimension = %v, want polygonGeometry", shape.dimension())
	}
	if !shape.HasInterior() {
		t.Errorf("laxPolygons should have interiors")
	}
	if shape.ReferencePoint().Contained {
		t.Errorf("laxPolygon with 1 point should not contain the reference point")
	}
}

func TestLaxPolygonShapeSingleLoopPolygon(t *testing.T) {
	// test laxPolygonFromPolygon
	vertices := parsePoints("0:0, 0:1, 1:1, 1:0")
	lenVerts := len(vertices)
	shape := laxPolygonFromPolygon(PolygonFromLoops([]*Loop{LoopFromPoints(vertices)}))

	if shape.numLoops != 1 {
		t.Errorf("laxPolygon.numLoops = %d, want %d", shape.numLoops, 1)
	}
	if lenVerts != shape.numVertices() {
		t.Errorf("laxPolygon.numVertices() = %d, want %d", shape.numVertices(), lenVerts)
	}
	if lenVerts != shape.numLoopVertices(0) {
		t.Errorf("laxPolygon.numLoopVertices(0) = %d, want %d", shape.numLoopVertices(0), lenVerts)
	}
	if lenVerts != shape.NumEdges() {
		t.Errorf("laxPolygon.NumEdges = %d, want %d", shape.NumEdges(), lenVerts)
	}
	if shape.NumChains() != 1 {
		t.Errorf("laxPolygon.NumChains = %d, want %d", shape.NumChains(), 1)
	}
	if shape.Chain(0).Start != 0 {
		t.Errorf("laxPolygon.Chain(0).Start = %d, want 0", shape.Chain(0).Start)
	}
	if lenVerts != shape.Chain(0).Length {
		t.Errorf("laxPolygon.Chain(0).Length = %d, want %d", shape.Chain(0).Length, lenVerts)
	}
	for i := 0; i < lenVerts; i++ {
		if vertices[i] != shape.loopVertex(0, i) {
			t.Errorf("shape.loopVertex(%d) = %v, want %v", i, shape.loopVertex(0, i), vertices[i])
		}

		edge := shape.Edge(i)
		if got, want := vertices[i], edge.V0; got != want {
			t.Errorf("laxPolygon.Edge(%d).V0 = %v, want %v", i, got, want)
		}
		if got, want := vertices[(i+1)%lenVerts], edge.V1; got != want {
			t.Errorf("laxPolygon.Edge(%d).V1 = %v, want %v", i, got, want)
		}
		if got, want := shape.ChainEdge(0, i).V0, edge.V0; got != want {
			t.Errorf("laxPolygon.ChainEdge(0, %d).V0 = %v, want %v", i, got, want)
		}
		if got, want := shape.ChainEdge(0, i).V1, edge.V1; got != want {
			t.Errorf("laxPolygon.ChainEdge(0, %d).V1 = %v, want %v", i, got, want)
		}
	}
	if polygonGeometry != shape.dimension() {
		t.Errorf("laxPolygons dimension = %v, want polygonGeometry", shape.dimension())
	}
	if !shape.HasInterior() {
		t.Errorf("laxPolygons should have interiors")
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

	if lenLoops != shape.numLoops {
		t.Errorf("laxPolygon.numLoops = %d, want %d", shape.numLoops, lenLoops)
	}

	numVertices := 0
	if lenLoops != shape.NumChains() {
		t.Errorf("laxPolygon.NumChains = %d, want %d", shape.NumChains(), lenLoops)
	}

	for i, loop := range loops {
		if len(loop) != shape.numLoopVertices(i) {
			t.Errorf("laxPolygon.numLoopVertices(%d) = %d, want %d", i, shape.numLoopVertices(i), len(loop))
		}
		if numVertices != shape.Chain(i).Start {
			t.Errorf("laxPolygon.Chain(%d).Start = %d, want %d", i, shape.Chain(i).Start, numVertices)
		}
		if len(loop) != shape.Chain(i).Length {
			t.Errorf("laxPolygon.Chain(%d).Length = %d, want %d", i, shape.Chain(i).Length, len(loop))
		}
		for j, pt := range loop {
			if pt != shape.loopVertex(i, j) {
				t.Errorf("laxPolygon.loopVertex(%d, %d) = %v, want %v", i, j, shape.loopVertex(i, j), pt)
			}
			edge := shape.Edge(numVertices + j)
			if pt != edge.V0 {
				t.Errorf("laxPolygon.Edge(%d).V0 = %v, want %v", numVertices+j, edge.V0, pt)
			}
			if got, want := loop[(j+1)%len(loop)], edge.V1; got != want {
				t.Errorf("laxPolygon.Edge(%d).V1 = %v, want %v", numVertices+j, got, want)
			}
		}
		numVertices += len(loop)
	}

	if numVertices != shape.numVertices() {
		t.Errorf("laxPolygon.numVertices() = %d, want %d", shape.numVertices(), numVertices)
	}
	if numVertices != shape.NumEdges() {
		t.Errorf("laxPolygon.NuymEdges() = %d, want %d", shape.NumEdges(), numVertices)
	}
	if polygonGeometry != shape.dimension() {
		t.Errorf("laxPolygons dimension = %v, want polygonGeometry", shape.dimension())
	}
	if !shape.HasInterior() {
		t.Errorf("laxPolygons should have interiors")
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
