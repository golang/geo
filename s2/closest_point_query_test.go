// Copyright 2015 Google Inc. All rights reserved.
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

	"github.com/golang/geo/s1"
)

func TestClosestPointQueryNoPoints(t *testing.T) {
	index := &PointIndex[int]{}
	query := NewClosestPointQuery(index, nil)
	target := NewMinDistanceToPointTarget(PointFromCoords(1, 0, 0))
	if got := query.FindClosestPoints(target); len(got) != 0 {
		t.Errorf("FindClosestPoints on empty index: got %d results, want 0", len(got))
	}
	if got := query.FindClosestPoint(target); !got.IsEmpty() {
		t.Error("FindClosestPoint on empty index should be empty")
	}
	if got := query.GetDistance(target); got != s1.InfChordAngle() {
		t.Errorf("GetDistance on empty index = %v, want InfChordAngle", got)
	}
}

func TestClosestPointQueryManyDuplicatePoints(t *testing.T) {
	const numPoints = 10000
	p := PointFromCoords(1, 0, 0)
	index := &PointIndex[int]{}
	for i := range numPoints {
		index.Add(p, i)
	}
	query := NewClosestPointQuery(index, nil)
	target := NewMinDistanceToPointTarget(p)
	results := query.FindClosestPoints(target)
	if got := len(results); got != numPoints {
		t.Errorf("FindClosestPoints on %d duplicates: got %d results, want %d", numPoints, got, numPoints)
	}
	for i, r := range results {
		if r.IsEmpty() {
			t.Errorf("result[%d].IsEmpty() = true", i)
		}
	}
}

func TestClosestPointQueryEmptyTarget(t *testing.T) {
	index := &PointIndex[int]{}
	for i := range 1000 {
		index.Add(randomPoint(), i)
	}
	query := NewClosestPointQuery(index, NewClosestPointQueryOptions().
		DistanceLimit(s1.ChordAngleFromAngle(1e-5*s1.Radian)))
	emptyTarget := NewMinDistanceToShapeIndexTarget(NewShapeIndex())
	if got := len(query.FindClosestPoints(emptyTarget)); got != 0 {
		t.Errorf("FindClosestPoints with empty target: got %d results, want 0", got)
	}
}

// testClosestPointQuery verifies that brute-force and optimized algorithms
// return equivalent results for the given target and query options.
func testClosestPointQuery[Data comparable](t *testing.T, target distanceTarget, query *ClosestPointQuery[Data]) {
	t.Helper()

	query.opts.useBruteForce = true
	expected := query.FindClosestPoints(target)
	query.opts.useBruteForce = false
	actual := query.FindClosestPoints(target)

	maxResults := query.opts.maxResults
	maxErr := query.opts.maxError
	distLimit := query.opts.distanceLimit

	if len(actual) > maxResults {
		t.Errorf("got %d results, want <= %d", len(actual), maxResults)
	}

	// If no distance limit and no region, the count must match exactly.
	if distLimit == s1.InfChordAngle() && query.opts.region == nil {
		want := min(maxResults, query.index.NumPoints())
		if len(actual) != want {
			t.Errorf("got %d results, want %d (maxResults=%d, numPoints=%d)", len(actual), want, maxResults, query.index.NumPoints())
		}
	}

	// All returned results must have distance < distLimit.
	for _, r := range actual {
		if r.Distance() >= distLimit {
			t.Errorf("result distance %v >= limit %v", r.Distance(), distLimit)
		}
	}

	// The brute-force minimum distance must be within max_error of the
	// optimized minimum distance.
	if len(expected) > 0 && len(actual) > 0 {
		minExpected := expected[0].Distance()
		minActual := actual[0].Distance()
		if minActual > minExpected+maxErr {
			t.Errorf("optimized min dist %v > brute force %v + maxError %v", minActual, minExpected, maxErr)
		}
	}
}

func TestClosestPointQueryBruteForceVsOptimized(t *testing.T) {
	const (
		numIndexes  = 10
		numPoints   = 100
		numQueries  = 50
		testCapKm   = 10.0
		earthRadius = 6371.0 // km
	)
	capAngle := s1.Angle(testCapKm/earthRadius) * s1.Radian

	for range numIndexes {
		center := randomPoint()
		indexCap := CapFromCenterAngle(center, capAngle)
		index := &PointIndex[int]{}
		for i := range numPoints {
			p := samplePointFromCap(indexCap)
			index.Add(p, i)
		}

		for range numQueries {
			queryRadius := 2 * capAngle
			queryCap := CapFromCenterAngle(center, queryRadius)
			query := NewClosestPointQuery(index, nil)

			// Vary the options.
			if randomUniformInt(5) != 0 {
				query.opts.maxResults = 1 + randomUniformInt(10)
			}
			if randomUniformInt(3) != 0 {
				frac := randomUniformFloat64(0, 1)
				query.opts.distanceLimit = s1.ChordAngleFromAngle(s1.Angle(frac) * queryRadius)
			}
			if randomUniformInt(2) != 0 {
				maxErrFrac := 1e-4 + math.Exp(randomUniformFloat64(0, 1)*math.Log(1.0))
				query.opts.maxError = s1.ChordAngleFromAngle(s1.Angle(maxErrFrac) * queryRadius)
			}

			targetType := randomUniformInt(3)
			switch targetType {
			case 0:
				p := samplePointFromCap(queryCap)
				testClosestPointQuery(t, NewMinDistanceToPointTarget(p), query)
			case 1:
				a := samplePointFromCap(queryCap)
				bCap := CapFromCenterAngle(a, s1.Angle(1e-4)*queryRadius)
				b := samplePointFromCap(bCap)
				testClosestPointQuery(t, NewMinDistanceToEdgeTarget(Edge{a, b}), query)
			case 2:
				minLevel := MaxLevel - 4
				level := minLevel + randomUniformInt(5)
				cellID := cellIDFromPoint(samplePointFromCap(queryCap)).Parent(level)
				testClosestPointQuery(t, NewMinDistanceToCellTarget(CellFromCellID(cellID)), query)
			}
		}
	}
}

