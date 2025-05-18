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

import (
	"errors"
	"sort"
)

// maxVertexID is the maximum possible vertex ID, used as a sentinel value.
const maxVertexID = int32(^uint32(0) >> 1)

// graphEdge is a tuple of edge IDs.
type graphEdge struct {
	first, second int32
}

// reverse returns a new graphEdge with the vertices in reverse order.
func (g graphEdge) reverse() graphEdge {
	return graphEdge{first: g.second, second: g.first}
}

// minGraphEdge returns the minimum of two edges in lexicographic order.
func minGraphEdge(a, b graphEdge) graphEdge {
	if a.first < b.first || (a.first == b.first && a.second <= b.second) {
		return a
	}
	return b
}

// stableLessThan compares two graphEdges for stable sorting.
// It uses the graphEdge IDs as a tiebreaker to ensure a stable sort.
func stableLessThan(a, b graphEdge, aID, bID int32) bool {
	if a.first != b.first {
		return a.first < b.first
	}
	if a.second != b.second {
		return a.second < b.second
	}
	return aID < bID
}

// edgeProcessor processes edges in a Graph to handle duplicates, siblings,
// and degenerate edges according to the specified GraphOptions.
type edgeProcessor struct {
	options      *graphOptions
	edges        []graphEdge
	inputIDs     []int32
	idSetLexicon *idSetLexicon
	outEdges     []int32
	inEdges      []int32
	newEdges     []graphEdge
	newInputIDs  []int32
}

// newedgeProcessor creates a new edgeProcessor with the given options and data.
func newEdgeProcessor(opts *graphOptions, edges []graphEdge, inputIDs []int32,
	idSetLexicon *idSetLexicon) *edgeProcessor {
	// opts should not be nil at this point, but just in case.
	if opts == nil {
		opts = defaultGraphOptions()
	}
	ep := &edgeProcessor{
		options:      opts,
		edges:        edges,
		inputIDs:     inputIDs,
		idSetLexicon: idSetLexicon,
		inEdges:      make([]int32, len(edges)),
		outEdges:     make([]int32, len(edges)),
	}

	// Sort the outgoing and incoming edges in lexicographic order.
	// We use a stable sort to ensure that each undirected edge becomes a sibling pair,
	// even if there are multiple identical input edges.

	// Fill the slice with a number sequence.
	for i := range ep.outEdges {
		ep.outEdges[i] = int32(i)
	}
	stableSortEdges(ep.outEdges, func(a, b int32) bool {
		return stableLessThan(ep.edges[a], ep.edges[b], a, b)
	})

	// Fill the slice with a number sequence.
	for i := range ep.inEdges {
		ep.inEdges[i] = int32(i)
	}
	stableSortEdges(ep.inEdges, func(a, b int32) bool {
		return stableLessThan(ep.edges[a].reverse(), ep.edges[b].reverse(), a, b)
	})

	ep.newEdges = make([]graphEdge, 0, len(edges))
	ep.newInputIDs = make([]int32, 0, len(edges))

	return ep
}

// stableSortEdges performs a stable sort on the given slice of EdgeIDs using the provided less function.
func stableSortEdges(edges []int32, less func(a, b int32) bool) {
	sort.SliceStable(edges, func(i, j int) bool {
		return less(edges[i], edges[j])
	})
}

// addEdge adds a single edge with its input edge ID set to the new edges.
func (ep *edgeProcessor) addEdge(edge graphEdge, inputEdgeIDSetID int32) {
	ep.newEdges = append(ep.newEdges, edge)
	ep.newInputIDs = append(ep.newInputIDs, inputEdgeIDSetID)
}

// addEdges adds multiple copies of the same edge with the same input edge ID set.
func (ep *edgeProcessor) addEdges(numEdges int, edge graphEdge, inputEdgeIDSetID int32) {
	for i := 0; i < numEdges; i++ {
		ep.addEdge(edge, inputEdgeIDSetID)
	}
}

// copyEdges copies a range of edges from the input edges to the new edges.
func (ep *edgeProcessor) copyEdges(outBegin, outEnd int) {
	for i := outBegin; i < outEnd; i++ {
		ep.addEdge(ep.edges[ep.outEdges[i]], ep.inputIDs[ep.outEdges[i]])
	}
}

