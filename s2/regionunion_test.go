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
)

func TestEmptyRegionUnionHasEmptyCap(t *testing.T) {
	var empty RegionUnion

	got := empty.CapBound()

	want := EmptyCap()
	if !got.ApproxEqual(want) {
		t.Errorf("empty region union cap = %v, want %v", got, want)
	}
}

func TestEmptyRegionUnionHasEmptyBound(t *testing.T) {
	var empty RegionUnion

	got := empty.RectBound()

	want := EmptyRect()
	if !got.ApproxEqual(want) {
		t.Errorf("empty region union cap = %v, want %v", got, want)
	}
}

func TestRegionUnionOfTwoPointsHasCorrectBound(t *testing.T) {
	got := twoPointsRegionUnion.RectBound()

	want := makeRect("-35:-40,35:40")
	if !got.ApproxEqual(want) {
		t.Errorf("%v.RectBound() = %v, want %v", twoPointsRegionUnion, got, want)
	}
}

var twoPointsRegionUnion = RegionUnion{
	PointFromLatLng(LatLngFromDegrees(35, 40)),
	PointFromLatLng(LatLngFromDegrees(-35, -40)),
}

func TestRegionUnionOfTwoPointsIntersectsFace0(t *testing.T) {
	got := twoPointsRegionUnion.IntersectsCell(face0Cell)

	if !got {
		t.Errorf("%v.IntersectsCell(%v) = %v, want true", twoPointsRegionUnion, face0Cell, got)
	}
}

func TestRegionUnionOfTwoPointsDoesNotContainFace0(t *testing.T) {
	got := twoPointsRegionUnion.ContainsCell(face0Cell)

	if got {
		t.Errorf("%v.ContainsCell(%v) = %v, want false", twoPointsRegionUnion, face0Cell, got)
	}
}

var face0Cell = CellFromCellID(CellIDFromFace(0))

func TestRegionUnionOfTwoContainsPoint(t *testing.T) {
	testCases := []struct {
		ll   LatLng
		want bool
	}{
		{LatLngFromDegrees(35, 40), true},
		{LatLngFromDegrees(-35, -40), true},
		{LatLngFromDegrees(0, 0), false},
	}

	for _, tc := range testCases {
		got := twoPointsRegionUnion.ContainsPoint(PointFromLatLng(tc.ll))

		if got != tc.want {
			t.Errorf("%v.ContainsPoint(%v) = %t, want %t", twoPointsRegionUnion, tc.ll, got, tc.want)
		}
	}
}

func TestTwoPointsRegionCovering(t *testing.T) {
	cov := NewRegionCoverer()
	cov.MaxCells = 1

	got := cov.Covering(twoPointsRegionUnion)

	const wantLen = 1
	if l := len(got); l != wantLen {
		t.Fatalf("covering = %v, len %v, want %v", got, l, wantLen)
	}
	want := CellIDFromFace(0)
	if g := got[0]; g != want {
		t.Errorf("covering[0] = %v, want %v", g, want)
	}
}

// Make sure RegionUnion implements Region.
var _ Region = RegionUnion{}
