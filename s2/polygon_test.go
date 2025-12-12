// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s2

import (
	"bytes"
	"math"
	"math/rand"
	"testing"

	"github.com/golang/geo/s1"
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

	// Rectangles that form a cross, with only shared vertices, no crossing
	// edges. Optional holes outside the intersecting region.
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
	shape := emptyPolygon

	if got, want := shape.NumEdges(), 0; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), 0; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Dimension(), 2; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if !shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = false, want true")
	}
	if shape.IsFull() {
		t.Errorf("shape.IsFull() = true, want false")
	}
	if shape.ReferencePoint().Contained {
		t.Errorf("shape.ReferencePoint().Contained = true, want false")
	}
}

func TestPolygonFull(t *testing.T) {
	shape := fullPolygon

	if got, want := shape.NumEdges(), 0; got != want {
		t.Errorf("shape.NumEdges() = %v, want %v", got, want)
	}
	if got, want := shape.NumChains(), 1; got != want {
		t.Errorf("shape.NumChains() = %v, want %v", got, want)
	}
	if got, want := shape.Chain(0).Start, 0; got != want {
		t.Errorf("fullPolygon.Chain(0).Start = %d, want %d", got, want)
	}
	if got, want := shape.Chain(0).Length, 0; got != want {
		t.Errorf("fullPolygon.Chain(0).Length = %d, want %d", got, want)
	}
	if got, want := shape.Dimension(), 2; got != want {
		t.Errorf("shape.Dimension() = %v, want %v", got, want)
	}
	if shape.IsEmpty() {
		t.Errorf("shape.IsEmpty() = true, want false")
	}
	if !shape.IsFull() {
		t.Errorf("shape.IsFull() = false, want true")
	}
	if !shape.ReferencePoint().Contained {
		t.Errorf("shape.ReferencePoint().Contained = false, want true")
	}
}

