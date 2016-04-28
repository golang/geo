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
	"math"
	"testing"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r3"
	"github.com/golang/geo/s1"
)

var (
	// The northern hemisphere, defined using two pairs of antipodal points.
	northHemi = LoopFromPoints(parsePoints("0:-180, 0:-90, 0:0, 0:90"))

	// The northern hemisphere, defined using three points 120 degrees apart.
	northHemi3 = LoopFromPoints(parsePoints("0:-180, 0:-60, 0:60"))

	// The southern hemisphere, defined using two pairs of antipodal points.
	southHemi = LoopFromPoints(parsePoints("0:90, 0:0, 0:-90, 0:-180"))

	// The western hemisphere, defined using two pairs of antipodal points.
	westHemi = LoopFromPoints(parsePoints("0:-180, -90:0, 0:0, 90:0"))

	// The eastern hemisphere, defined using two pairs of antipodal points.
	eastHemi = LoopFromPoints(parsePoints("90:0, 0:0, -90:0, 0:-180"))

	// The "near" hemisphere, defined using two pairs of antipodal points.
	nearHemi = LoopFromPoints(parsePoints("0:-90, -90:0, 0:90, 90:0"))

	// The "far" hemisphere, defined using two pairs of antipodal points.
	farHemi = LoopFromPoints(parsePoints("90:0, 0:90, -90:0, 0:-90"))

	// A spiral stripe that slightly over-wraps the equator.
	candyCane = LoopFromPoints(parsePoints("-20:150, -20:-70, 0:70, 10:-150, 10:70, -10:-70"))

	// A small clockwise loop in the northern & eastern hemisperes.
	smallNECW = LoopFromPoints(parsePoints("35:20, 45:20, 40:25"))

	// Loop around the north pole at 80 degrees.
	arctic80 = LoopFromPoints(parsePoints("80:-150, 80:-30, 80:90"))

	// Loop around the south pole at 80 degrees.
	antarctic80 = LoopFromPoints(parsePoints("-80:120, -80:0, -80:-120"))

	// A completely degenerate triangle along the equator that RobustCCW()
	// considers to be CCW.
	lineTriangle = LoopFromPoints(parsePoints("0:1, 0:2, 0:3"))

	// A nearly-degenerate CCW chevron near the equator with very long sides
	// (about 80 degrees).  Its area is less than 1e-640, which is too small
	// to represent in double precision.
	skinnyChevron = LoopFromPoints(parsePoints("0:0, -1e-320:80, 0:1e-320, 1e-320:80"))

	// A diamond-shaped loop around the point 0:180.
	loopA = LoopFromPoints(parsePoints("0:178, -1:180, 0:-179, 1:-180"))

	// Like loopA, but the vertices are at leaf cell centers.
	snappedLoopA = LoopFromPoints([]Point{
		CellIDFromLatLng(parseLatLngs("0:178")[0]).Point(),
		CellIDFromLatLng(parseLatLngs("-1:180")[0]).Point(),
		CellIDFromLatLng(parseLatLngs("0:-179")[0]).Point(),
		CellIDFromLatLng(parseLatLngs("1:-180")[0]).Point(),
	})

	// A different diamond-shaped loop around the point 0:180.
	loopB = LoopFromPoints(parsePoints("0:179, -1:180, 0:-178, 1:-180"))

	// The intersection of A and B.
	aIntersectB = LoopFromPoints(parsePoints("0:179, -1:180, 0:-179, 1:-180"))

	// The union of A and B.
	aUnionB = LoopFromPoints(parsePoints("0:178, -1:180, 0:-178, 1:-180"))

	// A minus B (concave).
	aMinusB = LoopFromPoints(parsePoints("0:178, -1:180, 0:179, 1:-180"))

	// B minus A (concave).
	bMinusA = LoopFromPoints(parsePoints("0:-179, -1:180, 0:-178, 1:-180"))

	// A shape gotten from A by adding a triangle to one edge, and
	// subtracting a triangle from the opposite edge.
	loopC = LoopFromPoints(parsePoints("0:178, 0:180, -1:180, 0:-179, 1:-179, 1:-180"))

	// A shape gotten from A by adding a triangle to one edge, and
	// adding another triangle to the opposite edge.
	loopD = LoopFromPoints(parsePoints("0:178, -1:178, -1:180, 0:-179, 1:-179, 1:-180"))

	//   3------------2
	//   |            |               ^
	//   |  7-8  b-c  |               |
	//   |  | |  | |  |      Latitude |
	//   0--6-9--a-d--1               |
	//   |  | |       |               |
	//   |  f-e       |               +----------->
	//   |            |                 Longitude
	//   4------------5
	//
	// Important: It is not okay to skip over collinear vertices when
	// defining these loops (e.g. to define loop E as "0,1,2,3") because S2
	// uses symbolic perturbations to ensure that no three vertices are
	// *ever* considered collinear (e.g., vertices 0, 6, 9 are not
	// collinear).  In other words, it is unpredictable (modulo knowing the
	// details of the symbolic perturbations) whether 0123 contains 06123
	// for example.

	// Loop E:  0,6,9,a,d,1,2,3
	// Loop F:  0,4,5,1,d,a,9,6
	// Loop G:  0,6,7,8,9,a,b,c,d,1,2,3
	// Loop H:  0,6,f,e,9,a,b,c,d,1,2,3
	// Loop I:  7,6,f,e,9,8
	loopE = LoopFromPoints(parsePoints("0:30, 0:34, 0:36, 0:39, 0:41, 0:44, 30:44, 30:30"))
	loopF = LoopFromPoints(parsePoints("0:30, -30:30, -30:44, 0:44, 0:41, 0:39, 0:36, 0:34"))
	loopG = LoopFromPoints(parsePoints("0:30, 0:34, 10:34, 10:36, 0:36, 0:39, 10:39, 10:41, 0:41, 0:44, 30:44, 30:30"))
	loopH = LoopFromPoints(parsePoints("0:30, 0:34, -10:34, -10:36, 0:36, 0:39, 10:39, 10:41, 0:41, 0:44, 30:44, 30:30"))

	loopI = LoopFromPoints(parsePoints("10:34, 0:34, -10:34, -10:36, 0:36, 10:36"))
)

