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

// Package s2intersect efficiently finds common areas shared by sets of S2
// CellUnions. Use this package instead of s2.CellUnionFromIntersection when
// there are 3 or more CellUnions to compare and all intersections are required,
// as the brute force alternative is exponential in time.
//
// Benchmarks show that direct comparison of a pair of CellUnions is more
// efficient with respect to both time and memory when using the standard
// s2.CellUnionFromIntersection function. The benefits of s2intersect arise as
// the number of potential intersections grows; for n CellUnions, there are
// 2^n - n - 1 possible intersections (power set excluding single-member sets
// and the empty set).
package s2intersect

import (
	"fmt"
	"sort"

	"github.com/golang/geo/geo/s2"
)

// Re nomenclature used throughout this file: an "intersection" is a 2D
// (spherical) concept in keeping with s2.CellUnionFromIntersection(). The 1D
// (Hilbert curve) equivalent is referred to as an "overlap", simply to
// differentiate between the two. An overlap interval is closed, and defined by
// RangeMin and RangeMax (i.e. leaf Cells) of the first and last CellIDs,
// respectively, in the intersection that it maps to—these are referred to as
// "limits". Wherever "indices" are referred to, these reference the slice of
// CellUnions passed in by a user of this package.

// An Intersection describes the intersection of a set of CellUnions.
//
// Intersections returned by Find() are disjoint; i.e. if the Indices of one are
// a sub set of the Indices of another, the sub-set Intersection will NOT
// include the CellIDs that are included in the super-set Intersection; for
// details, see Find() re disjoint Intersections.
type Intersection struct {
	// Indices refer to the slice of CellUnions passed to Find().
	Indices      []int
	Intersection s2.CellUnion
}

// Find returns the set of all disjoint intersections of the CellUnions. The
// ordering of the output is undefined.
//
// Whereas the naive approach of finding pairwise intersections is O(n^2) for n
// CellUnions, and a naive multi-way search is O(2^n), Find() is
// O(max(i log i, c)) for c Cells constituting i intervals on the Hilbert curve
// (contiguous Cells form a single interval).
//
// Find calls Normalize() on all CellUnions, the efficiency of which is not
// considered in the previous calculations.
//
// Note that returned Intersections are disjoint; Cells are only counted towards
// the most comprehensive Intersection. For example, the intersection of regions
// A and B will NOT include the area in which regions A, B, _and_ C intersect
// because {A,B,C} is a super set of {A,B}. Correcting this requires iterating
// over the power set of all resulting Intersections and "decomposing" them into
// constituent parts (power sets are O(2^n) for a set of size n).
//
// Consider the following CellUnions, shown as their corresponding intervals on
// the straightened Hilbert curve. Cells are denoted as = and grouped into 4 for
// ease of viewing (i.e. the pipes | have no meaning). The resulting
// Intersections are denoted with the letters X–Z and the disjoint nature of the
// output is such that Y and Z will not include the Cells that are instead
// reported in X, even though the respective CellUnions do indeed intersect
// there. Analogously, each area of the Venn diagram of regions can receive only
// one label.
//
//	0: |====|  ==|    |
//	1: |==  |   =|=== |
//	2: |====|====|  ==|
//	    XXYY   YX   Z
//
//	X: {0,1,2}
//	Y: {0,2}
//	Z: {1,2}
func Find(cus []s2.CellUnion) []Intersection {
	return overlapsToIntersections(cellUnionsToOverlaps(cus))
}

// An overlap describes a closed range on the Hilbert curve, throughout which,
// a set of CellUnions intersect.
type overlap struct {
	indices    []int
	start, end s2.CellID
}

func cellUnionsToOverlaps(cus []s2.CellUnion) []overlap {
	// We convert the 2-dimensional problem of finding intersections into a 1D
	// problem of overlapping intervals on a line (the Hilbert curve).
	var lims []*limit
	for i, cu := range cus {
		cu.Normalize()
		lims = append(lims, cellUnionToIntervalLimits(cu, i)...)
	}
	if len(lims) == 0 {
		return nil
	}

	lims = collapseLimits(lims)
	return intervalOverlaps(lims)
}

