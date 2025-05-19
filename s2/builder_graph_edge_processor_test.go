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

import (
	"slices"
	"testing"
)

func TestGraphEdgeProcessorStableLessThan(t *testing.T) {
	tests := []struct {
		name     string
		a        graphEdge
		b        graphEdge
		aInputID int32
		bInputID int32
		want     bool
	}{
		{
			name:     "a < b lexicographically",
			a:        graphEdge{first: 1, second: 2},
			b:        graphEdge{first: 2, second: 3},
			aInputID: 1,
			bInputID: 2,
			want:     true,
		},
		{
			name:     "a > b lexicographically",
			a:        graphEdge{first: 3, second: 4},
			b:        graphEdge{first: 1, second: 2},
			aInputID: 1,
			bInputID: 2,
			want:     false,
		},
		{
			name:     "a == b lexicographically, a.inputID < b.inputID",
			a:        graphEdge{first: 1, second: 2},
			b:        graphEdge{first: 1, second: 2},
			aInputID: 1,
			bInputID: 2,
			want:     true,
		},
		{
			name:     "a == b lexicographically, a.inputID > b.inputID",
			a:        graphEdge{first: 1, second: 2},
			b:        graphEdge{first: 1, second: 2},
			aInputID: 3,
			bInputID: 2,
			want:     false,
		},
		{
			name:     "a == b lexicographically, a.inputID == b.inputID",
			a:        graphEdge{first: 1, second: 2},
			b:        graphEdge{first: 1, second: 2},
			aInputID: 5,
			bInputID: 5,
			want:     false,
		},
		{
			name:     "first vertices equal, second vertices different",
			a:        graphEdge{first: 1, second: 2},
			b:        graphEdge{first: 1, second: 3},
			aInputID: 1,
			bInputID: 2,
			want:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := stableLessThan(test.a, test.b, test.aInputID, test.bInputID)
			if got != test.want {
				t.Errorf("stableLessThan() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGraphEdgeProcessorAddEdge(t *testing.T) {
	tests := []struct {
		name             string
		edge             graphEdge
		inputEdgeIDSetID int32
		wantEdges        int
		wantInputIDs     int
	}{
		{
			name:             "add single edge",
			edge:             graphEdge{first: 1, second: 2},
			inputEdgeIDSetID: 1,
			wantEdges:        1,
			wantInputIDs:     1,
		},
		{
			name:             "add second edge",
			edge:             graphEdge{first: 2, second: 3},
			inputEdgeIDSetID: 2,
			wantEdges:        2,
			wantInputIDs:     2,
		},
	}

	ep := &edgeProcessor{
		newEdges:    make([]graphEdge, 0),
		newInputIDs: make([]int32, 0),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ep.addEdge(test.edge, test.inputEdgeIDSetID)
			if len(ep.newEdges) != test.wantEdges {
				t.Errorf("addEdge() edges = %v, want %v", len(ep.newEdges), test.wantEdges)
			}
			if len(ep.newInputIDs) != test.wantInputIDs {
				t.Errorf("addEdge() inputIDs = %v, want %v", len(ep.newInputIDs), test.wantInputIDs)
			}
		})
	}
}

func TestGraphEdgeProcessorAddEdges(t *testing.T) {
	tests := []struct {
		name             string
		numEdges         int
		edge             graphEdge
		inputEdgeIDSetID int32
		wantEdges        int
		wantInputIDs     int
	}{
		{
			name:             "add single edge",
			numEdges:         1,
			edge:             graphEdge{first: 1, second: 2},
			inputEdgeIDSetID: 1,
			wantEdges:        1,
			wantInputIDs:     1,
		},
		{
			name:             "add multiple edges",
			numEdges:         3,
			edge:             graphEdge{first: 1, second: 2},
			inputEdgeIDSetID: 7,
			wantEdges:        3,
			wantInputIDs:     3,
		},
		{
			name:             "add zero edges",
			numEdges:         0,
			edge:             graphEdge{first: 1, second: 2},
			inputEdgeIDSetID: 8,
			wantEdges:        0, // Should remain unchanged
			wantInputIDs:     0, // Should remain unchanged
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ep := &edgeProcessor{
				newEdges:    make([]graphEdge, 0),
				newInputIDs: make([]int32, 0),
			}

			ep.addEdges(test.numEdges, test.edge, test.inputEdgeIDSetID)
			if len(ep.newEdges) != test.wantEdges {
				t.Errorf("addEdges() edges = %v, want %v", len(ep.newEdges), test.wantEdges)
			}

			if len(ep.newInputIDs) != test.wantInputIDs {
				t.Errorf("addEdges() inputIDs = %v, want %v", len(ep.newInputIDs), test.wantInputIDs)
			}

			// addEdges uses the same inputEdgeIDSetID for each repeated edge. Ensure
			// all the added ids match.
			for k, v := range ep.newInputIDs {
				if v != test.inputEdgeIDSetID {
					t.Errorf("in addEdges, newInputIDs[%d] = %d, want %d", k, v, test.inputEdgeIDSetID)
				}
			}
		})
	}
}

func TestGraphEdgeProcessorHandleDegenerateEdge(t *testing.T) {
	tests := []struct {
		name      string
		edge      graphEdge
		options   *graphOptions
		outBegin  int
		outEnd    int
		nOut      int
		nIn       int
		inBegin   int
		in        int
		wantErr   bool
		wantEdges int
	}{
		{
			name: "discard degenerate edges",
			edge: graphEdge{first: 1, second: 1},
			options: &graphOptions{
				degenerateEdges: degenerateEdgesDiscard,
			},
			outBegin:  0,
			outEnd:    1,
			nOut:      1,
			nIn:       1,
			inBegin:   0,
			in:        1,
			wantErr:   false,
			wantEdges: 0,
		},
		{
			name: "keep degenerate edges with merge",
			edge: graphEdge{first: 1, second: 1},
			options: &graphOptions{
				degenerateEdges: degenerateEdgesKeep,
				duplicateEdges:  duplicateEdgesMerge,
				edgeType:        edgeTypeDirected,
			},
			outBegin:  0,
			outEnd:    2,
			nOut:      2,
			nIn:       2,
			inBegin:   0,
			in:        2,
			wantErr:   false,
			wantEdges: 1,
		},
		{
			name: "keep degenerate edges without merge",
			edge: graphEdge{first: 1, second: 1},
			options: &graphOptions{
				degenerateEdges: degenerateEdgesKeep,
				duplicateEdges:  duplicateEdgesKeep,
				edgeType:        edgeTypeDirected,
			},
			outBegin:  0,
			outEnd:    2,
			nOut:      2,
			nIn:       2,
			inBegin:   0,
			in:        2,
			wantErr:   false,
			wantEdges: 2,
		},
		{
			name: "discard excess degenerate edges with incident edges",
			edge: graphEdge{first: 1, second: 1},
			options: &graphOptions{
				degenerateEdges: degenerateEdgesDiscardExcess,
				edgeType:        edgeTypeDirected,
			},
			outBegin:  1,
			outEnd:    2,
			nOut:      1,
			nIn:       1,
			inBegin:   1,
			in:        2,
			wantErr:   false,
			wantEdges: 0,
		},
		{
			name: "discard excess degenerate edges without incident edges",
			edge: graphEdge{first: 1, second: 1},
			options: &graphOptions{
				degenerateEdges: degenerateEdgesDiscardExcess,
				edgeType:        edgeTypeDirected,
				duplicateEdges:  duplicateEdgesMerge,
			},
			outBegin:  0,
			outEnd:    2,
			nOut:      2,
			nIn:       2,
			inBegin:   0,
			in:        2,
			wantErr:   false,
			wantEdges: 1,
		},
		{
			name: "undirected degenerate edges with require siblings",
			edge: graphEdge{first: 1, second: 1},
			options: &graphOptions{
				degenerateEdges: degenerateEdgesKeep,
				edgeType:        edgeTypeUndirected,
				siblingPairs:    siblingPairsRequire,
				duplicateEdges:  duplicateEdgesMerge,
			},
			outBegin:  0,
			outEnd:    2,
			nOut:      2,
			nIn:       2,
			inBegin:   0,
			in:        2,
			wantErr:   false,
			wantEdges: 1,
		},
		{
			name: "undirected degenerate edges with create siblings",
			edge: graphEdge{first: 1, second: 1},
			options: &graphOptions{
				degenerateEdges: degenerateEdgesKeep,
				edgeType:        edgeTypeUndirected,
				siblingPairs:    siblingPairsCreate,
				duplicateEdges:  duplicateEdgesKeep,
			},
			outBegin:  0,
			outEnd:    4,
			nOut:      4,
			nIn:       4,
			inBegin:   0,
			in:        4,
			wantErr:   false,
			wantEdges: 2,
		},
		{
			name: "inconsistent degenerate edges",
			edge: graphEdge{first: 1, second: 1},
			options: &graphOptions{
				degenerateEdges: degenerateEdgesKeep,
			},
			outBegin: 0,
			outEnd:   1,
			nOut:     1,
			nIn:      2, // Mismatched counts
			inBegin:  0,
			in:       2,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create edges and inputIDs arrays with the correct size
			edges := make([]graphEdge, test.outEnd)
			inputIDs := make([]int32, test.outEnd)
			for i := range edges {
				edges[i] = test.edge
				inputIDs[i] = int32(i + 1)
			}

			ep := newEdgeProcessor(test.options, edges, inputIDs, newIDSetLexicon())
			// Add the input IDs to the lexicon.
			for _, id := range inputIDs {
				ep.idSetLexicon.add(id)
			}

			gotErr := ep.handleDegenerateEdge(test.edge, test.outBegin, test.outEnd,
				test.nOut, test.nIn, test.inBegin, test.in)
			if (gotErr != nil) != test.wantErr {
				t.Errorf("handleDegenerateEdge() error = %v, wantErr %v",
					gotErr, test.wantErr)
			}
			if !test.wantErr && len(ep.newEdges) != test.wantEdges {
				t.Errorf("handleDegenerateEdge() added %d edges, want %d", len(ep.newEdges), test.wantEdges)
			}
		})
	}
}

