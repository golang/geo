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
	"testing"
)

func TestIncidentEdgeTrackerBasic(t *testing.T) {
	tests := []struct {
		index string
		want  int
	}{
		// These shapeindex strings came from validation query's test
		// corpus to determine which ones ended up actually getting
		// tracked edges.
		{
			// Has 0 tracked edges
			index: "## 0:0, 1:1",
			want:  0,
		},
		{
			// Has 1 tracked edges
			index: "## 2:0, 0:-2, -2:0, 0:2; 2:0, 0:-1, -1:0, 0:1",
			want:  1,
		},
		{
			// Has 2 tracked edges
			index: "## 2:0, 0:-2, -2:0, 0:2; 2:0, 0:-1, -2:0, 0:1",
			want:  2,
		},
	}

	for _, test := range tests {
		index := makeShapeIndex(test.index)
		index.Build()

		iter := index.Iterator()
		celldata := newIndexCellData()
		celldata.loadCell(index, iter.CellID(), iter.IndexCell())

		tracker := newIncidentEdgeTracker()

		for _, clipped := range celldata.indexCell.shapes {
			shapeID := clipped.shapeID
			tracker.startShape(shapeID)
			for _, e := range celldata.shapeEdges(shapeID) {
				tracker.addEdge(e.ID, e.Edge)
			}
			tracker.finishShape()
		}

		if got := len(tracker.edgeMap); got != test.want {
			t.Errorf("incidentEdgeTracker should have %d edges, got %d",
				test.want, got)
		}
	}
}
