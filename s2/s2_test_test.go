// Copyright 2014 Google Inc. All rights reserved.
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

	"github.com/golang/geo/s1"
)

func TestKmToAngle(t *testing.T) {
	const earthRadiusKm = 6371.01

	tests := []struct {
		have float64
		want s1.Angle
	}{
		{0.0, 0.0},
		{1.0, 0.00015696098420815537 * s1.Radian},
		{earthRadiusKm, 1.0 * s1.Radian},
		{-1.0, -0.00015696098420815537 * s1.Radian},
		{-10000.0, -1.5696098420815536300 * s1.Radian},
		{1e9, 156960.984208155363007 * s1.Radian},
	}
	for _, test := range tests {
		if got := kmToAngle(test.have); !float64Eq(float64(got), float64(test.want)) {
			t.Errorf("kmToAngle(%f) = %0.20f, want %0.20f", test.have, got, test.want)
		}
	}
}

// TODO(roberts): Remaining tests
// TriangleFractal
// TriangleMultiFractal
// SpaceFillingFractal
// KochCurveFractal
// KochCurveMultiFractal
// CesaroFractal
// CesaroMultiFractal