func TestGraphEdgeProcessorHandleNormalEdge(t *testing.T) {
	tests := []struct {
		name      string
		edge      graphEdge
		options   *graphOptions
		outBegin  int
		outEnd    int
		nOut      int
		nIn       int
		wantErr   bool
		wantEdges int
	}{
		{
			name: "keep sibling pairs with merge",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs:   siblingPairsKeep,
				edgeType:       edgeTypeDirected,
				duplicateEdges: duplicateEdgesMerge,
			},
			outBegin:  0,
			outEnd:    2,
			nOut:      2,
			nIn:       1,
			wantErr:   false,
			wantEdges: 1,
		},
		{
			name: "keep sibling pairs without merge",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs:   siblingPairsKeep,
				edgeType:       edgeTypeDirected,
				duplicateEdges: duplicateEdgesKeep,
			},
			outBegin:  0,
			outEnd:    2,
			nOut:      2,
			nIn:       1,
			wantErr:   false,
			wantEdges: 2,
		},
		{
			name: "discard sibling pairs directed balanced",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs:   siblingPairsDiscard,
				edgeType:       edgeTypeDirected,
				duplicateEdges: duplicateEdgesMerge,
			},
			outBegin:  0,
			outEnd:    2,
			nOut:      2,
			nIn:       2,
			wantErr:   false,
			wantEdges: 0,
		},
		{
			name: "discard sibling pairs directed unbalanced",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs:   siblingPairsDiscard,
				edgeType:       edgeTypeDirected,
				duplicateEdges: duplicateEdgesMerge,
			},
			outBegin:  0,
			outEnd:    3,
			nOut:      3,
			nIn:       1,
			wantErr:   false,
			wantEdges: 1,
		},
		{
			name: "discard sibling pairs undirected even",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs:   siblingPairsDiscard,
				edgeType:       edgeTypeUndirected,
				duplicateEdges: duplicateEdgesMerge,
			},
			outBegin:  0,
			outEnd:    2,
			nOut:      2,
			nIn:       2,
			wantErr:   false,
			wantEdges: 0,
		},
		{
			name: "discard sibling pairs undirected odd",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs:   siblingPairsDiscard,
				edgeType:       edgeTypeUndirected,
				duplicateEdges: duplicateEdgesMerge,
			},
			outBegin:  0,
			outEnd:    3,
			nOut:      3,
			nIn:       3,
			wantErr:   false,
			wantEdges: 1,
		},
		{
			name: "discard excess sibling pairs directed balanced",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs:   siblingPairsDiscardExcess,
				edgeType:       edgeTypeDirected,
				duplicateEdges: duplicateEdgesMerge,
			},
			outBegin:  0,
			outEnd:    2,
			nOut:      2,
			nIn:       2,
			wantErr:   false,
			wantEdges: 1,
		},
		{
			name: "discard excess sibling pairs directed unbalanced",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs:   siblingPairsDiscardExcess,
				edgeType:       edgeTypeDirected,
				duplicateEdges: duplicateEdgesMerge,
			},
			outBegin:  0,
			outEnd:    3,
			nOut:      3,
			nIn:       1,
			wantErr:   false,
			wantEdges: 1,
		},
		{
			name: "require sibling pairs directed balanced",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs:   siblingPairsRequire,
				edgeType:       edgeTypeDirected,
				duplicateEdges: duplicateEdgesMerge,
			},
			outBegin:  0,
			outEnd:    2,
			nOut:      2,
			nIn:       2,
			wantErr:   false,
			wantEdges: 1,
		},
		{
			name: "require sibling pairs directed unbalanced",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs:   siblingPairsRequire,
				edgeType:       edgeTypeDirected,
				duplicateEdges: duplicateEdgesMerge,
			},
			outBegin: 0,
			outEnd:   2,
			nOut:     2,
			nIn:      1,
			wantErr:  true,
		},
		{
			name: "create sibling pairs undirected",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs:   siblingPairsCreate,
				edgeType:       edgeTypeUndirected,
				duplicateEdges: duplicateEdgesKeep,
			},
			outBegin:  0,
			outEnd:    4,
			nOut:      4,
			nIn:       2,
			wantErr:   false,
			wantEdges: 2,
		},
		{
			name: "invalid sibling pairs option",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs: siblingPairs(255), // Use max uint8 value as invalid
				edgeType:     edgeTypeDirected,
			},
			outBegin: 0,
			outEnd:   1,
			nOut:     1,
			nIn:      1,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create edges and inputIDs arrays with the correct size
			edges := make([]graphEdge, test.outEnd)
			inputIDs := make([]int32, test.outEnd)
			for i := range edges {
				edges[i] = test.edge
				inputIDs[i] = int32(i + 1)
			}

			ep := newEdgeProcessor(test.options, edges, inputIDs, newIDSetLexicon())
			// Add the input IDs to the lexicon.
			for _, id := range inputIDs {
				ep.idSetLexicon.add(id)
			}

			gotErr := ep.handleNormalEdge(test.edge, test.outBegin, test.outEnd, test.nOut, test.nIn)
			if (gotErr != nil) != test.wantErr {
				t.Errorf("handleNormalEdge() error = %v, wantErr %v", gotErr, test.wantErr)
			}
			if !test.wantErr && len(ep.newEdges) != test.wantEdges {
				t.Errorf("handleNormalEdge() added %d edges, want %d", len(ep.newEdges), test.wantEdges)
			}
		})
	}
}

