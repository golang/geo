/*
Copyright 2015 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package s2

import (
	"testing"
)

func TestSimpleCrossing(t *testing.T) {
	tests := []struct {
		a, b, c, d Point
		want       bool
	}{
		{
			// Two regular edges that cross.
			PointFromCoords(1, 2, 1),
			PointFromCoords(1, -3, 0.5),
			PointFromCoords(1, -0.5, -3),
			PointFromCoords(0.1, 0.5, 3),
			true,
		},
		{
			// Two regular edges that cross antipodal points.
			PointFromCoords(1, 2, 1),
			PointFromCoords(1, -3, 0.5),
			PointFromCoords(-1, 0.5, 3),
			PointFromCoords(-0.1, -0.5, -3),
			false,
		},
		{
			// Two edges on the same great circle.
			PointFromCoords(0, 0, -1),
			PointFromCoords(0, 1, 0),
			PointFromCoords(0, 1, 1),
			PointFromCoords(0, 0, 1),
			false,
		},
		{
			// Two edges that cross where one vertex is the OriginPoint.
			PointFromCoords(1, 0, 0),
			OriginPoint(),
			PointFromCoords(1, -0.1, 1),
			PointFromCoords(1, 1, -0.1),
			true,
		},
		{
			// Two edges that cross antipodal points.
			PointFromCoords(1, 0, 0),
			PointFromCoords(0, 1, 0),
			PointFromCoords(0, 0, -1),
			PointFromCoords(-1, -1, 1),
			false,
		},
		{
			// Two edges that share an endpoint.  The Ortho() direction is (-4,0,2),
			// and edge CD is further CCW around (2,3,4) than AB.
			PointFromCoords(2, 3, 4),
			PointFromCoords(-1, 2, 5),
			PointFromCoords(7, -2, 3),
			PointFromCoords(2, 3, 4),
			false,
		},
	}

	for _, test := range tests {
		if got := SimpleCrossing(test.a, test.b, test.c, test.d); got != test.want {
			t.Errorf("SimpleCrossing(%v,%v,%v,%v) = %t, want %t",
				test.a, test.b, test.c, test.d, got, test.want)
		}
	}
}
