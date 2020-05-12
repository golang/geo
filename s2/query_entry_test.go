// Copyright 2020 Google Inc. All rights reserved.
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

	"github.com/golang/geo/s1"
)

func TestQueryQueueEntry(t *testing.T) {
	// This test case came from instrumenting one of the larger C++ unit tests
	// that queued up a number of cells in order to verify that this
	// priority queue implementation matches.
	var cells = []struct {
		dist distance
		cell string
	}{
		{dist: minDistance(s1.ChordAngleFromAngle(s1.Angle(0.220708))), cell: "1/3022"},
		{dist: minDistance(s1.ChordAngleFromAngle(s1.Angle(0.219459))), cell: "1/31"},
		{dist: minDistance(s1.ChordAngleFromAngle(s1.Angle(0.183294))), cell: "1/321"},
		{dist: minDistance(s1.ChordAngleFromAngle(s1.Angle(0.120563))), cell: "1/3311"},
		{dist: minDistance(s1.ChordAngleFromAngle(s1.Angle(0.0713617))), cell: "1/33103"},
		{dist: minDistance(s1.ChordAngleFromAngle(s1.Angle(0.0241046))), cell: "1/3201"},
		{dist: minDistance(s1.ChordAngleFromAngle(s1.Angle(0.119864))), cell: "1/3200"},
	}

	q := newQueryQueue()

	// Add the entries.
	for _, s := range cells {
		q.push(&queryQueueEntry{
			distance:  s.dist,
			id:        cellIDFromString(s.cell),
			indexCell: nil,
		})
	}

	// Insert one new item (should end up in position 2)
	item := &queryQueueEntry{
		distance:  minDistance(s1.ChordAngleFromAngle(s1.Angle(0.107601))),
		id:        cellIDFromString("1/3201003"),
		indexCell: nil,
	}
	q.push(item)

	expectedCellIDs := []CellID{
		cellIDFromString("1/3201"),
		cellIDFromString("1/33103"),
		cellIDFromString("1/3201003"),
		cellIDFromString("1/3200"),
		cellIDFromString("1/3311"),
		cellIDFromString("1/321"),
		cellIDFromString("1/31"),
		cellIDFromString("1/3022"),
	}

	if got, want := q.size(), len(expectedCellIDs); got != want {
		t.Errorf("number of elements different: got %d, want %d", got, want)
	}

	// Take the items out and verfy the queue has them in the preferred order.
	for i, cid := range expectedCellIDs {
		item := q.pop()
		if item.id != cid {
			t.Errorf("element %d: got %s, want %s", i, item.id, cid)
		}
	}
}