func TestGraphEdgeProcessorMergeInputIDs(t *testing.T) {
	tests := []struct {
		name           string
		edges          []graphEdge
		inputIDs       []int32
		outBegin       int
		outEnd         int
		wantInputIDSet []int32
	}{
		{
			name: "single edge",
			edges: []graphEdge{
				{first: 1, second: 2},
			},
			inputIDs:       []int32{1},
			outBegin:       0,
			outEnd:         1,
			wantInputIDSet: []int32{1},
		},
		{
			name: "multiple edges with same input ID should reduce to 1 output",
			edges: []graphEdge{
				{first: 1, second: 2},
				{first: 1, second: 2},
			},
			inputIDs:       []int32{1, 1},
			outBegin:       0,
			outEnd:         2,
			wantInputIDSet: []int32{1},
		},
		{
			name: "multiple edges with different input IDs should keep distinct ids",
			edges: []graphEdge{
				{first: 1, second: 2},
				{first: 1, second: 2},
				{first: 1, second: 2},
			},
			inputIDs:       []int32{1, 2, 3},
			outBegin:       0,
			outEnd:         3,
			wantInputIDSet: []int32{1, 2, 3},
		},
		{
			name: "subset of edges should return the smaller portion",
			edges: []graphEdge{
				{first: 1, second: 2},
				{first: 1, second: 2},
				{first: 1, second: 2},
			},
			inputIDs:       []int32{1, 2, 3},
			outBegin:       1,
			outEnd:         3,
			wantInputIDSet: []int32{2, 3},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ep := &edgeProcessor{
				edges:        test.edges,
				inputIDs:     test.inputIDs,
				idSetLexicon: newIDSetLexicon(),
				outEdges:     make([]int32, len(test.edges)),
			}

			// Initialize outEdges with sequential indices
			for i := range ep.outEdges {
				ep.outEdges[i] = int32(i)
			}

			// Add the input IDs to the lexicon first
			for _, id := range test.inputIDs {
				ep.idSetLexicon.add(id)
			}

			merged := ep.mergeInputIDs(test.outBegin, test.outEnd)

			// Get the actual set of IDs from the lexicon
			got := ep.idSetLexicon.idSet(merged)

			// Sort both slices for comparison
			slices.Sort(got)
			slices.Sort(test.wantInputIDSet)

			if !slices.Equal(got, test.wantInputIDSet) {
				t.Errorf("mergeInputIDs() = %v, want %v", got, test.wantInputIDSet)
			}
		})
	}
}

