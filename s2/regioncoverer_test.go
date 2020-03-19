// Copyright 2015 Google Inc. All rights reserved.
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
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func TestCovererRandomCells(t *testing.T) {
	rc := &RegionCoverer{MinLevel: 0, MaxLevel: 30, LevelMod: 1, MaxCells: 1}

	// Test random cell ids at all levels.
	for i := 0; i < 10000; i++ {
		id := randomCellID()
		covering := rc.Covering(Region(CellFromCellID(id)))
		if len(covering) != 1 {
			t.Errorf("Iteration %d, cell ID token %s, got covering size = %d, want covering size = 1", i, id.ToToken(), len(covering))
			// if the covering isn't len 1, the next check will panic
			break
		}
		if (covering)[0] != id {
			t.Errorf("Iteration %d, cell ID token %s, got covering = %v, want covering = %v", i, id.ToToken(), covering, id)
		}
	}
}

// checkCovering reports whether covering is a valid cover for the region.
func checkCovering(t *testing.T, rc *RegionCoverer, r Region, covering CellUnion, interior bool) {
	// Keep track of how many cells have the same rc.MinLevel ancestor.
	minLevelCells := map[CellID]int{}
	var tempCover CellUnion
	for _, ci := range covering {
		level := ci.Level()
		if level < rc.MinLevel {
			t.Errorf("CellID(%s).Level() = %d, want >= %d", ci.ToToken(), level, rc.MinLevel)
		}
		if level > rc.MaxLevel {
			t.Errorf("CellID(%s).Level() = %d, want <= %d", ci.ToToken(), level, rc.MaxLevel)
		}
		if rem := (level - rc.MinLevel) % rc.LevelMod; rem != 0 {
			t.Errorf("(CellID(%s).Level() - MinLevel) mod LevelMod = %d, want = %d", ci.ToToken(), rem, 0)
		}
		tempCover = append(tempCover, ci)
		minLevelCells[ci.Parent(rc.MinLevel)]++
	}
	if len(covering) > rc.MaxCells {
		// If the covering has more than the requested number of cells, then check
		// that the cell count cannot be reduced by using the parent of some cell.
		for ci, count := range minLevelCells {
			if count > 1 {
				t.Errorf("Min level CellID %s, count = %d, want = %d", ci.ToToken(), count, 1)
			}
		}
	}
	if interior {
		for _, ci := range covering {
			if !r.ContainsCell(CellFromCellID(ci)) {
				t.Errorf("Region(%v).ContainsCell(%v) = %t, want = %t", r, CellFromCellID(ci), false, true)
			}
		}
	} else {
		tempCover.Normalize()
		checkCoveringTight(t, r, tempCover, true, 0)
	}
}

// checkCoveringTight checks that "cover" completely covers the given region.
// If "checkTight" is true, also checks that it does not contain any cells that
// do not intersect the given region. ("id" is only used internally.)
func checkCoveringTight(t *testing.T, r Region, cover CellUnion, checkTight bool, id CellID) {
	if !id.IsValid() {
		for f := 0; f < 6; f++ {
			checkCoveringTight(t, r, cover, checkTight, CellIDFromFace(f))
		}
		return
	}

	if !r.IntersectsCell(CellFromCellID(id)) {
		// If region does not intersect id, then neither should the covering.
		if got := cover.IntersectsCellID(id); checkTight && got {
			t.Errorf("CellUnion(%v).IntersectsCellID(%s) = %t; want = %t", cover, id.ToToken(), got, false)
		}
	} else if !cover.ContainsCellID(id) {
		// The region may intersect id, but we can't assert that the covering
		// intersects id because we may discover that the region does not actually
		// intersect upon further subdivision.  (IntersectsCell is not exact.)
		if got := r.ContainsCell(CellFromCellID(id)); got {
			t.Errorf("Region(%v).ContainsCell(%v) = %t; want = %t", r, CellFromCellID(id), got, false)
		}
		if got := id.IsLeaf(); got {
			t.Errorf("CellID(%s).IsLeaf() = %t; want = %t", id.ToToken(), got, false)
		}

		for child := id.ChildBegin(); child != id.ChildEnd(); child = child.Next() {
			checkCoveringTight(t, r, cover, checkTight, child)
		}
	}
}