func TestPolygonInitLoopPropertiesGetsRightBounds(t *testing.T) {
	// Before the change to initLoopProperties to start the bounds as an
	// EmptyRect instead of it default to the zero rect, the bounds
	// computations failed. Lo was set to min (0, 12.55) and Hi was set to
	// max (0, -70.02).  So this poly would have a bounds of
	//   Lo: [0, -70.05],     Hi: [12.58, 0]]      instead of:
	//   Lo: [12.55, -70.05], Hi: [12.58, -70.02]]

	p := PolygonFromLoops([]*Loop{
		makeLoop("12.55:-70.05, 12.55:-70.02, 12.58:-70.02, 12.58:-70.05"),
		makeLoop("12.56:-70.04, 12.56:-70.03, 12.58:-70.03, 12.58:-70.04"),
	})
	want := rectFromDegrees(12.55, -70.05, 12.58, -70.02)
	if got := p.RectBound(); !rectsApproxEqual(got, want, 1e-6, 1e-6) {
		t.Errorf("%v.RectBound() = %v, want %v", p, got, want)
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

		if got, want := shape.NumEdges(), test.p.numVertices; got != want {
			t.Errorf("the number of vertices in a polygon should equal the number of edges. got %v, want %v", got, want)
		}
		if got, want := shape.NumChains(), test.p.NumLoops(); got != want {
			t.Errorf("the number of loops in a polygon should equal the number of chains. got %v, want %v", got, want)
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
		if got, want := shape.Dimension(), 2; got != want {
			t.Errorf("shape.Dimension() = %v, want %v", got, want)
		}
		if shape.IsEmpty() {
			t.Errorf("shape.IsEmpty() = true, want false")
		}
		if shape.IsFull() {
			t.Errorf("shape.IsFull() = true, want false")
		}

		if got, want := test.p.ContainsPoint(OriginPoint()), shape.ReferencePoint().Contained; got != want {
			t.Errorf("p.ContainsPoint(OriginPoint()) != shape.ReferencePoint().Contained")
		}
	}
}

// reverseLoopVertices reverses the order of all vertices in the given loop.
func reverseLoopVertices(l *Loop) {
	for i := 0; i < len(l.vertices)/2; i++ {
		l.vertices[i], l.vertices[len(l.vertices)-i-1] = l.vertices[len(l.vertices)-i-1], l.vertices[i]
	}
}

// shuffleLoops randomizes the slice of loops using Fisher-Yates shuffling.
func shuffleLoops(loops []*Loop) {
	n := len(loops)
	for i := range n {
		// choose index uniformly in [i, n-1]
		r := i + rand.Intn(n-i)
		loops[r], loops[i] = loops[i], loops[r]
	}
}

// modifyPolygonFunc declares a function that can tweak a Polygon for testing.
type modifyPolygonFunc func(p *Polygon)

// polygonSetInvalidLoopNesting flips a random loops orientation within the polygon.
func polygonSetInvalidLoopNesting(p *Polygon) {
	if len(p.loops) > 0 {
		i := randomUniformInt(len(p.loops))
		p.loops[i].Invert()
	}
}

// polygonSetInvalidLoopDepth randomly changes a loops depth value in the given polygon
func polygonSetInvalidLoopDepth(p *Polygon) {
	i := randomUniformInt(len(p.loops))
	if i == 0 || oneIn(3) {
		p.loops[i].depth = -1
	} else {
		p.loops[i].depth = p.loops[i-1].depth + 2
	}
}

// generatePolygonConcentricTestLoops creates the given number of nested regular
// loops around a common center point. All loops will have the same number of
// vertices (at least minVertices). Furthermore, the vertices at the same index
// position are collinear with the common center point of all the loops. The
// loop radii decrease exponentially in order to prevent accidental loop crossings
// when one of the loops is modified.
func generatePolygonConcentricTestLoops(numLoops, minVertices int) []*Loop {
	var loops []*Loop
	center := randomPoint()
	numVertices := minVertices + randomUniformInt(10)
	for i := range numLoops {
		radius := s1.Angle(80*math.Pow(0.1, float64(i))) * s1.Degree
		loops = append(loops, RegularLoop(center, radius, numVertices))
	}
	return loops
}

func checkPolygonInvalid(t *testing.T, label string, loops []*Loop, initOriented bool, f modifyPolygonFunc) {
	shuffleLoops(loops)
	var polygon *Polygon
	if initOriented {
		polygon = PolygonFromOrientedLoops(loops)
	} else {
		polygon = PolygonFromLoops(loops)
	}

	if f != nil {
		f(polygon)
	}

	if err := polygon.Validate(); err == nil {
		t.Errorf("%s: %v.Validate() = %v, want non-nil", label, polygon, err)
	}
}

func TestPolygonUninitializedIsValid(t *testing.T) {
	p := &Polygon{}
	if err := p.Validate(); err != nil {
		t.Errorf("an uninitialized polygon failed validation: %v", err)
	}
}

func TestPolygonIsValidLoopNestingInvalid(t *testing.T) {
	const iters = 1000

	for range iters {
		loops := generatePolygonConcentricTestLoops(2+randomUniformInt(4), 3)
		// Randomly invert all the loops in order to generate cases where the
		// outer loop encompasses almost the entire sphere. This tests different
		// code paths because bounding box checks are not as useful.
		if oneIn(2) {
			for _, loop := range loops {
				reverseLoopVertices(loop)
			}
		}
		checkPolygonInvalid(t, "invalid nesting", loops, false, polygonSetInvalidLoopNesting)
	}
}

// TODO(roberts): Implement remaining validity tests.
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

func TestPolygonParent(t *testing.T) {
	p1 := PolygonFromLoops([]*Loop{{}})
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
	p1 := PolygonFromLoops([]*Loop{{}})

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
		if test.intersects && (!test.contains && !test.contained) {
			testPolygonOverlappingPair(t, test.a, test.b)
		}
		testPolygonDestructiveUnion(t, test.a, test.b)
		testPolygonComplements(t, test.a, test.b)
	}

	testPolygonNestedPair(t, emptyPolygon, emptyPolygon)
	testPolygonNestedPair(t, fullPolygon, emptyPolygon)
	testPolygonNestedPair(t, fullPolygon, fullPolygon)
}

func TestPolygonArea(t *testing.T) {
	tests := []struct {
		have *Polygon
		want float64
	}{
		{have: emptyPolygon, want: 0},
		{have: fullPolygon, want: 4 * math.Pi},
		{have: southHemiPolygon, want: 2 * math.Pi},
		{have: farSouthHemiPolygon, want: math.Pi},
		{
			// compare the polygon of two shells to the sum of its loops.
			have: makePolygon(loopCross1SideHole+loopCrossCenterHole, true),
			// the strings for the loops contain ';' so copy and paste without it
			want: makeLoop("-1.5:0.5, -1.2:0.5, -1.2:-0.5, -1.5:-0.5").Area() +
				makeLoop("-0.5:0.5, 0.5:0.5, 0.5:-0.5, -0.5:-0.5").Area(),
		},
		{
			// test that polygon with a shell and a hole matches its loop parts.
			have: makePolygon(loopCross1+loopCrossCenterHole, true),
			// the strings for the loops contain ';' so copy and paste without it
			want: makeLoop("-2:1, -1:1, 1:1, 2:1, 2:-1, 1:-1, -1:-1, -2:-1").Area() -
				makeLoop("-0.5:0.5, 0.5:0.5, 0.5:-0.5, -0.5:-0.5").Area(),
		},
	}

	for _, test := range tests {
		if got := test.have.Area(); !float64Eq(got, test.want) {
			t.Errorf("%v.Area() = %v, want %v", test.have, got, test.want)
		}
	}
}

