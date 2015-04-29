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
	region := region1.CapBound()

	fmt.Printf("%s\n", region.String())

	cells := []CellID{}
	// cells := coverer.GetCoveringAsUnion(region)
	coverer.GetCovering(region, &cells)
	for i, cell := range cells {
		t.Errorf("cell %d: %x - %s", i, uint64(cell), cell.ToToken())
	}
}
