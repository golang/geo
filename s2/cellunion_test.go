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
	"math"
	"reflect"
	"testing"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/s1"
)

func TestCellUnionDuplicateCellsNotValid(t *testing.T) {
	id := cellIDFromPoint(PointFromCoords(1, 0, 0))
	cu := CellUnion([]CellID{id, id})
	if cu.IsValid() {
		t.Errorf("%v.IsValid() = true, want false", cu)
	}
}

func TestCellUnionUnsortedCellsNotValid(t *testing.T) {
	id := cellIDFromPoint(PointFromCoords(1, 0, 0)).Parent(10)
	cu := CellUnion([]CellID{id, id.Prev()})
	if cu.IsValid() {
		t.Errorf("%v.IsValid() = true, want false", cu)
	}
}

func TestCellUnionIsNormalized(t *testing.T) {
	id := cellIDFromPoint(PointFromCoords(1, 0, 0)).Parent(10)
	children := id.Children()
	cu := CellUnion([]CellID{children[0], children[1], children[2], children[3]})
	if !(cu.IsValid()) {
		t.Errorf("%v.IsValid() = false, want true", cu)
	}
	if cu.IsNormalized() {
		t.Errorf("%v.IsNormalized() = true, want false", cu)
	}
}

func TestCellUnionInvalidCellIdNotValid(t *testing.T) {
	cu := CellUnion([]CellID{CellID(0)})
	if cu.IsValid() {
		t.Error("CellUnion containing an invalid CellID should not be valid")
	}
}

func TestCellUnionAreSiblings(t *testing.T) {
	id := cellIDFromPoint(PointFromCoords(1, 0, 0)).Parent(10)
	children := id.Children()
	if siblings := areSiblings(children[0], children[1], children[2], children[3]); !siblings {
		t.Errorf("areSiblings(%v, %v, %v, %v) = false, want true", children[0], children[1], children[2], children[3])
	}

	if siblings := areSiblings(id, children[1], children[2], children[3]); siblings {
		t.Errorf("areSiblings(%v, %v, %v, %v) = true, want false", id, children[1], children[2], children[3])
	}
}