func TestCovererRandomCaps(t *testing.T) {
	rc := &RegionCoverer{MinLevel: 0, MaxLevel: 30, LevelMod: 1, MaxCells: 1}
	for i := 0; i < 1000; i++ {
		rc.MinLevel = int(rand.Int31n(maxLevel + 1))
		rc.MaxLevel = int(rand.Int31n(maxLevel + 1))
		for rc.MinLevel > rc.MaxLevel {
			rc.MinLevel = int(rand.Int31n(maxLevel + 1))
			rc.MaxLevel = int(rand.Int31n(maxLevel + 1))
		}
		rc.LevelMod = int(1 + rand.Int31n(3))
		rc.MaxCells = skewedInt(10)

		maxArea := math.Min(4*math.Pi, float64(3*rc.MaxCells+1)*AvgAreaMetric.Value(rc.MinLevel))
		r := Region(randomCap(0.1*AvgAreaMetric.Value(maxLevel), maxArea))

		covering := rc.Covering(r)
		checkCovering(t, rc, r, covering, false)
		interior := rc.InteriorCovering(r)
		checkCovering(t, rc, r, interior, true)

		// Check that Covering is deterministic.
		covering2 := rc.Covering(r)
		if !reflect.DeepEqual(covering, covering2) {
			t.Errorf("Iteration %d, got covering = %v, want covering = %v", i, covering2, covering)
		}

		// Also check Denormalize. The denormalized covering
		// may still be different and smaller than "covering" because
		// s2.RegionCoverer does not guarantee that it will not output all four
		// children of the same parent.
		covering.Denormalize(rc.MinLevel, rc.LevelMod)
		checkCovering(t, rc, r, covering, false)
	}
}

func TestRegionCovererInteriorCovering(t *testing.T) {
	// We construct the region the following way. Start with Cell of level l.
	// Remove from it one of its grandchildren (level l+2). If we then set
	//   minLevel < l + 1
	//   maxLevel > l + 2
	//   maxCells = 3
	// the best interior covering should contain 3 children of the initial cell,
	// that were not effected by removal of a grandchild.
	const level = 12
	smallCell := cellIDFromPoint(randomPoint()).Parent(level + 2)
	largeCell := smallCell.Parent(level)

	smallCellUnion := CellUnion([]CellID{smallCell})
	largeCellUnion := CellUnion([]CellID{largeCell})
	diff := CellUnionFromDifference(largeCellUnion, smallCellUnion)

	coverer := &RegionCoverer{
		MaxCells: 3,
		MaxLevel: level + 3,
		MinLevel: level,
	}

	interior := coverer.InteriorCovering(&diff)
	if len(interior) != 3 {
		t.Fatalf("len(coverer.InteriorCovering(%v)) = %v, want 3", diff, len(interior))
	}
	for i := 0; i < 3; i++ {
		if got, want := interior[i].Level(), level+1; got != want {
			t.Errorf("interior[%d].Level() = %v, want %v", i, got, want)
		}
	}
}

func TestRegionCovererSimpleRegionCovering(t *testing.T) {
	const maxLevel = maxLevel
	for i := 0; i < 100; i++ {
		level := randomUniformInt(maxLevel + 1)
		maxArea := math.Min(4*math.Pi, 1000.0*AvgAreaMetric.Value(level))
		c := randomCap(0.1*AvgAreaMetric.Value(maxLevel), maxArea)
		covering := SimpleRegionCovering(c, c.Center(), level)
		rc := &RegionCoverer{MaxLevel: level, MinLevel: level, MaxCells: math.MaxInt32, LevelMod: 1}
		checkCovering(t, rc, c, covering, false)
	}
}