func TestPolygonCentroid(t *testing.T) {
	tests := []struct {
		have *Polygon
		want Point
	}{
		{have: emptyPolygon, want: Point{}},
		{have: fullPolygon, want: Point{}},
		{
			// compare the polygon of two shells to the sum of its loops.
			have: makePolygon(loopCross1SideHole+loopCrossCenterHole, true),
			// the strings for the loops contain ';' so copy and paste without it
			want: Point{
				makeLoop("-1.5:0.5, -1.2:0.5, -1.2:-0.5, -1.5:-0.5").Centroid().Add(
					makeLoop("-0.5:0.5, 0.5:0.5, 0.5:-0.5, -0.5:-0.5").Centroid().Vector)},
		},
		{
			// test that polygon with a shell and a hole matches its loop parts.
			have: makePolygon(loopCross1+loopCrossCenterHole, true),
			// the strings for the loops contain ';' so copy and paste without it
			want: Point{
				makeLoop("-2:1, -1:1, 1:1, 2:1, 2:-1, 1:-1, -1:-1, -2:-1").Centroid().Sub(
					makeLoop("-0.5:0.5, 0.5:0.5, 0.5:-0.5, -0.5:-0.5").Centroid().Vector)},
		},
	}

	for _, test := range tests {
		if got := test.have.Centroid(); got.Cmp(test.want.Vector) != 0 {
			t.Errorf("%v.Centroid() = %v, want %v", test.have, got, test.want)
		}
	}
}

func TestPolygonInvert(t *testing.T) {
	origin := PointFromLatLng(LatLngFromDegrees(0, 0))
	pt := PointFromLatLng(LatLngFromDegrees(30, 30))
	p := PolygonFromLoops([]*Loop{
		RegularLoop(origin, 1000/earthRadiusKm, 100),
	})

	if p.ContainsPoint(pt) {
		t.Errorf("polygon contains point outside itself")
	}

	p.Invert()
	if !p.ContainsPoint(pt) {
		t.Errorf("inverted polygon does not contain point that is inside itself")
	}
}