func TestCellUnionNormalization(t *testing.T) {
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

func TestCellUnionBasic(t *testing.T) {
	empty := CellUnion{}
	empty.Normalize()
	if len(empty) != 0 {
		t.Errorf("empty CellUnion had %d cells, want 0", len(empty))
	}

	face1ID := CellIDFromFace(1)
	face1Cell := CellFromCellID(face1ID)
	face1Union := CellUnion{face1ID}
	face1Union.Normalize()
	if len(face1Union) != 1 {
		t.Errorf("%v had %d cells, want 1", face1Union, len(face1Union))
	}
	if face1ID != face1Union[0] {
		t.Errorf("%v[0] = %v, want %v", face1Union, face1Union[0], face1ID)
	}
	if got := face1Union.ContainsCell(face1Cell); !got {
		t.Errorf("%v.ContainsCell(%v) = %t, want %t", face1Union, face1Cell, got, true)
	}

	face2ID := CellIDFromFace(2)
	face2Cell := CellFromCellID(face2ID)
	face2Union := CellUnion{face2ID}
	face2Union.Normalize()
	if len(face2Union) != 1 {
		t.Errorf("%v had %d cells, want 1", face2Union, len(face2Union))
	}
	if face2ID != face2Union[0] {
		t.Errorf("%v[0] = %v, want %v", face2Union, face2Union[0], face2ID)
	}

	if got := face1Union.ContainsCell(face2Cell); got {
		t.Errorf("%v.ContainsCell(%v) = %t, want %t", face1Union, face2Cell, got, false)
	}

}

func TestCellUnion(t *testing.T) {
	tests := []struct {
		cells     []CellID // A test CellUnion.
		contained []CellID // List of cellIDs contained in the CellUnion.
		overlaps  []CellID // List of CellIDs that intersects the CellUnion but not contained in it.
		disjoint  []CellID // List of CellIDs that are disjoint from the CellUnion.
	}{
		{
			// Single cell around NYC, and some simple nearby probes
			cells: []CellID{0x89c25c0000000000},
			contained: []CellID{
				CellID(0x89c25c0000000000).ChildBegin(),
				CellID(0x89c25c0000000000).ChildBeginAtLevel(28),
			},
			overlaps: []CellID{
				CellID(0x89c25c0000000000).immediateParent(),
				CellIDFromFace(CellID(0x89c25c0000000000).Face()), // the whole face
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
			contained: []CellID{
				0x808f7ef300000000, // SFO
				0x808f7e5cf0000000, // SFO
				0x808587f000000000, // SFO
				0x89c25ac000000000, // NYC
				0x89c259a400000000, // NYC
				0x89c258fa10000000, // NYC
				0x89c258f174007000, // NYC
			},
			overlaps: []CellID{
				0x808c000000000000, // Big SFO
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
			contained: []CellID{
				0x808f400000000000,
				0x80eb118b00000000,
				0x8136a7a11d000000,
				0x8136a7a11dac0000,
				0x876c7c0000000000,
				0x87f96d0000000000,
				0x87ffffffff400000,
			},
			overlaps: []CellID{
				CellID(0x8100000000000000).immediateParent(),
				CellID(0x8740000000000000).immediateParent(),
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
		union := CellUnion(test.cells)
		union.Normalize()

		// Ensure self-containment tests are correct.
		for _, id := range test.cells {
			if !union.IntersectsCellID(id) {
				t.Errorf("CellUnion %v should self-intersect %v but does not", union, id)
			}
			if !union.ContainsCellID(id) {
				t.Errorf("CellUnion %v should self-contain %v but does not", union, id)
			}
		}
		// Test for containment specified in test case.
		for _, id := range test.contained {
			if !union.IntersectsCellID(id) {
				t.Errorf("CellUnion %v should intersect %v but does not", union, id)
			}
			if !union.ContainsCellID(id) {
				t.Errorf("CellUnion %v should contain %v but does not", union, id)
			}
		}
		// Make sure the CellUnion intersect these cells but do not contain.
		for _, id := range test.overlaps {
			if !union.IntersectsCellID(id) {
				t.Errorf("CellUnion %v should intersect %v but does not", union, id)
			}
			if union.ContainsCellID(id) {
				t.Errorf("CellUnion %v should not contain %v but does", union, id)
			}
		}
		// Negative cases make sure the CellUnion neither contain nor intersect these cells
		for _, id := range test.disjoint {
			if union.IntersectsCellID(id) {
				t.Errorf("CellUnion %v should not intersect %v but does", union, id)
			}
			if union.ContainsCellID(id) {
				t.Errorf("CellUnion %v should not contain %v but does", union, id)
			}
		}
	}
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

func TestCellUnionNormalizePseudoRandom(t *testing.T) {
	// Try a bunch of random test cases, and keep track of average statistics
	// for normalization (to see if they agree with the analysis above).

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

		// Test CapBound().
		cb := cellunion.CapBound()
		for _, ci := range cellunion {
			if !cb.ContainsCell(CellFromCellID(ci)) {
				t.Errorf("CapBound %v of union %v should contain cellID %v", cb, cellunion, ci)
			}
		}

		for _, j := range input {
			if !cellunion.ContainsCellID(j) {
				t.Errorf("Expected containment of CellID %v", j)
			}

			if cellunion.IntersectsCellID(j) == false {
				t.Errorf("Expected intersection with %v.", j)
			}

			if !j.isFace() {
				if cellunion.IntersectsCellID(j.immediateParent()) == false {
					t.Errorf("Expected intersection with parent cell %v.", j.immediateParent())
					if j.Level() > 1 {
						if cellunion.IntersectsCellID(j.immediateParent().immediateParent()) == false {
							t.Errorf("Expected intersection with parent's parent %v.",
								j.immediateParent().immediateParent())
						}
						if cellunion.IntersectsCellID(j.Parent(0)) == false {
							t.Errorf("Expected intersection with parent %v at level 0.", j.Parent(0))
						}
					}
				}
			}

			if !j.IsLeaf() {
				if cellunion.ContainsCellID(j.ChildBegin()) == false {
					t.Errorf("Expected containment of %v.", j.ChildBegin())
				}
				if cellunion.IntersectsCellID(j.ChildBegin()) == false {
					t.Errorf("Expected intersection with %v.", j.ChildBegin())
				}
				if cellunion.ContainsCellID(j.ChildEnd().Prev()) == false {
					t.Errorf("Expected containment of %v.", j.ChildEnd().Prev())
				}
				if cellunion.IntersectsCellID(j.ChildEnd().Prev()) == false {
					t.Errorf("Expected intersection with %v.", j.ChildEnd().Prev())
				}
				if cellunion.ContainsCellID(j.ChildBeginAtLevel(maxLevel)) == false {
					t.Errorf("Expected containment of %v.", j.ChildBeginAtLevel(maxLevel))
				}
				if cellunion.IntersectsCellID(j.ChildBeginAtLevel(maxLevel)) == false {
					t.Errorf("Expected intersection with %v.", j.ChildBeginAtLevel(maxLevel))
				}
			}
		}

		for _, exp := range expected {
			if !exp.isFace() {
				if cellunion.ContainsCellID(exp.Parent(exp.Level() - 1)) {
					t.Errorf("cellunion should not contain its parent %v", exp.Parent(exp.Level()-1))
				}
				if cellunion.ContainsCellID(exp.Parent(0)) {
					t.Errorf("cellunion should not contain the top level parent %v", exp.Parent(0))
				}
			}
		}

		var test []CellID
		var dummy []CellID
		addCells(CellID(0), false, &test, &dummy, t)
		for _, j := range test {
			intersects := false
			contains := false
			for _, k := range expected {
				if k.Contains(j) {
					contains = true
				}
				if k.Intersects(j) {
					intersects = true
				}
			}
			if cellunion.ContainsCellID(j) != contains {
				t.Errorf("Expected contains with %v.", (uint64)(j))
			}
			if cellunion.IntersectsCellID(j) != intersects {
				t.Errorf("Expected intersection with %v.", (uint64)(j))
			}
		}
	}
	t.Logf("avg in %.2f, avg out %.2f\n", (float64)(inSum)/(float64)(iters), (float64)(outSum)/(float64)(iters))
}

func TestCellUnionDenormalize(t *testing.T) {
	tests := []struct {
		name string
		minL int
		lMod int
		cu   *CellUnion
		exp  *CellUnion
	}{
		{
			"not expanded, level mod == 1",
			10,
			1,
			&CellUnion{
				CellIDFromFace(2).ChildBeginAtLevel(11),
				CellIDFromFace(2).ChildBeginAtLevel(11),
				CellIDFromFace(3).ChildBeginAtLevel(14),
				CellIDFromFace(0).ChildBeginAtLevel(10),
			},
			&CellUnion{
				CellIDFromFace(2).ChildBeginAtLevel(11),
				CellIDFromFace(2).ChildBeginAtLevel(11),
				CellIDFromFace(3).ChildBeginAtLevel(14),
				CellIDFromFace(0).ChildBeginAtLevel(10),
			},
		},
		{
			"not expanded, level mod > 1",
			10,
			2,
			&CellUnion{
				CellIDFromFace(2).ChildBeginAtLevel(12),
				CellIDFromFace(2).ChildBeginAtLevel(12),
				CellIDFromFace(3).ChildBeginAtLevel(14),
				CellIDFromFace(0).ChildBeginAtLevel(10),
			},
			&CellUnion{
				CellIDFromFace(2).ChildBeginAtLevel(12),
				CellIDFromFace(2).ChildBeginAtLevel(12),
				CellIDFromFace(3).ChildBeginAtLevel(14),
				CellIDFromFace(0).ChildBeginAtLevel(10),
			},
		},
		{
			"expended, (level - min_level) is not multiple of level mod",
			10,
			3,
			&CellUnion{
				CellIDFromFace(2).ChildBeginAtLevel(12),
				CellIDFromFace(5).ChildBeginAtLevel(11),
			},
			&CellUnion{
				CellIDFromFace(2).ChildBeginAtLevel(12).Children()[0],
				CellIDFromFace(2).ChildBeginAtLevel(12).Children()[1],
				CellIDFromFace(2).ChildBeginAtLevel(12).Children()[2],
				CellIDFromFace(2).ChildBeginAtLevel(12).Children()[3],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[0].Children()[0],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[0].Children()[1],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[0].Children()[2],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[0].Children()[3],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[1].Children()[0],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[1].Children()[1],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[1].Children()[2],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[1].Children()[3],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[2].Children()[0],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[2].Children()[1],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[2].Children()[2],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[2].Children()[3],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[3].Children()[0],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[3].Children()[1],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[3].Children()[2],
				CellIDFromFace(5).ChildBeginAtLevel(11).Children()[3].Children()[3],
			},
		},
		{
			"expended, level < min_level",
			10,
			3,
			&CellUnion{
				CellIDFromFace(2).ChildBeginAtLevel(9),
			},
			&CellUnion{
				CellIDFromFace(2).ChildBeginAtLevel(9).Children()[0],
				CellIDFromFace(2).ChildBeginAtLevel(9).Children()[1],
				CellIDFromFace(2).ChildBeginAtLevel(9).Children()[2],
				CellIDFromFace(2).ChildBeginAtLevel(9).Children()[3],
			},
		},
	}
	for _, test := range tests {
		if test.cu.Denormalize(test.minL, test.lMod); !reflect.DeepEqual(test.cu, test.exp) {
			t.Errorf("test: %s; got %v, want %v", test.name, test.cu, test.exp)
		}
	}
}

func TestCellUnionRectBound(t *testing.T) {
	tests := []struct {
		cu   *CellUnion
		want Rect
	}{
		{&CellUnion{}, EmptyRect()},
		{
			&CellUnion{CellIDFromFace(1)},
			Rect{
				r1.Interval{-math.Pi / 4, math.Pi / 4},
				s1.Interval{math.Pi / 4, 3 * math.Pi / 4},
			},
		},
		{
			&CellUnion{
				0x808c000000000000, // Big SFO
			},
			Rect{
				r1.Interval{
					float64(s1.Degree * 34.644220547108482),
					float64(s1.Degree * 38.011928357226651),
				},
				s1.Interval{
					float64(s1.Degree * -124.508522987668428),
					float64(s1.Degree * -121.628309835221216),
				},
			},
		},
		{
			&CellUnion{
				0x89c4000000000000, // Big NYC
			},
			Rect{
				r1.Interval{
					float64(s1.Degree * 38.794595155857657),
					float64(s1.Degree * 41.747046884651063),
				},
				s1.Interval{
					float64(s1.Degree * -76.456308667788633),
					float64(s1.Degree * -73.465162142654819),
				},
			},
		},
		{
			&CellUnion{
				0x89c4000000000000, // Big NYC
				0x808c000000000000, // Big SFO
			},
			Rect{
				r1.Interval{
					float64(s1.Degree * 34.644220547108482),
					float64(s1.Degree * 41.747046884651063),
				},
				s1.Interval{
					float64(s1.Degree * -124.508522987668428),
					float64(s1.Degree * -73.465162142654819),
				},
			},
		},
	}

	for _, test := range tests {
		if got := test.cu.RectBound(); !rectsApproxEqual(got, test.want, epsilon, epsilon) {
			t.Errorf("%v.RectBound() = %v, want %v", test.cu, got, test.want)
		}
	}
}

func TestCellUnionLeafCellsCovered(t *testing.T) {
	fiveFaces := CellUnion{CellIDFromFace(0)}
	fiveFaces.ExpandAtLevel(0)
	wholeWorld := CellUnion{CellIDFromFace(0)}
	wholeWorld.ExpandAtLevel(0)
	wholeWorld.ExpandAtLevel(0)

	tests := []struct {
		have []CellID
		want int64
	}{
		{},
		{
			have: []CellID{},
			want: 0,
		},
		{
			// One leaf cell on face 0.
			have: []CellID{
				CellIDFromFace(0).ChildBeginAtLevel(maxLevel),
			},
			want: 1,
		},
		{
			// Face 0 itself (which includes the previous leaf cell).
			have: []CellID{
				CellIDFromFace(0).ChildBeginAtLevel(maxLevel),
				CellIDFromFace(0),
			},
			want: 1 << 60,
		},
		{
			have: fiveFaces,
			want: 5 << 60,
		},
		{
			have: wholeWorld,
			want: 6 << 60,
		},
		{
			// Add some disjoint cells.
			have: []CellID{
				CellIDFromFace(0).ChildBeginAtLevel(maxLevel),
				CellIDFromFace(0),
				CellIDFromFace(1).ChildBeginAtLevel(1),
				CellIDFromFace(2).ChildBeginAtLevel(2),
				CellIDFromFace(2).ChildEndAtLevel(2).Prev(),
				CellIDFromFace(3).ChildBeginAtLevel(14),
				CellIDFromFace(4).ChildBeginAtLevel(27),
				CellIDFromFace(4).ChildEndAtLevel(15).Prev(),
				CellIDFromFace(5).ChildBeginAtLevel(30),
			},
			want: 1 + (1 << 6) + (1 << 30) + (1 << 32) +
				(2 << 56) + (1 << 58) + (1 << 60),
		},
	}

	for _, test := range tests {
		cu := CellUnion(test.have)
		cu.Normalize()
		if got := cu.LeafCellsCovered(); got != test.want {
			t.Errorf("CellUnion(%v).LeafCellsCovered() = %v, want %v", cu, got, test.want)
		}
	}
}

func TestCellUnionFromRange(t *testing.T) {
	for iter := 0; iter < 2000; iter++ {
		min := randomCellIDForLevel(maxLevel)
		max := randomCellIDForLevel(maxLevel)
		if min > max {
			min, max = max, min
		}

		cu := CellUnionFromRange(min, max.Next())
		if len(cu) <= 0 {
			t.Errorf("len(CellUnionFromRange(%v, %v)) = %d, want > 0", min, max.Next(), len(cu))
		}
		if min != cu[0].RangeMin() {
			t.Errorf("%v.RangeMin of CellUnion should not be below the minimum value it was created from %v", cu[0], min)
		}
		if max != cu[len(cu)-1].RangeMax() {
			t.Errorf("%v.RangeMax of CellUnion should not be above the maximum value it was created from %v", cu[len(cu)-1], max)
		}
		for i := 1; i < len(cu); i++ {
			if got, want := cu[i].RangeMin(), cu[i-1].RangeMax().Next(); got != want {
				t.Errorf("%v.RangeMin() = %v, want %v", cu[i], got, want)
			}
		}
	}

	// Focus on test cases that generate an empty or full range.

	// Test an empty range before the minimum CellID.
	idBegin := CellIDFromFace(0).ChildBeginAtLevel(maxLevel)
	cu := CellUnionFromRange(idBegin, idBegin)
	if len(cu) != 0 {
		t.Errorf("CellUnionFromRange with begin and end as the first CellID should be empty, got %d", len(cu))
	}

	// Test an empty range after the maximum CellID.
	idEnd := CellIDFromFace(5).ChildEndAtLevel(maxLevel)
	cu = CellUnionFromRange(idEnd, idEnd)
	if len(cu) != 0 {
		t.Errorf("CellUnionFromRange with begin and end as the last CellID should be empty, got %d", len(cu))
	}

	// Test the full sphere.
	cu = CellUnionFromRange(idBegin, idEnd)
	if len(cu) != 6 {
		t.Errorf("CellUnionFromRange from first CellID to last CellID should have 6 cells, got %d", len(cu))
	}

	for i := 0; i < len(cu); i++ {
		if !cu[i].isFace() {
			t.Errorf("CellUnionFromRange for full sphere cu[%d].isFace() = %t, want %t", i, cu[i].isFace(), true)
		}
	}
}

func TestCellUnionFromUnionDiffIntersection(t *testing.T) {
	const iters = 2000
	for i := 0; i < iters; i++ {
		input := []CellID{}
		expected := []CellID{}
		addCells(CellID(0), false, &input, &expected, t)

		var x, y, xOrY, xAndY []CellID
		for _, id := range input {
			inX := oneIn(2)
			inY := oneIn(2)

			if inX {
				x = append(x, id)
			}
			if inY {
				y = append(y, id)
			}
			if inX || inY {
				xOrY = append(xOrY, id)
			}
		}

		xcells := CellUnion(x)
		xcells.Normalize()
		ycells := CellUnion(y)
		ycells.Normalize()
		xOrYExpected := CellUnion(xOrY)
		xOrYExpected.Normalize()

		xOrYCells := CellUnionFromUnion(xcells, ycells)

		if !xOrYCells.Equal(xOrYExpected) {
			t.Errorf("CellUnionFromUnion(%v, %v) = %v, want %v", xcells, ycells, xOrYCells, xOrYExpected)
		}

		// Compute the intersection of x with each cell of y,
		// check that this intersection is correct, and append the
		// results to xAndYExpected.
		for _, yid := range ycells {
			u := CellUnionFromIntersectionWithCellID(xcells, yid)
			for _, xid := range xcells {
				if xid.Contains(yid) {
					if !(len(u) == 1 && u[0] == yid) {
						t.Errorf("CellUnionFromIntersectionWithCellID(%v, %v) = %v with len: %d, want len of 1.", xcells, yid, u, len(u))
					}
				} else if yid.Contains(xid) {
					if !u.ContainsCellID(xid) {
						t.Errorf("%v.ContainsCellID(%v) = false, want true", u, xid)
					}
				}
			}
			for _, uCellID := range u {
				if !xcells.ContainsCellID(uCellID) {
					t.Errorf("%v.ContainsCellID(%v) = false, but should contain CellID that was used to create CellUnion", xcells, uCellID)
				}
				if !yid.Contains(uCellID) {
					t.Errorf("%v.Contains(%v) = false, but should contain CellID that was used to create CellUnion", yid, uCellID)
				}
			}

			xAndY = append(xAndY, u...)
		}

		xAndYExpected := CellUnion(xAndY)
		xAndYExpected.Normalize()

		xAndYCells := CellUnionFromIntersection(xcells, ycells)
		if !xAndYCells.Equal(xAndYExpected) {
			t.Errorf("CellUnionFromIntersection(%v, %v) = %v, want %v", xcells, ycells, xAndYCells, xAndYExpected)
		}

		xMinusYCells := CellUnionFromDifference(xcells, ycells)
		yMinusXCells := CellUnionFromDifference(ycells, xcells)
		if !xcells.Contains(xMinusYCells) {
			t.Errorf("%v.Contains(%v) = false, want true", xcells, xMinusYCells)
		}
		if xMinusYCells.Intersects(ycells) {
			t.Errorf("%v.Intersects(%v) = true, want false", xMinusYCells, ycells)
		}
		if !ycells.Contains(yMinusXCells) {
			t.Errorf("%v.Contains(%v) = false, want true", ycells, yMinusXCells)
		}
		if yMinusXCells.Intersects(xcells) {
			t.Errorf("%v.Intersects(%v) = true, want false", yMinusXCells, xcells)
		}
		if xMinusYCells.Intersects(yMinusXCells) {
			t.Errorf("%v.Intersects(%v) = true, want false", xMinusYCells, yMinusXCells)
		}

		diffUnion := CellUnionFromUnion(xMinusYCells, yMinusXCells)
		diffIntersectionUnion := CellUnionFromUnion(diffUnion, xAndYCells)
		if !diffIntersectionUnion.Equal(xOrYCells) {
			t.Errorf("Union(%v, %v).Union(%v) = %v, want %v", xMinusYCells, yMinusXCells, xAndYCells, diffIntersectionUnion, xOrYCells)
		}
	}
}

// cellUnionDistanceFromAxis returns the maximum geodesic distance from axis to any point of
// the given CellUnion.
func cellUnionDistanceFromAxis(cu CellUnion, axis Point) float64 {
	var maxDist float64
	for _, cid := range cu {
		cell := CellFromCellID(cid)
		for j := 0; j < 4; j++ {
			a := cell.Vertex(j)
			b := cell.Vertex((j + 1) & 3)
			var dist float64
			// The maximum distance is not always attained at a cell vertex: if at
			// least one vertex is in the opposite hemisphere from axis then the
			// maximum may be attained along an edge. We solve this by computing
			// the minimum distance from the edge to (-axis) instead. We can't
			// simply do this all the time because DistanceFromSegment() has
			// poor accuracy when the result is close to Pi.
			//
			// TODO: Improve edgeutil's DistanceFromSegment accuracy near Pi.
			if a.Angle(axis.Vector) > math.Pi/2 || b.Angle(axis.Vector) > math.Pi/2 {
				dist = math.Pi - float64(DistanceFromSegment(Point{axis.Mul(-1)}, a, b).Radians())
			} else {
				dist = float64(a.Angle(axis.Vector))
			}
			maxDist = math.Max(maxDist, dist)
		}
	}
	return maxDist
}

func TestCellUnionExpand(t *testing.T) {
	// This test generates coverings for caps of random sizes, expands
	// the coverings by a random radius, and then make sure that the new
	// covering covers the expanded cap.  It also makes sure that the
	// new covering is not too much larger than expected.
	for i := 0; i < 5000; i++ {
		rndCap := randomCap(AvgAreaMetric.Value(maxLevel), 4*math.Pi)

		// Expand the cap area by a random factor whose log is uniformly
		// distributed between 0 and log(1e2).
		expandedCap := CapFromCenterHeight(
			rndCap.center, math.Min(2.0, math.Pow(1e2, randomFloat64())*rndCap.Height()))

		radius := (expandedCap.Radius() - rndCap.Radius()).Radians()
		maxLevelDiff := randomUniformInt(8)

		// Generate a covering for the original cap, and measure the maximum
		// distance from the cap center to any point in the covering.
		coverer := &RegionCoverer{
			MaxLevel: maxLevel,
			MaxCells: 1 + skewedInt(10),
			LevelMod: 1,
		}
		covering := coverer.CellUnion(rndCap)
		checkCellUnionCovering(t, rndCap, covering, true, 0)
		coveringRadius := cellUnionDistanceFromAxis(covering, rndCap.center)

		// This code duplicates the logic in Expand(min_radius, max_level_diff)
		// that figures out an appropriate cell level to use for the expansion.
		minLevel := maxLevel
		for _, cid := range covering {
			minLevel = minInt(minLevel, cid.Level())
		}
		expandLevel := minInt(minLevel+maxLevelDiff, MinWidthMetric.MaxLevel(radius))

		// Generate a covering for the expanded cap, and measure the new maximum
		// distance from the cap center to any point in the covering.
		covering.ExpandByRadius(s1.Angle(radius), maxLevelDiff)
		checkCellUnionCovering(t, expandedCap, covering, false, 0)
		expandedCoveringRadius := cellUnionDistanceFromAxis(covering, rndCap.center)

		// If the covering includes a tiny cell along the boundary, in theory the
		// maximum angle of the covering from the cap center can increase by up to
		// twice the maximum length of a cell diagonal.
		if expandedCoveringRadius-coveringRadius >= 2*MaxDiagMetric.Value(expandLevel) {
			t.Errorf("covering.ExpandByRadius(%v, %v) distance from center = %v want < %v", radius, maxLevelDiff, expandedCoveringRadius-coveringRadius, 2*MaxDiagMetric.Value(expandLevel))
		}
	}
}

// checkCellUnionCovering checks that the given covering completely covers the given region.
// If checkTight is true, it also checks that it does not contain any cells that do not
// intersect the given region. The id is the CellID to start at for the checks. If an
// invalid value is used as the ID, then all faces are checked.
func checkCellUnionCovering(t *testing.T, r Region, covering CellUnion, checkTight bool, id CellID) {
	if !id.IsValid() {
		for face := 0; face < 6; face++ {
			checkCellUnionCovering(t, r, covering, checkTight, CellIDFromFace(face))
		}
		return
	}

	if !r.IntersectsCell(CellFromCellID(id)) {
		// If region does not intersect the id, then neither should the covering.
		if checkTight {
			if covering.IntersectsCellID(id) {
				t.Errorf("%v.IntersectsCellID(%v) = true, want false", covering, id)
			}
		}
		return
	}
	if !covering.ContainsCellID(id) {
		// The region may intersect id, but we can't assert that the covering
		// intersects id because we may discover that the region does not actually
		// intersect upon further subdivision.  (IntersectsCell is not exact.)
		if r.ContainsCell(CellFromCellID(id)) {
			t.Errorf("%v.ContainsCell(%v) = true, want false", r, id)
			return
		}
		if id.IsLeaf() {
			t.Errorf("%v.IsLeaf() = true, want false", id)
			return
		}
		for child := id.ChildBegin(); child != id.ChildEnd(); child = child.Next() {
			checkCellUnionCovering(t, r, covering, checkTight, child)
		}
	}
}

func TestCellUnionEmpty(t *testing.T) {
	var empty CellUnion

	// Normalize()
	empty.Normalize()
	if len(empty) != 0 {
		t.Errorf("len(empty.Normalize()) = %d, want 0", len(empty))
	}

	// Denormalize(...)
	empty.Denormalize(0, 2)
	if len(empty) != 0 {
		t.Errorf("len(empty.Denormalize(0, 2)) = %d, want 0", len(empty))
	}

	face1ID := CellIDFromFace(1)
	// Contains(...)
	if empty.ContainsCellID(face1ID) {
		t.Errorf("empty.ContainsCellID(%v) = true, want false", face1ID)
	}
	if !empty.Contains(empty) {
		t.Errorf("empty.Contains(%v) = false, want true", empty)
	}

	// Intersects(...)
	if empty.IntersectsCellID(face1ID) {
		t.Errorf("empty.IntersectsCellID(%v) = true, want false", face1ID)
	}
	if empty.Intersects(empty) {
		t.Errorf("empty.Intersects(%v) = true, want false", empty)
	}

	// Union(...)
	cellUnion := CellUnionFromUnion(empty, empty)
	if len(cellUnion) != 0 {
		t.Errorf("CellUnionFromUnion(empty, empty) has %v cells, want 0", len(cellUnion))
	}

	// Intersection(...)
	intersection := CellUnionFromIntersectionWithCellID(empty, face1ID)
	if len(intersection) != 0 {
		t.Errorf("CellUnionFromIntersectionWithCellID(%v, %v) = %v cells, want 0", empty, face1ID, len(intersection))
	}

	intersection = CellUnionFromIntersection(empty, empty)
	if len(intersection) != 0 {
		t.Errorf("CellUnionFromIntersection(%v, %v) = %v cells, want 0", empty, empty, len(intersection))
	}

	// Difference(...)
	difference := CellUnionFromDifference(empty, empty)
	if len(difference) != 0 {
		t.Errorf("CellUnionFromDifference(%v, %v) = %v cells, want 0", empty, empty, len(difference))
	}

	// Expand(...)
	empty.ExpandByRadius(s1.Angle(1), 20)
	if len(empty) != 0 {
		t.Errorf("empty.ExpandByRadius(1, 20) = %v cells, want 0", len(empty))
	}

	empty.ExpandAtLevel(10)
	if len(empty) != 0 {
		t.Errorf("empty.ExpandAtLevel(10) = %v cells, want 0", len(empty))
	}
}

func BenchmarkCellUnionFromRange(b *testing.B) {
	x := CellIDFromFace(0).ChildBeginAtLevel(maxLevel)
	y := CellIDFromFace(5).ChildEndAtLevel(maxLevel)
	for i := 0; i < b.N; i++ {
		CellUnionFromRange(x, y)
	}
}