func intervalOverlaps(lims []*limit) []overlap {
	open := make(map[int]struct{})
	openIndices := func() []int {
		is := make([]int, 0, len(open))
		for i := range open {
			is = append(is, i)
		}
		sort.Ints(is)
		return is
	}

	var (
		lastStart s2.CellID
		overlaps  []overlap
	)

	for _, l := range lims {
		if len(open) > 1 {
			// The existence of a limit at this point constitutes a new overlap
			// regardless of the limit type. If it is an end limit then obviously the
			// overlap is ceasing, but if it's a start limit then the overlap until
			// here only applies to the already-open values but until the previous
			// leaf Cell.
			endLeaf := l.leaf
			if l.typ == start {
				endLeaf = endLeaf.Prev()
			}
			overlaps = append(overlaps, overlap{
				indices: openIndices(),
				start:   lastStart,
				end:     endLeaf,
			})
		}

		switch l.typ {
		case start:
			for _, i := range l.indices {
				open[i] = struct{}{}
			}
		case end:
			for _, i := range l.indices {
				delete(open, i)
			}
		}

		if len(open) > 1 {
			// As with the rationale at the beginning of the limit loop, we now have a
			// new interval regardless of the limit type. However, the starting leaf
			// is iterated if this is currently a close limit because the remaining
			// indices only overlap _after_ the current leaf by nature of the
			// intervals being closed.
			lastStart = l.leaf
			if l.typ == end {
				lastStart = lastStart.Next()
			}
		}
	}

	return overlaps
}

// A limit defines either the beginning or end of a closed interval on the
// Hilbert curve of leaf Cells.
type limit struct {
	leaf    s2.CellID
	typ     limitType // There are always pairs of {start,end}
	indices []int     // Indices within in the slice of CellUnions passed to Find().
}

type limitType bool

const (
	start limitType = false
	end   limitType = true // end is inclusive
)

// cellUnionToIntervalLimits converts all Cells into pairs of limits
// representing closed intervals on the Hilbert curve of leaf Cells. The value
// of idx must be the index of cu within the slice passed to Find().
func cellUnionToIntervalLimits(cu s2.CellUnion, idx int) []*limit {
	if len(cu) == 0 {
		return nil
	}

	var lims []*limit
	pushLeaf := func(cID s2.CellID, t limitType) {
		lims = append(lims, &limit{
			leaf:    cID,
			indices: []int{idx},
			typ:     t,
		})
	}

	// Intervals are defined by start and end leaf Cells. Instead of
	// converting every Cell into an interval, we merge contiguous ones by
	// checking if lastend.Next() is equal to the next Cell's start leaf.
	var lastend s2.CellID

	for _, cID := range cu {
		switch startLeaf := cID.RangeMin(); {
		case lastend == 0:
			// Current Cell represents the first interval
			pushLeaf(startLeaf, start)
		case lastend.Next() != startLeaf:
			// Current and last Cells are non-contiguous
			pushLeaf(lastend, end)
			pushLeaf(startLeaf, start)
		}
		lastend = cID.RangeMax()
	}

	pushLeaf(lastend, end)
	return lims
}

// collapseLimits returns a slice of limits such that all of those of the same
// type at the same leaf CellID are grouped with their indices in the same
// slice. The returned slice will be sorted by CellID and tyoe,
//
// e.g. {leaf(x),start,indices[2]} and {leaf(x),start,indices[7]} will become
// {leaf(x),start,indices[2,7]}.
//
// The original slice and its contents will be modified, but not in place;
// therefore only the returned value is valid.
func collapseLimits(lims []*limit) []*limit {
	if len(lims) == 0 {
		return nil
	}

	// Limits at the same leaf MUST be sorted such that start-types come before
	// end-types. This is because they are closed intervals so we can't have
	// ending intervals removed before the overlap is accounted for.
	sort.Slice(lims, func(i, j int) bool {
		lI, lJ := lims[i], lims[j]
		if lI.leaf != lJ.leaf {
			return lI.leaf < lJ.leaf
		}
		return lI.typ == start
	})

	out := []*limit{lims[0]}
	last := out[0]
	for _, l := range lims[1:] {
		if l.leaf == last.leaf && l.typ == last.typ {
			last.indices = append(last.indices, l.indices...)
			continue
		}
		sort.Ints(last.indices)
		last = l
		out = append(out, l)
	}
	return out
}

// indexKey converts a slice of sorted ints into a string for use as a map key.
func indexKey(is []int) string {
	return fmt.Sprintf("%d", is)
}

func overlapsToIntersections(overlaps []overlap) []Intersection {
	set := make(map[string]*Intersection)

	for _, o := range overlaps {
		k := indexKey(o.indices)
		i, ok := set[k]
		if !ok {
			i = &Intersection{Indices: o.indices}
			set[k] = i
		}
		// CellUnionFromRange assumes a half-open interval, so use Next() as
		// suggested in its comment.
		i.Intersection = append(i.Intersection, s2.CellUnionFromRange(o.start, o.end.Next())...)
	}

	result := make([]Intersection, 0, len(set))
	for _, i := range set {
		i.Intersection.Normalize()
		result = append(result, *i)
	}
	return result
}
