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
)

func TestPolylineMeasuresGreatCircles(t *testing.T) {
	// Construct random great circles and divide them randomly into segments.
	// Then make sure that the length and centroid are correct.  Note that
	// because of the way the centroid is computed, it does not matter how
	// we split the great circle into segments.
	for iter := 0; iter < 100; iter++ {
		// Choose a coordinate frame for the great circle.
		f := randomFrame()
		x := f.row(0)
		y := f.row(1)

		var line []Point
		for theta := 0.0; theta < 2*math.Pi; theta += math.Pow(randomFloat64(), 10) {
			line = append(line, Point{x.Mul(math.Cos(theta)).Add(y.Mul(math.Sin(theta)))})
		}

		// Close the circle.
		line = append(line, line[0])

		length := polylineLength(line)
		if got, want := math.Abs(length.Radians()-2*math.Pi), 2e-14; got > want {
			t.Errorf("polylineLength(%v) = %v, want < %v", line, got, want)
		}

		centroid := polylineCentroid(line)
		if got, want := centroid.Norm(), 2e-14; got > want {
			t.Errorf("%v.Norm() = %v, want < %v", centroid, got, want)
		}
	}
}
