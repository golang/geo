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

package s2

import (
	"reflect"
	"testing"
)

// Note that most of the actual testing is done in s2edge_query_{closest|furthest}_test.

func TestEdgeQueryMaxDistance(t *testing.T) {
	index := makeShapeIndex("0:0 | 1:0 | 2:0 | 3:0 # #")
	query := NewFurthestEdgeQuery(index, nil)
	target := NewMaxDistanceToPointTarget(parsePoint("4:0"))
	results := query.findEdges(target, NewFurthestEdgeQueryOptions().MaxResults(1).common)

	if len(results) != 1 {
		t.Errorf("len(results) = %v, want 1: %+v", len(results), results)
		return
	}

	if results[0].shapeID != 0 {
		t.Errorf("shapeID should be 0 got %v", results[0].shapeID)

	}

	if results[0].edgeID != 0 {
		t.Errorf("edgeID should be 0, got %v", results[0].edgeID)

	}

	if got, want := results[0].Distance().Angle().Degrees(), 4.0; !float64Near(got, want, 1e-13) {
		t.Errorf("results[0].Distance = %v, want ~%v", got, want)
	}
}

func TestEdgeQuerySortAndUnique(t *testing.T) {
	tests := []struct {
		have []EdgeQueryResult
		want []EdgeQueryResult
	}{
		{
			// one result gets doesn't change
			have: []EdgeQueryResult{
				{distance: minDistance(0.0790858), shapeID: 0, edgeID: 0},
			},
			want: []EdgeQueryResult{
				{distance: minDistance(0.0790858), shapeID: 0, edgeID: 0},
			},
		},
		{
			// tied result in same shape should be ordered by edge id.
			have: []EdgeQueryResult{
				{distance: minDistance(0.0639643), shapeID: 0, edgeID: 4},
				{distance: minDistance(0.0639643), shapeID: 0, edgeID: 2},
			},
			want: []EdgeQueryResult{
				{distance: minDistance(0.0639643), shapeID: 0, edgeID: 2},
				{distance: minDistance(0.0639643), shapeID: 0, edgeID: 4},
			},
		},
		{
			// more than one result in same shape.
			have: []EdgeQueryResult{
				{distance: minDistance(0.0639643), shapeID: 0, edgeID: 1},
				{distance: minDistance(0.0508695), shapeID: 0, edgeID: 11},
				{distance: minDistance(0.1181251), shapeID: 0, edgeID: 43},
			},
			want: []EdgeQueryResult{
				{distance: minDistance(0.0508695), shapeID: 0, edgeID: 11},
				{distance: minDistance(0.0639643), shapeID: 0, edgeID: 1},
				{distance: minDistance(0.1181251), shapeID: 0, edgeID: 43},
			},
		},
		{
			// more than one shape has the equal distance edge,
			// order should be by shape by edge.
			have: []EdgeQueryResult{
				{distance: minDistance(0.0639643), shapeID: 4, edgeID: 2},
				{distance: minDistance(0.0639643), shapeID: 3, edgeID: 8},
				{distance: minDistance(0.0508695), shapeID: 0, edgeID: 11},
				{distance: minDistance(0.1181251), shapeID: 0, edgeID: 43},
			},
			want: []EdgeQueryResult{
				{distance: minDistance(0.0508695), shapeID: 0, edgeID: 11},
				{distance: minDistance(0.0639643), shapeID: 3, edgeID: 8},
				{distance: minDistance(0.0639643), shapeID: 4, edgeID: 2},
				{distance: minDistance(0.1181251), shapeID: 0, edgeID: 43},
			},
		},
		{
			// larger set of results.
			have: []EdgeQueryResult{
				{distance: minDistance(0.0790858), shapeID: 0, edgeID: 0},
				{distance: minDistance(0.0639643), shapeID: 0, edgeID: 1},
				{distance: minDistance(0.0639643), shapeID: 0, edgeID: 2},
				{distance: minDistance(0.0677918), shapeID: 0, edgeID: 3},
				{distance: minDistance(0.0542300), shapeID: 0, edgeID: 4},
				{distance: minDistance(0.0455950), shapeID: 0, edgeID: 5},
				{distance: minDistance(0.0423160), shapeID: 0, edgeID: 6},
				{distance: minDistance(0.0320540), shapeID: 0, edgeID: 7},
				{distance: minDistance(0.0320540), shapeID: 0, edgeID: 8},
				{distance: minDistance(0.0404029), shapeID: 0, edgeID: 9},
				{distance: minDistance(0.0405702), shapeID: 0, edgeID: 10},
				{distance: minDistance(0.0508695), shapeID: 0, edgeID: 11},
				{distance: minDistance(0.0627421), shapeID: 0, edgeID: 12},
				{distance: minDistance(0.0539154), shapeID: 0, edgeID: 13},
				{distance: minDistance(0.1181251), shapeID: 0, edgeID: 43},
				{distance: minDistance(0.1061612), shapeID: 0, edgeID: 44},
				{distance: minDistance(0.1061612), shapeID: 0, edgeID: 45},
				{distance: minDistance(0.0957947), shapeID: 0, edgeID: 46},
			},
			want: []EdgeQueryResult{
				{distance: minDistance(0.0320540), shapeID: 0, edgeID: 7},
				{distance: minDistance(0.0320540), shapeID: 0, edgeID: 8},
				{distance: minDistance(0.0404029), shapeID: 0, edgeID: 9},
				{distance: minDistance(0.0405702), shapeID: 0, edgeID: 10},
				{distance: minDistance(0.0423160), shapeID: 0, edgeID: 6},
				{distance: minDistance(0.0455950), shapeID: 0, edgeID: 5},
				{distance: minDistance(0.0508695), shapeID: 0, edgeID: 11},
				{distance: minDistance(0.0539154), shapeID: 0, edgeID: 13},
				{distance: minDistance(0.0542300), shapeID: 0, edgeID: 4},
				{distance: minDistance(0.0627421), shapeID: 0, edgeID: 12},
				{distance: minDistance(0.0639643), shapeID: 0, edgeID: 1},
				{distance: minDistance(0.0639643), shapeID: 0, edgeID: 2},
				{distance: minDistance(0.0677918), shapeID: 0, edgeID: 3},
				{distance: minDistance(0.0790858), shapeID: 0, edgeID: 0},
				{distance: minDistance(0.0957947), shapeID: 0, edgeID: 46},
				{distance: minDistance(0.1061612), shapeID: 0, edgeID: 44},
				{distance: minDistance(0.1061612), shapeID: 0, edgeID: 45},
				{distance: minDistance(0.1181251), shapeID: 0, edgeID: 43},
			},
		},
	}

	for _, test := range tests {
		have := append([]EdgeQueryResult{}, test.have...)
		got := sortAndUniqueResults(have)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("sortAndUniqueResults(%v) =\n %v, \nwant %v", test.have, got, test.want)
		}
	}
}

