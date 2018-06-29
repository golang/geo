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
	"math/rand"
	"testing"
)

func TestPointVectorEmpty(t *testing.T) {
	var shape PointVector

	if got, want := shape.NumEdges(), 0; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), 0; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Dimension(), 0; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if !shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = false, want true")
	}
	if shape.IsFull() {
		t.Errorf("shape.IsFull() = true, want false")
	}
	if shape.ReferencePoint().Contained {
		t.Errorf("shape.ReferencePoint().Contained = true, want false")
	}
}

func TestPointVectorBasics(t *testing.T) {
	const seed = 8675309
	rand.Seed(seed)

	const numPoints = 100
	var p PointVector = make([]Point, numPoints)

	for i := 0; i < numPoints; i++ {
		p[i] = randomPoint()
	}

	shape := Shape(&p)
	if got, want := shape.NumEdges(), numPoints; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), numPoints; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Dimension(), 0; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = true, want false")
	}
	if shape.IsFull() {
		t.Errorf("shape.IsFull() = true, want false")
	}

	rand.Seed(seed)
	for i := 0; i < numPoints; i++ {
		if got, want := shape.Chain(i).Start, i; got != want {
			t.Errorf("shape.Chain(%d).Start = %d, want %d", i, got, want)
		}
		if got, want := shape.Chain(i).Length, 1; got != want {
			t.Errorf("shape.Chain(%d).Length = %v, want %d", i, got, want)
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
