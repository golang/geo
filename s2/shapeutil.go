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

// containsBruteForce reports whether the given shape contains the given point.
// Most clients should not use this method, since its running time is linear in
// the number of shape edges. Instead clients should create a ShapeIndex and use
// ContainsPointQuery, since this strategy is much more efficient when many
// points need to be tested.
//
// Polygon boundaries are treated as being semi-open (see ContainsPointQuery
// and VertexModel for other options).
func containsBruteForce(shape Shape, point Point) bool {
	if !shape.HasInterior() {
		return false
	}

	refPoint := shape.ReferencePoint()
	if refPoint.Point == point {
		return refPoint.Contained
	}

	crosser := NewEdgeCrosser(refPoint.Point, point)
	inside := refPoint.Contained
	for e := 0; e < shape.NumEdges(); e++ {
		edge := shape.Edge(e)
		inside = inside != crosser.EdgeOrVertexCrossing(edge.V0, edge.V1)
	}
	return inside
}
