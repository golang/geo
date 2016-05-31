/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package s2

import (
	"sort"
	"testing"

	"github.com/golang/geo/r2"
	"github.com/golang/geo/s1"
)

func TestCellIDFromFace(t *testing.T) {
	for face := 0; face < 6; face++ {
		fpl := CellIDFromFacePosLevel(face, 0, 0)
		f := CellIDFromFace(face)
		if fpl != f {
			t.Errorf("CellIDFromFacePosLevel(%d, 0, 0) != CellIDFromFace(%d), got %v wanted %v", face, face, f, fpl)
		}
	}
}

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

	if kid2 := ci.ChildBeginAtLevel(ci.Level() + 2).Pos(); kid2 != 0x12345610 {
		t.Errorf("child two levels down is 0x%X, want 0x12345610", kid2)
	}
	if kid0 := ci.ChildBegin().Pos(); kid0 != 0x12345640 {
		t.Errorf("first child is 0x%X, want 0x12345640", kid0)
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

	if uint64(ci.ChildBegin()) >= uint64(ci) {
		t.Errorf("ci.ChildBegin() is 0x%X, want < 0x%X", ci.ChildBegin(), ci)
	}
	if uint64(ci.ChildEnd()) <= uint64(ci) {
		t.Errorf("ci.ChildEnd() is 0x%X, want > 0x%X", ci.ChildEnd(), ci)
	}
	if ci.ChildEnd() != ci.ChildBegin().Next().Next().Next().Next() {
		t.Errorf("ci.ChildEnd() is 0x%X, want 0x%X", ci.ChildEnd(), ci.ChildBegin().Next().Next().Next().Next())
	}
	if ci.RangeMin() != ci.ChildBeginAtLevel(maxLevel) {
		t.Errorf("ci.RangeMin() is 0x%X, want 0x%X", ci.RangeMin(), ci.ChildBeginAtLevel(maxLevel))
	}
	if ci.RangeMax().Next() != ci.ChildEndAtLevel(maxLevel) {
		t.Errorf("ci.RangeMax().Next() is 0x%X, want 0x%X", ci.RangeMax().Next(), ci.ChildEndAtLevel(maxLevel))
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

type byCellID []CellID

func (v byCellID) Len() int           { return len(v) }
func (v byCellID) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v byCellID) Less(i, j int) bool { return uint64(v[i]) < uint64(v[j]) }

func TestVertexNeighbors(t *testing.T) {
	// Check the vertex neighbors of the center of face 2 at level 5.
	id := cellIDFromPoint(PointFromCoords(0, 0, 1))
	neighbors := id.VertexNeighbors(5)
	sort.Sort(byCellID(neighbors))

	for n, nbr := range neighbors {
		i, j := 1<<29, 1<<29
		if n < 2 {
			i--
		}
		if n == 0 || n == 3 {
			j--
		}
		want := cellIDFromFaceIJ(2, i, j).Parent(5)

		if nbr != want {
			t.Errorf("CellID(%s).VertexNeighbors()[%d] = %v, want %v", id, n, nbr, want)
		}
	}

	// Check the vertex neighbors of the corner of faces 0, 4, and 5.
	id = CellIDFromFacePosLevel(0, 0, maxLevel)
	neighbors = id.VertexNeighbors(0)
	sort.Sort(byCellID(neighbors))
	if len(neighbors) != 3 {
		t.Errorf("len(CellID(%d).VertexNeighbors()) = %d, wanted %d", id, len(neighbors), 3)
	}
	if neighbors[0] != CellIDFromFace(0) {
		t.Errorf("CellID(%d).VertexNeighbors()[0] = %d, wanted %d", id, neighbors[0], CellIDFromFace(0))
	}
	if neighbors[1] != CellIDFromFace(4) {
		t.Errorf("CellID(%d).VertexNeighbors()[1] = %d, wanted %d", id, neighbors[1], CellIDFromFace(4))
	}
}

func TestCellIDTokensNominal(t *testing.T) {
	tests := []struct {
		token string
		id    CellID
	}{
		{"1", 0x1000000000000000},
		{"3", 0x3000000000000000},
		{"14", 0x1400000000000000},
		{"41", 0x4100000000000000},
		{"094", 0x0940000000000000},
		{"537", 0x5370000000000000},
		{"3fec", 0x3fec000000000000},
		{"72f3", 0x72f3000000000000},
		{"52b8c", 0x52b8c00000000000},
		{"990ed", 0x990ed00000000000},
		{"4476dc", 0x4476dc0000000000},
		{"2a724f", 0x2a724f0000000000},
		{"7d4afc4", 0x7d4afc4000000000},
		{"b675785", 0xb675785000000000},
		{"40cd6124", 0x40cd612400000000},
		{"3ba32f81", 0x3ba32f8100000000},
		{"08f569b5c", 0x08f569b5c0000000},
		{"385327157", 0x3853271570000000},
		{"166c4d1954", 0x166c4d1954000000},
		{"96f48d8c39", 0x96f48d8c39000000},
		{"0bca3c7f74c", 0x0bca3c7f74c00000},
		{"1ae3619d12f", 0x1ae3619d12f00000},
		{"07a77802a3fc", 0x07a77802a3fc0000},
		{"4e7887ec1801", 0x4e7887ec18010000},
		{"4adad7ae74124", 0x4adad7ae74124000},
		{"90aba04afe0c5", 0x90aba04afe0c5000},
		{"8ffc3f02af305c", 0x8ffc3f02af305c00},
		{"6fa47550938183", 0x6fa4755093818300},
		{"aa80a565df5e7fc", 0xaa80a565df5e7fc0},
		{"01614b5e968e121", 0x01614b5e968e1210},
		{"aa05238e7bd3ee7c", 0xaa05238e7bd3ee7c},
		{"48a23db9c2963e5b", 0x48a23db9c2963e5b},
	}
	for _, test := range tests {
		ci := CellIDFromToken(test.token)
		if ci != test.id {
			t.Errorf("CellIDFromToken(%q) = %x, want %x", test.token, uint64(ci), uint64(test.id))
		}

		token := ci.ToToken()
		if token != test.token {
			t.Errorf("ci.ToToken = %q, want %q", token, test.token)
		}
	}
}

func TestCellIDFromTokensErrorCases(t *testing.T) {
	noneToken := CellID(0).ToToken()
	if noneToken != "X" {
		t.Errorf("CellID(0).Token() = %q, want X", noneToken)
	}
	noneID := CellIDFromToken(noneToken)
	if noneID != CellID(0) {
		t.Errorf("CellIDFromToken(%q) = %x, want 0", noneToken, uint64(noneID))
	}
	tests := []string{
		"876b e99",
		"876bee99\n",
		"876[ee99",
		" 876bee99",
	}
	for _, test := range tests {
		ci := CellIDFromToken(test)
		if uint64(ci) != 0 {
			t.Errorf("CellIDFromToken(%q) = %x, want 0", test, uint64(ci))
		}
	}
}

func TestIJLevelToBoundUV(t *testing.T) {
	maxIJ := 1<<maxLevel - 1

	tests := []struct {
		i     int
		j     int
		level int
		want  r2.Rect
	}{
		// The i/j space is [0, 2^30 - 1) which maps to [-1, 1] for the
		// x/y axes of the face surface. Results are scaled by the size of a cell
		// at the given level. At level 0, everything is one cell of the full size
		// of the space.  At maxLevel, the bounding rect is almost floating point
		// noise.

		// What should be out of bounds values, but passes the C++ code as well.
		{
			-1, -1, 0,
			r2.RectFromPoints(r2.Point{-5, -5}, r2.Point{-1, -1}),
		},
		{
			-1 * maxIJ, -1 * maxIJ, 0,
			r2.RectFromPoints(r2.Point{-5, -5}, r2.Point{-1, -1}),
		},
		{
			-1, -1, maxLevel,
			r2.RectFromPoints(r2.Point{-1.0000000024835267, -1.0000000024835267},
				r2.Point{-1, -1}),
		},
		{
			0, 0, maxLevel + 1,
			r2.RectFromPoints(r2.Point{-1, -1}, r2.Point{-1, -1}),
		},

		// Minimum i,j at different levels
		{
			0, 0, 0,
			r2.RectFromPoints(r2.Point{-1, -1}, r2.Point{1, 1}),
		},
		{
			0, 0, maxLevel / 2,
			r2.RectFromPoints(r2.Point{-1, -1},
				r2.Point{-0.999918621033430099, -0.999918621033430099}),
		},
		{
			0, 0, maxLevel,
			r2.RectFromPoints(r2.Point{-1, -1},
				r2.Point{-0.999999997516473060, -0.999999997516473060}),
		},

		// Just a hair off the outer bounds at different levels.
		{
			1, 1, 0,
			r2.RectFromPoints(r2.Point{-1, -1}, r2.Point{1, 1}),
		},
		{
			1, 1, maxLevel / 2,
			r2.RectFromPoints(r2.Point{-1, -1},
				r2.Point{-0.999918621033430099, -0.999918621033430099}),
		},
		{
			1, 1, maxLevel,
			r2.RectFromPoints(r2.Point{-0.9999999975164731, -0.9999999975164731},
				r2.Point{-0.9999999950329462, -0.9999999950329462}),
		},

		// Center point of the i,j space at different levels.
		{
			maxIJ / 2, maxIJ / 2, 0,
			r2.RectFromPoints(r2.Point{-1, -1}, r2.Point{1, 1})},
		{
			maxIJ / 2, maxIJ / 2, maxLevel / 2,
			r2.RectFromPoints(r2.Point{-0.000040691345930099, -0.000040691345930099},
				r2.Point{0, 0})},
		{
			maxIJ / 2, maxIJ / 2, maxLevel,
			r2.RectFromPoints(r2.Point{-0.000000001241763433, -0.000000001241763433},
				r2.Point{0, 0})},

		// Maximum i, j at different levels.
		{
			maxIJ, maxIJ, 0,
			r2.RectFromPoints(r2.Point{-1, -1}, r2.Point{1, 1}),
		},
		{
			maxIJ, maxIJ, maxLevel / 2,
			r2.RectFromPoints(r2.Point{0.999918621033430099, 0.999918621033430099},
				r2.Point{1, 1}),
		},
		{
			maxIJ, maxIJ, maxLevel,
			r2.RectFromPoints(r2.Point{0.999999997516473060, 0.999999997516473060},
				r2.Point{1, 1}),
		},
	}

	for _, test := range tests {
		uv := ijLevelToBoundUV(test.i, test.j, test.level)
		if !float64Eq(uv.X.Lo, test.want.X.Lo) ||
			!float64Eq(uv.X.Hi, test.want.X.Hi) ||
			!float64Eq(uv.Y.Lo, test.want.Y.Lo) ||
			!float64Eq(uv.Y.Hi, test.want.Y.Hi) {
			t.Errorf("ijLevelToBoundUV(%d, %d, %d), got %v, want %v",
				test.i, test.j, test.level, uv, test.want)
		}
	}
}

func TestCommonAncestorLevel(t *testing.T) {
	tests := []struct {
		ci     CellID
		other  CellID
		want   int
		wantOk bool
	}{
		// Identical cell IDs.
		{
			CellIDFromFace(0),
			CellIDFromFace(0),
			0,
			true,
		},
		{
			CellIDFromFace(0).ChildBeginAtLevel(30),
			CellIDFromFace(0).ChildBeginAtLevel(30),
			30,
			true,
		},
		// One cell is a descendant of the other.
		{
			CellIDFromFace(0).ChildBeginAtLevel(30),
			CellIDFromFace(0),
			0,
			true,
		},
		{
			CellIDFromFace(5),
			CellIDFromFace(5).ChildEndAtLevel(30).Prev(),
			0,
			true,
		},
		// No common ancestors.
		{
			CellIDFromFace(0),
			CellIDFromFace(5),
			0,
			false,
		},
		{
			CellIDFromFace(2).ChildBeginAtLevel(30),
			CellIDFromFace(3).ChildBeginAtLevel(20),
			0,
			false,
		},
		// Common ancestor distinct from both.
		{
			CellIDFromFace(5).ChildBeginAtLevel(9).Next().ChildBeginAtLevel(15),
			CellIDFromFace(5).ChildBeginAtLevel(9).ChildBeginAtLevel(20),
			8,
			true,
		},
		{
			CellIDFromFace(0).ChildBeginAtLevel(2).ChildBeginAtLevel(30),
			CellIDFromFace(0).ChildBeginAtLevel(2).Next().ChildBeginAtLevel(5),
			1,
			true,
		},
	}
	for _, test := range tests {
		if got, ok := test.ci.CommonAncestorLevel(test.other); ok != test.wantOk || got != test.want {
			t.Errorf("CellID(%v).VertexNeighbors(%v) = %d, %t; want %d, %t", test.ci, test.other, got, ok, test.want, test.wantOk)
		}
	}
}

func TestFindMSBSetNonZero64(t *testing.T) {
	testOne := uint64(0x8000000000000000)
	testAll := uint64(0xFFFFFFFFFFFFFFFF)
	testSome := uint64(0xFEDCBA9876543210)
	for i := 63; i >= 0; i-- {
		if got := findMSBSetNonZero64(testOne); got != i {
			t.Errorf("findMSBSetNonZero64(%x) = %d, want = %d", testOne, got, i)
		}
		if got := findMSBSetNonZero64(testAll); got != i {
			t.Errorf("findMSBSetNonZero64(%x) = %d, want = %d", testAll, got, i)
		}
		if got := findMSBSetNonZero64(testSome); got != i {
			t.Errorf("findMSBSetNonZero64(%x) = %d, want = %d", testSome, got, i)
		}
		testOne >>= 1
		testAll >>= 1
		testSome >>= 1
	}

	if got := findMSBSetNonZero64(1); got != 0 {
		t.Errorf("findMSBSetNonZero64(1) = %v, want 0", got)
	}

	if got := findMSBSetNonZero64(0); got != 0 {
		t.Errorf("findMSBSetNonZero64(0) = %v, want 0", got)
	}
}

func TestFindLSBSetNonZero64(t *testing.T) {
	testOne := uint64(0x0000000000000001)
	testAll := uint64(0xFFFFFFFFFFFFFFFF)
	testSome := uint64(0x0123456789ABCDEF)
	for i := 0; i < 64; i++ {
		if got := findLSBSetNonZero64(testOne); got != i {
			t.Errorf("findLSBSetNonZero64(%x) = %d, want = %d", testOne, got, i)
		}
		if got := findLSBSetNonZero64(testAll); got != i {
			t.Errorf("findLSBSetNonZero64(%x) = %d, want = %d", testAll, got, i)
		}
		if got := findLSBSetNonZero64(testSome); got != i {
			t.Errorf("findLSBSetNonZero64(%x) = %d, want = %d", testSome, got, i)
		}
		testOne <<= 1
		testAll <<= 1
		testSome <<= 1
	}

	if got := findLSBSetNonZero64(0); got != 0 {
		t.Errorf("findLSBSetNonZero64(0) = %v, want 0", got)
	}
}

func TestAdvance(t *testing.T) {
	tests := []struct {
		ci    CellID
		steps int64
		want  CellID
	}{
		{
			CellIDFromFace(0).ChildBeginAtLevel(0),
			7,
			CellIDFromFace(5).ChildEndAtLevel(0),
		},
		{
			CellIDFromFace(0).ChildBeginAtLevel(0),
			12,
			CellIDFromFace(5).ChildEndAtLevel(0),
		},
		{
			CellIDFromFace(5).ChildEndAtLevel(0),
			-7,
			CellIDFromFace(0).ChildBeginAtLevel(0),
		},
		{
			CellIDFromFace(5).ChildEndAtLevel(0),
			-12000000,
			CellIDFromFace(0).ChildBeginAtLevel(0),
		},
		{
			CellIDFromFace(0).ChildBeginAtLevel(5),
			500,
			CellIDFromFace(5).ChildEndAtLevel(5).Advance(500 - (6 << (2 * 5))),
		},
		{
			CellIDFromFacePosLevel(3, 0x12345678, maxLevel-4).ChildBeginAtLevel(maxLevel),
			256,
			CellIDFromFacePosLevel(3, 0x12345678, maxLevel-4).Next().ChildBeginAtLevel(maxLevel),
		},
		{
			CellIDFromFacePosLevel(1, 0, maxLevel),
			4 << (2 * maxLevel),
			CellIDFromFacePosLevel(5, 0, maxLevel),
		},
	}

	for _, test := range tests {
		if got := test.ci.Advance(test.steps); got != test.want {
			t.Errorf("CellID(%v).Advance(%d) = %v; want = %v", test.ci, test.steps, got, test.want)
		}
	}
}

func TestFaceSiTi(t *testing.T) {
	id := CellIDFromFacePosLevel(3, 0x12345678, maxLevel)
	// Check that the (si, ti) coordinates of the center end in a
	// 1 followed by (30 - level) 0's.
	for level := uint64(0); level <= maxLevel; level++ {
		l := maxLevel - int(level)
		want := 1 << level
		mask := 1<<(level+1) - 1

		_, si, ti := id.Parent(l).faceSiTi()
		if want != si&mask {
			t.Errorf("CellID.Parent(%d).faceSiTi(), si = %b, want %b", l, si&mask, want)
		}
		if want != ti&mask {
			t.Errorf("CellID.Parent(%d).faceSiTi(), ti = %b, want %b", l, ti&mask, want)
		}
	}
}
