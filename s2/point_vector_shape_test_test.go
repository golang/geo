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
	"math/rand"
	"testing"
)

func TestPointVectorShapeBasics(t *testing.T) {
	const seed = 8675309
	rand.Seed(seed)

	const numPoints = 100
	var p pointVectorShape = make([]Point, numPoints)

	for i := 0; i < numPoints; i++ {
		p[i] = randomPoint()
	}

	shape := Shape(p)

	if shape.NumEdges() != numPoints {
		t.Errorf("shape.NumEdges() = %v, want %v", shape.NumEdges(), numPoints)
	}
	if shape.NumChains() != numPoints {
		t.Errorf("shape.NumChains() = %v, want %v", shape.NumChains(), numPoints)
	}
	if shape.dimension() != pointGeometry {
		t.Errorf("shape.dimension() = %v, want %v", shape.dimension(), pointGeometry)
	}

	rand.Seed(seed)
	for i := 0; i < numPoints; i++ {
		if shape.Chain(i).Start != i {
			t.Errorf("shape.Chain(%d).Start = %v, want %v", i, shape.Chain(i).Start, i)
		}
		if shape.Chain(i).Length != 1 {
			t.Errorf("shape.Chain(%d).Length = %v, want 1", i, shape.Chain(i).Length)
		}
		edge := shape.Edge(i)
		pt := randomPoint()

		if !pt.ApproxEqual(edge.V0) {
			t.Errorf("shape.Edge(%d).V0 = %v, want %v", i, edge.V0, pt)
		}
		if !pt.ApproxEqual(edge.V1) {
			t.Errorf("shape.Edge(%d).V1 = %v, want %v", i, edge.V1, pt)
		}
	}
}
