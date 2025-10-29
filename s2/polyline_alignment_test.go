// Copyright 2023 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS-IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package s2

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPolylineAlignmentWindowCreateFromStrides(t *testing.T) {
	//    0 1 2 3 4 5
	//  0 * * * . . .
	//  1 . * * * . .
	//  2 . . * * . .
	//  3 . . . * * *
	//  4 . . . . * *
	strides := []columnStride{
		{0, 3},
		{1, 4},
		{2, 4},
		{3, 6},
		{4, 6},
	}
	w := windowFromStrides(strides)
	if got := w.columnStride(0).start; got != 0 {
		t.Errorf("%+v.columnStride(0).start = %d, want %d", w, got, 0)
	}
	if got := w.columnStride(0).end; got != 3 {
		t.Errorf("%+v.columnStride(0).end = %d, want %d", w, got, 3)
	}
	if got := w.columnStride(4).start; got != 4 {
		t.Errorf("%+v.columnStride(4).start = %d, want %d", w, got, 4)
	}
	if got := w.columnStride(4).end; got != 6 {
		t.Errorf("%+v.columnStride(4).end = %d, want %d", w, got, 6)
	}
}

func TestPolylineAlignmentTestWindowDebugString(t *testing.T) {
	strides := []columnStride{
		{0, 4},
		{0, 4},
		{0, 4},
		{0, 4},
	}
	w := windowFromStrides(strides)
	want := ` * * * *
 * * * *
 * * * *
 * * * *
`
	got := w.debugString()
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("w.debugString() = %q, want %q\ndiff: %s", got, want, diff)
	}
}

func TestPolylineAlignmentWindowUpsample(t *testing.T) {
	tests := []struct {
		strides      []columnStride
		upRow, upCol int
		want         string
	}{
		{
			// UpsampleWindowByFactorOfTwo
			//   0 1 2 3 4 5
			// 0 * * * . . .
			// 1 . * * * . .
			// 2 . . * * . .
			// 3 . . . * * *
			// 4 . . . . * *
			strides: []columnStride{
				{0, 3}, {1, 4}, {2, 4}, {3, 6}, {4, 6},
			},
			upRow: 10,
			upCol: 12,
			want: ` * * * * * * . . . . . .
 * * * * * * . . . . . .
 . . * * * * * * . . . .
 . . * * * * * * . . . .
 . . . . * * * * . . . .
 . . . . * * * * . . . .
 . . . . . . * * * * * *
 . . . . . . * * * * * *
 . . . . . . . . * * * *
 . . . . . . . . * * * *
`,
		},
		{
			// UpsamplesWindowXAxisByFactorOfThree
			//   0 1 2 3 4 5
			// 0 * * * . . .
			// 1 . * * * . .
			// 2 . . * * . .
			// 3 . . . * * *
			// 4 . . . . * *
			strides: []columnStride{
				{0, 3}, {1, 4}, {2, 4}, {3, 6}, {4, 6},
			},
			upRow: 5,
			upCol: 18,
			want: ` * * * * * * * * * . . . . . . . . .
 . . . * * * * * * * * * . . . . . .
 . . . . . . * * * * * * . . . . . .
 . . . . . . . . . * * * * * * * * *
 . . . . . . . . . . . . * * * * * *
`,
		},
		{
			//  UpsamplesWindowYAxisByFactorOfThree
			//   0 1 2 3 4 5
			// 0 * * * . . .
			// 1 . * * * . .
			// 2 . . * * . .
			// 3 . . . * * *
			// 4 . . . . * *
			strides: []columnStride{
				{0, 3}, {1, 4}, {2, 4}, {3, 6}, {4, 6},
			},
			upRow: 15,
			upCol: 6,
			want: ` * * * . . .
 * * * . . .
 * * * . . .
 . * * * . .
 . * * * . .
 . * * * . .
 . . * * . .
 . . * * . .
 . . * * . .
 . . . * * *
 . . . * * *
 . . . * * *
 . . . . * *
 . . . . * *
 . . . . * *
`,
		},
		{
			// UpsamplesWindowByNonInteger
			//   0 1 2 3 4 5
			// 0 * * * . . .
			// 1 . * * * . .
			// 2 . . * * . .
			// 3 . . . * * *
			// 4 . . . . * *
			strides: []columnStride{
				{0, 3}, {1, 4}, {2, 4}, {3, 6}, {4, 6},
			},
			upRow: 19,
			upCol: 23,
			want: ` * * * * * * * * * * * * . . . . . . . . . . .
 * * * * * * * * * * * * . . . . . . . . . . .
 * * * * * * * * * * * * . . . . . . . . . . .
 * * * * * * * * * * * * . . . . . . . . . . .
 . . . . * * * * * * * * * * * . . . . . . . .
 . . . . * * * * * * * * * * * . . . . . . . .
 . . . . * * * * * * * * * * * . . . . . . . .
 . . . . * * * * * * * * * * * . . . . . . . .
 . . . . . . . . * * * * * * * . . . . . . . .
 . . . . . . . . * * * * * * * . . . . . . . .
 . . . . . . . . * * * * * * * . . . . . . . .
 . . . . . . . . . . . . * * * * * * * * * * *
 . . . . . . . . . . . . * * * * * * * * * * *
 . . . . . . . . . . . . * * * * * * * * * * *
 . . . . . . . . . . . . * * * * * * * * * * *
 . . . . . . . . . . . . . . . * * * * * * * *
 . . . . . . . . . . . . . . . * * * * * * * *
 . . . . . . . . . . . . . . . * * * * * * * *
 . . . . . . . . . . . . . . . * * * * * * * *
`,
		},
	}

	for _, test := range tests {
		w := windowFromStrides(test.strides)
		wUp := w.upsample(test.upRow, test.upCol)
		got := wUp.debugString()
		if diff := cmp.Diff(got, test.want); diff != "" {
			t.Errorf("%+v.upsample(%d, %d) = %q, want %q\ndiff: %s",
				test.strides, test.upRow, test.upCol, got, test.want, diff)
		}
	}
}

