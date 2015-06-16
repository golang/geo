package s2

import (
	"math"
	"math/rand"
	"testing"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/s1"
)

var (
	// A stripe that slightly over-wraps the equator.
	candyCane *Loop = makeLoop("-20:150, -20:-70, 0:70, 10:-150, 10:70, -10:-70")

	// A small clockwise loop in the northern & eastern hemisperes.
	smallNeCw *Loop = makeLoop("35:20, 45:20, 40:25")

	// Loop around the north pole at 80 degrees.
	arctic80 *Loop = makeLoop("80:-150, 80:-30, 80:90")

	// Loop around the south pole at 80 degrees.
	antarctic80 *Loop = makeLoop("-80:120, -80:0, -80:-120")

	// The northern hemisphere, defined using two pairs of antipodal points.
	northHemi *Loop = makeLoop("0:-180, 0:-90, 0:0, 0:90")

	// The northern hemisphere, defined using three points 120 degrees apart.
	northHemi3 *Loop = makeLoop("0:-180, 0:-60, 0:60")

	// The western hemisphere, defined using two pairs of antipodal points.
	westHemi *Loop = makeLoop("0:-180, -90:0, 0:0, 90:0")

	// The "near" hemisphere, defined using two pairs of antipodal points.
	nearHemi *Loop = makeLoop("0:-90, -90:0, 0:90, 90:0")

	// A diamond-shaped loop around the point 0:180.
	loopA *Loop = makeLoop("0:178, -1:180, 0:-179, 1:-180")

	// Another diamond-shaped loop around the point 0:180.
	loopB *Loop = makeLoop("0:179, -1:180, 0:-178, 1:-180")

	// The intersection of A and B.
	aIntersectB *Loop = makeLoop("0:179, -1:180, 0:-179, 1:-180")

	// The union of A and B.
	aUnionB *Loop = makeLoop("0:178, -1:180, 0:-178, 1:-180")

	// A minus B (concave)
	aMinusB *Loop = makeLoop("0:178, -1:180, 0:179, 1:-180")

	// B minus A (concave)
	bMinusA *Loop = makeLoop("0:-179, -1:180, 0:-178, 1:-180")

	// A self-crossing loop with a duplicated vertex
	bowtie *Loop = makeLoop("0:0, 2:0, 1:1, 0:2, 2:2, 1:1")

	// Initialized below.
	southHemi *Loop
	eastHemi  *Loop
	farHemi   *Loop
)

func init() {
	southHemi = makeLoop("0:-180, 0:-90, 0:0, 0:90")
	southHemi.Invert()
	eastHemi = makeLoop("0:-180, -90:0, 0:0, 90:0")
	eastHemi.Invert()
	farHemi = makeLoop("0:-90, -90:0, 0:90, 90:0")
	farHemi.Invert()
}

func makeLoop(s string) *Loop {
	points := parsePoints(s)
	return LoopFromPoints(points)
}

func TestLoopBounds(t *testing.T) {
	if !(candyCane.RectBound().Lng.IsFull()) {
		t.Fatal("ttttt")
	}
	if !(s1.Angle(candyCane.RectBound().Lat.Lo).Degrees() < -20) {
		t.Fatal("")
	}
	if !(s1.Angle(candyCane.RectBound().Lat.Hi).Degrees() > 10) {
		t.Fatal("")
	}
	if !(smallNeCw.RectBound().IsFull()) {
		t.Fatal("")
	}
	if arctic80.RectBound() != RectFromLatLngLoHi(LatLngFromDegrees(80, -180), LatLngFromDegrees(90, 180)) {
		t.Fatal("")
	}
	if antarctic80.RectBound() != RectFromLatLngLoHi(LatLngFromDegrees(-90, -180), LatLngFromDegrees(-80, 180)) {
		t.Fatal("")
	}

	arctic80.Invert()
	// The highest latitude of each edge is attained at its midpoint.
	mid := Point{arctic80.Vertex(0).Add(arctic80.Vertex(1).Vector).Mul(0.5)}
	if math.Abs(s1.Angle(arctic80.RectBound().Lat.Hi).Radians()-LatLngFromPoint(mid).Lat.Radians()) > EPSILON {
		t.Fatal("")
	}
	arctic80.Invert()

	if !(southHemi.RectBound().Lng.IsFull()) {
		t.Fatal("")
	}
	if !(southHemi.RectBound().Lat == r1.IntervalFromPointPair(-math.Pi/2, 0)) {
		t.Fatal("")
	}
}

