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
	"fmt"
	"math"
	"testing"

	"github.com/golang/geo/s1"
)

func TestKmToAngle(t *testing.T) {
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

func numVerticesAtLevel(level int) int {
	// Sanity / overflow check
	if level < 0 || level > 14 {
		panic(fmt.Sprintf("level %d out of range for fractal tests", level))
	}
	return 3 * (1 << (2 * uint(level))) // 3*(4**level)
}

func TestTestingFractal(t *testing.T) {
	tests := []struct {
		label     string
		minLevel  int
		maxLevel  int
		dimension float64
	}{

		{
			label:     "TriangleFractal",
			minLevel:  7,
			maxLevel:  7,
			dimension: 1.0,
		},
		{
			label:     "TriangleMultiFractal",
			minLevel:  2,
			maxLevel:  6,
			dimension: 1.0,
		},
		{
			label:     "SpaceFillingFractal",
			minLevel:  4,
			maxLevel:  4,
			dimension: 1.999,
		},
		{
			label:     "KochCurveFractal",
			minLevel:  7,
			maxLevel:  7,
			dimension: math.Log(4) / math.Log(3),
		},
		{
			label:     "KochCurveMultiFractal",
			minLevel:  4,
			maxLevel:  8,
			dimension: math.Log(4) / math.Log(3),
		},
		{
			label:     "CesaroFractal",
			minLevel:  7,
			maxLevel:  7,
			dimension: 1.8,
		},
		{
			label:     "CesaroMultiFractal",
			minLevel:  3,
			maxLevel:  6,
			dimension: 1.8,
		},
	}

	// Constructs a fractal and then computes various metrics (number of
	// vertices, total length, minimum and maximum radius) and verifies that
	// they are within expected tolerances. Essentially this
	// directly verifies that the shape constructed *is* a fractal, i.e. the
	// total length of the curve increases exponentially with the level, while
	// the area bounded by the fractal is more or less constant.

	// The radius needs to be fairly small to avoid spherical distortions.
	const nominalRadius = 0.001 // radians, or about 6km
	const distortionError = 1e-5

	for _, test := range tests {
		f := newFractal()
		f.minLevel = test.minLevel
		f.maxLevel = test.maxLevel
		f.dimension = test.dimension

		frame := randomFrame()
		loop := f.makeLoop(frame, nominalRadius)

		if err := loop.Validate(); err != nil {
			t.Errorf("%s. fractal loop was not valid: %v", test.label, err)
		}

		// If minLevel and maxLevel are not equal, then the number of vertices and
		// the total length of the curve are subject to random variation.  Here we
		// compute an approximation of the standard deviation relative to the mean,
		// noting that most of the variance is due to the random choices about
		// whether to stop subdividing at minLevel or not. (The random choices
		// at higher levels contribute progressively less and less to the variance.)
		// The relativeError below corresponds to *one* standard deviation of
		// error; it can be increased to a higher multiple if necessary.
		//
		// Details: Let n=3*(4**minLevel) and k=(maxLevel-minLevel+1). Each of
		// the n edges at minLevel stops subdividing at that level with
		// probability (1/k). This gives a binomial distribution with mean u=(n/k)
		// and standard deviation s=sqrt((n/k)(1-1/k)). The relative error (s/u)
		// can be simplified to sqrt((k-1)/n).
		numLevels := test.maxLevel - test.minLevel + 1
		minVertices := numVerticesAtLevel(test.minLevel)
		relativeError := math.Sqrt((float64(numLevels) - 1.0) / float64(minVertices))

		// expansionFactor is the total fractal length at level n+1 divided by
		// the total fractal length at level n.
		expansionFactor := math.Pow(4, 1-1/test.dimension)
		expectedNumVertices := 0.0
		expectedLengthSum := 0.0

		// trianglePerim is the perimeter of the original equilateral triangle
		// before any subdivision occurs.
		trianglePerim := 3 * math.Sqrt(3) * math.Tan(nominalRadius)
		minLengthSum := trianglePerim * math.Pow(expansionFactor, float64(test.minLevel))
		for level := test.minLevel; level <= test.maxLevel; level++ {
			expectedNumVertices += float64(numVerticesAtLevel(level))
			expectedLengthSum += math.Pow(expansionFactor, float64(level))
		}
		expectedNumVertices /= float64(numLevels)
		expectedLengthSum *= trianglePerim / float64(numLevels)

		if got, want := loop.NumVertices(), minVertices; got < want {
			t.Errorf("%s. number of vertices = %d, should be more than %d", test.label, got, want)
		}
		if got, want := loop.NumVertices(), numVerticesAtLevel(test.maxLevel); got > want {
			t.Errorf("%s. number of vertices = %d, should be less than %d", test.label, got, want)
		}
		if got, want := expectedNumVertices, float64(loop.NumVertices()); !float64Near(got, want, relativeError*(expectedNumVertices-float64(minVertices))) {
			t.Errorf("%s. expected number of vertices %v should be close to %v, difference: %v", test.label, got, want, (got - want))
		}

		center := frame.col(2)
		minRadius := 2 * math.Pi
		maxRadius := 0.0
		lengthSum := s1.Angle(0.0)
		for i := 0; i < loop.NumVertices(); i++ {
			// Measure the radius of the fractal in the tangent plane at center.
			r := math.Tan(center.Angle(loop.Vertex(i).Vector).Radians())
			minRadius = math.Min(minRadius, r)
			maxRadius = math.Max(maxRadius, r)
			lengthSum += loop.Vertex(i).Angle(loop.Vertex(i + 1).Vector)
		}

		// vertexError is an approximate bound on the error when computing vertex
		// positions of the fractal (due to fromFrame, trig calculations, etc).
		const vertexError = 1e-14

		// Although minRadiusFactor() is only a lower bound in general, it happens
		// to be exact (to within numerical errors) unless the dimension is in the
		// range (1.0, 1.09).
		if test.dimension == 1.0 || test.dimension >= 1.09 {
			// Expect the min radius to match very closely.
			if got, want := f.minRadiusFactor()*nominalRadius, minRadius; !float64Near(got, want, vertexError) {
				t.Errorf("%s. minRadiusFactor()*nominalRadius = %v, want ~%v", test.label, got, want)
			}
		} else {
			// Expect the min radius to satisfy the lower bound.
			if got, want := f.minRadiusFactor()*nominalRadius-vertexError, minRadius; got < want {
				t.Errorf("%s. minRadiusFactor()*nominalRadius = %v, want >= %v", test.label, got, want)
			}
		}
		// maxRadiusFactor() is exact (modulo errors) for all dimensions.
		if got, want := f.maxRadiusFactor()*nominalRadius, maxRadius; !float64Near(got, want, vertexError) {
			t.Errorf("%s. maxRadiusFactor()*nominalRadius = %v, want >= %v", test.label, got, want)
		}

		if got, want := lengthSum.Radians(), expectedLengthSum; !float64Near(got, want, relativeError*(expectedLengthSum-minLengthSum)+distortionError*lengthSum.Radians()) {
			t.Errorf("%s. expected perimieter length = %v, want ~%v", test.label, got, want)
		}
	}
}

// TestChordAngleMaxPointError is located in here to work around circular
// import issues. This s1 test needs s2.Points which wont work with our
// packages. The test is in this file since while it uses Points, it's not
// part of Points methods so it shouldn't be in s2point_test.
func TestChordAngleMaxPointError(t *testing.T) {
	// Check that the error bound returned by s1.MaxPointError() is
	// large enough.
	const iters = 100000
	for iter := 0; iter < iters; iter++ {
		x := randomPoint()
		y := randomPoint()
		if oneIn(10) {
			// Occasionally test a point pair that is nearly identical or antipodal.
			r := s1.Angle(1e-15 * randomFloat64())
			y = InterpolateAtDistance(r, x, y)
			if oneIn(2) {
				y = Point{y.Mul(-1)}
			}
		}
		dist := ChordAngleBetweenPoints(x, y)
		err := dist.MaxPointError()
		if got, want := CompareDistance(x, y, dist.Expanded(err)), 0; got > 0 {
			t.Errorf("CompareDistance(%v, %v, %v.Expanded(%v)) = %v, want <= %v", x, y, dist, err, got, want)
		}
		if got, want := CompareDistance(x, y, dist.Expanded(-err)), 0; got < 0 {
			t.Errorf("CompareDistance(%v, %v, %v.Expanded(-%v)) = %v, want >= %v", x, y, dist, err, got, want)
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
