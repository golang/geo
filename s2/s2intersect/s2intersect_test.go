// Copyright 2023 Google Inc. All rights reserved.
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

package s2intersect

import (
	"sort"
	"testing"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// capCoverings returns coverings for a set of overlapping s2.Caps.
func capCoverings(rc *s2.RegionCoverer) []s2.CellUnion {
	var cus []s2.CellUnion
	for _, centre := range []s2.LatLng{
		s2.LatLngFromDegrees(-26.210955, 28.041440),
		s2.LatLngFromDegrees(-26.257569, 28.459799),
		s2.LatLngFromDegrees(-26.297932, 28.323583),
		s2.LatLngFromDegrees(-26.157067, 28.370976),
	} {
		cap := s2.CapFromCenterAngle(s2.PointFromLatLng(centre), s1.Degree/3)
		cus = append(cus, rc.Covering(cap))
	}
	return cus
}

func BenchmarkPair(b *testing.B) {
	rc := &s2.RegionCoverer{
		MaxLevel: 14,
		MaxCells: 100,
	}
	cus := capCoverings(rc)[:2]

	b.Run("Find", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Find(cus)
		}
	})

	b.Run("CellUnionFromIntersection", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s2.CellUnionFromIntersection(cus[0], cus[1])
		}
	})
}

func TestFind(t *testing.T) {
	// Create Caps from arbitrary centres, cover them, and then compare
	// intersections returned by Find() to those from a brute force method using
	// s2.CellUnionFromIntersection(). Although this adds logic to a test, it's
	// still clearer than looking at arbitrary lists of CellIDs.
	rc := &s2.RegionCoverer{
		MaxLevel: 14,
		MaxCells: 100,
	}
	testInput := capCoverings(rc)

	// Manual means of computing an intersection from the built-in
	// s2.CellUnionFromIntersection() function.
	intersection := func(indices []int) s2.CellUnion {
		if len(indices) == 0 {
			return nil
		}
		result := testInput[indices[0]]
		for _, idx := range indices[1:] {
			result = s2.CellUnionFromIntersection(result, testInput[idx])
		}
		return result
	}
	// Brute-force calculation of the Intersections requires subtraction of those
	// Cells assigned to super-set Intersections. See Find() re mutually exclusive
	// output.
	bruteForce := func(indices []int, subtract [][]int) Intersection {
		i := intersection(indices)
		for _, s := range subtract {
			i = s2.CellUnionFromDifference(i, intersection(s))
		}
		return Intersection{
			Indices:      indices,
			Intersection: i,
		}
	}
	got := Find(testInput)
	want := []Intersection{
		// Iterate over the elements of the power set of {0,1,2,3} that have >=2
		// elements in them.
		//
		// For each, the []int slice is the actual set of indices we are expecting,
		// whereas the 2d [][]int slice captures the other intersections that
		// "steal" from the value because they have a super set of indices.
		bruteForce([]int{0, 1}, [][]int{{0, 1, 2}, {0, 1, 3}, {0, 1, 2, 3}}),
		bruteForce([]int{0, 2}, [][]int{{0, 1, 2}, {0, 2, 3}, {0, 1, 2, 3}}),
		bruteForce([]int{0, 3}, [][]int{{0, 1, 3}, {0, 2, 3}, {0, 1, 2, 3}}),
		bruteForce([]int{1, 2}, [][]int{{0, 1, 2}, {1, 2, 3}, {0, 1, 2, 3}}),
		bruteForce([]int{1, 3}, [][]int{{0, 1, 3}, {1, 2, 3}, {0, 1, 2, 3}}),
		bruteForce([]int{2, 3}, [][]int{{0, 2, 3}, {1, 2, 3}, {0, 1, 2, 3}}),
		bruteForce([]int{0, 1, 2}, [][]int{{0, 1, 2, 3}}),
		bruteForce([]int{0, 1, 3}, [][]int{{0, 1, 2, 3}}),
		bruteForce([]int{0, 2, 3}, [][]int{{0, 1, 2, 3}}),
		bruteForce([]int{1, 2, 3}, [][]int{{0, 1, 2, 3}}),
		bruteForce([]int{0, 1, 2, 3}, nil),
	}

	// Some of the manual calculation may have resulted in empty intersecting
	// CellUnions (i.e. not actually an Intersection). This is merely because
	// we're iterating over every single option by hand.
	var wantNonEmpty []Intersection
	for _, w := range want {
		if len(w.Intersection) > 0 {
			wantNonEmpty = append(wantNonEmpty, w)
		}
	}

	sortIntersections(t, got)
	sortIntersections(t, wantNonEmpty)

	if diff := cmp.Diff(wantNonEmpty, got); diff != "" {
		t.Errorf("Find([four caps; see code] diff (-want +got):\n%s)", diff)
	}
}

