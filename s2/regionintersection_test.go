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

import (
	"testing"
)

func TestEmptyRegionIntersectionHasFullCap(t *testing.T) {
	var empty RegionIntersection

	got := empty.CapBound()

	want := FullCap()
	if !got.ApproxEqual(want) {
		t.Errorf("empty region intersection cap = %v, want %v", got, want)
	}
}

func TestEmptyRegionIntersectionHasFullBound(t *testing.T) {
	var empty RegionIntersection

	got := empty.RectBound()

	want := FullRect()
	if !got.ApproxEqual(want) {
		t.Errorf("empty region intersection rect = %v, want %v", got, want)
	}
}

func TestEmptyRegionIntersectionContainsCell(t *testing.T) {
	var empty RegionIntersection

	// Empty intersection covers the whole sphere, so it contains every cell.
	if !empty.ContainsCell(CellFromCellID(CellIDFromFace(0))) {
		t.Error("empty RegionIntersection.ContainsCell = false, want true")
	}
}

func TestEmptyRegionIntersectionIntersectsCell(t *testing.T) {
	var empty RegionIntersection

	// Empty intersection covers the whole sphere, so it intersects every cell.
	if !empty.IntersectsCell(CellFromCellID(CellIDFromFace(0))) {
		t.Error("empty RegionIntersection.IntersectsCell = false, want true")
	}
}

// twoRectsRegionIntersection is the intersection of two overlapping rects.
// Their overlap is the region 20:20 to 60:60.
var twoRectsRegionIntersection = RegionIntersection{
	makeRect("0:0, 60:60"),
	makeRect("20:20, 80:80"),
}

func TestRegionIntersectionRectBound(t *testing.T) {
	got := twoRectsRegionIntersection.RectBound()

	want := makeRect("20:20, 60:60")
	if !got.ApproxEqual(want) {
		t.Errorf("%v.RectBound() = %v, want %v", twoRectsRegionIntersection, got, want)
	}
}

func TestRegionIntersectionContainsPoint(t *testing.T) {
	testCases := []struct {
		ll   LatLng
		want bool
	}{
		{LatLngFromDegrees(30, 30), true},  // inside both rects
		{LatLngFromDegrees(10, 10), false}, // inside first rect, outside second
		{LatLngFromDegrees(70, 70), false}, // inside second rect, outside first
		{LatLngFromDegrees(90, 90), false}, // outside both rects
	}

	for _, tc := range testCases {
		got := twoRectsRegionIntersection.ContainsPoint(PointFromLatLng(tc.ll))
		if got != tc.want {
			t.Errorf("%v.ContainsPoint(%v) = %t, want %t", twoRectsRegionIntersection, tc.ll, got, tc.want)
		}
	}
}

func TestRegionIntersectionContainsCell(t *testing.T) {
	// A small cell well inside the intersection should be contained.
	center := CellIDFromLatLng(LatLngFromDegrees(40, 40)).Parent(10)
	if !twoRectsRegionIntersection.ContainsCell(CellFromCellID(center)) {
		t.Errorf("%v.ContainsCell(center) = false, want true", twoRectsRegionIntersection)
	}

	// A face cell spans a huge area and is not contained in either rect.
	face := CellFromCellID(CellIDFromFace(0))
	if twoRectsRegionIntersection.ContainsCell(face) {
		t.Errorf("%v.ContainsCell(face0) = true, want false", twoRectsRegionIntersection)
	}

	// A cell straddling the southern boundary of the intersection (lat≈20)
	// extends outside the intersection and must not be reported as contained.
	boundary := CellFromCellID(CellIDFromLatLng(LatLngFromDegrees(20, 40)).Parent(5))
	if twoRectsRegionIntersection.ContainsCell(boundary) {
		t.Errorf("%v.ContainsCell(boundary) = true, want false", twoRectsRegionIntersection)
	}
}

func TestRegionIntersectionIntersectsCell(t *testing.T) {
	// A cell well inside the intersection area should intersect.
	inside := CellIDFromLatLng(LatLngFromDegrees(40, 40)).Parent(5)
	if !twoRectsRegionIntersection.IntersectsCell(CellFromCellID(inside)) {
		t.Errorf("%v.IntersectsCell(inside) = false, want true", twoRectsRegionIntersection)
	}

	// Face 5 covers the south-pole hemisphere, fully outside both positive-latitude rects.
	outside := CellFromCellID(CellIDFromFace(5))
	if twoRectsRegionIntersection.IntersectsCell(outside) {
		t.Errorf("%v.IntersectsCell(face5) = true, want false", twoRectsRegionIntersection)
	}

	// A cell straddling the southern boundary (lat≈20) overlaps the intersection.
	boundary := CellFromCellID(CellIDFromLatLng(LatLngFromDegrees(20, 40)).Parent(5))
	if !twoRectsRegionIntersection.IntersectsCell(boundary) {
		t.Errorf("%v.IntersectsCell(boundary) = false, want true", twoRectsRegionIntersection)
	}
}

func TestSingleRegionIntersectionMatchesRegion(t *testing.T) {
	rect := makeRect("20:20, 60:60")
	ri := RegionIntersection{rect}

	if !ri.RectBound().ApproxEqual(rect.RectBound()) {
		t.Errorf("single-region RectBound = %v, want %v", ri.RectBound(), rect.RectBound())
	}

	testCases := []struct {
		ll   LatLng
		want bool
	}{
		{LatLngFromDegrees(40, 40), true},
		{LatLngFromDegrees(0, 0), false},
	}
	for _, tc := range testCases {
		p := PointFromLatLng(tc.ll)
		if got := ri.ContainsPoint(p); got != tc.want {
			t.Errorf("single-region ContainsPoint(%v) = %t, want %t", tc.ll, got, tc.want)
		}
	}
}

func TestRegionIntersectionCellUnionBound(t *testing.T) {
	got := twoRectsRegionIntersection.CellUnionBound()

	if len(got) == 0 {
		t.Fatal("CellUnionBound of non-empty intersection should not be empty")
	}
	// CellUnionBound returns unsorted cells; normalize before containment check.
	center := PointFromLatLng(LatLngFromDegrees(40, 40))
	cu := CellUnion(got)
	cu.Normalize()
	if !cu.ContainsPoint(center) {
		t.Errorf("CellUnionBound %v does not contain intersection center", got)
	}
}

func TestRegionIntersectionCovering(t *testing.T) {
	cov := NewRegionCoverer()
	cov.MaxCells = 8

	got := cov.Covering(twoRectsRegionIntersection)

	if len(got) == 0 {
		t.Fatalf("covering of non-empty intersection should not be empty")
	}
	center := PointFromLatLng(LatLngFromDegrees(40, 40))
	cu := CellUnion(got)
	if !cu.ContainsPoint(center) {
		t.Errorf("covering %v does not contain intersection center %v", got, center)
	}
}

// Make sure RegionIntersection implements Region.
var _ Region = RegionIntersection{}