func TestGraphEdgeProcessorRun(t *testing.T) {
	tests := []struct {
		name      string
		edges     []graphEdge
		inputIDs  []int32
		options   *graphOptions
		wantErr   bool
		wantEdges int
	}{
		{
			name:      "empty graph",
			edges:     []graphEdge{},
			inputIDs:  []int32{},
			options:   defaultGraphOptions(),
			wantErr:   false,
			wantEdges: 0,
		},
		{
			name: "single edge",
			edges: []graphEdge{
				{first: 1, second: 2},
			},
			inputIDs:  []int32{1},
			options:   defaultGraphOptions(),
			wantErr:   false,
			wantEdges: 1,
		},
		{
			name: "degenerate edge with discard",
			edges: []graphEdge{
				{first: 1, second: 1},
			},
			inputIDs: []int32{1},
			options: &graphOptions{
				edgeType:        edgeTypeDirected,
				duplicateEdges:  duplicateEdgesKeep,
				degenerateEdges: degenerateEdgesDiscard,
				siblingPairs:    siblingPairsKeep,
			},
			wantErr:   false,
			wantEdges: 0,
		},
		{
			name: "duplicate edges with merge",
			edges: []graphEdge{
				{first: 1, second: 2},
				{first: 1, second: 2},
			},
			inputIDs: []int32{1, 2},
			options: &graphOptions{
				edgeType:        edgeTypeDirected,
				duplicateEdges:  duplicateEdgesMerge,
				degenerateEdges: degenerateEdgesKeep,
				siblingPairs:    siblingPairsKeep,
			},
			wantErr:   false,
			wantEdges: 1,
		},
		{
			name: "sibling pairs with discard",
			edges: []graphEdge{
				{first: 1, second: 2},
				{first: 2, second: 1},
			},
			inputIDs: []int32{1, 2},
			options: &graphOptions{
				edgeType:        edgeTypeDirected,
				duplicateEdges:  duplicateEdgesKeep,
				degenerateEdges: degenerateEdgesKeep,
				siblingPairs:    siblingPairsDiscard,
			},
			wantErr:   false,
			wantEdges: 0,
		},
		{
			name: "undirected edges with require siblings",
			edges: []graphEdge{
				{first: 1, second: 2},
				{first: 2, second: 1},
			},
			inputIDs: []int32{1, 2},
			options: &graphOptions{
				edgeType:        edgeTypeUndirected,
				duplicateEdges:  duplicateEdgesKeep,
				degenerateEdges: degenerateEdgesKeep,
				siblingPairs:    siblingPairsRequire,
			},
			wantErr:   true,
			wantEdges: 0,
		},
		{
			name: "undirected edges with create siblings",
			edges: []graphEdge{
				{first: 1, second: 2},
				{first: 2, second: 1},
			},
			inputIDs: []int32{1, 2},
			options: &graphOptions{
				edgeType:        edgeTypeUndirected,
				duplicateEdges:  duplicateEdgesKeep,
				degenerateEdges: degenerateEdgesKeep,
				siblingPairs:    siblingPairsCreate,
			},
			wantErr:   false,
			wantEdges: 2,
		},
		{
			name: "require siblings with missing sibling",
			edges: []graphEdge{
				{first: 1, second: 2},
			},
			inputIDs: []int32{1},
			options: &graphOptions{
				edgeType:        edgeTypeUndirected,
				duplicateEdges:  duplicateEdgesKeep,
				degenerateEdges: degenerateEdgesKeep,
				siblingPairs:    siblingPairsRequire,
			},
			wantErr:   true,
			wantEdges: 0,
		},
		{
			name: "multiple edges with various options",
			edges: []graphEdge{
				{first: 1, second: 2},
				{first: 2, second: 1},
				{first: 1, second: 2},
				{first: 3, second: 3},
			},
			inputIDs: []int32{1, 2, 3, 4},
			options: &graphOptions{
				edgeType:        edgeTypeDirected,
				duplicateEdges:  duplicateEdgesMerge,
				degenerateEdges: degenerateEdgesDiscard,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			wantErr:   false,
			wantEdges: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lexicon := newIDSetLexicon()
			ep := newEdgeProcessor(test.options, test.edges, test.inputIDs, lexicon)
			err := ep.Run()
			if (err != nil) != test.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && len(ep.edges) != test.wantEdges {
				t.Errorf("Run() produced %d edges, want %d", len(ep.edges), test.wantEdges)
			}
		})
	}
}
