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

func TestCentroidsPlanarCentroid(t *testing.T) {
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

func TestCentroidsTrueCentroid(t *testing.T) {
	// Test TrueCentroid with very small triangles. This test assumes that
	// the triangle is small enough so that it is nearly planar.
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

		// The centroid of a planar triangle is at the intersection of its
		// medians, which is two-thirds of the way along each median.
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

func TestCentroidsEdgeTrueCentroidSemiCircles(t *testing.T) {
	// Test the centroid of polyline ABC that follows the equator and consists
	// of two 90 degree edges (i.e., C = -A).  The centroid (multiplied by
	// length) should point toward B and have a norm of 2.0.  (The centroid
	// itself has a norm of 2/Pi, and the total edge length is Pi.)
	a := PointFromCoords(0, -1, 0)
	b := PointFromCoords(1, 0, 0)
	c := PointFromCoords(0, 1, 0)
	centroid := Point{EdgeTrueCentroid(a, b).Add(EdgeTrueCentroid(b, c).Vector)}

	if !b.ApproxEqual(Point{centroid.Normalize()}) {
		t.Errorf("EdgeTrueCentroid(%v, %v) + EdgeTrueCentroid(%v, %v) = %v, want %v", a, b, b, c, centroid, b)
	}
	if got, want := centroid.Norm(), 2.0; !float64Eq(got, want) {
		t.Errorf("%v.Norm() = %v, want %v", centroid, got, want)
	}
}

func TestCentroidsEdgeTrueCentroidGreatCircles(t *testing.T) {
	// Construct random great circles and divide them randomly into segments.
	// Then make sure that the centroid is approximately at the center of the
	// sphere.  Note that because of the way the centroid is computed, it does
	// not matter how we split the great circle into segments.
	//
	// Note that this is a direct test of the properties that the centroid
	// should have, rather than a test that it matches a particular formula.
	for iter := 0; iter < 100; iter++ {
		f := randomFrameAtPoint(randomPoint())
		x := f.col(0)
		y := f.col(1)

		var centroid Point

		v0 := x
		for theta := 0.0; theta < 2*math.Pi; theta += math.Pow(randomFloat64(), 10) {
			v1 := Point{x.Mul(math.Cos(theta)).Add(y.Mul(math.Sin(theta)))}
			centroid = Point{centroid.Add(EdgeTrueCentroid(v0, v1).Vector)}
			v0 = v1
		}
		// Close the circle.
		centroid = Point{centroid.Add(EdgeTrueCentroid(v0, x).Vector)}
		if got, want := centroid.Norm(), 2e-14; got > want {
			t.Errorf("%v.Norm() = %v, want <= %v", centroid, got, want)
		}
	}
}
