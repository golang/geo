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

	"github.com/golang/geo/r3"
)

func TestPointPlanarCentroid(t *testing.T) {
	tests := []struct {
		name             string
		p0, p1, p2, want Point
	}{
		{
			name: "xyz axis",
			p0:   Point{r3.Vector{0, 0, 1}},
			p1:   Point{r3.Vector{0, 1, 0}},
			p2:   Point{r3.Vector{1, 0, 0}},
			want: Point{r3.Vector{1. / 3, 1. / 3, 1. / 3}},
		},
		{
			name: "Same point",
			p0:   Point{r3.Vector{1, 0, 0}},
			p1:   Point{r3.Vector{1, 0, 0}},
			p2:   Point{r3.Vector{1, 0, 0}},
			want: Point{r3.Vector{1, 0, 0}},
		},
	}

	for _, test := range tests {
		got := PlanarCentroid(test.p0, test.p1, test.p2)
		if !got.ApproxEqual(test.want) {
			t.Errorf("%s: PlanarCentroid(%v, %v, %v) = %v, want %v", test.name, test.p0, test.p1, test.p2, got, test.want)
		}
	}
}

func TestPointTrueCentroid(t *testing.T) {
	// Test TrueCentroid with very small triangles. This test assumes that
	// the triangle is small enough so that it is nearly planar.
	// The centroid of a planar triangle is at the intersection of its
	// medians, which is two-thirds of the way along each median.
	for i := 0; i < 100; i++ {
		f := randomFrame()
		p := f.col(0)
		x := f.col(1)
		y := f.col(2)
		d := 1e-4 * math.Pow(1e-4, randomFloat64())

		// Make a triangle with two equal sides.
		p0 := Point{p.Sub(x.Mul(d)).Normalize()}
		p1 := Point{p.Add(x.Mul(d)).Normalize()}
		p2 := Point{p.Add(y.Mul(d * 3)).Normalize()}
		want := Point{p.Add(y.Mul(d)).Normalize()}

		got := TrueCentroid(p0, p1, p2).Normalize()
		if got.Distance(want.Vector) >= 2e-8 {
			t.Errorf("TrueCentroid(%v, %v, %v).Normalize() = %v, want %v", p0, p1, p2, got, want)
		}

		// Make a triangle with a right angle.
		p0 = p
		p1 = Point{p.Add(x.Mul(d * 3)).Normalize()}
		p2 = Point{p.Add(y.Mul(d * 6)).Normalize()}
		want = Point{p.Add(x.Add(y.Mul(2)).Mul(d)).Normalize()}

		got = TrueCentroid(p0, p1, p2).Normalize()
		if got.Distance(want.Vector) >= 2e-8 {
			t.Errorf("TrueCentroid(%v, %v, %v).Normalize() = %v, want %v", p0, p1, p2, got, want)
		}
	}
}