func TestClosestPointQueryFindClosestPoint(t *testing.T) {
	p0 := parsePoint("0:0")
	p1 := parsePoint("1:0")
	p2 := parsePoint("2:0")

	index := &PointIndex[int]{}
	index.Add(p0, 0)
	index.Add(p1, 1)
	index.Add(p2, 2)

	query := NewClosestPointQuery(index, nil)
	target := NewMinDistanceToPointTarget(parsePoint("1.1:0"))

	result := query.FindClosestPoint(target)
	if result.IsEmpty() {
		t.Fatal("FindClosestPoint returned empty result")
	}
	if result.Data() != 1 {
		t.Errorf("closest point data = %d, want 1", result.Data())
	}
	if result.Point() != p1 {
		t.Errorf("closest point = %v, want %v", result.Point(), p1)
	}
}

func TestClosestPointQueryDistanceLimitAndMaxResults(t *testing.T) {
	index := &PointIndex[int]{}
	for i := range 100 {
		index.Add(randomPoint(), i)
	}

	query := NewClosestPointQuery(index, NewClosestPointQueryOptions().
		MaxResults(5).
		DistanceLimit(s1.ChordAngleFromAngle(s1.InfAngle())))
	target := NewMinDistanceToPointTarget(randomPoint())

	results := query.FindClosestPoints(target)
	if len(results) > 5 {
		t.Errorf("got %d results, want <= 5", len(results))
	}
	// Verify results are sorted by distance.
	for i := 1; i < len(results); i++ {
		if results[i].Distance() < results[i-1].Distance() {
			t.Errorf("results not sorted: results[%d].Distance=%v < results[%d].Distance=%v",
				i, results[i].Distance(), i-1, results[i-1].Distance())
		}
	}
}

func TestClosestPointQueryIsDistanceLess(t *testing.T) {
	p0 := parsePoint("23:12")
	p1 := parsePoint("47:11")

	index := &PointIndex[int]{}
	index.Add(p0, 0)

	query := NewClosestPointQuery(index, nil)
	target := NewMinDistanceToPointTarget(p0)

	// Distance to p0 is zero.
	zeroAngle := s1.ChordAngle(0)
	if query.IsDistanceLess(target, zeroAngle) {
		t.Error("IsDistanceLess(p0, 0): want false for distance 0")
	}
	if !query.IsDistanceLessOrEqual(target, zeroAngle) {
		t.Error("IsDistanceLessOrEqual(p0, 0): want true")
	}
	if !query.IsConservativeDistanceLessOrEqual(target, zeroAngle) {
		t.Error("IsConservativeDistanceLessOrEqual(p0, 0): want true")
	}

	// Distance to p1 is positive.
	target1 := NewMinDistanceToPointTarget(p1)
	d01 := ChordAngleBetweenPoints(p0, p1)
	if !query.IsDistanceLess(target1, d01.Successor()) {
		t.Error("IsDistanceLess(p1, d01.Successor): want true")
	}
	if query.IsDistanceLess(target1, d01) {
		t.Error("IsDistanceLess(p1, d01): want false (equal, not less)")
	}
	if !query.IsDistanceLessOrEqual(target1, d01) {
		t.Error("IsDistanceLessOrEqual(p1, d01): want true")
	}
	if !query.IsConservativeDistanceLessOrEqual(target1, d01) {
		t.Error("IsConservativeDistanceLessOrEqual(p1, d01): want true")
	}
}

func TestClosestPointQueryReInit(t *testing.T) {
	p := PointFromCoords(1, 0, 0)
	index := &PointIndex[int]{}
	query := NewClosestPointQuery(index, nil)
	target := NewMinDistanceToPointTarget(p)

	if got := len(query.FindClosestPoints(target)); got != 0 {
		t.Errorf("empty index: got %d results, want 0", got)
	}

	index.Add(p, 42)
	query.ReInit()

	results := query.FindClosestPoints(target)
	if got := len(results); got != 1 {
		t.Fatalf("after ReInit: got %d results, want 1", got)
	}
	if results[0].Data() != 42 {
		t.Errorf("result.Data() = %d, want 42", results[0].Data())
	}
}
