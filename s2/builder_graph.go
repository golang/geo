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

// A Graph represents a collection of snapped edges that is passed
// to a Layer for assembly. (Example layers include polygons, polylines, and
// polygon meshes.)
//
// The graph consists of vertices and directed edges. Vertices are numbered
// sequentially starting from zero. An edge is represented as a pair of
// vertex IDs. The edges are sorted in lexicographic order, therefore all of
// the outgoing edges from a particular vertex form a contiguous range.
//
// TODO(rsned): Consider pulling out the methods that are helper functions for
// Layer implementations (such as getDirectedLoops) into a builder_graph_util.go.
type graph struct {
	opts                   *graphOptions
	numVertices            int32
	vertices               []Point
	edges                  []graphEdge
	inputEdgeIDSetIDs      []int32
	inputEdgeIDSetLexicon  *idSetLexicon
	labelSetIDs            []int32
	labelSetLexicon        *idSetLexicon
	isFullPolygonPredicate isFullPolygonPredicate
}

// newGraph returns a new graph instance initialized with the given data.
func newGraph(opts *graphOptions,
	vertices []Point,
	edges []graphEdge,
	inputEdgeIDSetIDs []int32,
	inputEdgeIDSetLexicon *idSetLexicon,
	labelSetIDs []int32,
	labelSetLexicon *idSetLexicon,
	isFullPolygonPredicate isFullPolygonPredicate) *graph {
	g := &graph{
		opts:                   opts,
		vertices:               vertices,
		edges:                  edges,
		inputEdgeIDSetIDs:      inputEdgeIDSetIDs,
		inputEdgeIDSetLexicon:  inputEdgeIDSetLexicon,
		labelSetIDs:            labelSetIDs,
		labelSetLexicon:        labelSetLexicon,
		isFullPolygonPredicate: isFullPolygonPredicate,
	}

	return g
}

// processGraphEdges transform an unsorted collection of graphEdges according
// to the given set of GraphOptions. This includes actions such as discarding
// degenerate edges; merging duplicate edges; and canonicalizing sibling
// edge pairs in several possible ways (e.g. discarding or creating them).
// The output is suitable for passing to the newGraph method.
//
// If options.edgeType == EdgeTypeUndirected, then all input edges
// should already have been transformed into a pair of directed edges.
//
// "inputIDs" is a slice of the same length as "edges" that indicates
// which input edges were snapped to each edge, by mapping each edge ID to a
// set of input edge IDs in idSetLexicon. This slice and the lexicon are
// also updated appropriately as edges are discarded, merged, etc.
//
// Note that the options may be modified by this method: in particular, if
// edgeType is edgeTypeUndirected and siblingPairs is siblingPairsCreate or
// siblingPairsRequire, then half of the edges in each direction will be
// discarded and edgeType will be changed to edgeTypeDirected the comments
// on siblingPairs for more details).
func processGraphEdges(opts *graphOptions, edges []graphEdge, inputIds []int32,
	idSetLexicon *idSetLexicon) (newEdges []graphEdge, newInputIDs []int32, err error) {
	// graphEdgeProcessor discards the edges and inputIDs slices passed in and
	// replaces them with new slices, so we need to return whatever it ends
	// up with.
	ep := newGraphEdgeProcessor(opts, edges, inputIds, idSetLexicon)
	err = ep.Run()

	// Certain values of siblingPairs discard half of the edges and change
	// the edgeType to edgeTypeDirected (see the description of GraphOptions).
	if opts.siblingPairs == siblingPairsRequire ||
		opts.siblingPairs == siblingPairsCreate {
		opts.edgeType = edgeTypeDirected
	}
	return ep.edges, ep.inputIDs, err
}

// degenerateEdges controls how degenerate edges (i.e., an edge from a vertex to
// itself) are handled. Such edges may be present in the input, or they may be
// created when both endpoints of an edge are snapped to the same output vertex.
// The options available are:
type degenerateEdges uint8

const (
	// degenerateEdgesKeep: Keeps all degenerate edges. Be aware that this
	// may create many redundant edges when simplifying geometry (e.g., a
	// polyline of the form AABBBBBCCCCCCDDDD). degenerateEdgesKeep is mainly
	// useful for algorithms that require an output edge for every input edge.
	degenerateEdgesKeep degenerateEdges = iota
	// degenerateEdgesDiscard discards all degenerate edges. This is useful for
	// layers that/do not support degeneracies, such as PolygonLayer.
	degenerateEdgesDiscard
	// degenerateEdgesDiscardExcess discards all degenerate edges that are
	// connected to/non-degenerate edges and merges any remaining
	// duplicate/degenerate edges. This is useful for simplifying/polygons
	// while ensuring that loops that collapse to a/single point do not disappear.
	degenerateEdgesDiscardExcess
)

// duplicateEdges controls how duplicate edges (i.e., edges that are present
// multiple times) are handled. Such edges may be present in the input, or they
// can be created when vertices are snapped together. When several edges are
// merged, the result is a single edge labelled with all of the original input
// edge IDs.
type duplicateEdges uint8

const (
	duplicateEdgesKeep duplicateEdges = iota
	duplicateEdgesMerge
)

