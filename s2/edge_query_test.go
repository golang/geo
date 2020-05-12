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
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/golang/geo/s1"
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

type queryTargetType int

const (
	queryTypePoint queryTargetType = iota
	queryTypeEdge
	queryTypeCell
	queryTypeIndex
)

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

// benchmarkEdgeQueryFindClosest calls FindEdges the given number of times on
// a ShapeIndex with approximately numIndexEdges edges generated by the given
// generator. The geometry is generated within a Cap of the radius given.
//
// Each query uses a target of the given targetType.
//
//   - If maxDistanceFraction > 0, then DistanceLimit is set to the given
//     fraction of the index radius.
//
//   - If maxErrorFraction > 0, then MaxError is set to the given
//     fraction of the index radius.
//
// TODO(roberts): If there is a need to benchmark Furthest as well, this will need
// some changes to not use just the Closest variants of parts.
// Furthest isn't doing anything different under the covers than Closest, so there
// isn't really a huge need for benchmarking both.
func benchmarkEdgeQueryFindClosest(b *testing.B, bmOpts *edgeQueryBenchmarkOptions) {
	const numIndexSamples = 8

	b.StopTimer()
	index := NewShapeIndex()
	opts := NewClosestEdgeQueryOptions().MaxResults(1).IncludeInteriors(bmOpts.includeInteriors)

	radius := kmToAngle(bmOpts.radiusKm.Radians())
	if bmOpts.maxDistanceFraction > 0 {
		opts.DistanceLimit(s1.ChordAngleFromAngle(s1.Angle(bmOpts.maxDistanceFraction) * radius))
	}
	if bmOpts.maxErrorFraction > 0 {
		opts.MaxError(s1.ChordAngleFromAngle(s1.Angle(bmOpts.maxErrorFraction) * radius))
	}

	opts.UseBruteForce(*benchmarkBruteForce)
	query := NewClosestEdgeQuery(index, opts)

	delta := 0 // Bresenham-type algorithm for geometry sampling.
	var targets []distanceTarget

	// To follow the sizing on the C++ tests to ease comparisons, the number of
	// edges in the index range on 3 * 4^n (up to 16384).
	bmOpts.numIndexEdges = 3
	for n := 1; n <= 7; n++ {
		bmOpts.numIndexEdges *= 4
		b.Run(fmt.Sprintf("%d", bmOpts.numIndexEdges),
			func(b *testing.B) {
				iTarget := 0
				for i := 0; i < b.N; i++ {
					delta -= numIndexSamples
					if delta < 0 {
						// Generate a new index and a new set of
						// targets to go with it. Reset the random
						// number seed so that we use the same sequence
						// of indexed shapes no matter how many
						// iterations are specified.
						b.StopTimer()
						delta += i
						targets, _ = generateEdgeQueryWithTargets(bmOpts, query, index)
						b.StartTimer()
					}
					query.FindEdges(targets[iTarget])
					iTarget++
					if iTarget == len(targets) {
						iTarget = 0
					}
				}
			})
	}
}

// edgeQueryBenchmarkOptions holds the various parameters than can be adjusted by the
// benchmarking runners.
type edgeQueryBenchmarkOptions struct {
	iters                    int
	fact                     shapeIndexGeneratorFunc
	numIndexEdges            int
	includeInteriors         bool
	targetType               queryTargetType
	numTargetEdges           int
	chooseTargetFromIndex    bool
	radiusKm                 s1.Angle
	maxDistanceFraction      float64
	maxErrorFraction         float64
	targetRadiusFraction     float64
	centerSeparationFraction float64
	randomSeed               int64
}

