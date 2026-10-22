// Copyright 2006 Google Inc. All rights reserved.
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

// RegionIntersection represents the intersection of a set of regions.
// It is convenient for computing a covering of the intersection of a set of
// regions. Note that an intersection of no regions covers the entire sphere.
//
// ContainsCell may return false even when the cell is fully contained by the
// intersection, because each member region's ContainsCell is allowed to be
// conservative. This conservatism compounds: if any region reports false for
// a cell it actually contains, the overall result is false.
type RegionIntersection []Region

// CapBound returns a bounding cap for this RegionIntersection.
func (ri RegionIntersection) CapBound() Cap { return ri.RectBound().CapBound() }

// RectBound returns a bounding latitude-longitude rectangle for this RegionIntersection.
func (ri RegionIntersection) RectBound() Rect {
	ret := FullRect()
	for _, reg := range ri {
		ret = ret.Intersection(reg.RectBound())
	}
	return ret
}

// ContainsCell reports whether the given Cell is contained by this RegionIntersection.
func (ri RegionIntersection) ContainsCell(c Cell) bool {
	for _, reg := range ri {
		if !reg.ContainsCell(c) {
			return false
		}
	}
	return true
}

// IntersectsCell reports whether this RegionIntersection may intersect the given cell.
// It returns false only if at least one region definitively does not intersect the cell;
// a true result means all regions may intersect, but does not guarantee actual intersection.
func (ri RegionIntersection) IntersectsCell(c Cell) bool {
	for _, reg := range ri {
		if !reg.IntersectsCell(c) {
			return false
		}
	}
	return true
}

// ContainsPoint reports whether this RegionIntersection contains the Point.
func (ri RegionIntersection) ContainsPoint(p Point) bool {
	for _, reg := range ri {
		if !reg.ContainsPoint(p) {
			return false
		}
	}
	return true
}

// CellUnionBound computes a covering of the RegionIntersection.
func (ri RegionIntersection) CellUnionBound() []CellID {
	return ri.CapBound().CellUnionBound()
}