// siblingPairs controls how sibling edge pairs (i.e., pairs consisting
// of an edge and its reverse edge) are handled.  Layer types that
// define an interior (e.g., polygons) normally discard such edge pairs
// since they do not affect the result (i.e., they define a "loop" with
// no interior).
//
// If edgeType is edgeTypeUndirected, a sibling edge pair is considered
// to consist of four edges (two duplicate edges and their siblings), since
// only two of these four edges will be used in the final output.
//
// Furthermore, since the options REQUIRE and CREATE guarantee that all
// edges will have siblings, Builder implements these options for
// undirected edges by discarding half of the edges in each direction and
// changing the edgeType to edgeTypeDirected.  For example, two
// undirected input edges between vertices A and B would first be converted
// into two directed edges in each direction, and then one edge of each pair
// would be discarded leaving only one edge in each direction.
//
// Degenerate edges are considered not to have siblings. If such edges are
// present, they are passed through unchanged by siblingPairsDiscard. For
// siblingPairsRequire or siblingPairsCreate with undirected edges, the
// number of copies of each degenerate edge is reduced by a factor of two.
// Any of the options that discard edges (DISCARD, DISCARDEXCESS, and
// REQUIRE/CREATE in the case of undirected edges) have the side effect that
// when duplicate edges are present, all of the corresponding edge labels
// are merged together and assigned to the remaining edges. (This avoids
// the problem of having to decide which edges are discarded.) Note that
// this merging takes place even when all copies of an edge are kept. For
// example, consider the graph {AB1, AB2, AB3, BA4, CD5, CD6} (where XYn
// denotes an edge from X to Y with label "n"). With siblingPairsDiscard,
// we need to discard one of the copies of AB. But which one? Rather than
// choosing arbitrarily, instead we merge the labels of all duplicate edges
// (even ones where no sibling pairs were discarded), yielding {AB123,
// AB123, CD45, CD45} (assuming that duplicate edges are being kept).
// Notice that the labels of duplicate edges are merged even if no siblings
// were discarded (such as CD5, CD6 in this example), and that this would
// happen even with duplicate degenerate edges (e.g. the edges EE7, EE8).
type siblingPairs uint8

const (
	// siblingPairsKeep keeps sibling pairs. This can be used to create
	// polylines that double back on themselves, or degenerate loops (with
	// a layer type such as LaxPolygon).
	siblingPairsKeep siblingPairs = iota
	// siblingPairsDiscard discards all sibling edge pairs.
	siblingPairsDiscard
	// siblingPairsDiscardExcess is like siblingPairsDiscard, except that a
	// single sibling pair is kept if the result would otherwise be empty.
	// This is useful for polygons with degeneracies (LaxPolygon), and for
	// simplifying polylines while ensuring that they are not split into
	// multiple disconnected pieces.
	siblingPairsDiscardExcess
	// siblingPairsRequire requires that all edges have a sibling (and returns
	// an error otherwise). This is useful with layer types that create a
	// collection of adjacent polygons (a polygon mesh).
	siblingPairsRequire
	// siblingPairsCreate ensures that all edges have a sibling edge by
	// creating them if necessary. This is useful with polygon meshes where
	// the input polygons do not cover the entire sphere. Such edges always
	// have an empty set of labels and do not have an associated InputEdgeID.
	siblingPairsCreate
)

// graphOptions is only needed by Layer implementations. A layer is
// responsible for assembling a Graph of snapped edges into the
// desired output format (e.g., an Polygon). The graphOptions allows
// each Layer type to specify requirements on its input graph: for example, if
// degenerateEdgesDiscard is specified, then Builder will ensure that all
// degenerate edges are removed before passing the graph to Layer's Build
// method.
//
// A default graphOptions value specifies that all edges should be kept,
// since this produces the least surprising output and makes it easier
// to diagnose the problem when an option is left unspecified.
type graphOptions struct {
	// edgeType specifies whether the Builder input edges should be treated as
	// undirected. If true, then all input edges are duplicated into pairs
	// consisting of an edge and a sibling (reverse) edge. Note that the
	// automatically created sibling edge has an empty set of labels and does
	// not have an associated inputEdgeID.
	//
	// The layer implementation is responsible for ensuring that exactly one
	// edge from each pair is used in the output, i.e. *only half* of the graph
	// edges will be used. (Note that some values of the siblingPairs option
	// automatically take care of this issue by removing half of the edges and
	// changing edgeType to Directed.)
	edgeType        edgeType
	degenerateEdges degenerateEdges
	duplicateEdges  duplicateEdges
	siblingPairs    siblingPairs

	// This is a specialized option that is only needed by clients that want to
	// work with the graphs for multiple layers at the same time (e.g., in order
	// to check whether the same edge is present in two different graphs). [Note
	// that if you need to do this, usually it is easier just to build a single
	// graph with suitable edge labels.]
	//
	// When there are a large number of layers, then by default Builder builds
	// a minimal subgraph for each layer containing only the vertices needed by
	// the edges in that layer. This ensures that layer types that iterate over
	// the vertices run in time proportional to the size of that layer rather
	// than the size of all layers combined. (For example, if there are a
	// million layers with one edge each, then each layer would be passed a
	// graph with 2 vertices rather than 2 million vertices.)
	//
	// If this option is set to false, this optimization is disabled. Instead
	// the graph passed to this layer will contain the full set of vertices.
	// (This is not recommended when the number of layers could be large.)
	//
	// Default is false.
	disableVertexFiltering bool
}