// mergeInputIDs merges the input edge ID sets for a range of edges.
func (ep *edgeProcessor) mergeInputIDs(outBegin, outEnd int) int32 {
	if outEnd-outBegin == 1 {
		return ep.inputIDs[ep.outEdges[outBegin]]
	}

	var tmpIDs []int32
	for i := outBegin; i < outEnd; i++ {
		tmpIDs = append(tmpIDs, ep.idSetLexicon.idSet(ep.inputIDs[ep.outEdges[i]])...)
	}
	return int32(ep.idSetLexicon.add(tmpIDs...))
}

// Run processes the edges according to the specified options.
func (ep *edgeProcessor) Run() error {
	numEdges := len(ep.edges)
	if numEdges == 0 {
		return nil
	}

	// Walk through the two sorted arrays performing a merge join. For each
	// edge, gather all the duplicate copies of the edge in both directions
	// (outgoing and incoming). Then decide what to do based on options and
	// how many copies of the edge there are in each direction.
	out, in := 0, 0
	outEdge := ep.edges[ep.outEdges[out]]
	inEdge := ep.edges[ep.inEdges[in]]
	sentinel := graphEdge{first: maxVertexID, second: maxVertexID}

	for {
		edge := minGraphEdge(outEdge, inEdge.reverse())
		if edge == sentinel {
			break
		}

		outBegin := out
		inBegin := in
		for outEdge == edge {
			out++
			if out == numEdges {
				outEdge = sentinel
			} else {
				outEdge = ep.edges[ep.outEdges[out]]
			}
		}
		for inEdge.reverse() == edge {
			in++
			if in == numEdges {
				inEdge = sentinel
			} else {
				inEdge = ep.edges[ep.inEdges[in]]
			}
		}
		nOut := out - outBegin
		nIn := in - inBegin

		if edge.first == edge.second {
			// This is a degenerate edge.
			if err := ep.handleDegenerateEdge(edge, outBegin, out, nOut, nIn, inBegin, in); err != nil {
				return err
			}
		} else if err := ep.handleNormalEdge(edge, outBegin, out, nOut, nIn); err != nil {
			return err
		}
	}

	// Replace the old edges with the new ones.
	ep.edges = ep.newEdges
	ep.inputIDs = ep.newInputIDs
	return nil
}

// handleDegenerateEdge handles a degenerate edge (an edge from a vertex to itself).
func (ep *edgeProcessor) handleDegenerateEdge(edge graphEdge, outBegin, outEnd int, nOut, nIn, inBegin, in int) error {
	// This is a degenerate edge.
	if nOut != nIn {
		return errors.New("inconsistent number of degenerate edges")
	}
	if ep.options.degenerateEdges == degenerateEdgesDiscard {
		return nil
	}
	if ep.options.degenerateEdges == degenerateEdgesDiscardExcess {
		// Check if there are any non-degenerate incident edges.
		if (outBegin > 0 && ep.edges[ep.outEdges[outBegin-1]].first == edge.first) ||
			(outEnd < len(ep.edges) && ep.edges[ep.outEdges[outEnd]].first == edge.first) ||
			(inBegin > 0 && ep.edges[ep.inEdges[inBegin-1]].second == edge.first) ||
			(in < len(ep.edges) && ep.edges[ep.inEdges[in]].second == edge.first) {
			return nil // There were non-degenerate incident edges, so discard.
		}
	}

	// degenerateEdgesDiscardExcess also merges degenerate edges.
	merge := ep.options.duplicateEdges == duplicateEdgesMerge ||
		ep.options.degenerateEdges == degenerateEdgesDiscardExcess

	if ep.options.edgeType == edgeTypeUndirected &&
		(ep.options.siblingPairs == siblingPairsRequire ||
			ep.options.siblingPairs == siblingPairsCreate) {
		// When we have undirected edges and are guaranteed to have siblings,
		// we cut the number of edges in half (see Builder).
		if nOut&1 != 0 {
			return errors.New("odd number of undirected degenerate edges")
		}
		if merge {
			ep.addEdges(1, edge, ep.mergeInputIDs(outBegin, outEnd))
		} else {
			ep.addEdges(nOut/2, edge, ep.mergeInputIDs(outBegin, outEnd))
		}
	} else if merge {
		if ep.options.edgeType == edgeTypeUndirected {
			ep.addEdges(2, edge, ep.mergeInputIDs(outBegin, outEnd))
		} else {
			ep.addEdges(1, edge, ep.mergeInputIDs(outBegin, outEnd))
		}
	} else if ep.options.siblingPairs == siblingPairsDiscard ||
		ep.options.siblingPairs == siblingPairsDiscardExcess {
		// Any SiblingPair option that discards edges causes the labels of all
		// duplicate edges to be merged together (see Builder).
		ep.addEdges(nOut, edge, ep.mergeInputIDs(outBegin, outEnd))
	} else {
		ep.copyEdges(outBegin, outEnd)
	}
	return nil
}