func TestEmptyFullLoops(t *testing.T) {
	emptyLoop := EmptyLoop()

	if !emptyLoop.IsEmpty() {
		t.Errorf("empty loop should be empty")
	}
	if emptyLoop.IsFull() {
		t.Errorf("empty loop should not be full")
	}
	if !emptyLoop.isEmptyOrFull() {
		t.Errorf("empty loop should pass IsEmptyOrFull")
	}

	fullLoop := FullLoop()

	if fullLoop.IsEmpty() {
		t.Errorf("full loop should not be empty")
	}
	if !fullLoop.IsFull() {
		t.Errorf("full loop should be full")
	}
	if !fullLoop.isEmptyOrFull() {
		t.Errorf("full loop should pass IsEmptyOrFull")
	}
}

func TestLoopRectBound(t *testing.T) {
	if !EmptyLoop().RectBound().IsEmpty() {
		t.Errorf("empty loop's RectBound should be empty")
	}
	if !FullLoop().RectBound().IsFull() {
		t.Errorf("full loop's RectBound should be full")
	}
	if !candyCane.RectBound().Lng.IsFull() {
		t.Errorf("candy cane loop's RectBound should have a full longitude range")
	}
	if got := candyCane.RectBound().Lat.Lo; got >= -0.349066 {
		t.Errorf("candy cane loop's RectBound should have a lower latitude (%v) under -0.349066 radians", got)
	}
	if got := candyCane.RectBound().Lat.Hi; got <= 0.174533 {
		t.Errorf("candy cane loop's RectBound should have an upper latitude (%v) over 0.174533 radians", got)
	}
	if !smallNECW.RectBound().IsFull() {
		t.Errorf("small northeast clockwise loop's RectBound should be full")
	}
	if got, want := arctic80.RectBound(), rectFromDegrees(80, -180, 90, 180); !rectsApproxEqual(got, want, rectErrorLat, rectErrorLng) {
		t.Errorf("arctic 80 loop's RectBound (%v) should be %v", got, want)
	}
	if got, want := antarctic80.RectBound(), rectFromDegrees(-90, -180, -80, 180); !rectsApproxEqual(got, want, rectErrorLat, rectErrorLng) {
		t.Errorf("antarctic 80 loop's RectBound (%v) should be %v", got, want)
	}
	if !southHemi.RectBound().Lng.IsFull() {
		t.Errorf("south hemi loop's RectBound should have a full longitude range")
	}
	got, want := southHemi.RectBound().Lat, r1.Interval{-math.Pi / 2, 0}
	if !got.ApproxEqual(want) {
		t.Errorf("south hemi loop's RectBound latitude interval (%v) should be %v", got, want)
	}

	// Create a loop that contains the complement of the arctic80 loop.
	arctic80Inv := invert(arctic80)
	// The highest latitude of each edge is attained at its midpoint.
	mid := Point{arctic80Inv.vertices[0].Vector.Add(arctic80Inv.vertices[1].Vector).Mul(.5)}
	if got, want := arctic80Inv.RectBound().Lat.Hi, float64(LatLngFromPoint(mid).Lat); math.Abs(got-want) > 10*dblEpsilon {
		t.Errorf("arctic 80 inverse loop's RectBound should have a latutude hi of %v, got %v", got, want)
	}
}

