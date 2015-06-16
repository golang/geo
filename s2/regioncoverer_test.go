package s2

import (
	"math"
	"math/rand"
	"testing"
)

func TestCoveringCap(t *testing.T) {
	expectedResultsFromJavaLibrary := []string{
		"80c297c50b",
		"80c297c50d",
		"80c297c512aaaaab",
		"80c297c57",
		"80c297c582aaaaab",
		"80c297c59d",
		"80c297c59f",
		"80c297c5a1",
	}
	coverer := NewRegionCoverer()
	coverer.SetMaxCells(8)
	cell := CellFromCellID(CellIDFromToken("80c297c574"))
	cap := cell.CapBound()

	cells := []CellID{}
	coverer.GetCovering(cap, &cells)
	for i, cell := range cells {
		if cell.ToToken() != expectedResultsFromJavaLibrary[i] {
			t.Fatalf("TestCoveringLoop result %d got %s expected %s", i, cell.ToToken(), expectedResultsFromJavaLibrary[i])
		}
	}
}

func TestCoveringPolyline(t *testing.T) {
	expectedResultsFromJavaLibrary := []string{
		"80c2bea0418c",
		"80c2bea04194",
		"80c2bea041c4",
		"80c2bea041eb",
		"80c2bea0423",
		"80c2bea0434",
		"80c2bea043b",
		"80c2bea043caac",
	}

	points := []Point{
		PointFromLatLng(LatLngFromDegrees(34.0909533022671600, -118.3914214745164100)),
		PointFromLatLng(LatLngFromDegrees(34.0906409358360560, -118.3911871165037200)),
	}
	polyline := PolylineFromPoints(points)

	coverer := NewRegionCoverer()
	coverer.SetMaxCells(8)

	cells := []CellID{}
	coverer.GetCovering(polyline, &cells)
	for i, cell := range cells {
		if cell.ToToken() != expectedResultsFromJavaLibrary[i] {
			t.Fatalf("TestCoveringLoop result %d got %s expected %s", i, cell.ToToken(), expectedResultsFromJavaLibrary[i])
		}
	}
}

func TestCoveringLoop(t *testing.T) {
	expectedResultsFromJavaLibrary := []string{
		"80c2c7b46204",
		"80c2c7b4620c",
		"80c2c7b4899",
		"80c2c7b489c4",
		"80c2c7b489dc",
		"80c2c7b489f",
		"80c2c7b48a04",
		"80c2c7b48a1d",
	}
	vertices := []Point{
		PointFromLatLng(LatLngFromDegrees(34.0487325747361496, -118.2554703578353070)),
		PointFromLatLng(LatLngFromDegrees(34.0486331233724258, -118.2555538415908671)),
		PointFromLatLng(LatLngFromDegrees(34.0486803488948837, -118.2556309551000595)),
		PointFromLatLng(LatLngFromDegrees(34.0487739664704705, -118.2555437833070613)),
	}
	loop := LoopFromPoints(vertices)
	loop.Normalize()

	coverer := NewRegionCoverer()
	coverer.SetMaxCells(8)

	cells := []CellID{}
	coverer.GetCovering(loop, &cells)
	for i, cell := range cells {
		if cell.ToToken() != expectedResultsFromJavaLibrary[i] {
			t.Fatalf("TestCoveringLoop result %d got %s expected %s", i, cell.ToToken(), expectedResultsFromJavaLibrary[i])
		}
	}
}

func random(n int32) int32 {
	if n == 0 {
		return 0
	}
	return rand.Int31n(n)
}

/**
 * Checks that "covering" completely covers the given region. If "check_tight"
 * is true, also checks that it does not contain any cells that do not
 * intersect the given region. ("id" is only used internally.)
 */
func checkCoveringRegion(region Region, covering CellUnion, checkTight bool, id CellID) {
	if !id.IsValid() {
		for face := 0; face < 6; face++ {
			checkCoveringRegion(region, covering, checkTight, CellIDFromFacePosLevel(face, 0, 0))
		}
		return
	}

	if !region.IntersectsCell(CellFromCellID(id)) {
		// If region does not intersect id, then neither should the covering.
		if checkTight {
			if !(!covering.Intersects(id)) {
				panic("")
			}
		}

	} else if !covering.Contains(id) {
		// The region may intersect id, but we can't assert that the covering
		// intersects id because we may discover that the region does not actually
		// intersect upon further subdivision. (MayIntersect is not exact.)
		if region.ContainsCell(CellFromCellID(id)) {
			panic("")
		}
		if id.IsLeaf() {
			panic("")
		}
		end := id.ChildEnd()
		for child := id.ChildBegin(); child != end; child = child.Next() {
			checkCoveringRegion(region, covering, checkTight, child)
		}
	}
}

func checkCovering(coverer *RegionCoverer, region Region, covering []CellID, interior bool) {
	// Keep track of how many cells have the same coverer.min_level() ancestor.
	minLevelCells := make(map[CellID]int)
	for i := 0; i < len(covering); i++ {
		level := covering[i].Level()
		if !(level >= coverer.MinLevel()) {
			panic("")
		}
		if !(level <= coverer.MaxLevel()) {
			panic("")
		}
		if (level-coverer.MinLevel())%coverer.LevelMod() != 0 {
			panic("")
		}

		key := covering[i].Parent(coverer.MinLevel())
		if i, ok := minLevelCells[key]; !ok {
			minLevelCells[key] = 1
		} else {
			minLevelCells[key] = i + 1
		}
	}
	if len(covering) > coverer.MaxCells() {
		// If the covering has more than the requested number of cells, then check
		// that the cell count cannot be reduced by using the parent of some cell.
		for _, i := range minLevelCells {
			if i != 1 {
				panic("")
			}
		}
	}

	if interior {
		for i := 0; i < len(covering); i++ {
			if !(region.ContainsCell(CellFromCellID(covering[i]))) {
				panic("")
			}
		}
	} else {
		cellUnion := CellUnionFromCellIDs(covering)
		checkCoveringRegion(region, *cellUnion, true, CellIDNone())
	}
}

func TestSimpleCoverings(t *testing.T) {
	coverer := NewRegionCoverer()
	coverer.SetMaxCells(math.MaxInt32)
	for i := 0; i < 1000; i++ {
		level := int(random(MAX_LEVEL + 1))
		coverer.SetMinLevel(level)
		coverer.SetMaxLevel(level)
		maxArea := math.Min(4*math.Pi, 1000*AverageArea(level))
		cap := randomCap(0.1*AverageArea(MAX_LEVEL), maxArea)
		covering := []CellID{}
		GetSimpleCovering(cap, cap.Center(), level, &covering)
		checkCovering(coverer, cap, covering, false)
	}
}
