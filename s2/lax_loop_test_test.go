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

import (
	"testing"
)

func TestLaxLoopEmptyLoop(t *testing.T) {
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

func TestLaxLoopNonEmptyLoop(t *testing.T) {
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
