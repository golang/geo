package s2

import (
	"testing"
)

func TestCovering(t *testing.T) {
	coverer := NewRegionCoverer()
	coverer.SetMaxCells(8)
	cells := coverer.GetCovering(nil)
	for _, cell := range *cells {
		if cell.Level() != 1 {
			t.Errorf("Level not as expected %s", cell.ToToken())
		}
		if cell.Face() != 4 {
			t.Error("Face not as expected")
		}
	}
}
