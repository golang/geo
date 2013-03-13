package s2

import (
	"reflect"
	"testing"
)

func TestNormalization(t *testing.T) {
	cu := CellUnion{
		0x80855c0000000000, // A: a cell over Pittsburg CA
		0x80855d0000000000, // B, a child of A
		0x8085634000000000, // first child of X, disjoint from A
		0x808563c000000000, // second child of X
		0x80855dc000000000, // a child of B
		0x808562c000000000, // third child of X
		0x8085624000000000, // fourth child of X
		0x80855d0000000000, // B again
	}
	exp := CellUnion{
		0x80855c0000000000, // A
		0x8085630000000000, // X
	}
	cu.Normalize()
	if !reflect.DeepEqual(cu, exp) {
		t.Errorf("got %v, want %v", cu, exp)
	}

	// add a redundant cell
	/* TODO(dsymonds)
	cu.Add(0x808562c000000000)
	if !reflect.DeepEqual(cu, exp) {
		t.Errorf("after redundant add, got %v, want %v", cu, exp)
	}
	*/
}
