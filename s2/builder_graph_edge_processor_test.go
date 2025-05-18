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
		name           string
		edge           graphEdge
		options        *graphOptions
		outBegin       int
		outEnd         int
		nOut           int
		nIn            int
		inBegin        int
		in             int
		wantErr        bool
		wantErrMessage string
	}{
		{
			name: "discard degenerate edges",
			edge: graphEdge{first: 1, second: 1},
			options: &graphOptions{
				degenerateEdges: degenerateEdgesDiscard,
			},
			outBegin: 0,
			outEnd:   1,
			nOut:     1,
			nIn:      1,
			inBegin:  0,
			in:       1,
			wantErr:  false,
		},
		{
			name: "inconsistent degenerate edges",
			edge: graphEdge{first: 1, second: 1},
			options: &graphOptions{
				degenerateEdges: degenerateEdgesKeep,
			},
			outBegin:       0,
			outEnd:         1,
			nOut:           1,
			nIn:            2, // Mismatched counts
			inBegin:        0,
			in:             2,
			wantErr:        true,
			wantErrMessage: "Inconsistent number of degenerate edges",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ep := &edgeProcessor{
				options:      test.options,
				edges:        []graphEdge{test.edge},
				inputIDs:     []int32{1},
				idSetLexicon: newIDSetLexicon(),
				newEdges:     make([]graphEdge, 0),
				newInputIDs:  make([]int32, 0),
				outEdges:     []int32{0}, // Initialize with index 0
				inEdges:      []int32{0}, // Initialize with index 0
			}
			gotErr := ep.handleDegenerateEdge(test.edge, test.outBegin, test.outEnd,
				test.nOut, test.nIn, test.inBegin, test.in)
			if (gotErr != nil) != test.wantErr {
				t.Errorf("handleDegenerateEdge() error = %v, wantErr %v",
					gotErr, test.wantErr)
			}
		})
	}
}

func TestGraphEdgeProcessorHandleNormalEdge(t *testing.T) {
	tests := []struct {
		name           string
		edge           graphEdge
		options        *graphOptions
		outBegin       int
		outEnd         int
		nOut           int
		nIn            int
		wantErr        bool
		wantErrMessage string
	}{
		{
			name: "keep sibling pairs",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs: siblingPairsKeep,
				edgeType:     edgeTypeDirected,
			},
			outBegin: 0,
			outEnd:   1,
			nOut:     1,
			nIn:      1,
			wantErr:  false,
		},
		{
			name: "invalid sibling pairs option",
			edge: graphEdge{first: 1, second: 2},
			options: &graphOptions{
				siblingPairs: siblingPairs(255), // Use max uint8 value as invalid
				edgeType:     edgeTypeDirected,
			},
			outBegin:       0,
			outEnd:         1,
			nOut:           1,
			nIn:            1,
			wantErr:        true,
			wantErrMessage: "Invalid sibling pairs option",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ep := &edgeProcessor{
				options:      test.options,
				edges:        []graphEdge{test.edge},
				inputIDs:     []int32{1},
				idSetLexicon: newIDSetLexicon(),
				newEdges:     make([]graphEdge, 0),
				newInputIDs:  make([]int32, 0),
				outEdges:     []int32{0}, // Initialize with index 0
				inEdges:      []int32{0}, // Initialize with index 0
			}

			gotErr := ep.handleNormalEdge(test.edge, test.outBegin, test.outEnd, test.nOut, test.nIn)
			if (gotErr != nil) != test.wantErr {
				t.Errorf("handleNormalEdge() error = %v, wantErr %v", gotErr, test.wantErr)
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
		name     string
		edges    []graphEdge
		inputIDs []int32
		options  *graphOptions
		wantErr  bool
	}{
		{
			name:     "empty graph",
			edges:    []graphEdge{},
			inputIDs: []int32{},
			options: &graphOptions{
				edgeType:        edgeTypeDirected,
				duplicateEdges:  duplicateEdgesKeep,
				degenerateEdges: degenerateEdgesKeep,
				siblingPairs:    siblingPairsKeep,
			},
			wantErr: false,
		},
		{
			name: "single edge",
			edges: []graphEdge{
				{first: 1, second: 2},
			},
			inputIDs: []int32{1},
			options: &graphOptions{
				edgeType:        edgeTypeDirected,
				duplicateEdges:  duplicateEdgesKeep,
				degenerateEdges: degenerateEdgesKeep,
				siblingPairs:    siblingPairsKeep,
			},
			wantErr: false,
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
			wantErr: false,
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
		})
	}
}