func TestFindLeaves(t *testing.T) {
	leaf := func(idx int64) s2.CellID {
		return sydney.RangeMin().Advance(idx)
	}

	cus := []s2.CellUnion{
		/*
		 * The cell leaf(0) is denoted with #.
		 * Note that this is constructed such that (a) [1] starts where [2] ends,
		 * and ends where [0] begins again; and (b) it includes intervals of length
		 * 1. Both scenarios failed prior to the changes made along with this test.
		 *
		 *          #
		 * 0: |====| = =|
		 * 1: |  ==|== =|
		 * 2: | == |   =|
		 *      ABC  C B
		 */
		{
			leaf(-4), leaf(-3), leaf(-2), leaf(-1),
			leaf(1), leaf(3),
		},
		{
			leaf(-2), leaf(-1), leaf(0), leaf(1), leaf(3),
		},
		{
			leaf(-3), leaf(-2), leaf(3),
		},
	}

	got := Find(cus)
	want := []Intersection{
		// A
		{
			Indices:      []int{0, 2},
			Intersection: s2.CellUnion{leaf(-3)},
		},
		// B
		{
			Indices:      []int{0, 1, 2},
			Intersection: s2.CellUnion{leaf(-2), leaf(3)},
		},
		// C
		{
			Indices:      []int{0, 1},
			Intersection: s2.CellUnion{leaf(-1), leaf(1)},
		},
	}

	sortIntersections(t, got)
	sortIntersections(t, want)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Find([overlaps with leaf Cells; see code] diff (-want +got):\n%s)", diff)
	}
}

// sorts lexicographically by Indices slice.
func sortIntersections(t *testing.T, ints []Intersection) {
	t.Helper()

	sort.Slice(ints, func(i, j int) bool {
		iI, iJ := ints[i], ints[j]
		if nI, nJ := len(iI.Indices), len(iJ.Indices); nI != nJ {
			return nI < nJ
		}

		for k, idx := range iI.Indices {
			if idx != iJ.Indices[k] {
				return idx < iJ.Indices[k]
			}
		}

		t.Fatalf("Sorting %T found two elements with identical Indices %d", ints, iI.Indices)
		return true
	})
}

// Improves error output.
func (t limitType) String() string {
	if t == start {
		return "start"
	}
	return "end"
}

func TestEmptyOutput(t *testing.T) {
	tests := []struct {
		name string
		cus  []s2.CellUnion
	}{
		{
			name: "nil input",
			cus:  nil,
		},
		{
			name: "non-nil but empty input",
			cus:  []s2.CellUnion{},
		},
		{
			name: "single CellUnion input",
			cus:  []s2.CellUnion{{sydney}},
		},
		{
			name: "disjoint CellUnion input",
			cus:  []s2.CellUnion{{sydney.Prev()}, {sydney.Next()}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Find(tt.cus); len(got) != 0 {
				t.Errorf("Find(%v) got %+v; want empty slice", tt.cus, got)
			}
		})
	}
}

const sydney = s2.CellID(0x6b12b00000000000)

func TestCellUnionToIntervalLimits(t *testing.T) {
	tests := []struct {
		name string
		cu   s2.CellUnion
		want []*limit // Note that limit.indices are ignored
	}{
		{
			name: "empty",
			want: nil,
		},
		{
			name: "single Cell",
			cu:   s2.CellUnion{sydney},
			want: []*limit{
				{
					leaf: sydney.RangeMin(),
					typ:  start,
				},
				{
					leaf: sydney.RangeMax(),
					typ:  end,
				},
			},
		},
		{
			name: "two contiguous Cells are merged into single interval",
			cu:   s2.CellUnion{sydney, sydney.Next()},
			want: []*limit{
				{
					leaf: sydney.RangeMin(),
					typ:  start,
				},
				{
					leaf: sydney.Next().RangeMax(),
					typ:  end,
				},
			},
		},
		{
			name: "three contiguous Cells are merged into single interval",
			cu:   s2.CellUnion{sydney.Prev(), sydney, sydney.Next()},
			want: []*limit{
				{
					leaf: sydney.Prev().RangeMin(),
					typ:  start,
				},
				{
					leaf: sydney.Next().RangeMax(),
					typ:  end,
				},
			},
		},
		{
			name: "two contiguous Cells at different levels are merged into single interval",
			cu:   s2.CellUnion{sydney, sydney.Next().ChildBegin()},
			want: []*limit{
				{
					leaf: sydney.RangeMin(),
					typ:  start,
				},
				{
					leaf: sydney.Next().ChildBegin().RangeMax(),
					typ:  end,
				},
			},
		},
		{
			name: "non-contiguous Cells are separate intervals",
			cu:   s2.CellUnion{sydney.Prev(), sydney.Next()},
			want: []*limit{
				{
					leaf: sydney.Prev().RangeMin(),
					typ:  start,
				},
				{
					leaf: sydney.Prev().RangeMax(),
					typ:  end,
				},
				{
					leaf: sydney.Next().RangeMin(),
					typ:  start,
				},
				{
					leaf: sydney.Next().RangeMax(),
					typ:  end,
				},
			},
		},
		{
			name: "mix of contiguous and non-contiguous Cells",
			cu: s2.CellUnion{
				// Those expected to be grouped into a single interval are on the same line
				sydney.Prev().Prev(),                                             // A
				sydney, sydney.Next().Children()[0], sydney.Next().Children()[1], // B
				// Note the gap that excludes Children()[2]
				sydney.Next().Children()[3], sydney.Next().Next(), // C
			},
			want: []*limit{
				// A
				{
					leaf: sydney.Prev().Prev().RangeMin(),
					typ:  start,
				},
				{
					leaf: sydney.Prev().Prev().RangeMax(),
					typ:  end,
				},
				// B
				{
					leaf: sydney.RangeMin(),
					typ:  start,
				},
				{
					leaf: sydney.Next().Children()[1].RangeMax(),
					typ:  end,
				},
				// C
				{
					leaf: sydney.Next().Children()[3].RangeMin(),
					typ:  start,
				},
				{
					leaf: sydney.Next().Next().RangeMax(),
					typ:  end,
				},
			},
		},
	}

	opts := []cmp.Option{
		cmp.AllowUnexported(limit{}),
		cmpopts.IgnoreFields(limit{}, "indices"),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cellUnionToIntervalLimits(tt.cu, 0)
			if diff := cmp.Diff(tt.want, got, opts...); diff != "" {
				t.Errorf("cellUnionToIntervalLimits(%v) diff (-want +got):\n%s", tt.cu, diff)
			}
		})
	}
}