func TestLoopCapBound(t *testing.T) {
	if !EmptyLoop().CapBound().IsEmpty() {
		t.Errorf("empty loop's CapBound should be empty")
	}
	if !FullLoop().CapBound().IsFull() {
		t.Errorf("full loop's CapBound should be full")
	}
	if !smallNECW.CapBound().IsFull() {
		t.Errorf("small northeast clockwise loop's CapBound should be full")
	}
	if got, want := arctic80.CapBound(), rectFromDegrees(80, -180, 90, 180).CapBound(); !got.ApproxEqual(want) {
		t.Errorf("arctic 80 loop's CapBound (%v) should be %v", got, want)
	}
	if got, want := antarctic80.CapBound(), rectFromDegrees(-90, -180, -80, 180).CapBound(); !got.ApproxEqual(want) {
		t.Errorf("antarctic 80 loop's CapBound (%v) should be %v", got, want)
	}
}

func invert(l *Loop) *Loop {
	vertices := make([]Point, 0, len(l.vertices))
	for i := len(l.vertices) - 1; i >= 0; i-- {
		vertices = append(vertices, l.vertices[i])
	}
	return LoopFromPoints(vertices)
}

func TestOriginInside(t *testing.T) {
	if !northHemi.originInside {
		t.Errorf("north hemisphere polygon should include origin")
	}
	if !northHemi3.originInside {
		t.Errorf("north hemisphere 3 polygon should include origin")
	}
	if southHemi.originInside {
		t.Errorf("south hemisphere polygon should not include origin")
	}
	if westHemi.originInside {
		t.Errorf("west hemisphere polygon should not include origin")
	}
	if !eastHemi.originInside {
		t.Errorf("east hemisphere polygon should include origin")
	}
	if nearHemi.originInside {
		t.Errorf("near hemisphere polygon should not include origin")
	}
	if !farHemi.originInside {
		t.Errorf("far hemisphere polygon should include origin")
	}
	if candyCane.originInside {
		t.Errorf("candy cane polygon should not include origin")
	}
	if !smallNECW.originInside {
		t.Errorf("smallNECW polygon should include origin")
	}
	if !arctic80.originInside {
		t.Errorf("arctic 80 polygon should include origin")
	}
	if antarctic80.originInside {
		t.Errorf("antarctic 80 polygon should not include origin")
	}
	if loopA.originInside {
		t.Errorf("loop A polygon should not include origin")
	}
}

func TestLoopContainsPoint(t *testing.T) {
	north := PointFromCoords(0, 0, 1)
	south := PointFromCoords(0, 0, -1)

	if EmptyLoop().ContainsPoint(north) {
		t.Errorf("empty loop should not not have any points")
	}
	if !FullLoop().ContainsPoint(south) {
		t.Errorf("full loop should have full point vertex")
	}

	for _, tc := range []struct {
		name string
		l    *Loop
		in   Point
		out  Point
	}{
		{
			"north hemisphere",
			northHemi,
			PointFromCoords(0, 0, 1),
			PointFromCoords(0, 0, -1),
		},
		{
			"south hemisphere",
			southHemi,
			PointFromCoords(0, 0, -1),
			PointFromCoords(0, 0, 1),
		},
		{
			"west hemisphere",
			westHemi,
			PointFromCoords(0, -1, 0),
			PointFromCoords(0, 1, 0),
		},
		{
			"east hemisphere",
			eastHemi,
			PointFromCoords(0, 1, 0),
			PointFromCoords(0, -1, 0),
		},
		{
			"candy cane",
			candyCane,
			PointFromLatLng(LatLngFromDegrees(5, 71)),
			PointFromLatLng(LatLngFromDegrees(-8, 71)),
		},
	} {
		l := tc.l
		for i := 0; i < 4; i++ {
			if !l.ContainsPoint(tc.in) {
				t.Errorf("%s loop should contain %v at rotation %d", tc.name, tc.in, i)
			}
			if l.ContainsPoint(tc.out) {
				t.Errorf("%s loop shouldn't contain %v at rotation %d", tc.name, tc.out, i)
			}
			l = rotate(l)
		}
	}
}