func TestPolygonOfCellUnionBorder(t *testing.T) {
	t.Run("empty cell union", func(t *testing.T) {
		p, err := PolygonOfCellUnionBorder(CellUnion{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !p.IsEmpty() {
			t.Errorf("expected empty polygon, got %d loops", p.NumLoops())
		}
	})

	t.Run("single cell", func(t *testing.T) {
		cellID := CellIDFromFace(0).ChildBeginAtLevel(5)
		cells := CellUnion{cellID}

		p, err := PolygonOfCellUnionBorder(cells)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if p.NumLoops() != 1 {
			t.Errorf("expected 1 loop, got %d", p.NumLoops())
		}
		if p.Loop(0).NumVertices() != 4 {
			t.Errorf("expected 4 vertices, got %d", p.Loop(0).NumVertices())
		}
		if !p.ContainsPoint(CellFromCellID(cellID).Center()) {
			t.Error("polygon should contain cell center")
		}
	})

	t.Run("two adjacent cells", func(t *testing.T) {
		cell := CellIDFromFace(0).ChildBeginAtLevel(5)
		neighbor := cell.EdgeNeighbors()[0]
		cells := CellUnion{cell, neighbor}
		cells.Normalize()

		p, err := PolygonOfCellUnionBorder(cells)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if p.NumLoops() != 1 {
			t.Errorf("expected 1 loop, got %d", p.NumLoops())
		}
		// Two adjacent cells share one edge, so 4+4-2 = 6 vertices
		if p.Loop(0).NumVertices() != 6 {
			t.Errorf("expected 6 vertices, got %d", p.Loop(0).NumVertices())
		}
	})

	t.Run("merged siblings (4 cells = square)", func(t *testing.T) {
		parent := CellIDFromFace(1).ChildBeginAtLevel(8)
		children := parent.Children()
		cells := CellUnion{children[0], children[1], children[2], children[3]}
		cells.Normalize()

		p, err := PolygonOfCellUnionBorder(cells)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if p.NumLoops() != 1 {
			t.Fatalf("NumLoops() = %d, want 1", p.NumLoops())
		}
		// 4 sibling cells form a square with 4 vertices (internal edges cancel)
		if p.Loop(0).NumVertices() != 4 {
			t.Errorf("expected 4 vertices for 2x2 cell block, got %d", p.Loop(0).NumVertices())
		}
	})

	t.Run("full sphere", func(t *testing.T) {
		cells := make(CellUnion, NumFaces)
		for face := 0; face < NumFaces; face++ {
			cells[face] = CellIDFromFace(face)
		}
		cells.Normalize()

		p, err := PolygonOfCellUnionBorder(cells)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !p.IsFull() {
			t.Error("expected full polygon for complete sphere coverage")
		}
	})
}

func TestPolygonOfCellUnionBorder_Hole(t *testing.T) {
	// Create a shape with a hole: a ring of cells surrounding an empty center.
	center := CellIDFromFace(0).ChildBeginAtLevel(5)
	neighbors := center.AllNeighbors(5)

	// Use only the neighbors, not the center - creates a ring with a hole.
	cells := CellUnion(neighbors)
	cells.Normalize()

	p, err := PolygonOfCellUnionBorder(cells)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.NumLoops() != 2 {
		t.Fatalf("NumLoops() = %d, want 2 (outer + hole)", p.NumLoops())
	}

	// The hole center should NOT be contained by the polygon.
	centerPoint := CellFromCellID(center).Center()
	if p.ContainsPoint(centerPoint) {
		t.Error("polygon contains the hole center, should not")
	}

	// Each neighbor cell center should be contained by the polygon.
	for _, neighborID := range neighbors {
		neighborCenter := CellFromCellID(neighborID).Center()
		if !p.ContainsPoint(neighborCenter) {
			t.Errorf("polygon does not contain neighbor %v", neighborID)
		}
	}
}

func TestPolygonOfCellUnionBorder_FullSphereWithHole(t *testing.T) {
	// Full sphere = 24 cells at level 1 (6 faces * 4 children)
	var cells CellUnion
	for face := 0; face < NumFaces; face++ {
		for _, child := range CellIDFromFace(face).Children() {
			cells = append(cells, child)
		}
	}

	// Remove one cell to create a hole
	hole := cells[0]
	cells = cells[1:]
	cells.Normalize()

	p, err := PolygonOfCellUnionBorder(cells)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.NumLoops() != 2 {
		t.Fatalf("NumLoops() = %d, want 2", p.NumLoops())
	}

	// The hole center should NOT be contained
	holeCenter := CellFromCellID(hole).Center()
	if p.ContainsPoint(holeCenter) {
		t.Error("polygon contains the hole center, should not")
	}

	// A cell in the union should be contained
	testCell := cells[(len(cells) / 2)]
	testCenter := CellFromCellID(testCell).Center()
	if !p.ContainsPoint(testCenter) {
		t.Error("polygon does not contain a cell that should be in the union")
	}

	// Area should be close to full sphere minus one cell
	expectedArea := FullPolygon().Area() - CellFromCellID(hole).ExactArea()
	if math.Abs(p.Area()-expectedArea) > 0.001 {
		t.Errorf("Area() = %.6f, want %.6f", p.Area(), expectedArea)
	}
}

func TestPolygonOfCellUnionBorder_MixedLevels(t *testing.T) {
	// Test that cells at different levels properly merge without internal
	// boundary artifacts. This creates a region where a large cell (level 3)
	// is adjacent to smaller cells (level 5).

	// Start with a level 3 cell
	largeCell := CellIDFromFace(0).ChildBeginAtLevel(3)

	// Get an adjacent cell at level 3, then break it into level 5 children,
	// but OMIT ONE to prevent Normalize() from merging them back.
	adjacentLarge := largeCell.Next()
	var smallCells CellUnion
	for i, child := range adjacentLarge.Children() {
		for j, grandchild := range child.Children() {
			// Skip one grandchild to prevent normalization merging
			if i == 0 && j == 0 {
				continue
			}
			smallCells = append(smallCells, grandchild)
		}
	}

	// Combine the large cell with the small cells
	cells := CellUnion{largeCell}
	cells = append(cells, smallCells...)
	cells.Normalize()

	// Verify we actually have mixed levels
	minLevel, maxLevel := 30, 0
	for _, cell := range cells {
		level := cell.Level()
		if level < minLevel {
			minLevel = level
		}
		if level > maxLevel {
			maxLevel = level
		}
	}
	if minLevel == maxLevel {
		t.Fatalf(
			"test setup error: all cells at level %d, expected mixed levels",
			minLevel,
		)
	}

	p, err := PolygonOfCellUnionBorder(cells)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.NumLoops() < 1 {
		t.Fatal("expected at least 1 loop")
	}

	// All cell centers should be contained
	for _, cellID := range cells {
		center := CellFromCellID(cellID).Center()
		if !p.ContainsPoint(center) {
			t.Errorf("polygon does not contain cell %v center", cellID)
		}
	}
}

func TestPolygonOfCellUnionBorder_CrossFaceBoundary(t *testing.T) {
	// Test that adjacent cells crossing a face boundary form a single polygon.
	// Face 0 and Face 5 share an edge, so cells at their shared boundary
	// should produce a merged polygon with 6 vertices (two squares sharing an edge).
	cellA := CellIDFromFace(0).ChildBeginAtLevel(2)
	cellB := CellIDFromFace(5).ChildEndAtLevel(2).Prev()

	cells := CellUnion{cellA, cellB}
	cells.Normalize()

	p, err := PolygonOfCellUnionBorder(cells)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.NumLoops() != 1 {
		t.Fatalf("NumLoops() = %d, want 1", p.NumLoops())
	}

	// Two adjacent cells share one edge, so 4+4-2 = 6 vertices
	if p.Loop(0).NumVertices() != 6 {
		t.Errorf("NumVertices() = %d, want 6", p.Loop(0).NumVertices())
	}
}

func TestPolygonOfCellUnionBorder_LTJunction(t *testing.T) {
	// Test L-junction: 3 cells where two coarse cells meet one fine cell.
	// Cell A (level 3), Cell B (level 3) adjacent to A, Cell C (level 9) also
	// adjacent to A on a different edge. Internal edges should not leak.
	cellA := CellIDFromFace(0).ChildBeginAtLevel(3)
	cellB := cellA.EdgeNeighbors()[0]
	cellC := cellA.EdgeNeighbors()[1]
	for cellC.Level() < 9 {
		cellC = cellC.ChildBegin()
	}

	cells := CellUnion{cellA, cellB, cellC}
	cells.Normalize()

	p, err := PolygonOfCellUnionBorder(cells)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.NumLoops() > 2 {
		t.Errorf("NumLoops() = %d, want <= 2", p.NumLoops())
	}

	for _, cellID := range cells {
		center := CellFromCellID(cellID).Center()
		if !p.ContainsPoint(center) {
			t.Errorf("polygon does not contain cell %v center", cellID)
		}
	}
}

func TestPolygonOfCellUnionBorder_ManyLevelTransitions(t *testing.T) {
	// Test cells at multiple levels (11-15) adjacent to each other, simulating
	// real-world coastline data. The bug produced many spurious loops due to
	// edge level mismatches.
	var cells CellUnion

	seed := CellIDFromFace(0).ChildBeginAtLevel(11)
	cells = append(cells, seed)
	for _, n := range seed.AllNeighbors(11) {
		cells = append(cells, n)
	}

	// Add finer cells to create level transitions
	fineArea := seed.EdgeNeighbors()[0]
	for fineArea.Level() < 15 {
		fineArea = fineArea.ChildBegin()
	}
	cells = append(cells, fineArea)
	for _, n := range fineArea.AllNeighbors(15) {
		cells = append(cells, n)
	}

	cells.Normalize()

	p, err := PolygonOfCellUnionBorder(cells)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.NumLoops() > 15 {
		t.Errorf("NumLoops() = %d, want <= 15", p.NumLoops())
	}
}

// TODO(roberts): Remaining Tests
// TestInit
// TestMultipleInit
// TestInitSingleLoop
// TestCellConstructorAndContains
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

// go test -fuzz=FuzzDecodePolygon github.com/golang/geo/s2
func FuzzDecodePolygon(f *testing.F) {
	for _, p := range []*Polygon{near0Polygon, near01Polygon, near30Polygon, near23Polygon, far01Polygon, far21Polygon, south0abPolygon} {
		buf := new(bytes.Buffer)
		if err := p.Encode(buf); err != nil {
			f.Errorf("error encoding %v: ", err)
		}
		f.Add(buf.Bytes())
	}

	f.Fuzz(func(t *testing.T, encoded []byte) {
		p := &Polygon{}
		if err := p.Decode(bytes.NewReader(encoded)); err != nil {
			// Construction failed, no need to test further.
			return
		}
		if got := p.Area(); got < 0 {
			t.Errorf("Area() = %v, want >= 0. Polygon: %v", got, p)
		}
		// TODO: Test more methods on Polygon.
	})
}
