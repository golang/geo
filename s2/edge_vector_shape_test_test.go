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

// TODO(roberts): TestEdgeVectorShapeEdgeAccess
