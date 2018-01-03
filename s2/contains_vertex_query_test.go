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

	"github.com/golang/geo/s1"
)

func TestContainsVertexQueryUndetermined(t *testing.T) {
	q := NewContainsVertexQuery(parsePoint("1:2"))
	q.AddEdge(parsePoint("3:4"), 1)
	q.AddEdge(parsePoint("3:4"), -1)
	if got := q.ContainsVertex(); got != 0 {
		t.Errorf("ContainsVertex() = %v, want 0 for vertex with undetermined containment", got)
	}
}

func TestContainsVertexQueryContainedWithDuplicates(t *testing.T) {
	// The Ortho reference direction points approximately due west.
	// Containment is determined by the unmatched edge immediately clockwise.
	q := NewContainsVertexQuery(parsePoint("0:0"))
	q.AddEdge(parsePoint("3:-3"), -1)
	q.AddEdge(parsePoint("1:-5"), 1)
	q.AddEdge(parsePoint("2:-4"), 1)
	q.AddEdge(parsePoint("1:-5"), -1)
	if got := q.ContainsVertex(); got != 1 {
		t.Errorf("ContainsVertex() = %v, want 1 for vertex that is contained", got)
	}
}

func TestContainsVertexQueryNotContainedWithDuplicates(t *testing.T) {
	// The Ortho reference direction points approximately due west.
	// Containment is determined by the unmatched edge immediately clockwise.
	q := NewContainsVertexQuery(parsePoint("1:1"))
	q.AddEdge(parsePoint("1:-5"), 1)
	q.AddEdge(parsePoint("2:-4"), -1)
	q.AddEdge(parsePoint("3:-3"), 1)
	q.AddEdge(parsePoint("1:-5"), -1)
	if got := q.ContainsVertex(); got != -1 {
		t.Errorf("ContainsVertex() = %v, want -1 for vertex that is not contained", got)
	}
}

func TestContainsVertexQueryMatchesLoopContainment(t *testing.T) {
	// Check that the containment function defined is compatible with Loop
	loop := RegularLoop(parsePoint("89:-179"), s1.Angle(10)*s1.Degree, 1000)
	for i := 1; i <= loop.NumVertices(); i++ {
		q := NewContainsVertexQuery(loop.Vertex(i))
		q.AddEdge(loop.Vertex(i-1), -1)
		q.AddEdge(loop.Vertex(i+1), 1)
		if got, want := q.ContainsVertex() > 0, loop.ContainsPoint(loop.Vertex(i)); got != want {
			t.Errorf("ContainsVertex() = %v, loop.ContainsPoint(%v) = %v, should be the same", got, loop.Vertex(i), want)
		}
	}
}
