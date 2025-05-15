// Copyright 2025 The S2 Geometry Project Authors. All rights reserved.
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

// incidentEdgeKey is a tuple of (shape id, vertex) that compares by shape id.
type incidentEdgeKey struct {
	shapeID int32
	vertex  Point
}

// We need a strict ordering to be a valid key for an ordered container, but
// we don't actually care about the ordering of the vertices (as long as
// they're grouped by shape id). Vertices are 3D points so they don't have a
// natural ordering, so we'll just compare them lexicographically.
func (i incidentEdgeKey) Cmp(o incidentEdgeKey) int {
	if i.shapeID < o.shapeID {
		return -1
	}
	if i.shapeID > o.shapeID {
		return 1
	}

	return i.vertex.Cmp(o.vertex.Vector)
}

// vertexEdge is a tuple of vertex and edgeID for processing incident edges.
type vertexEdge struct {
	vertex Point
	edgeID int32
}

// incidentEdgeTracker is a used for detecting and tracking shape edges that
// are incident on the same vertex. Edges of multiple shapes may be tracked,
// but lookup is by shape id and vertex: there is no facility to get all
// edges of all shapes at a vertex. Edge vertices must compare exactly equal
// to be considered the same vertex, no tolerance is applied as this isn't
// intended for e.g.: snapping shapes together, which Builder does better
// and more robustly.
//
// To use, instantiate and then add edges with one or more sequences of calls,
// where each sequence begins with startShape(), followed by addEdge() calls to
// add edges for that shape, and ends with finishShape(). Those sequences do
// not need to visit shapes or edges in order. Then, call incidentEdges() to get
// the resulting map from incidentEdgeKeys (which are shapeId, vertex pairs) to
// a set of edgeIds of the shape that are incident to that vertex..
//
// This works on a block of edges at a time, meaning that to detect incident
// edges on a particular vertex, we must have at least three edges incident
// at that vertex when finishShape() is called. We don't maintain partial
// information between calls. However, subject to this constraint, a single
// shape's edges may be defined with multiple sequences of startShape(),
// addEdge()... , finishShape() calls.
//
// The reason for this is simple: most edges don't have more than two incident
// edges (the incoming and outgoing edge). If we had to maintain information
// between calls, we'd end up with a map that contains every vertex, to no
// benefit. Instead, when finishShape() is called, we discard vertices that
// contain two or fewer incident edges.
//
// In principle this isn't a real limitation because generally we process a
// ShapeIndex cell at a time, and, if a vertex has multiple edges, we'll see
// all the edges in the same cell as the vertex, and, in general, it's possible
// to aggregate edges before calling.
//
// The tracker maintains incident edges until it's cleared. If you call it with
// each cell from an ShapeIndex, then at the end you will have all the
// incident edge information for the whole index. If only a subset is needed,
// call reset() when you're done.
type incidentEdgeTracker struct {
	currentShapeID int32

	nursery []vertexEdge

	// We can and do encounter the same edges multiple times, so we need to
	// deduplicate edges as they're inserted.
	edgeMap map[incidentEdgeKey]map[int32]bool
}

// newIncidentEdgeTracker returns a new tracker.
func newIncidentEdgeTracker() *incidentEdgeTracker {
	return &incidentEdgeTracker{
		currentShapeID: -1,
		nursery:        []vertexEdge{},
		edgeMap:        make(map[incidentEdgeKey]map[int32]bool),
	}
}

// startShape is used to start adding edges to the edge tracker. After calling,
// any vertices with multiple (> 2) incident edges will appear in the
// incident edge map.
func (t *incidentEdgeTracker) startShape(id int32) {
	t.currentShapeID = id
	t.nursery = t.nursery[:0]
}

// addEdge adds the given edges start to the nursery, and if not degenerate,
// adds it second endpoint as well.
func (t *incidentEdgeTracker) addEdge(edgeID int32, e Edge) {
	if t.currentShapeID < 0 {
		return
	}

	// Add non-degenerate edges to the nursery.
	t.nursery = append(t.nursery, vertexEdge{vertex: e.V0, edgeID: edgeID})
	if !e.IsDegenerate() {
		t.nursery = append(t.nursery, vertexEdge{vertex: e.V1, edgeID: edgeID})
	}
}

func (t *incidentEdgeTracker) finishShape() {
	// We want to keep any vertices with more than two incident edges. We could
	// sort the array by vertex and remove any with fewer, but that would require
	// shifting the array and could turn quadratic quickly.
	//
	// Instead we'll scan forward from each vertex, swapping entries with the same
	// vertex into a contiguous range. Once we've done all the swapping we can
	// just make sure that we have at least three edges in the range.
	nurserySize := len(t.nursery)
	for start := 0; start < nurserySize; {
		end := start + 1

		// Scan to the end of the array, swap entries so that entries with
		// the same vertex as the start are adjacent.
		next := start
		currVertex := t.nursery[start].vertex
		for next+1 < nurserySize {
			next++
			if t.nursery[next].vertex == currVertex {
				t.nursery[next], t.nursery[end] = t.nursery[end], t.nursery[next]
				end++
			}
		}

		// Most vertices will have two incident edges (the incoming edge and the
		// outgoing edge), which aren't interesting, skip them.
		numEdges := end - start
		if numEdges <= 2 {
			start = end
			continue
		}

		key := incidentEdgeKey{
			shapeID: t.currentShapeID,
			vertex:  t.nursery[start].vertex,
		}

		// If we don't have this key yet, create it manually.
		if _, ok := t.edgeMap[key]; !ok {
			t.edgeMap[key] = map[int32]bool{}
		}

		for ; start != end; start++ {
			t.edgeMap[key][t.nursery[start].edgeID] = true
		}
	}
}

// reset removes all incident edges from the tracker.
func (t *incidentEdgeTracker) reset() {
	t.edgeMap = make(map[incidentEdgeKey]map[int32]bool)
}