func TestCellUnionsToOverlaps(t *testing.T) {
	tests := []struct {
		name string
		cus  []s2.CellUnion
		want []overlap
	}{
		{
			name: "same Cell",
			cus: []s2.CellUnion{
				{sydney},
				{sydney},
			},
			want: []overlap{
				{
					indices: []int{0, 1},
					start:   sydney.RangeMin(),
					end:     sydney.RangeMax(),
				},
			},
		},
		{
			name: "same 2 Cells",
			cus: []s2.CellUnion{
				{sydney, sydney.Next()},
				{sydney, sydney.Next()},
			},
			want: []overlap{
				{
					indices: []int{0, 1},
					start:   sydney.RangeMin(),
					end:     sydney.Next().RangeMax(),
				},
			},
		},
		{
			name: "partial overlap between two CellUnions",
			cus: []s2.CellUnion{
				{sydney},
				{sydney.Prev(), sydney, sydney.Next()},
			},
			want: []overlap{
				{
					indices: []int{0, 1},
					start:   sydney.RangeMin(),
					end:     sydney.RangeMax(),
				},
			},
		},
		{
			name: "arbitrary mix (see code comments for diagram)",
			/*
			 * 0: |====|  ==|    |
			 * 1: |==  |   =|=== |
			 * 2: |====|====|  ==|
			 *     AABB   CD   E
			 */
			cus: []s2.CellUnion{
				{
					sydney.Prev(),
					sydney.Children()[2], sydney.Children()[3],
				},
				{
					sydney.Prev().Children()[0], sydney.Prev().Children()[1],
					sydney.Children()[3],
					sydney.Next().Children()[0], sydney.Next().Children()[1], sydney.Next().Children()[2],
				},
				{
					sydney.Prev(),
					sydney,
					sydney.Next().Children()[2], sydney.Next().Children()[3],
				},
			},
			want: []overlap{
				// A
				{
					indices: []int{0, 1, 2},
					start:   sydney.Prev().RangeMin(),
					end:     sydney.Prev().Children()[1].RangeMax(),
				},
				// B
				{
					indices: []int{0, 2},
					start:   sydney.Prev().Children()[2].RangeMin(),
					end:     sydney.Prev().Children()[3].RangeMax(),
				},
				// C
				{
					indices: []int{0, 2},
					start:   sydney.Children()[2].RangeMin(),
					end:     sydney.Children()[2].RangeMax(),
				},
				// D
				{
					indices: []int{0, 1, 2},
					start:   sydney.Children()[3].RangeMin(),
					end:     sydney.Children()[3].RangeMax(),
				},
				// E
				{
					indices: []int{1, 2},
					start:   sydney.Next().Children()[2].RangeMin(),
					end:     sydney.Next().Children()[2].RangeMax(),
				},
			},
		},
	}

	opts := []cmp.Option{
		cmp.AllowUnexported(overlap{}),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cellUnionsToOverlaps(tt.cus)
			if diff := cmp.Diff(tt.want, got, opts...); diff != "" {
				t.Errorf("cellUnionsToOverlaps(%v) diff (-want +got):\n%s", tt.cus, diff)
			}
		})
	}
}
