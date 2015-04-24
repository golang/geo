package s2

import (
	"testing"
)

func TestCovering(t *testing.T) {
	coverer := NewRegionCoverer()
	coverer.SetMaxCells(8)
	region := CellFromCellID(CellIDFromToken("80c297c53c"))
	cells := coverer.GetCovering(region)
	for _, cell := range *cells {
		if cell.Level() != 1 {
			t.Errorf("Level not as expected %s", cell.ToToken())
		}
		if cell.Face() != 4 {
			t.Error("Face not as expected")
		}
	}
}