func TestPolylineAlignmentWindowDilate(t *testing.T) {
	tests := []struct {
		strides []columnStride
		dilate  int
		want    string
	}{
		{
			// DilatesWindowByRadiusZero
			//   0 1 2 3 4 5
			// 0 * * * . . .
			// 1 . . * . . .
			// 2 . . * . . .
			// 3 . . * * . .
			// 4 . . . * * *
			strides: []columnStride{
				{0, 3}, {2, 3}, {2, 3}, {2, 4}, {3, 6},
			},
			dilate: 0,
			want: ` * * * . . .
 . . * . . .
 . . * . . .
 . . * * . .
 . . . * * *
`,
		},
		{
			// DilatesWindowByRadiusOne
			//   0 1 2 3 4 5 (x's are the spots that we dilate into)
			// 0 * * * x . .
			// 1 x x * x . .
			// 2 . x * x x .
			// 3 . x * * x x
			// 4 . x x * * *
			strides: []columnStride{
				{0, 3}, {2, 3}, {2, 3}, {2, 4}, {3, 6},
			},
			dilate: 1,
			want: ` * * * * . .
 * * * * . .
 . * * * * .
 . * * * * *
 . * * * * *
`,
		},
		{
			// DilatesWindowByRadiusTwo
			//   0 1 2 3 4 5 (x's are the spots that we dilate into)
			// 0 * * * x x .
			// 1 x x * x x x
			// 2 x x * x x x
			// 3 x x * * x x
			// 4 x x x * * *
			strides: []columnStride{
				{0, 3}, {2, 3}, {2, 3}, {2, 4}, {3, 6},
			},
			dilate: 2,
			want: ` * * * * * .
 * * * * * *
 * * * * * *
 * * * * * *
 * * * * * *
`,
		},
		{
			// DilatesWindowByRadiusVeryLarge
			//   0 1 2 3 4 5 (x's are the spots that we dilate into)
			// 0 * * * x x .
			// 1 x x * x x x
			// 2 x x * x x x
			// 3 x x * * x x
			// 4 x x x * * *
			strides: []columnStride{
				{0, 3}, {2, 3}, {2, 3}, {2, 4}, {3, 6},
			},
			dilate: 100,
			want: ` * * * * * *
 * * * * * *
 * * * * * *
 * * * * * *
 * * * * * *
`,
		},
	}

	for _, test := range tests {
		w := windowFromStrides(test.strides)
		wUp := w.dilate(test.dilate)
		got := wUp.debugString()
		if diff := cmp.Diff(got, test.want); diff != "" {
			t.Errorf("%+v.dilate(%d) = %q, want %q\ndiff: %s",
				test.strides, test.dilate, got, test.want, diff)
		}
	}
}

func TestPolylineAlignmentHalfResolution(t *testing.T) {
	tests := []struct {
		have string
		want string
	}{
		{
			// ZeroLength
			have: "",
			want: "",
		},
		{
			// EvenLength
			have: "0:0, 0:1, 0:2, 1:2",
			want: "0:0, 0:2",
		},
		{
			// OddLength
			have: "0:0, 0:1, 0:2, 1:2, 3:5",
			want: "0:0, 0:2, 3:5",
		},
	}

	for _, test := range tests {
		got := halfResolution(makePolyline(test.have))
		if gotS := pointsToString(*got, false); gotS != test.want {
			t.Errorf("halfResolution(%s) = %s, want %s", test.have, gotS, test.want)
		}
	}
}

