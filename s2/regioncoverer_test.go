package s2

import (
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
			t.Fatal("TestCoveringLoop result %d got %s expected %s", i, cell.ToToken(), expectedResultsFromJavaLibrary[i])
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
			t.Fatal("TestCoveringLoop result %d got %s expected %s", i, cell.ToToken(), expectedResultsFromJavaLibrary[i])
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
			t.Fatal("TestCoveringLoop result %d got %s expected %s", i, cell.ToToken(), expectedResultsFromJavaLibrary[i])
		}
	}
}
