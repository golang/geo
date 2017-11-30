/*
Copyright 2015 Google Inc. All rights reserved.

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
	"testing"
)

const (
	// A set of nested loops around the LatLng point 0:0.
	// Every vertex of nearLoop0 is also a vertex of nearLoop1.
	nearPoint    = "0:0"
	nearLoop0    = "-1:0, 0:1, 1:0, 0:-1;"
	nearLoop1    = "-1:-1, -1:0, -1:1, 0:1, 1:1, 1:0, 1:-1, 0:-1;"
	nearLoop2    = "-1:-2, -2:5, 5:-2;"
	nearLoop3    = "-2:-2, -3:6, 6:-3;"
	nearLoopHemi = "0:-90, -90:0, 0:90, 90:0;"

	// A set of nested loops around the LatLng point 0:180. Every vertex of
	// farLoop0 and farLoop2 belongs to farLoop1, and all the loops except
	// farLoop2 are non-convex.
	farPoint    = "0:180"
	farLoop0    = "0:179, 1:180, 0:-179, 2:-180;"
	farLoop1    = "0:179, -1:179, 1:180, -1:-179, 0:-179, 3:-178, 2:-180, 3:178;"
	farLoop2    = "3:-178, 3:178, -1:179, -1:-179;"
	farLoop3    = "-3:-178, 4:-177, 4:177, -3:178, -2:179;"
	farLoopHemi = "0:-90, 60:90, -60:90;"

	// A set of nested loops around the LatLng point -90:0.
	southLoopPoint = "-89.9999:0.001"
	southLoop0a    = "-90:0, -89.99:0.01, -89.99:0;"
	southLoop0b    = "-90:0, -89.99:0.03, -89.99:0.02;"
	southLoop0c    = "-90:0, -89.99:0.05, -89.99:0.04;"
	southLoop1     = "-90:0, -89.9:0.1, -89.9:-0.1;"
	southLoop2     = "-90:0, -89.8:0.2, -89.8:-0.2;"
	southLoopHemi  = "0:-180, 0:60, 0:-60;"

	// Two different loops that surround all the near and far loops except
	// for the hemispheres.
	nearFarLoop1 = "-1:-9, -9:-9, -9:9, 9:9, 9:-9, 1:-9, " +
		"1:-175, 9:-175, 9:175, -9:175, -9:-175, -1:-175;"
	nearFarLoop2 = "-2:15, -2:170, -8:-175, 8:-175, " +
		"2:170, 2:15, 8:-4, -8:-4;"

	// Loop that results from intersection of other loops.
	farHemiSouthHemiLoop = "0:-180, 0:90, -60:90, 0:-90;"

	// Rectangles that form a cross, with only shared vertices, no crossing edges.
	// Optional holes outside the intersecting region. 1 is the horizontal rectangle,
	// and 2 is the vertical. The intersections are shared vertices.
	//       x---x
	//       | 2 |
	//   +---*---*---+
	//   | 1 |1+2| 1 |
	//   +---*---*---+
	//       | 2 |
	//       x---x
	loopCross1          = "-2:1, -1:1, 1:1, 2:1, 2:-1, 1:-1, -1:-1, -2:-1;"
	loopCross1SideHole  = "-1.5:0.5, -1.2:0.5, -1.2:-0.5, -1.5:-0.5;"
	loopCrossCenterHole = "-0.5:0.5, 0.5:0.5, 0.5:-0.5, -0.5:-0.5;"
	loopCross2SideHole  = "0.5:-1.5, 0.5:-1.2, -0.5:-1.2, -0.5:-1.5;"
	loopCross2          = "1:-2, 1:-1, 1:1, 1:2, -1:2, -1:1, -1:-1, -1:-2;"

	// Two rectangles that intersect, but no edges cross and there's always
	// local containment (rather than crossing) at each shared vertex.
	// In this ugly ASCII art, 1 is A+B, 2 is B+C:
	//   +---+---+---+
	//   | A | B | C |
	//   +---+---+---+
	loopOverlap1          = "0:1, 1:1, 2:1, 2:0, 1:0, 0:0;"
	loopOverlap1SideHole  = "0.2:0.8, 0.8:0.8, 0.8:0.2, 0.2:0.2;"
	loopOverlapCenterHole = "1.2:0.8, 1.8:0.8, 1.8:0.2, 1.2:0.2;"
	loopOverlap2SideHole  = "2.2:0.8, 2.8:0.8, 2.8:0.2, 2.2:0.2;"
	loopOverlap2          = "1:1, 2:1, 3:1, 3:0, 2:0, 1:0;"

	// By symmetry, the intersection of the two polygons has almost half the area
	// of either polygon.
	//   +---+
	//   | 3 |
	//   +---+---+
	//   |3+4| 4 |
	//   +---+---+
	loopOverlap3 = "-10:10, 0:10, 0:-10, -10:-10, -10:0"
	loopOverlap4 = "-10:0, 10:0, 10:-10, -10:-10"
)

var (
	// Some standard polygons to use in the tests.
	emptyPolygon = &Polygon{}
	fullPolygon  = FullPolygon()

	near0Polygon     = makePolygon(nearLoop0, true)
	near01Polygon    = makePolygon(nearLoop0+nearLoop1, true)
	near30Polygon    = makePolygon(nearLoop3+nearLoop0, true)
	near23Polygon    = makePolygon(nearLoop2+nearLoop3, true)
	near0231Polygon  = makePolygon(nearLoop0+nearLoop2+nearLoop3+nearLoop1, true)
	near023H1Polygon = makePolygon(nearLoop0+nearLoop2+nearLoop3+nearLoopHemi+nearLoop1, true)

	far01Polygon    = makePolygon(farLoop0+farLoop1, true)
	far21Polygon    = makePolygon(farLoop2+farLoop1, true)
	far231Polygon   = makePolygon(farLoop2+farLoop3+farLoop1, true)
	far2H0Polygon   = makePolygon(farLoop2+farLoopHemi+farLoop0, true)
	far2H013Polygon = makePolygon(farLoop2+farLoopHemi+farLoop0+farLoop1+farLoop3, true)

	south0abPolygon     = makePolygon(southLoop0a+southLoop0b, true)
	south2Polygon       = makePolygon(southLoop2, true)
	south20b1Polygon    = makePolygon(southLoop2+southLoop0b+southLoop1, true)
	south2H1Polygon     = makePolygon(southLoop2+southLoopHemi+southLoop1, true)
	south20bH0acPolygon = makePolygon(southLoop2+southLoop0b+southLoopHemi+
		southLoop0a+southLoop0c, true)

	nf1N10F2S10abcPolygon = makePolygon(southLoop0c+farLoop2+nearLoop1+
		nearFarLoop1+nearLoop0+southLoop1+southLoop0b+southLoop0a, true)

	nf2N2F210S210abPolygon = makePolygon(farLoop2+southLoop0a+farLoop1+
		southLoop1+farLoop0+southLoop0b+nearFarLoop2+southLoop2+nearLoop2, true)

	f32n0Polygon  = makePolygon(farLoop2+nearLoop0+farLoop3, true)
	n32s0bPolygon = makePolygon(nearLoop3+southLoop0b+nearLoop2, true)

	cross1Polygon           = makePolygon(loopCross1, true)
	cross1SideHolePolygon   = makePolygon(loopCross1+loopCross1SideHole, true)
	cross1CenterHolePolygon = makePolygon(loopCross1+loopCrossCenterHole, true)
	cross2Polygon           = makePolygon(loopCross2, true)
	cross2SideHolePolygon   = makePolygon(loopCross2+loopCross2SideHole, true)
	cross2CenterHolePolygon = makePolygon(loopCross2+loopCrossCenterHole, true)

	overlap1Polygon           = makePolygon(loopOverlap1, true)
	overlap1SideHolePolygon   = makePolygon(loopOverlap1+loopOverlap1SideHole, true)
	overlap1CenterHolePolygon = makePolygon(loopOverlap1+loopOverlapCenterHole, true)
	overlap2Polygon           = makePolygon(loopOverlap2, true)
	overlap2SideHolePolygon   = makePolygon(loopOverlap2+loopOverlap2SideHole, true)
	overlap2CenterHolePolygon = makePolygon(loopOverlap2+loopOverlapCenterHole, true)

	overlap3Polygon = makePolygon(loopOverlap3, true)
	overlap4Polygon = makePolygon(loopOverlap4, true)

	farHemiPolygon      = makePolygon(farLoopHemi, true)
	southHemiPolygon    = makePolygon(southLoopHemi, true)
	farSouthHemiPolygon = makePolygon(farHemiSouthHemiLoop, true)
)

func TestPolygonInitSingleLoop(t *testing.T) {
	if !PolygonFromLoops([]*Loop{EmptyLoop()}).IsEmpty() {
		t.Errorf("polygon from Empty Loop should make an EmptyPolygon")
	}
	if !PolygonFromLoops([]*Loop{FullLoop()}).IsFull() {
		t.Errorf("polygon from Full Loop should make a FullPolygon")
	}
	p := PolygonFromLoops([]*Loop{makeLoop("0:0, 0:10, 10:0")})
	if got, want := p.numVertices, 3; got != want {
		t.Errorf("%v.numVertices = %v, want %v", p, got, want)
	}
}

func TestPolygonEmpty(t *testing.T) {
	if !emptyPolygon.IsEmpty() {
		t.Errorf("empty polygon should be empty")
	}
	if emptyPolygon.IsFull() {
		t.Errorf("empty polygon should not be full")
	}

	if emptyPolygon.ReferencePoint().Contained {
		t.Errorf("emptyPolygon.ReferencePoint().Contained = true, want false")
	}
	if got, want := emptyPolygon.NumEdges(), 0; got != want {
		t.Errorf("emptyPolygon.NumEdges() = %v, want %v", got, want)
	}
	if got := emptyPolygon.dimension(); got != polygonGeometry {
		t.Errorf("emptyPolygon.dimension() = %v, want %v", got, polygonGeometry)
	}
	if got, want := emptyPolygon.NumChains(), 0; got != want {
		t.Errorf("emptyPolygon.NumChains() = %v, want %v", got, want)
	}
}

func TestPolygonFull(t *testing.T) {
	if fullPolygon.IsEmpty() {
		t.Errorf("full polygon should not be emtpy")
	}
	if !fullPolygon.IsFull() {
		t.Errorf("full polygon should be full")
	}

	if !fullPolygon.ReferencePoint().Contained {
		t.Errorf("fullPolygon.ReferencePoint().Contained = false, want true")
	}
	if got, want := fullPolygon.NumEdges(), 0; got != want {
		t.Errorf("fullPolygon.NumEdges() = %v, want %v", got, want)
	}
	if got := fullPolygon.dimension(); got != polygonGeometry {
		t.Errorf("fullPolygon.dimension() = %v, want %v", got, polygonGeometry)
	}
	if got, want := fullPolygon.NumChains(), 0; got != want {
		t.Errorf("fullPolygon.NumChains() = %v, want %v", got, want)
	}
}

func TestPolygonShape(t *testing.T) {
	const numLoops = 100
	const numVerticesPerLoop = 6
	concentric := concentricLoopsPolygon(PointFromCoords(1, 0, 0), numLoops, numVerticesPerLoop)

	tests := []struct {
		p *Polygon
	}{
		{near0Polygon},    // one loop polygon
		{near0231Polygon}, // several loops polygon
		{concentric},      // many loops polygon
	}

	for _, test := range tests {
		shape := Shape(test.p)

		if test.p.numVertices != shape.NumEdges() {
			t.Errorf("the number of vertices in a polygon should equal the number of edges. got %v, want %v", test.p.numVertices, shape.NumEdges())
		}

		if test.p.NumLoops() != shape.NumChains() {
			t.Errorf("the number of loops in a polygon should equal the number of chains. got %v, want %v", test.p.NumLoops(), shape.NumChains())
		}

		edgeID := 0
		for i, l := range test.p.loops {
			if edgeID != shape.Chain(i).Start {
				t.Errorf("the edge id of the start of loop(%d) should equal the sum of vertices so far in the polygon. got %d, want %d", i, shape.Chain(i).Start, edgeID)
			}
			if len(l.vertices) != shape.Chain(i).Length {
				t.Errorf("the length of Chain(%d) should equal the length of loop(%d), got %v, want %v", i, i, shape.Chain(i).Length, len(l.vertices))
			}
			for j := 0; j < len(l.Vertices()); j++ {
				edge := shape.Edge(edgeID)
				if l.OrientedVertex(j) != edge.V0 {
					t.Errorf("l.Vertex(%d) = %v, want %v", j, l.Vertex(j), edge.V0)
				}
				if l.OrientedVertex(j+1) != edge.V1 {
					t.Errorf("l.Vertex(%d) = %v, want %v", j+1, l.Vertex(j+1), edge.V1)
				}
				edgeID++
			}
		}
		if shape.dimension() != polygonGeometry {
			t.Errorf("dimension = %v, want %v", shape.dimension(), polygonGeometry)
		}
		if !shape.HasInterior() {
			t.Errorf("polygons should always have interiors")
		}
		if got, want := test.p.ContainsPoint(OriginPoint()), shape.ReferencePoint().Contained; got != want {
			t.Errorf("p.ContainsPoint(OriginPoint()) != shape.ReferencePoint().Contained")
		}
	}
}

func TestPolygonParent(t *testing.T) {
	p1 := PolygonFromLoops([]*Loop{&Loop{}})
	tests := []struct {
		p    *Polygon
		have int
		want int
		ok   bool
	}{
		{fullPolygon, 0, -1, false},
		{p1, 0, -1, false},

		// TODO: When multiple loops are supported, add more test cases to
		// more fully show the parent levels.
	}

	for _, test := range tests {
		if got, ok := test.p.Parent(test.have); ok != test.ok || got != test.want {
			t.Errorf("%v.Parent(%d) = %d,%v, want %d,%v", test.p, test.have, got, ok, test.want, test.ok)
		}
	}
}

func TestPolygonLastDescendant(t *testing.T) {
	p1 := PolygonFromLoops([]*Loop{&Loop{}})

	tests := []struct {
		p    *Polygon
		have int
		want int
	}{
		{fullPolygon, 0, 0},
		{fullPolygon, -1, 0},

		{p1, 0, 0},
		{p1, -1, 0},

		// TODO: When multiple loops are supported, add more test cases.
	}

	for _, test := range tests {
		if got := test.p.LastDescendant(test.have); got != test.want {
			t.Errorf("%v.LastDescendant(%d) = %d, want %d", test.p, test.have, got, test.want)
		}
	}
}

func TestPolygonContainsPoint(t *testing.T) {
	tests := []struct {
		polygon string
		point   string
	}{
		{nearLoop0, nearPoint},
		{nearLoop1, nearPoint},
		{nearLoop2, nearPoint},
		{nearLoop3, nearPoint},
		{nearLoopHemi, nearPoint},
		{southLoop0a, southLoopPoint},
		{southLoop1, southLoopPoint},
		{southLoop2, southLoopPoint},
		{southLoopHemi, southLoopPoint},
	}

	for _, test := range tests {
		poly := makePolygon(test.polygon, true)
		pt := parsePoint(test.point)
		if !poly.ContainsPoint(pt) {
			t.Errorf("%v.ContainsPoint(%v) = false, want true", test.polygon, test.point)
		}
	}
}

// Given a pair of polygons where A contains B, check that various identities
// involving union, intersection, and difference operations hold true.
func testPolygonOneNestedPair(t *testing.T, a, b *Polygon) {
	if !a.Contains(b) {
		t.Errorf("%v.Contains(%v) = false, want true", a, b)
	}
	if got, want := a.Intersects(b), !b.IsEmpty(); got != want {
		t.Errorf("%v.Intersects(%v) = %v, want %v", a, b, got, want)
	}
	if got, want := b.Intersects(a), !b.IsEmpty(); got != want {
		t.Errorf("%v.Intersects(%v) = %v, want %v", b, a, got, want)
	}

	// TODO(roberts): Add the remaining checks related to construction
	// via union, intersection, and difference.
}

// Given a pair of disjoint polygons A and B, check that various identities
// involving union, intersection, and difference operations hold true.
func testPolygonOneDisjointPair(t *testing.T, a, b *Polygon) {
	if a.Intersects(b) {
		t.Errorf("%v.Intersects(%v) = true, want false", a, b)
	}
	if b.Intersects(a) {
		t.Errorf("%v.Intersects(%v) = true, want false", b, a)
	}
	if got, want := a.Contains(b), b.IsEmpty(); got != want {
		t.Errorf("%v.Contains(%v) = %v, want %v", b, a, got, want)
	}

	if got, want := b.Contains(a), a.IsEmpty(); got != want {
		t.Errorf("%v.Contains(%v) = %v, want %v", b, a, got, want)
	}

	// TODO(roberts): Add the remaining checks related to construction
	// via builder, union, intersection, and difference.
}

// Given polygons A and B whose union covers the sphere, check that various
// identities involving union, intersection, and difference hold true.
func testPolygonOneCoveringPair(t *testing.T, a, b *Polygon) {
	if got, want := a.Contains(b), a.IsFull(); got != want {
		t.Errorf("%v.Contains(%v) = %v, want %v", a, b, got, want)
	}
	if got, want := b.Contains(a), b.IsFull(); got != want {
		t.Errorf("%v.Contains(%v) = %v, want %v", a, b, got, want)
	}
	// TODO(roberts): Add the remaining checks related to construction via union

}

// Given polygons A and B such that both A and its complement intersect both B
// and its complement, check that various identities involving union,
// intersection, and difference hold true.
func testPolygonOneOverlappingPair(t *testing.T, a, b *Polygon) {
	if a.Contains(b) {
		t.Errorf("%v.Contains(%v) = true, want false", a, b)
	}
	if b.Contains(a) {
		t.Errorf("%v.Contains(%v) = true, want false", b, a)
	}
	if !a.Intersects(b) {
		t.Errorf("%v.Intersects(%v) = false, want true", a, b)
	}

	// TODO(roberts): Add the remaining checks related to construction
	// via builder, union, intersection, and difference.
}

// Given a pair of polygons where A contains B, test various identities
// involving A, B, and their complements.
func testPolygonNestedPair(t *testing.T, a, b *Polygon) {
	// TODO(roberts): Uncomment once complement is completed
	// a1 := InitToComplement(a)
	// b1 := InitToComplement(b)

	testPolygonOneNestedPair(t, a, b)
	// testPolygonOneNestedPair(t, b1, a1)
	// testPolygonOneDisjointPair(t, a1, b)
	// testPolygonOneCoveringPair(t, a, b1)
}

// Given a pair of disjoint polygons A and B, test various identities
// involving A, B, and their complements.
func testPolygonDisjointPair(t *testing.T, a, b *Polygon) {
	// TODO(roberts): Uncomment once complement is completed
	// a1 := InitToComplement(a)
	// b1 := InitToComplement(b)

	testPolygonOneDisjointPair(t, a, b)
	// testPolygonOneCoveringPair(t, a1, b1)
	// testPolygonOneNestedPair(t, a1, b)
	// testPolygonOneNestedPair(t, b1, a)
}

// Given polygons A and B such that both A and its complement intersect both B
// and its complement, test various identities involving these four polygons.
func testPolygonOverlappingPair(t *testing.T, a, b *Polygon) {
	// TODO(roberts): Uncomment once complement is completed
	// a1 := InitToComplement(a);
	// b1 := InitToComplement(b);

	testPolygonOneOverlappingPair(t, a, b)
	// testPolygonOneOverlappingPair(t, a1, b1);
	// testPolygonOneOverlappingPair(t, a1, b);
	// testPolygonOneOverlappingPair(t, a, b1);
}

// Test identities that should hold for any pair of polygons A, B and their
// complements.
func testPolygonComplements(t *testing.T, a, b *Polygon) {
	// TODO(roberts): Uncomment once complement is completed
	// a1 := InitToComplement(a);
	// b1 := InitToComplement(b);

	// testOneComplementPair(t, a, a1, b, b1)
	// testOneComplementPair(t, a1, a, b, b1)
	// testOneComplementPair(t, a, a1, b1, b)
	// testOneComplementPair(t, a1, a, b1, b)

	// TODO(roberts): Add the checks related to construction via union, etc.
}

func testPolygonDestructiveUnion(t *testing.T, a, b *Polygon) {
	// TODO(roberts): Add the checks related to construction via union, etc.
}

func TestPolygonRelations(t *testing.T) {
	tests := []struct {
		a, b       *Polygon
		contains   bool // A contains B
		contained  bool // B contains A
		intersects bool // A and B intersect
	}{
		{
			a:          near01Polygon,
			b:          emptyPolygon,
			contains:   true,
			contained:  false,
			intersects: false,
		},
		{
			a:          near01Polygon,
			b:          near01Polygon,
			contains:   true,
			contained:  true,
			intersects: true,
		},
		{
			a:          fullPolygon,
			b:          near01Polygon,
			contains:   true,
			contained:  false,
			intersects: true,
		},
		{
			a:          near01Polygon,
			b:          near30Polygon,
			contains:   false,
			contained:  true,
			intersects: true,
		},
		{
			a:          near01Polygon,
			b:          near23Polygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},
		{
			a:          near01Polygon,
			b:          near0231Polygon,
			contains:   false,
			contained:  true,
			intersects: true,
		},
		{
			a:          near01Polygon,
			b:          near023H1Polygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},
		{
			a:          near30Polygon,
			b:          near23Polygon,
			contains:   true,
			contained:  false,
			intersects: true,
		},
		{
			a:          near30Polygon,
			b:          near0231Polygon,
			contains:   true,
			contained:  false,
			intersects: true,
		},
		{
			a:          near30Polygon,
			b:          near023H1Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          near23Polygon,
			b:          near0231Polygon,
			contains:   false,
			contained:  true,
			intersects: true,
		},
		{
			a:          near23Polygon,
			b:          near023H1Polygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},
		{
			a:          near0231Polygon,
			b:          near023H1Polygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},

		{
			a:          far01Polygon,
			b:          far21Polygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},
		{
			a:          far01Polygon,
			b:          far231Polygon,
			contains:   false,
			contained:  true,
			intersects: true,
		},
		{
			a:          far01Polygon,
			b:          far2H0Polygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},
		{
			a:          far01Polygon,
			b:          far2H013Polygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},
		{
			a:          far21Polygon,
			b:          far231Polygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},
		{
			a:          far21Polygon,
			b:          far2H0Polygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},
		{
			a:          far21Polygon,
			b:          far2H013Polygon,
			contains:   false,
			contained:  true,
			intersects: true,
		},
		{
			a:          far231Polygon,
			b:          far2H0Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          far231Polygon,
			b:          far2H013Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          far2H0Polygon,
			b:          far2H013Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},

		{
			a:          south0abPolygon,
			b:          south2Polygon,
			contains:   false,
			contained:  true,
			intersects: true,
		},
		{
			a:          south0abPolygon,
			b:          south20b1Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          south0abPolygon,
			b:          south2H1Polygon,
			contains:   false,
			contained:  true,
			intersects: true,
		},
		{
			a:          south0abPolygon,
			b:          south20bH0acPolygon,
			contains:   false,
			contained:  true,
			intersects: true,
		},
		{
			a:          south2Polygon,
			b:          south20b1Polygon,
			contains:   true,
			contained:  false,
			intersects: true,
		},
		{
			a:          south2Polygon,
			b:          south2H1Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          south2Polygon,
			b:          south20bH0acPolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          south20b1Polygon,
			b:          south2H1Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          south20b1Polygon,
			b:          south20bH0acPolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          south2H1Polygon,
			b:          south20bH0acPolygon,
			contains:   true,
			contained:  false,
			intersects: true,
		},

		{
			a:          nf1N10F2S10abcPolygon,
			b:          nf2N2F210S210abPolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          nf1N10F2S10abcPolygon,
			b:          near23Polygon,
			contains:   true,
			contained:  false,
			intersects: true,
		},
		{
			a:          nf1N10F2S10abcPolygon,
			b:          far21Polygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},
		{
			a:          nf1N10F2S10abcPolygon,
			b:          south0abPolygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},
		{
			a:          nf1N10F2S10abcPolygon,
			b:          f32n0Polygon,
			contains:   true,
			contained:  false,
			intersects: true,
		},

		{
			a:          nf2N2F210S210abPolygon,
			b:          near01Polygon,
			contains:   false,
			contained:  false,
			intersects: false,
		},
		{
			a:          nf2N2F210S210abPolygon,
			b:          far01Polygon,
			contains:   true,
			contained:  false,
			intersects: true,
		},
		{
			a:          nf2N2F210S210abPolygon,
			b:          south20b1Polygon,
			contains:   true,
			contained:  false,
			intersects: true,
		},
		{
			a:          nf2N2F210S210abPolygon,
			b:          south0abPolygon,
			contains:   true,
			contained:  false,
			intersects: true,
		},
		{
			a:          nf2N2F210S210abPolygon,
			b:          n32s0bPolygon,
			contains:   true,
			contained:  false,
			intersects: true,
		},
		{
			a:          cross1Polygon,
			b:          cross2Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          cross1SideHolePolygon,
			b:          cross2Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          cross1CenterHolePolygon,
			b:          cross2Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          cross1Polygon,
			b:          cross2SideHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          cross1Polygon,
			b:          cross2CenterHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          cross1SideHolePolygon,
			b:          cross2SideHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          cross1CenterHolePolygon,
			b:          cross2SideHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          cross1SideHolePolygon,
			b:          cross2CenterHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          cross1CenterHolePolygon,
			b:          cross2CenterHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		// These cases, when either polygon has a hole, test a different code path
		// from the other cases.
		{
			a:          overlap1Polygon,
			b:          overlap2Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          overlap1SideHolePolygon,
			b:          overlap2Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          overlap1CenterHolePolygon,
			b:          overlap2Polygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          overlap1Polygon,
			b:          overlap2SideHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          overlap1Polygon,
			b:          overlap2CenterHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          overlap1SideHolePolygon,
			b:          overlap2SideHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          overlap1CenterHolePolygon,
			b:          overlap2SideHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          overlap1SideHolePolygon,
			b:          overlap2CenterHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
		{
			a:          overlap1CenterHolePolygon,
			b:          overlap2CenterHolePolygon,
			contains:   false,
			contained:  false,
			intersects: true,
		},
	}

	for i, test := range tests {
		if got := test.a.Contains(test.b); got != test.contains {
			t.Errorf("%d. %v.Contains(%v) = %v, want %v", i, test.a, test.b, got, test.contains)
		}

		if got := test.b.Contains(test.a); got != test.contained {
			t.Errorf("%d. %v.Contains(%v) = %v, want %v", i, test.b, test.a, got, test.contained)
		}

		if got := test.a.Intersects(test.b); got != test.intersects {
			t.Errorf("%v.Intersects(%v) = %v, want %v", test.a, test.b, got, test.intersects)
		}

		if test.contains {
			testPolygonNestedPair(t, test.a, test.b)
		}
		if test.contained {
			testPolygonNestedPair(t, test.b, test.a)
		}
		if !test.intersects {
			testPolygonDisjointPair(t, test.a, test.b)
		}
		if test.intersects && !(test.contains || test.contained) {
			testPolygonOverlappingPair(t, test.a, test.b)
		}
		testPolygonDestructiveUnion(t, test.a, test.b)
		testPolygonComplements(t, test.a, test.b)
	}

	testPolygonNestedPair(t, emptyPolygon, emptyPolygon)
	testPolygonNestedPair(t, fullPolygon, emptyPolygon)
	testPolygonNestedPair(t, fullPolygon, fullPolygon)
}

// TODO(roberts): Remaining Tests
// TestInit
// TestMultipleInit
// TestInitSingleLoop
// TestCellConstructorAndContains
// TestUninitializedIsValid
// TestOverlapFractions
// TestOriginNearPole
// TestTestApproxContainsAndDisjoint
// TestOneLoopPolygonShape
// TestSeveralLoopPolygonShape
// TestManyLoopPolygonShape
// TestPointInBigLoop
// TestOperations
// TestIntersectionSnapFunction
// TestIntersectionPreservesLoopOrder
// TestLoopPointers
// TestBug1 - Bug14
// TestPolylineIntersection
// TestSplitting
// TestInitToCellUnionBorder
// TestUnionWithAmbgiuousCrossings
// TestInitToSloppySupportsEmptyPolygons
// TestInitToSnappedDoesNotRotateVertices
// TestInitToSnappedWithSnapLevel
// TestInitToSnappedIsValid_A
// TestInitToSnappedIsValid_B
// TestInitToSnappedIsValid_C
// TestInitToSnappedIsValid_D
// TestProject
// TestDistance
//
// IsValidTests
//   TestUnitLength
//   TestVertexCount
//   TestDuplicateVertex
//   TestSelfIntersection
//   TestEmptyLoop
//   TestFullLoop
//   TestLoopsCrossing
//   TestDuplicateEdge
//   TestInconsistentOrientations
//   TestLoopDepthNegative
//   TestLoopNestingInvalid
//   TestFuzzTest
//
// PolygonSimplifier
//   TestNoSimplification
//   TestSimplifiedLoopSelfIntersects
//   TestNoSimplificationManyLoops
//   TestTinyLoopDisappears
//   TestStraightLinesAreSimplified
//   TestEdgeSplitInManyPieces
//   TestEdgesOverlap
//   TestLargeRegularPolygon
//
// InitToSimplifiedInCell
//   TestPointsOnCellBoundaryKept
//   TestPointsInsideCellSimplified
//   TestCellCornerKept
//   TestNarrowStripRemoved
//   TestNarrowGapRemoved
//   TestCloselySpacedEdgeVerticesKept
//   TestPolylineAssemblyBug
