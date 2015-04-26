package s2

import (
	"testing"
)

func TestCovering(t *testing.T) {
	coverer := NewRegionCoverer()
	coverer.SetMaxCells(8)
	region1 := CellFromCellID(CellIDFromToken("80c297c574"))
	region := region1.CapBound()

	cells := []CellID{}
	// cells := coverer.GetCoveringAsUnion(region)
	coverer.GetCovering(region, &cells)
	for i, cell := range cells {
		t.Errorf("cell %d: %x - %s", i, uint64(cell), cell.ToToken())
	}
}
