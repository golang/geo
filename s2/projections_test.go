// Copyright 2018 Google Inc. All rights reserved.
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
	"math"
	"testing"

	"github.com/golang/geo/r2"
	"github.com/golang/geo/r3"
)

func TestPlateCarreeProjectionInterpolate(t *testing.T) {
	a := r2.Point{1.234, -5.456e-20}
	b := r2.Point{2.1234e-20, 7.456}
	tests := []struct {
		dist float64
		a, b r2.Point
		want r2.Point
	}{
		{
			// Test that coordinates and/or arguments are not accidentally reversed.
			0.25,
			r2.Point{1, 5},
			r2.Point{3, 9},
			r2.Point{1.5, 6},
		},
		{
			// Test extrapolation.
			-2,
			r2.Point{1, 0},
			r2.Point{3, 0},
			r2.Point{-3, 0},
		},
		// Check that interpolation is exact at both endpoints.
		{0, a, b, a},
		{1, a, b, b},
	}
	proj := NewPlateCarreeProjection(180)

	for _, test := range tests {
		if got := proj.Interpolate(test.dist, test.a, test.b); got != test.want {
			t.Errorf("proj.Interpolate(%v, %v, %v) = %v, want %v", test.dist, test.a, test.b, got, test.want)
		}
	}
}

func TestPlateCarreeProjectionProjectUnproject(t *testing.T) {
	tests := []struct {
		have Point
		want r2.Point
	}{
		{Point{r3.Vector{1, 0, 0}}, r2.Point{0, 0}},
		{Point{r3.Vector{-1, 0, 0}}, r2.Point{180, 0}},
		{Point{r3.Vector{0, 1, 0}}, r2.Point{90, 0}},
		{Point{r3.Vector{0, -1, 0}}, r2.Point{-90, 0}},
		{Point{r3.Vector{0, 0, 1}}, r2.Point{0, 90}},
		{Point{r3.Vector{0, 0, -1}}, r2.Point{0, -90}},
	}

	proj := NewPlateCarreeProjection(180)

	for _, test := range tests {
		if got := proj.Project(test.have); !r2PointsApproxEqual(test.want, got, epsilon) {
			t.Errorf("proj.Project(%v) = %v, want %v", test.have, got, test.want)
		}
		if got := proj.Unproject(test.want); !got.ApproxEqual(test.have) {
			t.Errorf("proj.Unproject(%v) = %v, want %v", test.want, got, test.have)
		}
	}
}

func TestMercatorProjectionProjectUnproject(t *testing.T) {
	tests := []struct {
		have Point
		want r2.Point
	}{
		{Point{r3.Vector{1, 0, 0}}, r2.Point{0, 0}},
		{Point{r3.Vector{-1, 0, 0}}, r2.Point{180, 0}},
		{Point{r3.Vector{0, 1, 0}}, r2.Point{90, 0}},
		{Point{r3.Vector{0, -1, 0}}, r2.Point{-90, 0}},
		// Test one arbitrary point as a sanity check.
		{PointFromLatLng(LatLng{1, 0}), r2.Point{0, 70.255578967830246}},
	}

	proj := NewMercatorProjection(180)

	for _, test := range tests {
		if got := proj.Project(test.have); !r2PointsApproxEqual(test.want, got, epsilon) {
			t.Errorf("proj.Project(%v) = %v, want %v", test.have, got, test.want)
		}
		if got := proj.Unproject(test.want); !got.ApproxEqual(test.have) {
			t.Errorf("proj.Unproject(%v) = %v, want %v", test.want, got, test.have)
		}
	}

	// two cases have values that should be infinities, so the equality tests
	// need to be inf ready.
	testsInf := []struct {
		have Point
		want r2.Point
	}{
		{Point{r3.Vector{0, 0, 1}}, r2.Point{0, math.Inf(1)}},
		{Point{r3.Vector{0, 0, -1}}, r2.Point{0, math.Inf(-1)}},
	}
	for _, test := range testsInf {
		got := proj.Project(test.have)
		if ((math.IsInf(test.want.X, 1) && !math.IsInf(got.X, 1)) ||
			(math.IsInf(test.want.X, -1) && !math.IsInf(got.X, -1))) ||
			((math.IsInf(test.want.Y, 1) && !math.IsInf(got.Y, 1)) ||
				(math.IsInf(test.want.Y, -1) && !math.IsInf(got.Y, -1))) {
			t.Errorf("proj.Project(%v) = %v, want %v", test.have, got, test.want)
		}
	}
}