// For various tests and benchmarks on the edge query code, there are a number of
// ShapeIndex generators that can be used.
type shapeIndexGeneratorFunc func(c Cap, numEdges int, index *ShapeIndex)

// loopShapeIndexGenerator generates a regular loop that approximately fills
// the given Cap.
//
// Regular loops are nearly the worst case for distance calculations, since
// many edges are nearly equidistant from any query point that is not
// immediately adjacent to the loop.
func loopShapeIndexGenerator(c Cap, numEdges int, index *ShapeIndex) {
	index.Add(RegularLoop(c.Center(), c.Radius(), numEdges))
}

// fractalLoopShapeIndexGenerator generates a fractal loop that approximately
// fills the given Cap.
func fractalLoopShapeIndexGenerator(c Cap, numEdges int, index *ShapeIndex) {
	fractal := newFractal()
	fractal.setLevelForApproxMaxEdges(numEdges)
	index.Add(fractal.makeLoop(randomFrameAtPoint(c.Center()), c.Radius()))
}

// pointCloudShapeIndexGenerator generates a cloud of points that approximately
// fills the given Cap.
func pointCloudShapeIndexGenerator(c Cap, numPoints int, index *ShapeIndex) {
	var points PointVector
	for i := 0; i < numPoints; i++ {
		points = append(points, samplePointFromCap(c))
	}
	index.Add(&points)
}

