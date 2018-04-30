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
	"github.com/golang/geo/s1"
)

var (
	pz   = Point{r3.Vector{0, 0, 1}}
	p000 = Point{r3.Vector{1, 0, 0}}
	p045 = Point{r3.Vector{1, 1, 0}.Normalize()}
	p090 = Point{r3.Vector{0, 1, 0}}
	p180 = Point{r3.Vector{-1, 0, 0}}
	// Degenerate triangles.
	pr = Point{r3.Vector{0.257, -0.5723, 0.112}}
	pq = Point{r3.Vector{-0.747, 0.401, 0.2235}}

	// For testing the Girard area fall through case.
	g1 = Point{r3.Vector{1, 1, 1}}
	g2 = Point{g1.Add(pr.Mul(1e-15)).Normalize()}
	g3 = Point{g1.Add(pq.Mul(1e-15)).Normalize()}
)

func TestPointMeasuresPointArea(t *testing.T) {
	epsilon := 1e-10
	tests := []struct {
		a, b, c  Point
		want     float64
		nearness float64
	}{
		{p000, p090, pz, math.Pi / 2.0, 0},
		// This test case should give 0 as the epsilon, but either Go or C++'s value for Pi,
		// or the accuracy of the multiplications along the way, cause a difference ~15 decimal
		// places into the result, so it is not quite a difference of 0.
		{p045, pz, p180, 3.0 * math.Pi / 4.0, 1e-14},
		// Make sure that Area has good *relative* accuracy even for very small areas.
		{Point{r3.Vector{epsilon, 0, 1}}, Point{r3.Vector{0, epsilon, 1}}, pz, 0.5 * epsilon * epsilon, 1e-14},
		// Make sure that it can handle degenerate triangles.
		{pr, pr, pr, 0.0, 0},
		{pr, pq, pr, 0.0, 1e-15},
		{p000, p045, p090, 0.0, 0},
		// Try a very long and skinny triangle.
		{p000, Point{r3.Vector{1, 1, epsilon}}, p090, 5.8578643762690495119753e-11, 1e-9},
		{g1, g2, g3, 0.0, 1e-15},
	}
	for _, test := range tests {
		if got := PointArea(test.a, test.b, test.c); !float64Near(got, test.want, test.nearness) {
			t.Errorf("PointArea(%v, %v, %v), got %v want %v", test.a, test.b, test.c, got, test.want)
		}
	}

	maxGirard := 0.0
	for i := 0; i < 10000; i++ {
		p0 := randomPoint()
		d1 := randomPoint()
		d2 := randomPoint()
		p1 := Point{p0.Add(d1.Mul(1e-15)).Normalize()}
		p2 := Point{p0.Add(d2.Mul(1e-15)).Normalize()}
		// The actual displacement can be as much as 1.2e-15 due to roundoff.
		// This yields a maximum triangle area of about 0.7e-30.
		if got := PointArea(p0, p1, p2); got > 0.7e-30 {
			t.Errorf("PointArea(%v, %v, %v) = %v, want <= %v", p1, p1, p2, got, 0.7e-30)
		}
		if a := GirardArea(p0, p1, p2); a > maxGirard {
			maxGirard = a
		}
	}
	// This check only passes if GirardArea uses PointCross.
	if maxGirard > 1e-14 {
		t.Errorf("maximum GirardArea = %v, want <= %v", maxGirard, 1e-14)
	}
}

func TestPointMeasuresPointAreaQuarterHemisphere(t *testing.T) {
	tests := []struct {
		a, b, c, d, e Point
		want          float64
	}{
		// Triangles with near-180 degree edges that sum to a quarter-sphere.
		{PointFromCoords(1, 0.1*epsilon, epsilon), p000, p045, p180, pz, math.Pi},
		// Four other triangles that sum to a quarter-sphere.
		{PointFromCoords(1, 1, epsilon), p000, p045, p180, pz, math.Pi},
	}
	for _, test := range tests {
		area := PointArea(test.a, test.b, test.c) +
			PointArea(test.a, test.c, test.d) +
			PointArea(test.a, test.d, test.e) +
			PointArea(test.a, test.e, test.b)

		if !float64Eq(area, test.want) {
			t.Errorf("Adding up 4 quarter hemispheres with PointArea(), got %v want %v", area, test.want)
		}
	}

	// Compute the area of a hemisphere using four triangles with one near-180
	// degree edge and one near-degenerate edge.
	for i := 0; i < 100; i++ {
		lng := s1.Angle(2 * math.Pi * randomFloat64())
		p2Lng := lng + s1.Angle(randomFloat64())
		p0 := PointFromLatLng(LatLng{1e-20, lng}.Normalized())
		p1 := PointFromLatLng(LatLng{0, lng}.Normalized())
		p2 := PointFromLatLng(LatLng{0, p2Lng}.Normalized())
		p3 := PointFromLatLng(LatLng{0, lng + math.Pi}.Normalized())
		p4 := PointFromLatLng(LatLng{0, lng + 5.0}.Normalized())
		area := PointArea(p0, p1, p2) + PointArea(p0, p2, p3) + PointArea(p0, p3, p4) + PointArea(p0, p4, p1)
		if !float64Near(area, 2*math.Pi, 2e-15) {
			t.Errorf("hemisphere area of %v, %v, %v, %v, %v = %v, want %v", p1, p1, p2, p3, p4, area, 2*math.Pi)
		}
	}
}

func TestPointMeasuresAngleMethods(t *testing.T) {

	tests := []struct {
		a, b, c       Point
		wantAngle     s1.Angle
		wantTurnAngle s1.Angle
	}{
		{p000, pz, p045, math.Pi / 4, -3 * math.Pi / 4},
		{p045, pz, p180, 3 * math.Pi / 4, -math.Pi / 4},
		{p000, pz, p180, math.Pi, 0},
		{pz, p000, p045, math.Pi / 2, math.Pi / 2},
		{pz, p000, pz, 0, -math.Pi},
	}

	for _, test := range tests {
		if got := Angle(test.a, test.b, test.c); math.Abs(float64(got-test.wantAngle)) > epsilon {
			t.Errorf("Angle(%v, %v, %v) = %v, want %v", test.a, test.b, test.c, got, test.wantAngle)
		}
		if got := TurnAngle(test.a, test.b, test.c); math.Abs(float64(got-test.wantTurnAngle)) > epsilon {
			t.Errorf("TurnAngle(%v, %v, %v) = %v, want %v", test.a, test.b, test.c, got, test.wantTurnAngle)
		}
	}
}

func BenchmarkPointArea(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PointArea(p000, p090, pz)
	}
}

func BenchmarkPointAreaGirardCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PointArea(g1, g2, g3)
	}
}
