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

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

func ExampleRect_DistanceToLatLng() {
	r := s2.RectFromLatLng(s2.LatLngFromDegrees(-1, -1)).AddPoint(s2.LatLngFromDegrees(1, 1))

	printDist := func(lat, lng float64) {
		fmt.Printf("%f\n", r.DistanceToLatLng(s2.LatLngFromDegrees(lat, lng))/s1.Degree)
	}

	fmt.Println("Distances next to the rectangle.")
	printDist(-2, 0)
	printDist(0, -2)
	printDist(2, 0)
	printDist(0, 2)

	fmt.Println("Distances beyond the corners of the rectangle.")
	printDist(-2, -2)
	printDist(-2, 2)
	printDist(2, 2)
	printDist(2, -2)

	fmt.Println("Distance within the rectangle.")
	printDist(0, 0)
	printDist(0.5, 0)
	printDist(0, 0.5)
	printDist(-0.5, 0)
	printDist(0, -0.5)

	// Output:
	// Distances next to the rectangle.
	// 1.000000
	// 1.000000
	// 1.000000
	// 1.000000
	// Distances beyond the corners of the rectangle.
	// 1.413962
	// 1.413962
	// 1.413962
	// 1.413962
	// Distance within the rectangle.
	// 0.000000
	// 0.000000
	// 0.000000
	// 0.000000
	// 0.000000
}

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

func ExampleEdgeQuery_FindEdges_findClosestEdges() {
	// Let's start with one or more Polylines that we wish to compare against.
	polylines := []s2.Polyline{
		// This is an iteration = 3 Koch snowflake centered at the
		// center of the continental US.
		s2.Polyline{
			s2.PointFromLatLng(s2.LatLngFromDegrees(47.5467, -103.6035)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(45.9214, -103.7320)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(45.1527, -105.8000)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(44.2866, -103.8538)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(42.6450, -103.9695)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(41.8743, -105.9314)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(42.7141, -107.8226)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(41.0743, -107.8377)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(40.2486, -109.6869)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(39.4333, -107.8521)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(37.7936, -107.8658)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(38.5849, -106.0503)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(37.7058, -104.2841)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(36.0638, -104.3793)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(35.3062, -106.1585)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(34.4284, -104.4703)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(32.8024, -104.5573)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(33.5273, -102.8163)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(32.6053, -101.1982)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(34.2313, -101.0361)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(34.9120, -99.2189)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(33.9382, -97.6134)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(32.3185, -97.8489)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(32.9481, -96.0510)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(31.9449, -94.5321)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(33.5521, -94.2263)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(34.1285, -92.3780)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(35.1678, -93.9070)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(36.7893, -93.5734)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(37.3529, -91.6381)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(36.2777, -90.1050)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(37.8824, -89.6824)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(38.3764, -87.7108)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(39.4869, -89.2407)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(41.0883, -88.7784)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(40.5829, -90.8289)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(41.6608, -92.4765)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(43.2777, -92.0749)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(43.7961, -89.9408)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(44.8865, -91.6533)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(46.4844, -91.2100)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(45.9512, -93.4327)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(46.9863, -95.2792)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(45.3722, -95.6237)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(44.7496, -97.7776)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(45.7189, -99.6629)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(47.3422, -99.4244)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(46.6523, -101.6056)),
		},
	}

	// We will use a point that we wish to find the edges which are closest to it.
	point := s2.PointFromLatLng(s2.LatLngFromDegrees(37.7, -122.5))

	// Load them into a ShapeIndex.
	index := s2.NewShapeIndex()
	for _, l := range polylines {
		index.Add(&l)
	}

	// Create a ClosestEdgeQuery and specify that we want the 7 closest.
	//
	// Note that if you were to request all results, and compare to the results
	// of a FurthestEdgeQuery, the results will not be a complete reversal. This
	// is because the distances being reported are to the closest end of a given
	// edge, while the Furthest query is reporting distances to the farthest end
	// of a given edge.
	q := s2.NewClosestEdgeQuery(index, s2.NewClosestEdgeQueryOptions().MaxResults(7))
	target := s2.NewMinDistanceToPointTarget(point)

	for _, result := range q.FindEdges(target) {
		polylineIndex := result.ShapeID()
		edgeIndex := result.EdgeID()
		distance := result.Distance()
		fmt.Printf("Polyline %d, Edge %d is %0.4f degrees from Point (%0.6f, %0.6f, %0.6f)\n",
			polylineIndex, edgeIndex,
			distance.Angle().Degrees(), point.X, point.Y, point.Z)
	}
	// Output:
	// Polyline 0, Edge 7 is 10.2718 degrees from Point (-0.425124, -0.667311, 0.611527)
	// Polyline 0, Edge 8 is 10.2718 degrees from Point (-0.425124, -0.667311, 0.611527)
	// Polyline 0, Edge 9 is 11.5362 degrees from Point (-0.425124, -0.667311, 0.611527)
	// Polyline 0, Edge 10 is 11.5602 degrees from Point (-0.425124, -0.667311, 0.611527)
	// Polyline 0, Edge 6 is 11.8071 degrees from Point (-0.425124, -0.667311, 0.611527)
	// Polyline 0, Edge 5 is 12.2577 degrees from Point (-0.425124, -0.667311, 0.611527)
	// Polyline 0, Edge 11 is 12.9502 degrees from Point (-0.425124, -0.667311, 0.611527)

}