const edgeQueryTestNumIndexes = 50
const edgeQueryTestNumEdges = 100
const edgeQueryTestNumQueries = 200

// The approximate radius of Cap from which query edges are chosen.
var testCapRadius = kmToAngle(10)

/*
// testEdgeQueryWithGenerator is used to perform high volume random testing on EdqeQuery
// using a variety of index generation methods and varying sizes.
//
// The running time of this test is proportional to
//    (numIndexes + numQueries) * numEdges.
// Every query is checked using the brute force algorithm.
func testEdgeQueryWithGenerator(t *testing.T,
	newQueryFunc func(si *ShapeIndex, opts *EdgeQueryOptions) *EdgeQuery,
	newOptsFunc func() *EdgeQueryOptions,
	gen shapeIndexGeneratorFunc,
	numIndexes, numEdges, numQueries int) {

	// Build a set of ShapeIndexes containing the desired geometry.
	var indexCaps []Cap
	var indexes []*ShapeIndex
	for i := 0; i < numIndexes; i++ {
		rand.Seed(int64(i))
		indexCaps = append(indexCaps, CapFromCenterAngle(randomPoint(), testCapRadius))
		indexes = append(indexes, NewShapeIndex())
		gen(indexCaps[i], numEdges, indexes[i])
	}

	for i := 0; i < numQueries; i++ {
		rand.Seed(int64(i))
		iIndex := randomUniformInt(numIndexes)
		indexCap := indexCaps[iIndex]

		// Choose query points from an area approximately 4x larger than the
		// geometry being tested.
		queryRadius := 2 * indexCap.Radius()
		// Exercise the opposite-hemisphere code 1/5 of the time.
		antipodal := 1.0
		if oneIn(5) {
			antipodal = -1
		}
		//queryCap := CapFromCenterAngle(indexCap.Center(), queryRadius)
		queryCap := CapFromCenterAngle(Point{indexCap.Center().Mul(antipodal)}, queryRadius)

		opts := newOptsFunc()

		// Occasionally we don't set any limit on the number of result edges.
		// (This may return all edges if we also don't set a distance limit.)
		if oneIn(5) {
			opts.MaxResults(1 + randomUniformInt(10))
		}

		// We set a distance limit 1/3 of the time.
		if oneIn(3) {
			opts.DistanceLimit(s1.ChordAngleFromAngle(s1.Angle(randomFloat64()) * queryRadius))
		}
		if oneIn(2) {
			// Choose a maximum error whose logarithm is uniformly distributed over
			// a reasonable range, except that it is sometimes zero.
			opts.MaxError(s1.ChordAngleFromAngle(s1.Angle(math.Pow(1e-4, randomFloat64()) * queryRadius.Radians())))
		}
		opts.IncludeInteriors(oneIn(2))

		query := newQueryFunc(indexes[iIndex], opts)

		switch randomUniformInt(4) {
		case 0:
			// Find the edges furthest from a given point.
			point := samplePointFromCap(queryCap)
			target := NewMaxDistanceToPointTarget(point)
			testFindEdges(target, query)
		case 1:
			// Find the edges furthest from a given edge.
			a := samplePointFromCap(queryCap)
			b := samplePointFromCap(
				CapFromCenterAngle(a, s1.Angle(math.Pow(1e-4, randomFloat64()))*queryRadius))
			target := NewMaxDistanceToEdgeTarget(Edge{a, b})
			testFindEdges(target, query)

		case 2:
			// Find the edges furthest from a given cell.
			minLevel := MaxDiagMetric.MinLevel(queryRadius.Radians())
			level := minLevel + randomUniformInt(maxLevel-minLevel+1)
			a := samplePointFromCap(queryCap)
			cell := CellFromCellID(cellIDFromPoint(a).Parent(level))
			target := NewMaxDistanceToCellTarget(cell)
			testFindEdges(target, query)

		case 3:
			// Use another one of the pre-built indexes as the target.
			jIndex := randomUniformInt(numIndexes)
			target := NewMaxDistanceToShapeIndexTarget(indexes[jIndex])
			target.setIncludeInteriors(oneIn(2))
			testFindEdges(target, query)

		}
	}
}
*/
