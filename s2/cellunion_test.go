/*
Copyright 2014 Google Inc. All rights reserved.

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

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestNormalization(t *testing.T) {
	cu := CellUnion{
		0x80855c0000000000, // A: a cell over Pittsburg CA
		0x80855d0000000000, // B, a child of A
		0x8085634000000000, // first child of X, disjoint from A
		0x808563c000000000, // second child of X
		0x80855dc000000000, // a child of B
		0x808562c000000000, // third child of X
		0x8085624000000000, // fourth child of X
		0x80855d0000000000, // B again
	}
	exp := CellUnion{
		0x80855c0000000000, // A
		0x8085630000000000, // X
	}
	cu.Normalize()
	if !reflect.DeepEqual(cu, exp) {
		t.Errorf("got %v, want %v", cu, exp)
	}

	// add a redundant cell
	/* TODO(dsymonds)
	cu.Add(0x808562c000000000)
	if !reflect.DeepEqual(cu, exp) {
		t.Errorf("after redundant add, got %v, want %v", cu, exp)
	}
	*/
}

func TestIntersects(t *testing.T) {
	tests := []struct {
		cells    []CellID
		overlaps []CellID
		disjoint []CellID
	}{
		{
			// Singe cell around NYC, and some simple nearby probes
			cells: []CellID{0x89c25c0000000000},
			overlaps: []CellID{
				CellID(0x89c25c0000000000).immediateParent(),
				CellIDFromFace(CellID(0x89c25c0000000000).Face()), // the whole face
				CellID(0x89c25c0000000000).ChildBegin(),
				CellID(0x89c25c0000000000).ChildBeginAtLevel(28),
			},
			disjoint: []CellID{
				CellID(0x89c25c0000000000).Next(),                       // Cell next to this one at same level
				CellID(0x89c25c0000000000).Next().ChildBeginAtLevel(28), // Cell next to this one at deep level
				0x89c2700000000000,                                      // Big(er) neighbor cell
				0x89e9000000000000,                                      // Very big next door cell.
				0x89c1000000000000,                                      // Very big cell, smaller value than probe
			},
		},

		{
			// NYC and SFO:
			cells: []CellID{
				0x89c25b0000000000, // NYC
				0x89c2590000000000, // NYC
				0x89c2f70000000000, // NYC
				0x89c2f50000000000, // NYC
				0x8085870000000000, // SFO
				0x8085810000000000, // SFO
				0x808f7d0000000000, // SFO
				0x808f7f0000000000, // SFO
			},
			overlaps: []CellID{
				0x808f7ef300000000, // SFO
				0x808f7e5cf0000000, // SFO
				0x808587f000000000, // SFO
				0x808c000000000000, // Big SFO

				0x89c25ac000000000, // NYC
				0x89c259a400000000, // NYC
				0x89c258fa10000000, // NYC
				0x89c258f174007000, // NYC
				0x89c4000000000000, // Big NYC
			},
			disjoint: []CellID{
				0x89c15a4fcb1bb000, // outside NYC
				0x89c15a4e4aa95000, // outside NYC
				0x8094000000000000, // outside SFO (big)
				0x8096f10000000000, // outside SFO (smaller)

				0x87c0000000000000, // Midwest very big
			},
		},
		{
			// CellUnion with cells at many levels:
			cells: []CellID{
				0x8100000000000000, // starting around california
				0x8740000000000000, // adjacent cells at increasing
				0x8790000000000000, // levels, moving eastward.
				0x87f4000000000000,
				0x87f9000000000000, // going down across the midwest
				0x87ff400000000000,
				0x87ff900000000000,
				0x87fff40000000000,
				0x87fff90000000000,
				0x87ffff4000000000,
				0x87ffff9000000000,
				0x87fffff400000000,
				0x87fffff900000000,
				0x87ffffff40000000,
				0x87ffffff90000000,
				0x87fffffff4000000,
				0x87fffffff9000000,
				0x87ffffffff400000, // to a very small cell in Wisconsin
			},
			overlaps: []CellID{
				0x808f400000000000,
				0x80eb118b00000000,
				0x8136a7a11d000000,
				0x8136a7a11dac0000,
				0x876c7c0000000000,
				0x87f96d0000000000,
				0x87ffffffff400000,
			},
			disjoint: []CellID{
				0x52aaaaaaab300000,
				0x52aaaaaaacd00000,
				0x87fffffffa100000,
				0x87ffffffed500000,
				0x87ffffffa0100000,
				0x87fffffed5540000,
				0x87fffffed6240000,
				0x52aaaacccb340000,
				0x87a0000400000000,
				0x87a000001f000000,
				0x87a0000029d00000,
				0x9500000000000000,
			},
		},
	}
	for _, test := range tests {
		var union CellUnion = test.cells
		union.Normalize()

		// Ensure self-intersecting tests are always correct:
		for _, id := range test.cells {
			if union.Intersects(id) != true {
				t.Errorf("CellUnion %v Should self-intersect %v but does not", union, id)
			}
		}
		// Test for other intersections specified in test case.
		for _, id := range test.overlaps {
			if union.Intersects(id) != true {
				t.Errorf("CellUnion %v Should contain %v but does not", union, id)
			}
		}
		// Negative cases make sure we don't intersect these cells
		for _, id := range test.disjoint {
			if union.Intersects(id) != false {
				t.Errorf("CellUnion %v Should NOT contain %v but does not", union, id)
			}
		}
	}
}

