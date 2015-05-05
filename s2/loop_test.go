package s2

import (
	// "runtime/debug"
	"testing"
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
	// if !loopA.IsValid() {
	// 	t.Errorf("loopA should be valid")
	// }
	// if !loopB.IsValid() {
	// 	t.Errorf("loopB should be valid")
	// }
	// if bowtie.IsValid() {
	// 	t.Errorf("bowtie should not be valid")
	// }
}
