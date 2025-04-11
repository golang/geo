// Copyright 2019 Google Inc. All rights reserved.
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

package s1_test

import (
	"fmt"
	"math"

	"github.com/golang/geo/s1"
)

func ExampleInterval_DirectedHausdorffDistance() {
	// Small interval around the midpoints between quadrants, such that
	// the center of each interval is offset slightly CCW from the midpoint.
	mid := s1.IntervalFromEndpoints(math.Pi/2-0.01, math.Pi/2+0.02)
	fmt.Println("empty to empty: ", s1.EmptyInterval().DirectedHausdorffDistance(s1.EmptyInterval()))
	fmt.Println("empty to mid12: ", s1.EmptyInterval().DirectedHausdorffDistance(mid))
	fmt.Println("mid12 to empty: ", mid.DirectedHausdorffDistance(s1.EmptyInterval()))

	// Quadrant pair.
	quad2 := s1.IntervalFromEndpoints(0, -math.Pi)
	// Quadrant triple.
	quad3 := s1.IntervalFromEndpoints(0, -math.Pi/2)
	fmt.Println("quad12 to quad123 ", quad2.DirectedHausdorffDistance(quad3))

	// An interval whose complement center is 0.
	in := s1.IntervalFromEndpoints(3, -3)

	ivs := []s1.Interval{s1.IntervalFromEndpoints(-0.1, 0.2), s1.IntervalFromEndpoints(0.1, 0.2), s1.IntervalFromEndpoints(-0.2, -0.1)}
	for _, iv := range ivs {
		fmt.Printf("dist from %v to in: %f\n", iv, iv.DirectedHausdorffDistance(in))
	}
	// Output:
	// empty to empty:  0.0000000
	// empty to mid12:  0.0000000
	// mid12 to empty:  180.0000000
	// quad12 to quad123  0.0000000
	// dist from [-0.1000000, 0.2000000] to in: 3.000000
	// dist from [0.1000000, 0.2000000] to in: 2.900000
	// dist from [-0.2000000, -0.1000000] to in: 2.900000
}