// handleNormalEdge handles a non-degenerate edge.
func (ep *edgeProcessor) handleNormalEdge(edge graphEdge, outBegin, outEnd int, nOut, nIn int) error {
	if ep.options.siblingPairs == siblingPairsKeep {
		if nOut > 1 && ep.options.duplicateEdges == duplicateEdgesMerge {
			ep.addEdge(edge, ep.mergeInputIDs(outBegin, outEnd))
		} else {
			ep.copyEdges(outBegin, outEnd)
		}
	} else if ep.options.siblingPairs == siblingPairsDiscard {
		if ep.options.edgeType == edgeTypeDirected {
			// If nOut == nIn: balanced sibling pairs
			// If nOut < nIn:  unbalanced siblings, in the form AB, BA, BA
			// If nOut > nIn:  unbalanced siblings, in the form AB, AB, BA
			if nOut <= nIn {
				return nil
			}
			// Any option that discards edges causes the labels of all duplicate
			// edges to be merged together (see Builder).
			if ep.options.duplicateEdges == duplicateEdgesMerge {
				ep.addEdges(1, edge, ep.mergeInputIDs(outBegin, outEnd))
			} else {
				ep.addEdges(nOut-nIn, edge, ep.mergeInputIDs(outBegin, outEnd))
			}
		} else {
			if nOut&1 == 0 {
				return nil
			}
			ep.addEdge(edge, ep.mergeInputIDs(outBegin, outEnd))
		}
	} else if ep.options.siblingPairs == siblingPairsDiscardExcess {
		if ep.options.edgeType == edgeTypeDirected {
			// See comments above. The only difference is that if there are
			// balanced sibling pairs, we want to keep one such pair.
			if nOut < nIn {
				return nil
			}
			if ep.options.duplicateEdges == duplicateEdgesMerge {
				ep.addEdges(1, edge, ep.mergeInputIDs(outBegin, outEnd))
			} else {
				ep.addEdges(maxInt(1, nOut-nIn), edge, ep.mergeInputIDs(outBegin, outEnd))
			}
		} else {
			ep.addEdges((nOut&1)+1, edge, ep.mergeInputIDs(outBegin, outEnd))
		}
	} else {
		if ep.options.siblingPairs != siblingPairsRequire &&
			ep.options.siblingPairs != siblingPairsCreate {
			return errors.New("invalid sibling pairs option")
		}
		// In C++, this check also checked the state of the S2Error passed in
		// to make sure no previous errors had occured before now.
		if ep.options.siblingPairs == siblingPairsRequire &&
			(ep.options.edgeType == edgeTypeDirected && nOut != nIn ||
				ep.options.edgeType == edgeTypeUndirected && nOut&1 != 0) {
			return errors.New("expected all input edges to have siblingsa but some were missing")
		}

		if ep.options.duplicateEdges == duplicateEdgesMerge {
			ep.addEdge(edge, ep.mergeInputIDs(outBegin, outEnd))
		} else if ep.options.edgeType == edgeTypeUndirected {
			// Convert graph to use directed edges instead (see documentation of
			// REQUIRE/CREATE for undirected edges).
			ep.addEdges((nOut+1)/2, edge, ep.mergeInputIDs(outBegin, outEnd))
		} else {
			ep.copyEdges(outBegin, outEnd)
			if nIn > nOut {
				// Automatically created edges have no input edge ids or labels.
				ep.addEdges(nIn-nOut, edge, emptySetID)
			}
		}
	}
	return nil
}