// generateEdgeQueryWithTargets generates and adds geometry to a ShapeIndex for
// use in an edge query.
//
// Approximately numIndexEdges will be generated by the requested generator and
// inserted. The geometry is generated within a Cap of the radius specified
// by radiusKm (the index radius). Parameters with fraction in their
// names are expressed as a fraction of this radius.
//
// Also generates a set of target geometries for the query, based on the
// targetType and the input parameters. If targetType is INDEX, then:
//   (i) the target will have approximately numTargetEdges edges.
//   (ii) includeInteriors will be set on the target index.
//
//   - If chooseTargetFromIndex is true, then the target will be chosen
//     from the geometry in the index itself, otherwise it will be chosen
//     randomly according to the parameters below:
//
//   - If targetRadiusFraction > 0, the target radius will be approximately
//     the given fraction of the index radius; if targetRadiusFraction < 0,
//     it will be chosen randomly up to corresponding positive fraction.
//
//   - If centerSeparationFraction > 0, then the centers of index and target
//     bounding caps will be separated by the given fraction of the index
//     radius; if centerSeparationFraction < 0, they will be separated by up
//     to the corresponding positive fraction.
//
//   - The randomSeed is used to initialize an internal seed, which is
//     incremented at the start of each call to generateEdgeQueryWithTargets.
//     This is for debugging purposes.
//
func generateEdgeQueryWithTargets(opts *edgeQueryBenchmarkOptions, query *EdgeQuery, queryIndex *ShapeIndex) (targets []distanceTarget, targetIndexes []*ShapeIndex) {

	// To save time, we generate at most this many distinct targets per index.
	const maxTargetsPerIndex = 100

	// Set a specific seed to allow repeatabilty
	rand.Seed(opts.randomSeed)
	opts.randomSeed++
	indexCap := CapFromCenterAngle(randomPoint(), opts.radiusKm)

	query.Reset()
	queryIndex.Reset()
	opts.fact(indexCap, opts.numIndexEdges, queryIndex)

	targets = make([]distanceTarget, 0)
	targetIndexes = make([]*ShapeIndex, 0)

	numTargets := maxTargetsPerIndex
	if opts.targetType == queryTypeIndex {
		// Limit the total number of target edges to reduce the benchmark running times.
		numTargets = minInt(numTargets, 500000/opts.numTargetEdges)
	}

	for i := 0; i < numTargets; i++ {
		targetDist := fractionToRadius(opts.centerSeparationFraction, opts.radiusKm.Radians())
		targetCap := CapFromCenterAngle(
			sampleBoundaryFromCap(CapFromCenterAngle(indexCap.Center(), targetDist)),
			fractionToRadius(opts.targetRadiusFraction, opts.radiusKm.Radians()))

		switch opts.targetType {
		case queryTypePoint:
			var pt Point
			if opts.chooseTargetFromIndex {
				pt = sampleEdgeFromIndex(queryIndex).V0
			} else {
				pt = targetCap.Center()
			}
			targets = append(targets, NewMinDistanceToPointTarget(pt))

		case queryTypeEdge:
			var edge Edge
			if opts.chooseTargetFromIndex {
				edge = sampleEdgeFromIndex(queryIndex)
			} else {
				edge.V0 = sampleBoundaryFromCap(targetCap)
				edge.V1 = sampleBoundaryFromCap(targetCap)
			}
			targets = append(targets, NewMinDistanceToEdgeTarget(edge))
		case queryTypeCell:
			var cellID CellID
			if opts.chooseTargetFromIndex {
				cellID = sampleCellFromIndex(queryIndex)
			} else {
				cellID = cellIDFromPoint(targetCap.Center()).Parent(
					MaxDiagMetric.ClosestLevel(targetCap.Radius().Radians()))
			}
			targets = append(targets, NewMinDistanceToCellTarget(CellFromCellID(cellID)))
		case queryTypeIndex:
			targetIndex := NewShapeIndex()
			if opts.chooseTargetFromIndex {
				var shape edgeVectorShape
				for i := 0; i < opts.numTargetEdges; i++ {
					edge := sampleEdgeFromIndex(queryIndex)
					shape.Add(edge.V0, edge.V1)
				}
				targetIndex.Add(&shape)
			} else {
				opts.fact(targetCap, opts.numTargetEdges, targetIndex)
			}
			target := NewMinDistanceToShapeIndexTarget(targetIndex)
			target.setIncludeInteriors(opts.includeInteriors)
			targets = append(targets, target)
		default:
			panic(fmt.Sprintf("unknown query target type %v", opts.targetType))
		}
	}

	return targets, targetIndexes
}

func sampleBoundaryFromCap(c Cap) Point {
	return InterpolateAtDistance(c.Radius(), c.Center(), randomPoint())
}

func sampleEdgeFromIndex(index *ShapeIndex) Edge {
	e := randomUniformInt(index.NumEdges())

	for _, shape := range index.shapes {
		if e < shape.NumEdges() {
			return shape.Edge(e)
		}
		e -= shape.NumEdges()
	}
	// This should only happen if the index has no edges at all.
	panic("index with no edges")
}

func sampleCellFromIndex(index *ShapeIndex) CellID {
	iter := index.Iterator()
	for i := randomUniformInt(len(index.cells)); i >= 0; i-- {
		iter.Next()
		continue
	}
	return iter.CellID()
}

func fractionToRadius(fraction, radiusKm float64) s1.Angle {
	if fraction < 0 {
		fraction = -randomFloat64() * fraction
	}
	return s1.Angle(fraction) * kmToAngle(radiusKm)
}