func TestVertex(t *testing.T) {
	tests := []struct {
		loop   *Loop
		vertex int
		want   Point
	}{
		{EmptyLoop(), 0, PointFromCoords(0, 0, 1)},
		{EmptyLoop(), 1, PointFromCoords(0, 0, 1)},
		{FullLoop(), 0, PointFromCoords(0, 0, -1)},
		{FullLoop(), 1, PointFromCoords(0, 0, -1)},
		{arctic80, 0, parsePoint("80:-150")},
		{arctic80, 1, parsePoint("80:-30")},
		{arctic80, 2, parsePoint("80:90")},
		{arctic80, 3, parsePoint("80:-150")},
	}

	for _, test := range tests {
		if got := test.loop.Vertex(test.vertex); !pointsApproxEquals(got, test.want, epsilon) {
			t.Errorf("%v.Vertex(%d) = %v, want %v", test.loop, test.vertex, got, test.want)
		}
	}

	// Check that wrapping is correct.
	if !pointsApproxEquals(arctic80.Vertex(2), arctic80.Vertex(5), epsilon) {
		t.Errorf("Vertex should wrap values. %v.Vertex(2) = %v != %v.Vertex(5) = %v",
			arctic80, arctic80.Vertex(2), arctic80, arctic80.Vertex(5))
	}

	loopAroundThrice := 2 + 3*len(arctic80.vertices)
	if !pointsApproxEquals(arctic80.Vertex(2), arctic80.Vertex(loopAroundThrice), epsilon) {
		t.Errorf("Vertex should wrap values. %v.Vertex(2) = %v != %v.Vertex(%d) = %v",
			arctic80, arctic80.Vertex(2), arctic80, loopAroundThrice, arctic80.Vertex(loopAroundThrice))
	}
}

func TestNumEdges(t *testing.T) {
	tests := []struct {
		loop *Loop
		want int
	}{
		{EmptyLoop(), 0},
		{FullLoop(), 0},
		{farHemi, 4},
		{candyCane, 6},
		{smallNECW, 3},
		{arctic80, 3},
		{antarctic80, 3},
		{lineTriangle, 3},
		{skinnyChevron, 4},
	}

	for _, test := range tests {
		if got := test.loop.NumEdges(); got != test.want {
			t.Errorf("%v.NumEdges() = %v, want %v", test.loop, got, test.want)
		}
	}
}

func TestEdge(t *testing.T) {
	tests := []struct {
		loop  *Loop
		edge  int
		wantA Point
		wantB Point
	}{
		{
			loop:  farHemi,
			edge:  2,
			wantA: Point{r3.Vector{0, 0, -1}},
			wantB: Point{r3.Vector{0, -1, 0}},
		},
		{
			loop: candyCane,
			edge: 0,

			wantA: parsePoint("-20:150"),
			wantB: parsePoint("-20:-70"),
		},
		{
			loop:  candyCane,
			edge:  1,
			wantA: parsePoint("-20:-70"),
			wantB: parsePoint("0:70"),
		},
		{
			loop:  candyCane,
			edge:  2,
			wantA: parsePoint("0:70"),
			wantB: parsePoint("10:-150"),
		},
		{
			loop:  candyCane,
			edge:  3,
			wantA: parsePoint("10:-150"),
			wantB: parsePoint("10:70"),
		},
		{
			loop:  candyCane,
			edge:  4,
			wantA: parsePoint("10:70"),
			wantB: parsePoint("-10:-70"),
		},
		{
			loop:  candyCane,
			edge:  5,
			wantA: parsePoint("-10:-70"),
			wantB: parsePoint("-20:150"),
		},
		{
			loop:  skinnyChevron,
			edge:  2,
			wantA: parsePoint("0:1e-320"),
			wantB: parsePoint("1e-320:80"),
		},
		{
			loop:  skinnyChevron,
			edge:  3,
			wantA: parsePoint("1e-320:80"),
			wantB: parsePoint("0:0"),
		},
	}

	for _, test := range tests {
		if a, b := test.loop.Edge(test.edge); !(pointsApproxEquals(a, test.wantA, epsilon) && pointsApproxEquals(b, test.wantB, epsilon)) {
			t.Errorf("%v.Edge(%d) = (%v, %v), want (%v, %v)", test.loop, test.edge, a, b, test.wantA, test.wantB)
		}
	}
}

func rotate(l *Loop) *Loop {
	vertices := make([]Point, 0, len(l.vertices))
	for i := 1; i < len(l.vertices); i++ {
		vertices = append(vertices, l.vertices[i])
	}
	vertices = append(vertices, l.vertices[0])
	return LoopFromPoints(vertices)
}

func TestLoopFromCell(t *testing.T) {
	cell := CellFromCellID(CellIDFromLatLng(LatLng{40.565459 * s1.Degree, -74.645276 * s1.Degree}))
	loopFromCell := LoopFromCell(cell)

	// Demonstrates the reason for this test; the cell bounds are more
	// conservative than the resulting loop bounds.
	if loopFromCell.RectBound().Contains(cell.RectBound()) {
		t.Errorf("loopFromCell's RectBound countains the original cells RectBound, but should not")
	}
}