func oneIn(n int) bool {
	return rand.Intn(n) == 0
}

func addCells(id CellID, selected bool, input *[]CellID, expected *[]CellID, t *testing.T) {
	// Decides whether to add "id" and/or some of its descendants to the test case.  If "selected"
	// is true, then the region covered by "id" *must* be added to the test case (either by adding
	// "id" itself, or some combination of its descendants, or both).  If cell ids are to the test
	// case "input", then the corresponding expected result after simplification is added to
	// "expected".

	if id == 0 {
		// Initial call: decide whether to add cell(s) from each face.
		for face := 0; face < 6; face++ {
			addCells(CellIDFromFace(face), false, input, expected, t)
		}
		return
	}

	if id.IsLeaf() {
		// The oneIn() call below ensures that the parent of a leaf cell will always be selected (if
		// we make it that far down the hierarchy).
		if selected != true {
			t.Errorf("id IsLeaf() and not selected")
		}
		*input = append(*input, id)
		return
	}

	// The following code ensures that the probability of selecting a cell at each level is
	// approximately the same, i.e. we test normalization of cells at all levels.
	if !selected && oneIn(maxLevel-id.Level()) {
		//  Once a cell has been selected, the expected output is predetermined.  We then make sure
		//  that cells are selected that will normalize to the desired output.
		*expected = append(*expected, id)
		selected = true

	}

	// With the rnd.OneIn() constants below, this function adds an average
	// of 5/6 * (kMaxLevel - level) cells to "input" where "level" is the
	// level at which the cell was first selected (level 15 on average).
	// Therefore the average number of input cells in a test case is about
	// (5/6 * 15 * 6) = 75.  The average number of output cells is about 6.

	// If a cell is selected, we add it to "input" with probability 5/6.
	added := false
	if selected && !oneIn(6) {
		*input = append(*input, id)
		added = true
	}
	numChildren := 0
	for child := id.ChildBegin(); child != id.ChildEnd(); child = child.Next() {
		// If the cell is selected, on average we recurse on 4/12 = 1/3 child.
		// This intentionally may result in a cell and some of its children
		// being included in the test case.
		//
		// If the cell is not selected, on average we recurse on one child.
		// We also make sure that we do not recurse on all 4 children, since
		// then we might include all 4 children in the input case by accident
		// (in which case the expected output would not be correct).
		recurse := false
		if selected {
			recurse = oneIn(12)
		} else {
			recurse = oneIn(4)
		}
		if recurse && numChildren < 3 {
			addCells(child, selected, input, expected, t)
			numChildren++
		}
		// If this cell was selected but the cell itself was not added, we
		// must ensure that all 4 children (or some combination of their
		// descendants) are added.

		if selected && !added {
			addCells(child, selected, input, expected, t)
		}
	}
}

func TestNormalizePseudoRandom(t *testing.T) {
	// Try a bunch of random test cases, and keep track of average statistics for normalization (to
	// see if they agree with the analysis above).

	inSum := 0
	outSum := 0
	iters := 2000

	for i := 0; i < iters; i++ {
		input := []CellID{}
		expected := []CellID{}
		addCells(CellID(0), false, &input, &expected, t)
		inSum += len(input)
		outSum += len(expected)
		cellunion := CellUnion(input)
		cellunion.Normalize()

		if len(expected) != len(cellunion) {
			t.Errorf("Expected size of union to be %d, but got %d.",
				len(expected), len(cellunion))
		}

		for _, j := range input {
			if cellunion.Intersects(j) == false {
				t.Errorf("Expected intersection with %v.", j)
			}

			if !j.isFace() {
				if cellunion.Intersects(j.immediateParent()) == false {
					t.Errorf("Expected intersection with parent cell %v.", j.immediateParent())
					if j.Level() > 1 {
						if cellunion.Intersects(j.immediateParent().immediateParent()) == false {
							t.Errorf("Expected intersection with parent's parent %v.",
								j.immediateParent().immediateParent())
						}
						if cellunion.Intersects(j.Parent(0)) == false {
							t.Errorf("Expected intersection with parent %v at level 0.", j.Parent(0))
						}
					}
				}
			}

			if !j.IsLeaf() {
				if cellunion.Intersects(j.ChildBegin()) == false {
					t.Errorf("Expected intersection with %v.", j.ChildBegin())
				}
				if cellunion.Intersects(j.ChildEnd().Prev()) == false {
					t.Errorf("Expected intersection with %v.", j.ChildEnd().Prev())
				}
				if cellunion.Intersects(j.ChildBeginAtLevel(maxLevel)) == false {
					t.Errorf("Expected intersection with %v.", j.ChildBeginAtLevel(maxLevel))
				}
			}
		}

		var test []CellID
		var dummy []CellID
		addCells(CellID(0), false, &test, &dummy, t)
		for _, j := range test {
			intersects := false
			for _, k := range expected {
				if k.Intersects(j) {
					intersects = true
				}
			}
			if cellunion.Intersects(j) != intersects {
				t.Errorf("Expected intersection with %v.", (uint64)(j))
			}
		}
	}
	t.Logf("avg in %.2f, avg out %.2f\n", (float64)(inSum)/(float64)(iters), (float64)(outSum)/(float64)(iters))
}
