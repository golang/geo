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

package s2_test

import (
	"fmt"

	"github.com/golang/geo/s2"
)

func ExamplePolygonFromOrientedLoops() {
	// Let's define three loops, in format World Geodetic System 1984,
	// the format that geoJSON uses. The third loop is a hole in the second,
	// the first loop is remote from the others. Loops 1 and 2 are counter-clockwise,
	// while loop 3 is clockwise.
	l1 := [][]float64{
		{102.0, 2.0},
		{103.0, 2.0},
		{103.0, 3.0},
		{102.0, 3.0},
	}
	l2 := [][]float64{
		{100.0, 0.0},
		{101.0, 0.0},
		{101.0, 1.0},
		{100.0, 1.0},
	}
	l3 := [][]float64{
		{100.2, 0.2},
		{100.2, 0.8},
		{100.8, 0.8},
		{100.8, 0.2},
	}
	toLoop := func(points [][]float64) *s2.Loop {
		var pts []s2.Point
		for _, pt := range points {
			pts = append(pts, s2.PointFromLatLng(s2.LatLngFromDegrees(pt[1], pt[0])))
		}
		return s2.LoopFromPoints(pts)
	}
	// We can combine all loops into a single polygon:
	p := s2.PolygonFromOrientedLoops([]*s2.Loop{toLoop(l1), toLoop(l2), toLoop(l3)})

	for i, loop := range p.Loops() {
		fmt.Printf("loop %d is hole: %t\n", i, loop.IsHole())
	}
	fmt.Printf("Combined area: %.7f\n", p.Area())

	// Note how the area of the polygon is the area of l1 + l2 - invert(l3), because l3 is a hole:
	p12 := s2.PolygonFromOrientedLoops([]*s2.Loop{toLoop(l1), toLoop(l2)})
	p3 := s2.PolygonFromOrientedLoops([]*s2.Loop{toLoop(l3)})
	p3.Invert()
	fmt.Printf("l1+l2 = %.7f, inv(l3) = %.7f; l1+l2 - inv(l3) = %.7f\n", p12.Area(), p3.Area(), p12.Area()-p3.Area())
	// Output:
	// loop 0 is hole: false
	// loop 1 is hole: false
	// loop 2 is hole: true
	// Combined area: 0.0004993
	// l1+l2 = 0.0006089, inv(l3) = 0.0001097; l1+l2 - inv(l3) = 0.0004993
}
