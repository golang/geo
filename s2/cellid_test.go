package s2

import (
	"testing"

	"code.google.com/p/gos2/s1"
)

func TestParentChildRelationships(t *testing.T) {
	ci := CellIDFromFacePosLevel(3, 0x12345678, maxLevel-4)

	if !ci.IsValid() {
		t.Errorf("CellID %v should be valid", ci)
	}
	if f := ci.Face(); f != 3 {
		t.Errorf("ci.Face() is %v, want 3", f)
	}
	if p := ci.Pos(); p != 0x12345700 {
		t.Errorf("ci.Pos() is 0x%X, want 0x12345700", p)
	}
	if l := ci.Level(); l != 26 { // 26 is maxLevel - 4
		t.Errorf("ci.Level() is %v, want 26", l)
	}
	if ci.IsLeaf() {
		t.Errorf("CellID %v should not be a leaf", ci)
	}

	if kid0 := ci.Children()[0].Pos(); kid0 != 0x12345640 {
		t.Errorf("first child is 0x%X, want 0x12345640", kid0)
	}
	if parent := ci.immediateParent().Pos(); parent != 0x12345400 {
		t.Errorf("ci.immediateParent().Pos() = 0x%X, want 0x12345400", parent)
	}
	if parent := ci.Parent(ci.Level() - 2).Pos(); parent != 0x12345000 {
		t.Errorf("ci.Parent(l-2).Pos() = 0x%X, want 0x12345000", parent)
	}
}

func TestContainment(t *testing.T) {
	a := CellID(0x80855c0000000000) // Pittsburg
	b := CellID(0x80855d0000000000) // child of a
	c := CellID(0x80855dc000000000) // child of b
	d := CellID(0x8085630000000000) // part of Pittsburg disjoint from a
	tests := []struct {
		x, y                                 CellID
		xContainsY, yContainsX, xIntersectsY bool
	}{
		{a, a, true, true, true},
		{a, b, true, false, true},
		{a, c, true, false, true},
		{a, d, false, false, false},
		{b, b, true, true, true},
		{b, c, true, false, true},
		{b, d, false, false, false},
		{c, c, true, true, true},
		{c, d, false, false, false},
		{d, d, true, true, true},
	}
	should := func(b bool) string {
		if b {
			return "should"
		}
		return "should not"
	}
	for _, test := range tests {
		if test.x.Contains(test.y) != test.xContainsY {
			t.Errorf("%v %s contain %v", test.x, should(test.xContainsY), test.y)
		}
		if test.x.Intersects(test.y) != test.xIntersectsY {
			t.Errorf("%v %s intersect %v", test.x, should(test.xIntersectsY), test.y)
		}
		if test.y.Contains(test.x) != test.yContainsX {
			t.Errorf("%v %s contain %v", test.y, should(test.yContainsX), test.x)
		}
	}

	// TODO(dsymonds): Test Contains, Intersects better, such as with adjacent cells.
}

func TestCellIDString(t *testing.T) {
	ci := CellID(0xbb04000000000000)
	if s, exp := ci.String(), "5/31200"; s != exp {
		t.Errorf("ci.String() = %q, want %q", s, exp)
	}
}

func TestLatLng(t *testing.T) {
	// You can generate these with the s2cellid2latlngtestcase C++ program in this directory.
	tests := []struct {
		id       CellID
		lat, lng float64
	}{
		{0x47a1cbd595522b39, 49.703498679, 11.770681595},
		{0x46525318b63be0f9, 55.685376759, 12.588490937},
		{0x52b30b71698e729d, 45.486546517, -93.449700022},
		{0x46ed8886cfadda85, 58.299984854, 23.049300056},
		{0x3663f18a24cbe857, 34.364439040, 108.330699969},
		{0x10a06c0a948cf5d, -30.694551352, -30.048758753},
		{0x2b2bfd076787c5df, -25.285264027, 133.823116966},
		{0xb09dff882a7809e1, -75.000000031, 0.000000133},
		{0x94daa3d000000001, -24.694439215, -47.537363213},
		{0x87a1000000000001, 38.899730392, -99.901813021},
		{0x4fc76d5000000001, 81.647200334, -55.631712940},
		{0x3b00955555555555, 10.050986518, 78.293170610},
		{0x1dcc469991555555, -34.055420593, 18.551140038},
		{0xb112966aaaaaaaab, -69.219262171, 49.670072392},
	}
	for _, test := range tests {
		l1 := LatLngFromDegrees(test.lat, test.lng)
		l2 := test.id.LatLng()
		if l1.Distance(l2) > 1e-9*s1.Degree { // ~0.1mm on earth.
			t.Errorf("LatLng() for CellID %x (%s) : got %s, want %s", uint64(test.id), test.id, l2, l1)
		}
		c1 := test.id
		c2 := CellIDFromLatLng(l1)
		if c1 != c2 {
			t.Errorf("CellIDFromLatLng(%s) = %x (%s), want %s", l1, uint64(c2), c2, c1)
		}
	}
}

func TestEdgeNeighbors(t *testing.T) {
	// Check the edge neighbors of face 1.
	faces := []int{5, 3, 2, 0}
	for i, nbr := range cellIDFromFaceIJ(1, 0, 0).Parent(0).EdgeNeighbors() {
		if !nbr.isFace() {
			t.Errorf("CellID(%d) is not a face", nbr)
		}
		if got, want := nbr.Face(), faces[i]; got != want {
			t.Errorf("CellID(%d).Face() = %d, want %d", nbr, got, want)
		}
	}
	// Check the edge neighbors of the corner cells at all levels.  This case is
	// trickier because it requires projecting onto adjacent faces.
	const maxIJ = maxSize - 1
	for level := 1; level <= maxLevel; level++ {
		id := cellIDFromFaceIJ(1, 0, 0).Parent(level)
		// These neighbors were determined manually using the face and axis
		// relationships.
		levelSizeIJ := sizeIJ(level)
		want := []CellID{
			cellIDFromFaceIJ(5, maxIJ, maxIJ).Parent(level),
			cellIDFromFaceIJ(1, levelSizeIJ, 0).Parent(level),
			cellIDFromFaceIJ(1, 0, levelSizeIJ).Parent(level),
			cellIDFromFaceIJ(0, maxIJ, 0).Parent(level),
		}
		for i, nbr := range id.EdgeNeighbors() {
			if nbr != want[i] {
				t.Errorf("CellID(%d).EdgeNeighbors()[%d] = %v, want %v", id, i, nbr, want[i])
			}
		}
	}
}