func TestLoopAreaCentroid(t *testing.T) {
	if !(northHemi.GetArea() == 2*math.Pi) {
		t.Fatal("")
	}
	if !(eastHemi.GetArea() == 2*math.Pi) {
		t.Fatal("")
	}

	// Construct spherical caps of random height, and approximate their boundary
	// with closely spaces vertices. Then check that the area and centroid are
	// correct.

	for i := 0; i < 100; i++ {
		// Choose a coordinate frame for the spherical cap.
		x := randomPoint()
		y := x.Cross(randomPoint().Vector).Normalize()
		z := x.Cross(y).Normalize()

		// Given two points at latitude phi and whose longitudes differ by dtheta,
		// the geodesic between the two points has a maximum latitude of
		// atan(tan(phi) / cos(dtheta/2)). This can be derived by positioning
		// the two points at (-dtheta/2, phi) and (dtheta/2, phi).
		//
		// We want to position the vertices close enough together so that their
		// maximum distance from the boundary of the spherical cap is kMaxDist.
		// Thus we want fabs(atan(tan(phi) / cos(dtheta/2)) - phi) <= kMaxDist.
		kMaxDist := 1e-6
		height := 2 * rand.Float64()
		phi := math.Asin(1 - height)
		maxDtheta := 2 * math.Acos(math.Tan(math.Abs(phi))/math.Tan(math.Abs(phi)+kMaxDist))
		maxDtheta = math.Min(math.Pi, maxDtheta) // At least 3 vertices.

		vertices := []Point{}
		for theta := 0.0; theta < 2*math.Pi; theta += rand.Float64() * maxDtheta {

			xCosThetaCosPhi := x.Mul((math.Cos(theta) * math.Cos(phi)))
			ySinThetaCosPhi := y.Mul((math.Sin(theta) * math.Cos(phi)))
			zSinPhi := z.Mul(math.Sin(phi))

			sum := Point{xCosThetaCosPhi.Add(ySinThetaCosPhi.Add(zSinPhi))}

			vertices = append(vertices, sum)
		}

		loop := LoopFromPoints(vertices)
		areaCentroid := loop.GetAreaAndCentroid()

		area := loop.GetArea()
		centroid := loop.GetCentroid()
		expectedArea := 2 * math.Pi * height
		if areaCentroid.GetArea() != area {
			t.Fatal("")
		}
		if !centroid.Equals(areaCentroid.GetCentroid()) {
			t.Fatal("")
		}
		if !(math.Abs(area-expectedArea) <= 2*math.Pi*kMaxDist) {
			t.Fatal("")
		}

		// high probability
		if !(math.Abs(area-expectedArea) >= 0.01*kMaxDist) {
			t.Fatal("")
		}

		expectedCentroid := z.Mul(expectedArea * (1 - 0.5*height))

		if !(centroid.Sub(expectedCentroid).Norm() <= 2*kMaxDist) {
			t.Fatal("")
		}
	}
}

func rotate(loop *Loop) *Loop {
	vertices := make([]Point, 0, loop.NumVertices())
	vertices = append(vertices, loop.vertices[1:loop.NumVertices()]...)
	vertices = append(vertices, loop.vertices[0])
	return LoopFromPoints(vertices)
}