func TestRegionCovererIsCanonical(t *testing.T) {
	tests := []struct {
		cells []string
		cov   *RegionCoverer
		want  bool
	}{
		// InvalidCellID
		{cells: []string{"1/"}, cov: NewRegionCoverer(), want: true},
		{cells: []string{"invalid"}, cov: NewRegionCoverer(), want: false},
		// Unsorted
		{cells: []string{"1/1", "1/3"}, cov: NewRegionCoverer(), want: true},
		{cells: []string{"1/3", "1/1"}, cov: NewRegionCoverer(), want: false},

		// Overlapping
		{cells: []string{"1/2", "1/33"}, cov: NewRegionCoverer(), want: true},
		{cells: []string{"1/3", "1/33"}, cov: NewRegionCoverer(), want: false},

		// MinLevel
		{
			cells: []string{"1/31"},
			cov:   &RegionCoverer{MinLevel: 2, MaxLevel: 30, LevelMod: 1, MaxCells: 8},
			want:  true,
		},
		{
			cells: []string{"1/3"},
			cov:   &RegionCoverer{MinLevel: 2, MaxLevel: 30, LevelMod: 1, MaxCells: 8},
			want:  false,
		},

		// MaxLevel
		{
			cells: []string{"1/31"},
			cov:   &RegionCoverer{MinLevel: 0, MaxLevel: 2, LevelMod: 1, MaxCells: 8},
			want:  true,
		},
		{
			cells: []string{"1/312"},
			cov:   &RegionCoverer{MinLevel: 0, MaxLevel: 2, LevelMod: 1, MaxCells: 8},
			want:  false,
		},

		// LevelMod
		{
			cells: []string{"1/31"},
			cov:   &RegionCoverer{MinLevel: 0, MaxLevel: 30, LevelMod: 2, MaxCells: 8},
			want:  true,
		},
		{
			cells: []string{"1/312"},
			cov:   &RegionCoverer{MinLevel: 0, MaxLevel: 30, LevelMod: 2, MaxCells: 8},
			want:  false,
		},

		// MaxCells
		{cells: []string{"1/1", "1/3"}, cov: &RegionCoverer{MinLevel: 0, MaxLevel: 30, LevelMod: 1, MaxCells: 2}, want: true},
		{cells: []string{"1/1", "1/3", "2/"}, cov: &RegionCoverer{MinLevel: 0, MaxLevel: 30, LevelMod: 1, MaxCells: 2}, want: false},
		{cells: []string{"1/123", "2/1", "3/0122"}, cov: &RegionCoverer{MinLevel: 0, MaxLevel: 30, LevelMod: 1, MaxCells: 2}, want: true},

		// Normalized
		// Test that no sequence of cells could be replaced by an ancestor.
		{
			cells: []string{"1/01", "1/02", "1/03", "1/10", "1/11"},
			cov:   NewRegionCoverer(),
			want:  true,
		},
		{
			cells: []string{"1/00", "1/01", "1/02", "1/03", "1/10"},
			cov:   NewRegionCoverer(),
			want:  false,
		},

		{
			cells: []string{"0/22", "1/01", "1/02", "1/03", "1/10"},
			cov:   NewRegionCoverer(),
			want:  true,
		},
		{
			cells: []string{"0/22", "1/00", "1/01", "1/02", "1/03"},
			cov:   NewRegionCoverer(),
			want:  false,
		},

		{
			cells: []string{"1/1101", "1/1102", "1/1103", "1/1110", "1/1111", "1/1112",
				"1/1113", "1/1120", "1/1121", "1/1122", "1/1123", "1/1130",
				"1/1131", "1/1132", "1/1133", "1/1200"},
			cov:  &RegionCoverer{MinLevel: 0, MaxLevel: 30, LevelMod: 2, MaxCells: 20},
			want: true,
		},
		{
			cells: []string{"1/1100", "1/1101", "1/1102", "1/1103", "1/1110", "1/1111",
				"1/1112", "1/1113", "1/1120", "1/1121", "1/1122", "1/1123",
				"1/1130", "1/1131", "1/1132", "1/1133"},
			cov:  &RegionCoverer{MinLevel: 0, MaxLevel: 30, LevelMod: 2, MaxCells: 20},
			want: false,
		},
	}
	for _, test := range tests {
		cu := makeCellUnion(test.cells...)
		if got := test.cov.IsCanonical(cu); got != test.want {
			t.Errorf("IsCanonical(%v) = %t, want %t", cu, got, test.want)
		}
	}
}

