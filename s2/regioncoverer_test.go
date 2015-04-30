package s2

import (
	"container/heap"
	"fmt"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	pq := newPriorityQueue(4)
	heap.Init(&pq)
	for i := -4; i < 0; i++ {
		heap.Push(&pq, newQueueEntry(i, nil))
	}
	for i := -1; i >= -4; i-- {
		entry := heap.Pop(&pq).(*queueEntry)
		if entry.priority != i {
			t.Errorf("expected %d, got %d", i, entry.priority)
		}
	}
}

func TestCovering(t *testing.T) {
	coverer := NewRegionCoverer()
	coverer.SetMaxCells(8)
	region1 := CellFromCellID(CellIDFromToken("80c297c574"))
	rect1 := region1.RectBound()
	fmt.Printf("%s\n", rect1.String())

	region := region1.CapBound()
	rect := region.RectBound()
	fmt.Printf("%s\n", rect.String())

	fmt.Printf("%s\n", region.String())

	cells := []CellID{}
	// cells := coverer.GetCoveringAsUnion(region)
	coverer.GetCovering(region, &cells)
	for i, cell := range cells {
		t.Errorf("cell %d: %x - %s", i, uint64(cell), cell.ToToken())
	}
}

func TestCoveringPolyline(t *testing.T) {
	coverer := NewRegionCoverer()
	coverer.SetMaxCells(8)

	points := []Point{
		PointFromLatLng(LatLngFromDegrees(34.0909533022671600, -118.3914214745164100)),
		PointFromLatLng(LatLngFromDegrees(34.0906409358360560, -118.3911871165037200)),
	}

	for _, point := range points {
		fmt.Printf("point: %v\n", point.String())
	}

	// region := RectFromLatLng(LatLngFromDegrees(34.0909533022671600, -118.3914214745164100))
	// region = region.AddPoint(LatLngFromDegrees(34.0906409358360560, -118.3911871165037200))

	polyline := PolylineFromPoints(points)
	// region := polyline.RectBound()
	// fmt.Printf("%s\n", region.String())

	cells := []CellID{}
	coverer.GetCovering(polyline, &cells)
	for i, cell := range cells {
		t.Errorf("cell %d: %x - %s", i, uint64(cell), cell.ToToken())
	}
}