func ExampleEdgeQuery_FindEdges_findFurthestEdges() {
	// Let's start with one or more Polylines that we wish to compare against.
	polylines := []s2.Polyline{
		// This is an iteration = 3 Koch snowflake centered at the
		// center of the continental US.
		s2.Polyline{
			s2.PointFromLatLng(s2.LatLngFromDegrees(47.5467, -103.6035)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(45.9214, -103.7320)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(45.1527, -105.8000)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(44.2866, -103.8538)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(42.6450, -103.9695)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(41.8743, -105.9314)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(42.7141, -107.8226)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(41.0743, -107.8377)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(40.2486, -109.6869)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(39.4333, -107.8521)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(37.7936, -107.8658)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(38.5849, -106.0503)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(37.7058, -104.2841)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(36.0638, -104.3793)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(35.3062, -106.1585)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(34.4284, -104.4703)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(32.8024, -104.5573)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(33.5273, -102.8163)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(32.6053, -101.1982)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(34.2313, -101.0361)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(34.9120, -99.2189)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(33.9382, -97.6134)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(32.3185, -97.8489)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(32.9481, -96.0510)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(31.9449, -94.5321)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(33.5521, -94.2263)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(34.1285, -92.3780)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(35.1678, -93.9070)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(36.7893, -93.5734)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(37.3529, -91.6381)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(36.2777, -90.1050)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(37.8824, -89.6824)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(38.3764, -87.7108)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(39.4869, -89.2407)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(41.0883, -88.7784)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(40.5829, -90.8289)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(41.6608, -92.4765)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(43.2777, -92.0749)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(43.7961, -89.9408)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(44.8865, -91.6533)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(46.4844, -91.2100)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(45.9512, -93.4327)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(46.9863, -95.2792)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(45.3722, -95.6237)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(44.7496, -97.7776)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(45.7189, -99.6629)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(47.3422, -99.4244)),
			s2.PointFromLatLng(s2.LatLngFromDegrees(46.6523, -101.6056)),
		},
	}

	// We will use a point that we want to find the edges which are furthest from it.
	point := s2.PointFromLatLng(s2.LatLngFromDegrees(37.7, -122.5))

	// Load them into a ShapeIndex
	index := s2.NewShapeIndex()
	for _, l := range polylines {
		index.Add(&l)
	}

	// Create a FurthestEdgeQuery and specify that we want the 3 furthest.
	q := s2.NewFurthestEdgeQuery(index, s2.NewFurthestEdgeQueryOptions().MaxResults(3))
	target := s2.NewMaxDistanceToPointTarget(point)

	for _, result := range q.FindEdges(target) {
		polylineIndex := result.ShapeID()
		edgeIndex := result.EdgeID()
		distance := result.Distance()
		fmt.Printf("Polyline %d, Edge %d is %0.3f degrees from Point (%0.6f, %0.6f, %0.6f)\n",
			polylineIndex, edgeIndex,
			distance.Angle().Degrees(), point.X, point.Y, point.Z)
	}
	// Output:
	// Polyline 0, Edge 31 is 27.245 degrees from Point (-0.425124, -0.667311, 0.611527)
	// Polyline 0, Edge 32 is 27.245 degrees from Point (-0.425124, -0.667311, 0.611527)
	// Polyline 0, Edge 33 is 26.115 degrees from Point (-0.425124, -0.667311, 0.611527)

}