func distanceMatrix(a, b *Polyline) costTable {
	aN := len(*a)
	bN := len(*b)
	table := costTable(make([][]float64, aN))
	for i := range aN {
		table[i] = make([]float64, bN)
		for j := range bN {
			table[i][j] = (*a)[i].Sub((*b)[j].Vector).Norm()
		}
	}
	return table
}

// Do some testing against random sequences with a brute-force solver.
// Returns the optimal cost of alignment up until vertex i, j.
func bruteForceCost(table costTable, i, j int) float64 {
	if i == 0 && j == 0 {
		return table[0][0]
	} else if i == 0 {
		return bruteForceCost(table, i, j-1) + table[i][j]
	} else if j == 0 {
		return bruteForceCost(table, i-1, j) + table[i][j]
	} else {
		return min(bruteForceCost(table, i-1, j-1),
			bruteForceCost(table, i-1, j),
			bruteForceCost(table, i, j-1)) +
			table[i][j]
	}
}

func TestPolylineAlignmentExactAlignmentCost(t *testing.T) {
	tests := []struct {
		label    string
		a, b     string
		wantPath warpPath
	}{
		// Test cases that should cause panic and crash
		/*
			{
				label:    "ExactLengthZeroInputs",
				a:        "",
				b:        "",
				wantPath: warpPath{},
			},
			{
				label:    "ExactLengthZeroInputA",
				a:        "",
				b:        "0:0, 1:1, 2:2",
				wantPath: warpPath{},
			},
			{
				label:    "ExactLengthZeroInputB",
				a:        "0:0, 1:1, 2:2",
				b:        "",
				wantPath: warpPath{},
			},
		*/
		{
			label:    "ExactLengthOneInputs",
			a:        "1:1",
			b:        "2:2",
			wantPath: warpPath{{0, 0}},
		},
		{
			label:    "ExactLengthOneInputA",
			a:        "0:0",
			b:        "0:0, 1:1, 2:2",
			wantPath: warpPath{{0, 0}, {0, 1}, {0, 2}},
		},
		{
			label:    "ExactLengthOneInputB",
			a:        "0:0, 1:1, 2:2",
			b:        "0:0",
			wantPath: warpPath{{0, 0}, {1, 0}, {2, 0}},
		},
		{
			label:    "ExactHeaderFileExample",
			a:        "1:0, 5:0, 6:0, 9:0",
			b:        "2:0, 7:0, 8:0",
			wantPath: warpPath{{0, 0}, {1, 1}, {2, 1}, {3, 2}},
		},
		{
			// Tests that we get the correct path in the case where we have polylines at
			// right angles, that would get a different matching of points for distance
			// cost versus squared distance cost. If we had used squared distance for the
			// cost the path would be {{0, 0}, {1, 0}, {2, 0}, {3, 1}, {3, 2}}; See
			// https://screenshot.googleplex.com/7eeMjdSc5HeSeTD for the costs between the
			// different pairs for distance and squared distance
			//
			// A0---A1---A2
			// B0       |
			// |        A3
			// B1-------B2
			label:    "DifferentPathForDistanceVersusSquaredDistance",
			a:        "0.1:-0.1, 0.1:0, 0.1:0.1, -0.1:0.1",
			b:        "0.1:-0.1, -0.1:-0.1, -0.1:0.1",
			wantPath: warpPath{{0, 0}, {1, 0}, {2, 1}, {3, 2}},
		},
	}

	for _, test := range tests {
		// Use Brute Force solver to verify exact Dynamic Programming solvers.
		a := makePolyline(test.a)
		b := makePolyline(test.b)
		aN := len(*a)
		bN := len(*b)

		bruteCost := bruteForceCost(distanceMatrix(a, b), aN-1, bN-1)
		exactCost := ExactVertexAlignmentCost(a, b)
		if !float64Eq(bruteCost, exactCost) {
			t.Errorf("%s: ExactVertexAlignmentCost(%v, %v) = %f, want %f",
				test.label, a, b, exactCost, bruteCost)
		}

		exactAlignment := ExactVertexAlignment(a, b)
		if !float64Eq(bruteCost, exactAlignment.alignmentCost) {
			t.Errorf("%s: ExactVertexAlignment(%v, %v) = %f, want %f",
				test.label, a, b, exactAlignment.alignmentCost, bruteCost)
		}
	}

	// TODO(rsned): Add FuzzWithBruteForce to this.
}

// TODO()rsned): Differences from C++
// Medoid tests
// Consensus tests