func TestLoopContains(t *testing.T) {
	if !(candyCane.ContainsPoint(PointFromLatLng(LatLngFromDegrees(5, 71)))) {
		t.FailNow()
	}
	for i := 0; i < 4; i++ {
		if !(northHemi.ContainsPoint(PointFromCoordsRaw(0, 0, 1))) {
			t.FailNow()
		}
		if !(!northHemi.ContainsPoint(PointFromCoordsRaw(0, 0, -1))) {
			t.FailNow()
		}
		if !(!southHemi.ContainsPoint(PointFromCoordsRaw(0, 0, 1))) {
			t.FailNow()
		}
		if !(southHemi.ContainsPoint(PointFromCoordsRaw(0, 0, -1))) {
			t.FailNow()
		}
		if !(!westHemi.ContainsPoint(PointFromCoordsRaw(0, 1, 0))) {
			t.FailNow()
		}
		if !(westHemi.ContainsPoint(PointFromCoordsRaw(0, -1, 0))) {
			t.FailNow()
		}
		if !(eastHemi.ContainsPoint(PointFromCoordsRaw(0, 1, 0))) {
			t.FailNow()
		}
		if !(!eastHemi.ContainsPoint(PointFromCoordsRaw(0, -1, 0))) {
			t.FailNow()
		}
		northHemi = rotate(northHemi)
		southHemi = rotate(southHemi)
		eastHemi = rotate(eastHemi)
		westHemi = rotate(westHemi)
	}

	// This code checks each cell vertex is contained by exactly one of
	// the adjacent cells.
	for level := 0; level < 3; level++ {
		loops := []*Loop{}
		loopVertices := []Point{}
		points := make(map[Point]bool)

		for id := CellIDBegin(level); id != CellIDEnd(level); id = id.Next() {
			cell := CellFromCellID(id)
			points[cell.Id().Point()] = true
			for k := 0; k < 4; k++ {
				loopVertices = append(loopVertices, cell.Vertex(k))
				points[cell.Vertex(k)] = true
			}
			loops = append(loops, LoopFromPoints(loopVertices))
			loopVertices = []Point{}
		}
		for point := range points {
			count := 0
			for _, loop := range loops {
				if loop.ContainsPoint(point) {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("Failed at level %d with count %d, loops:%d, points:%d", level, count, len(loops), len(points))
			}
		}
	}
}

func testLoopRelation(t *testing.T, a, b *Loop, containsOrCrosses int, intersects, nestable bool) {
	if a.ContainsLoop(b) != (containsOrCrosses == 1) {
		if containsOrCrosses == 1 {
			t.Fatalf("loop should be contained or crossing")
		} else {
			t.Fatalf("loop should not be contained or crossing")
		}
	}
	if a.IntersectsLoop(b) != intersects {
		if intersects {
			t.Fatalf("loops should intersect")
		} else {
			t.Fatalf("loops should not intersect")
		}
	}
	if nestable {
		if a.ContainsNested(b) != a.ContainsLoop(b) {
			t.Fatalf("loops should be nested")
		}
	}
	if containsOrCrosses >= -1 {
		if a.ContainsOrCrosses(b) != containsOrCrosses {
			t.Fatalf("loops should contain or cross %d", containsOrCrosses)
		}
	}
}

func TestLoopRelations(t *testing.T) {
	testLoopRelation(t, northHemi, northHemi, 1, true, false)
	testLoopRelation(t, northHemi, southHemi, 0, false, false)
	testLoopRelation(t, northHemi, eastHemi, -1, true, false)
	testLoopRelation(t, northHemi, arctic80, 1, true, true)
	testLoopRelation(t, northHemi, antarctic80, 0, false, true)
	testLoopRelation(t, northHemi, candyCane, -1, true, false)

	// // We can't compare northHemi3 vs. northHemi or southHemi.
	testLoopRelation(t, northHemi3, northHemi3, 1, true, false)
	testLoopRelation(t, northHemi3, eastHemi, -1, true, false)
	testLoopRelation(t, northHemi3, arctic80, 1, true, true)
	testLoopRelation(t, northHemi3, antarctic80, 0, false, true)
	testLoopRelation(t, northHemi3, candyCane, -1, true, false)

	testLoopRelation(t, southHemi, northHemi, 0, false, false)
	testLoopRelation(t, southHemi, southHemi, 1, true, false)
	testLoopRelation(t, southHemi, farHemi, -1, true, false)
	testLoopRelation(t, southHemi, arctic80, 0, false, true)
	testLoopRelation(t, southHemi, antarctic80, 1, true, true)
	testLoopRelation(t, southHemi, candyCane, -1, true, false)

	testLoopRelation(t, candyCane, northHemi, -1, true, false)
	testLoopRelation(t, candyCane, southHemi, -1, true, false)
	testLoopRelation(t, candyCane, arctic80, 0, false, true)
	testLoopRelation(t, candyCane, antarctic80, 0, false, true)
	testLoopRelation(t, candyCane, candyCane, 1, true, false)

	testLoopRelation(t, nearHemi, westHemi, -1, true, false)

	testLoopRelation(t, smallNeCw, southHemi, 1, true, false)
	testLoopRelation(t, smallNeCw, westHemi, 1, true, false)
	testLoopRelation(t, smallNeCw, northHemi, -2, true, false)
	testLoopRelation(t, smallNeCw, eastHemi, -2, true, false)

	testLoopRelation(t, loopA, loopA, 1, true, false)
	testLoopRelation(t, loopA, loopB, -1, true, false)
	testLoopRelation(t, loopA, aIntersectB, 1, true, false)
	testLoopRelation(t, loopA, aUnionB, 0, true, false)
	testLoopRelation(t, loopA, aMinusB, 1, true, false)
	testLoopRelation(t, loopA, bMinusA, 0, false, false)

	testLoopRelation(t, loopB, loopA, -1, true, false)
	testLoopRelation(t, loopB, loopB, 1, true, false)
	testLoopRelation(t, loopB, aIntersectB, 1, true, false)
	testLoopRelation(t, loopB, aUnionB, 0, true, false)
	testLoopRelation(t, loopB, aMinusB, 0, false, false)
	testLoopRelation(t, loopB, bMinusA, 1, true, false)

	testLoopRelation(t, aIntersectB, loopA, 0, true, false)
	testLoopRelation(t, aIntersectB, loopB, 0, true, false)
	testLoopRelation(t, aIntersectB, aIntersectB, 1, true, false)
	testLoopRelation(t, aIntersectB, aUnionB, 0, true, true)
	testLoopRelation(t, aIntersectB, aMinusB, 0, false, false)
	testLoopRelation(t, aIntersectB, bMinusA, 0, false, false)

	testLoopRelation(t, aUnionB, loopA, 1, true, false)
	testLoopRelation(t, aUnionB, loopB, 1, true, false)
	testLoopRelation(t, aUnionB, aIntersectB, 1, true, true)
	testLoopRelation(t, aUnionB, aUnionB, 1, true, false)
	testLoopRelation(t, aUnionB, aMinusB, 1, true, false)
	testLoopRelation(t, aUnionB, bMinusA, 1, true, false)

	testLoopRelation(t, aMinusB, loopA, 0, true, false)
	testLoopRelation(t, aMinusB, loopB, 0, false, false)
	testLoopRelation(t, aMinusB, aIntersectB, 0, false, false)
	testLoopRelation(t, aMinusB, aUnionB, 0, true, false)
	testLoopRelation(t, aMinusB, aMinusB, 1, true, false)
	testLoopRelation(t, aMinusB, bMinusA, 0, false, true)

	testLoopRelation(t, bMinusA, loopA, 0, false, false)
	testLoopRelation(t, bMinusA, loopB, 0, true, false)
	testLoopRelation(t, bMinusA, aIntersectB, 0, false, false)
	testLoopRelation(t, bMinusA, aUnionB, 0, true, false)
	testLoopRelation(t, bMinusA, aMinusB, 0, false, true)
	testLoopRelation(t, bMinusA, bMinusA, 1, true, false)
}

/**
 * Tests that nearly colinear points pass S2Loop.isValid()
 */
func TestLoopRoundingError(t *testing.T) {
	points := []Point{
		PointFromCoordsRaw(-0.9190364081111774, 0.17231932652084575, 0.35451111445694833),
		PointFromCoordsRaw(-0.92130667053206, 0.17274500072476123, 0.3483578383756171),
		PointFromCoordsRaw(-0.9257244057938284, 0.17357332608634282, 0.3360158106235289),
		PointFromCoordsRaw(-0.9278712595449962, 0.17397586116468677, 0.32982923679138537),
	}
	loop := LoopFromPoints(points)
	if !loop.IsValid() {
		t.Errorf("loop should be valid")
	}
}

func TestLoopIsValid(t *testing.T) {
	if !loopA.IsValid() {
		t.Errorf("loopA should be valid")
	}
	if !loopB.IsValid() {
		t.Errorf("loopB should be valid")
	}
	if bowtie.IsValid() {
		t.Errorf("bowtie should not be valid")
	}
}

/**
 * Tests {@link S2Loop#compareTo(S2Loop)}.
 */
func TestLoopComparisons(t *testing.T) {
	abc := makeLoop("0:1, 0:2, 1:2")
	abcd := makeLoop("0:1, 0:2, 1:2, 1:1")
	abcde := makeLoop("0:1, 0:2, 1:2, 1:1, 1:0")
	if !(abc.CompareTo(abcd) < 0) {
		t.FailNow()
	}
	if !(abc.CompareTo(abcde) < 0) {
		t.FailNow()
	}
	if !(abcd.CompareTo(abcde) < 0) {
		t.FailNow()
	}
	if !(abcd.CompareTo(abc) > 0) {
		t.FailNow()
	}
	if !(abcde.CompareTo(abc) > 0) {
		t.FailNow()
	}
	if !(abcde.CompareTo(abcd) > 0) {
		t.FailNow()
	}

	bcda := makeLoop("0:2, 1:2, 1:1, 0:1")
	if 0 != abcd.CompareTo(bcda) {
		t.FailNow()
	}
	if 0 != bcda.CompareTo(abcd) {
		t.FailNow()
	}

	wxyz := makeLoop("10:11, 10:12, 11:12, 11:11")
	if !(abcd.CompareTo(wxyz) > 0) {
		t.FailNow()
	}
	if !(wxyz.CompareTo(abcd) < 0) {
		t.FailNow()
	}
}
func TestLoopGetDistance(t *testing.T) {
	// Error margin since we're doing numerical computations
	epsilon := 1e-15

	// A square with (lat,lng) vertices (0,1), (1,1), (1,2) and (0,2)
	// Tests the case where the shortest distance is along a normal to an edge,
	// onto a vertex
	s1 := makeLoop("0:1, 1:1, 1:2, 0:2")

	// A square with (lat,lng) vertices (-1,1), (1,1), (1,2) and (-1,2)
	// Tests the case where the shortest distance is along a normal to an edge,
	// not onto a vertex
	s2 := makeLoop("-1:1, 1:1, 1:2, -1:2")

	// A diamond with (lat,lng) vertices (1,0), (2,1), (3,0) and (2,-1)
	// Test the case where the shortest distance is NOT along a normal to an
	// edge
	s3 := makeLoop("1:0, 2:1, 3:0, 2:-1")

	// All the vertices should be distance 0
	for i := 0; i < s1.NumVertices(); i++ {
		if math.Abs(s1.GetDistance(s1.Vertex(i)).Radians()) > epsilon {
			t.FailNow()
		}
	}

	// A point on one of the edges should be distance 0
	if math.Abs(s1.GetDistance(PointFromLatLng(LatLngFromDegrees(0.5, 1))).Radians()) > epsilon {
		t.FailNow()
	}

	// In all three cases, the closest point to the origin is (0,1), which is at
	// a distance of 1 degree.
	// Note: all of these are intentionally distances measured along the
	// equator, since that makes the math significantly simpler. Otherwise, the
	// distance wouldn't actually be 1 degree.
	origin := PointFromLatLng(LatLngFromDegrees(0, 0))
	if math.Abs(1-s1.GetDistance(origin).Degrees()) > epsilon {
		t.FailNow()
	}
	if math.Abs(1-s2.GetDistance(origin).Degrees()) > epsilon {
		t.FailNow()
	}
	if math.Abs(1-s3.GetDistance(origin).Degrees()) > epsilon {
		t.FailNow()
	}
}