const numCoveringBMRegions = 1000

func BenchmarkRegionCovererCoveringCap(b *testing.B) {
	benchmarkRegionCovererCovering(b, func(n int) string {
		return fmt.Sprintf("Cap%d", n)
	},
		func(n int) []Region {
			regions := make([]Region, numCoveringBMRegions)
			for i := 0; i < numCoveringBMRegions; i++ {
				regions[i] = randomCap(0.1*AvgAreaMetric.Value(maxLevel), 4*math.Pi)
			}
			return regions
		})
}

func BenchmarkRegionCovererCoveringCell(b *testing.B) {
	benchmarkRegionCovererCovering(b, func(n int) string {
		return fmt.Sprintf("Cell%d", n)
	},
		func(n int) []Region {
			regions := make([]Region, numCoveringBMRegions)
			for i := 0; i < numCoveringBMRegions; i++ {
				regions[i] = CellFromCellID(randomCellIDForLevel(maxLevel - randomUniformInt(n)))
			}
			return regions
		})
}

func BenchmarkRegionCovererCoveringLoop(b *testing.B) {
	benchmarkRegionCovererCovering(b, func(n int) string {
		return fmt.Sprintf("Loop-%d-edges", int(math.Pow(2.0, float64(n))))
	},
		func(n int) []Region {
			size := int(math.Pow(2.0, float64(n)))
			regions := make([]Region, numCoveringBMRegions)
			for i := 0; i < numCoveringBMRegions; i++ {
				regions[i] = RegularLoop(randomPoint(), kmToAngle(10.0), size)
			}
			return regions
		})
}

func BenchmarkRegionCovererCoveringCellUnion(b *testing.B) {
	benchmarkRegionCovererCovering(b, func(n int) string {
		return fmt.Sprintf("CellUnion-%d-cells", int(math.Pow(2.0, float64(n))))
	},
		func(n int) []Region {
			size := int(math.Pow(2.0, float64(n)))
			regions := make([]Region, numCoveringBMRegions)
			for i := 0; i < numCoveringBMRegions; i++ {
				cu := randomCellUnion(size)
				regions[i] = &cu
			}
			return regions
		})
}

// TODO(roberts): Add more benchmarking that changes the values in the coverer (min/max level, # cells).

// benchmark Covering using the supplied func to generate a slice of random Regions of
// the given type to choose from for the benchmark.
//
// e.g. Loops with [4..~2^n] edges, CellUnions of 2^n random Cells, random Cells and Caps
func benchmarkRegionCovererCovering(b *testing.B, genLabel func(n int) string, genRegions func(n int) []Region) {
	rc := &RegionCoverer{MinLevel: 0, MaxLevel: 30, LevelMod: 1, MaxCells: 8}

	// Range over a variety of region complexities.
	for n := 2; n <= 16; n++ {
		b.Run(genLabel(n),
			func(b *testing.B) {
				b.StopTimer()
				regions := genRegions(n)
				l := len(regions)
				b.StartTimer()
				for i := 0; i < b.N; i++ {
					rc.Covering(regions[i%l])
				}
			})
	}
}

// TODO(roberts): Differences from C++
//  func TestRegionCovererAccuracy(t *testing.T) {
//  func TestRegionCovererFastCoveringHugeFixedLevelCovering(t *testing.T) {
//  func TestRegionCovererCanonicalizeCoveringUnsortedDuplicateCells(t *testing.T) {
//  func TestRegionCovererCanonicalizeCoveringMaxLevelExceeded(t *testing.T) {
//  func TestRegionCovererCanonicalizeCoveringWrongLevelMod(t *testing.T) {
//  func TestRegionCovererCanonicalizeCoveringReplacedByParent(t *testing.T) {
//  func TestRegionCovererCanonicalizeCoveringDenormalizedCellUnion(t *testing.T) {
//  func TestRegionCovererCanonicalizeCoveringMaxCellsMergesSmallest(t *testing.T) {
//  func TestRegionCovererCanonicalizeCoveringMaxCellsMergesRepeatedly(t *testing.T) {
