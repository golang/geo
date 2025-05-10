// Copyright 2025 The S2 Geometry Project Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package s2

// PolylineType Indicates whether polylines should be "paths" (which don't
// allow duplicate vertices, except possibly the first and last vertex) or
// "walks" (which allow duplicate vertices and edges).
type PolylineType uint8

const (
	PolylineTypePath PolylineType = iota
	PolylineTypeWalk
)

// graphEdge is a tuple of edge IDs.
type graphEdge struct {
	first, second int32
}

// A Graph represents a collection of snapped edges that is passed
// to a Layer for assembly. (Example layers include polygons, polylines, and
// polygon meshes.) The Graph object does not own any of its underlying data;
// it is simply a view of data that is stored elsewhere. You will only
// need this interface if you want to implement a new Layer subtype.
//
// The graph consists of vertices and directed edges. Vertices are numbered
// sequentially starting from zero. An edge is represented as a pair of
// vertex ids. The edges are sorted in lexicographic order, therefore all of
// the outgoing edges from a particular vertex form a contiguous range.
//
// TODO(rsned): Consider pulling out the methods that are helper functions for
// Layer implementations (such as getDirectedLoops) into a builder_util_graph.go.
type graph struct {
	opts                   *graphOptions
	numVertices            int
	vertices               []Point
	edges                  []graphEdge
	inputEdgeIDSetIDs      []int32
	inputEdgeIDSetLexicon  *idSetLexicon
	labelSetIDs            []int32
	labelSetLexicon        *idSetLexicon
	isFullPolygonPredicate isFullPolygonPredicate
}
