package s2

import (
	"slices"
	"testing"
)

type testGraphEdge struct {
	edge     graphEdge
	inputIDs []int32
}

func TestGraphProcessGraphEdges(t *testing.T) {
	tests := []struct {
		name                string
		opts                *graphOptions
		have                []testGraphEdge
		want                []testGraphEdge
		wantErr             bool
		wantChangedEdgeType bool
		wantEdgeType        edgeType
	}{
		{
			name: "Discards Degenerate Edges",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsKeep,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 0, second: 0}},
			},
			want:    []testGraphEdge{},
			wantErr: false,
		},
		{
			name: "Keep Duplicate Degenerate Edges",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesKeep,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsKeep,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 0, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 0, second: 0}},
			},
			wantErr: false,
		},
		{
			name: "MergeDuplicateDegenerateEdges",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesKeep,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsKeep,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{2}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2}},
			},
			wantErr: false,
		},
		{
			name: "Merge Undirected Duplicate Degenerate Edges",
			// Edge count should be reduced to 2 (i.e., one undirected edge), and all
			// labels should be merged.
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesKeep,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsKeep,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1}},
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{2}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2}},
			},
			wantErr: false,
		},
		{
			name: "Converted Undirected Degenerate Edges",
			// Converting from edgeTypeUndirected to edgeTypeDirected cuts the edge
			// count in half and merges any edge labels.
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesKeep,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsRequire,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1}},
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{2}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2}},
			},
			wantErr:             false,
			wantChangedEdgeType: true,
			wantEdgeType:        edgeTypeDirected,
		},
		{
			// Like the previous test case, except that we also merge duplicates.
			name: "Merge Converted Undirected Duplicate Degenerate Edges",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesKeep,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsRequire,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1}},
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{2}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2}},
			},
			wantErr:             false,
			wantChangedEdgeType: true,
			wantEdgeType:        edgeTypeDirected,
		},

		// Test that degenerate edges are discarded if they are connected to any
		// non-degenerate edges (whether they are incoming or outgoing, and whether
		// they are lexicographically before or after the degenerate edge).
		{
			name: "Discard Excess Connected Degenerate Edges_1",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscardExcess,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsKeep,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 0, second: 1}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "Discard Excess Connected Degenerate Edges_2",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscardExcess,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsKeep,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "Discard Excess Connected Degenerate Edges_3",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscardExcess,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsKeep,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 1}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "Discard Excess Connected Degenerate Edges_4",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscardExcess,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsKeep,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 1}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		// Test that degenerateEdgesDiscardExcess merges any duplicate undirected
		// degenerate edges together.
		{
			name: "Discard Excess Undirected Isolated Degenerate Edges",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesDiscardExcess,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsKeep,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1}},
				{edge: graphEdge{first: 0, second: 0}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{2}},
				{edge: graphEdge{first: 0, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		// Test that degenerateEdgesDiscardExcess with SiblingPairsRequire merges any duplicate
		// edges together and converts the edges from edgeTypeUndirected to edgeTypeDirected.
		{
			name: "DiscardExcessConvertedUndirectedIsolatedDegenerateEdges",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesDiscardExcess,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsRequire,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{2}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{3}},
				{edge: graphEdge{first: 0, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2, 3}},
			},
			wantChangedEdgeType: true,
			wantEdgeType:        edgeTypeDirected,
		},
		// Test that when either SiblingPairsDiscard or SiblingPairsDiscardExcess
		// are specified, the edge labels of degenerate edges are merged together
		// (for consistency, since these options merge the labels of all
		// non-degenerate edges as well).
		{
			name: "SiblingPairsDiscardMergesDegenerateEdgeLabels_1",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesKeep,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscard,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{2}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{3}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2, 3}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2, 3}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2, 3}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "SiblingPairsDiscardMergesDegenerateEdgeLabels_2",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesKeep,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{2}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{3}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2, 3}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2, 3}},
				{edge: graphEdge{first: 0, second: 0}, inputIDs: []int32{1, 2, 3}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "KeepSiblingPairs",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsKeep,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "MergeDuplicateSiblingPairs",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsKeep,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		// Check that matched pairs are discarded, leaving behind any excess edges.
		{
			name: "DiscardSiblingPairs_1",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscard,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want:                []testGraphEdge{},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardSiblingPairs_2",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscard,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want:                []testGraphEdge{},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardSiblingPairs_3",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscard,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardSiblingPairs_4",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscard,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		// Check that matched pairs are discarded, and then any remaining edges
		// are merged.
		{
			name: "DiscardSiblingPairsMergeDuplicates_1",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsDiscard,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want:                []testGraphEdge{},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardSiblingPairsMergeDuplicates_2",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsDiscard,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardSiblingPairsMergeDuplicates_3",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsDiscard,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		// An undirected sibling pair consists of four edges, two in each direction
		// (see Builder). Since undirected edges always come in pairs, this
		// means that the result always consists of either 0 or 2 edges.
		{
			name: "DiscardUndirectedSiblingPairs_1",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscard,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardUndirectedSiblingPairs_2",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscard,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want:                []testGraphEdge{},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardUndirectedSiblingPairs_3",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscard,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		// Like SiblingPairsDiscard, except that one sibling pair is kept if the
		// result would otherwise be empty.
		{
			name: "DiscardExcessSiblingPairs_1",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardExcessSiblingPairs_2",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardExcessSiblingPairs_3",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardExcessSiblingPairs_4",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		// Like SiblingPairsDiscard, except that one sibling pair is kept if the
		// result would otherwise be empty.
		{
			name: "DiscardExcessSiblingPairsMergeDuplicates_1",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardExcessSiblingPairsMergeDuplicates_2",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardExcessSiblingPairsMergeDuplicates_3",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		// Like SiblingPairsDiscard, except that one undirected sibling pair
		// (4 edges) is kept if the result would otherwise be empty.
		{
			name: "DiscardExcessUndirectedSiblingPairs_1",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardExcessUndirectedSiblingPairs_2",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "DiscardExcessUndirectedSiblingPairs_3",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsDiscardExcess,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "CreateSiblingPairs_1",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsCreate,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "CreateSiblingPairs_2",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsCreate,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "CreateSiblingPairs_3",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsCreate,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		// Like the test case "Create Sibling Pairs", but should generate an error.
		{
			name: "RequireSiblingPairs_1",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsRequire,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			// Requires sibling pairs but only has one edge, so should generate an error.
			name: "RequireSiblingPairs_2",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsRequire,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             true,
			wantChangedEdgeType: false,
		},
		// An undirected sibling pair consists of 4 edges, but SiblingPairsCreate
		// also converts the graph to EdgeTypeDirected and cuts the number of
		// edges in half.
		{
			name: "CreateUndirectedSiblingPairs_1",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsCreate,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: true,
			wantEdgeType:        edgeTypeDirected,
		},
		{
			name: "CreateUndirectedSiblingPairs_2",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsCreate,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: true,
			wantEdgeType:        edgeTypeDirected,
		},
		{
			name: "CreateUndirectedSiblingPairs_3",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesKeep,
				siblingPairs:    siblingPairsCreate,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: true,
			wantEdgeType:        edgeTypeDirected,
		},
		{
			name: "CreateSiblingPairsMergeDuplicates_1",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsCreate,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "CreateSiblingPairsMergeDuplicates_2",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsCreate,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: false,
		},
		{
			name: "CreateUndirectedSiblingPairsMergeDuplicates_1",
			opts: &graphOptions{
				edgeType:        edgeTypeDirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsCreate,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: true,
			wantEdgeType:        edgeTypeDirected,
		},
		{
			name: "CreateUndirectedSiblingPairsMergeDuplicates_2",
			opts: &graphOptions{
				edgeType:        edgeTypeUndirected,
				degenerateEdges: degenerateEdgesDiscard,
				duplicateEdges:  duplicateEdgesMerge,
				siblingPairs:    siblingPairsCreate,
			},
			have: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			want: []testGraphEdge{
				{edge: graphEdge{first: 0, second: 1}},
				{edge: graphEdge{first: 1, second: 0}},
			},
			wantErr:             false,
			wantChangedEdgeType: true,
			wantEdgeType:        edgeTypeDirected,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			edges := make([]graphEdge, len(test.have))
			idSetLexicon := newIDSetLexicon()
			inputIDSetIDs := []int32{}
			for i, e := range test.have {
				edges[i] = e.edge
				inputIDSetIDs = append(inputIDSetIDs, idSetLexicon.add(e.inputIDs...))
			}
			gotEdges, gotIDs, err := processGraphEdges(test.opts, edges, inputIDSetIDs, idSetLexicon)

			if (err != nil) != test.wantErr {
				t.Errorf("err != nil = %v, wanted %v", err != nil, test.wantErr)
			}
			if len(gotEdges) != len(gotIDs) {
				t.Errorf("Num edges (%d) != num IDs (%d)", len(edges), len(inputIDSetIDs))
			}

			for i, want := range test.want {
				if i > len(gotEdges) {
					t.Errorf("Not enough output edges")
				}
				if want.edge != gotEdges[i] {
					t.Errorf("got[%d] = %+v, want %+v", i, gotEdges[i], want.edge)
				}

				actualIDs := idSetLexicon.idSet(gotIDs[i])
				if !slices.Equal(want.inputIDs, actualIDs) {
					t.Errorf("edge %d:  got: %+v, want: %+v", i, actualIDs, gotIDs)
				}
			}
			if len(test.want) != len(gotEdges) {
				t.Errorf("Too many output edges %d", len(gotEdges))
			}

			if test.wantChangedEdgeType {
				if test.wantEdgeType != test.opts.edgeType {
					t.Errorf("tested option and input combination should have changed edgeType but didn't")
				}
			}
		})
	}
}
