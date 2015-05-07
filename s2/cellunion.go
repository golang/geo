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

import "sort"

// A CellUnion is a collection of CellIDs.
//
// It is normalized if it is sorted, and does not contain redundancy.
// Specifically, it may not contain the same CellID twice, nor a CellID that is contained by another,
// nor the four sibling CellIDs that are children of a single higher level CellID.
type CellUnion []CellID

func CellUnionFromCellIDs(ids []CellID) *CellUnion {
	union := &CellUnion{}
	*union = append(*union, ids...)
	union.Normalize()
	return union
}

func CellUnionFromArrayAndSwap(ids *[]CellID) *CellUnion {
	union := &CellUnion{}
	*union = append(*union, *ids...)
	union.Normalize()
	*ids = []CellID{}
	return union
}

func (cu *CellUnion) DeNormalize(minLevel, levelMod int, output *[]CellID) {
	*output = make([]CellID, 0, len(*cu))
	for _, ci := range *cu {
		level := ci.Level()
		newLevel := max(minLevel, level)
		if levelMod > 1 {
			// Round up so that (new_level - min_level) is a multiple of level_mod.
			// (Note that S2CellId::kMaxLevel is a multiple of 1, 2, and 3.)
			newLevel += (MAX_LEVEL - (newLevel - minLevel)) % levelMod
			newLevel = min(MAX_LEVEL, newLevel)
		}
		if newLevel == level {
			*output = append(*output, ci)
		} else {
			end := ci.ChildEndAtLevel(newLevel)
			for id := ci.ChildBeginAtLevel(newLevel); id != end; id = id.Next() {
				*output = append(*output, id)
			}
		}
	}
}

// Normalize normalizes the CellUnion.
func (cu *CellUnion) Normalize() {
	sort.Sort(byID(*cu))

	output := make([]CellID, 0, len(*cu)) // the list of accepted cells
	// Loop invariant: output is a sorted list of cells with no redundancy.
	for _, ci := range *cu {
		// The first two passes here either ignore this new candidate,
		// or remove previously accepted cells that are covered by this candidate.

		// Ignore this cell if it is contained by the previous one.
		// We only need to check the last accepted cell. The ordering of the
		// cells implies containment (but not the converse), and output has no redundancy,
		// so if this candidate is not contained by the last accepted cell
		// then it cannot be contained by any previously accepted cell.
		if len(output) > 0 && output[len(output)-1].Contains(ci) {
			continue
		}

		// Discard any previously accepted cells contained by this one.
		// This could be any contiguous trailing subsequence, but it can't be
		// a discontiguous subsequence because of the containment property of
		// sorted S2 cells mentioned above.
		j := len(output) - 1 // last index to keep
		for j >= 0 {
			if !ci.Contains(output[j]) {
				break
			}
			j--
		}
		output = output[:j+1]

		// See if the last three cells plus this one can be collapsed.
		// We loop because collapsing three accepted cells and adding a higher level cell
		// could cascade into previously accepted cells.
		for len(output) >= 3 {
			fin := output[len(output)-3:]

			// fast XOR test; a necessary but not sufficient condition
			if fin[0]^fin[1]^fin[2]^ci != 0 {
				break
			}

			// more expensive test; exact.
			// Compute the two bit mask for the encoded child position,
			// then see if they all agree.
			mask := CellID(ci.lsb() << 1)
			mask = ^(mask + mask<<1)
			should := ci & mask
			if (fin[0]&mask != should) || (fin[1]&mask != should) || (fin[2]&mask != should) || ci.isFace() {
				break
			}

			output = output[:len(output)-3]
			ci = ci.immediateParent() // checked !ci.isFace above
		}
		output = append(output, ci)
	}
	*cu = output
}

/**
 * Return true if the cell union contains the given cell id. Containment is
 * defined with respect to regions, e.g. a cell contains its 4 children. This
 * is a fast operation (logarithmic in the size of the cell union).
 */
func (cu *CellUnion) Contains(id CellID) bool {
	// This function requires that Normalize has been called first.
	//
	// This is an exact test. Each cell occupies a linear span of the S2
	// space-filling curve, and the cell id is simply the position at the center
	// of this span. The cell union ids are sorted in increasing order along
	// the space-filling curve. So we simply find the pair of cell ids that
	// surround the given cell id (using binary search). There is containment
	// if and only if one of these two cell ids contains this cell.

	i := sort.Search(len(*cu), func(i int) bool { return id < (*cu)[i] })

	if i != len(*cu) && (*cu)[i].RangeMin() <= id {
		return true
	}
	return i != 0 && (*cu)[i-1].RangeMax() >= id
}

// Intersects reports whether this cell union intersects the given cell ID.
//
// This method assumes that the CellUnion has been normalized.
func (cu *CellUnion) Intersects(id CellID) bool {
	// Find index of array item that occurs directly after our probe cell:
	i := sort.Search(len(*cu), func(i int) bool { return id < (*cu)[i] })

	if i != len(*cu) && (*cu)[i].RangeMin() <= id.RangeMax() {
		return true
	}
	return i != 0 && (*cu)[i-1].RangeMax() >= id.RangeMin()
}

type byID []CellID

func (cu byID) Len() int           { return len(cu) }
func (cu byID) Less(i, j int) bool { return cu[i] < cu[j] }
func (cu byID) Swap(i, j int)      { cu[i], cu[j] = cu[j], cu[i] }
