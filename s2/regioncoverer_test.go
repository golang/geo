package s2

import (
	"testing"
)

func TestCovering(t *testing.T) {
	coverer := NewRegionCoverer()
	coverer.SetMaxCells(8)
	region := CellFromCellID(CellIDFromToken("80c297c53"))
	cells := []CellID{}
	// cells := coverer.GetCoveringAsUnion(region)
	coverer.GetCovering(region, &cells)
	for _, cell := range cells {
		t.Errorf("cell: %s", cell.ToToken())
		if cell.Level() != 1 {
			t.Errorf("Level not as expected %s", cell.ToToken())
		}
		if cell.Face() != 4 {
			t.Error("Face not as expected")
		}
	}
}
