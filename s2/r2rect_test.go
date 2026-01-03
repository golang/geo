package s2

import (
	"testing"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
)

func TestR2RectEmptyRectangles(t *testing.T) {
	empty := R2Rect{r2.EmptyRect()}

	if got := empty.IsValid(); !got {
		t.Errorf("empty.IsValid() = %v, want true", got)
	}
	if got := empty.IsEmpty(); !got {
		t.Errorf("empty.IsEmpty() = %v, want true", got)
	}
	if empty != empty {
		t.Errorf("empty != empty; want equal")
	}
}

func TestR2RectConstructorsAndAccessors(t *testing.T) {
	d1 := R2Rect{r2.RectFromPoints(r2.Point{X: 0.1, Y: 0}, r2.Point{X: 0.25, Y: 1})}
	if got, want := d1.X.Lo, 0.1; got != want {
		t.Errorf("X.Lo = %v, want %v", got, want)
	}
	if got, want := d1.X.Hi, 0.25; got != want {
		t.Errorf("X.Hi = %v, want %v", got, want)
	}
	if got, want := d1.Y.Lo, 0.0; got != want {
		t.Errorf("Y.Lo = %v, want %v", got, want)
	}
	if got, want := d1.Y.Hi, 1.0; got != want {
		t.Errorf("Y.Hi = %v, want %v", got, want)
	}
	if got, want := d1.X, (r1.Interval{Lo: 0.1, Hi: 0.25}); got != want {
		t.Errorf("X interval = %v, want %v", got, want)
	}
	if got, want := d1.Y, (r1.Interval{Lo: 0, Hi: 1}); got != want {
		t.Errorf("Y interval = %v, want %v", got, want)
	}

	if got, want := d1.VertexIJ(0, 0), d1.Lo(); got != want {
		t.Errorf("VertexIJ(0, 0) = %v, want %v", got, want)
	}
	if got, want := d1.VertexIJ(1, 1), d1.Hi(); got != want {
		t.Errorf("VertexIJ(1, 1) = %v, want %v", got, want)
	}
	if d1 != d1 {
		t.Errorf("d1 != d1; want equal")
	}
	if d1 == (R2Rect{r2.EmptyRect()}) {
		t.Errorf("d1 == EmptyRect(); want non-empty")
	}
}

func TestR2RectFromCell(t *testing.T) {
	want := R2Rect{r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 0.5, Y: 0.5})}
	if got := R2RectFromCell(CellFromCellID(CellIDFromFacePosLevel(0, 0, 1))); got != want {
		t.Errorf("R2RectFromCell(face=0,pos=0,level=1) = %v, want %v", got, want)
	}

	want = R2Rect{r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 1, Y: 1})}
	if got := R2RectFromCellID(CellIDFromFacePosLevel(0, 0, 0)); got != want {
		t.Errorf("R2RectFromCellID(face=0,pos=0,level=0) = %v, want %v", got, want)
	}
}

func TestR2RectFromCenterSize(t *testing.T) {
	center := r2.Point{X: 0.3, Y: 0.5}
	size := r2.Point{X: 0.2, Y: 0.4}
	got := R2Rect{r2.RectFromCenterSize(center, size)}
	want := R2Rect{r2.RectFromPoints(r2.Point{X: 0.2, Y: 0.3}, r2.Point{X: 0.4, Y: 0.7})}
	if !got.ApproxEqual(want) {
		t.Errorf("RectFromCenterSize(%v, %v) = %v, want approx %v", center, size, got, want)
	}

	center = r2.Point{X: 1, Y: 0.1}
	size = r2.Point{X: 0, Y: 2}
	got = R2Rect{r2.RectFromCenterSize(center, size)}
	want = R2Rect{r2.RectFromPoints(r2.Point{X: 1, Y: -0.9}, r2.Point{X: 1, Y: 1.1})}
	if !got.ApproxEqual(want) {
		t.Errorf("RectFromCenterSize(%v, %v) = %v, want approx %v", center, size, got, want)
	}
}

func TestR2RectFromPoint(t *testing.T) {
	d1 := R2Rect{r2.RectFromPoints(r2.Point{X: 0.1, Y: 0}, r2.Point{X: 0.25, Y: 1})}
	got := R2Rect{r2.RectFromPoints(d1.Lo(), d1.Lo())}
	want := R2Rect{r2.RectFromPoints(d1.Lo())}
	if got != want {
		t.Errorf("RectFromPoints(%v, %v) = %v, want %v", d1.Lo(), d1.Lo(), got, want)
	}

	got = R2Rect{r2.RectFromPoints(r2.Point{X: 0.15, Y: 0.3}, r2.Point{X: 0.35, Y: 0.9})}
	want = R2Rect{r2.RectFromPoints(r2.Point{X: 0.15, Y: 0.9}, r2.Point{X: 0.35, Y: 0.3})}
	if got != want {
		t.Errorf("RectFromPoints with swapped Y endpoints = %v, want %v", got, want)
	}

	got = R2Rect{r2.RectFromPoints(r2.Point{X: 0.12, Y: 0}, r2.Point{X: 0.83, Y: 0.5})}
	want = R2Rect{r2.RectFromPoints(r2.Point{X: 0.83, Y: 0}, r2.Point{X: 0.12, Y: 0.5})}
	if got != want {
		t.Errorf("RectFromPoints with swapped X endpoints = %v, want %v", got, want)
	}
}

func TestR2RectSimplePredicates(t *testing.T) {
	sw1 := r2.Point{X: 0, Y: 0.25}
	ne1 := r2.Point{X: 0.5, Y: 0.75}
	rect1 := R2Rect{r2.RectFromPoints(sw1, ne1)}

	if got, want := rect1.Center(), (r2.Point{X: 0.25, Y: 0.5}); got != want {
		t.Errorf("Center() = %v, want %v", got, want)
	}
	if got, want := rect1.Vertex(0), (r2.Point{X: 0, Y: 0.25}); got != want {
		t.Errorf("Vertex(0) = %v, want %v", got, want)
	}
	if got, want := rect1.Vertex(1), (r2.Point{X: 0.5, Y: 0.25}); got != want {
		t.Errorf("Vertex(1) = %v, want %v", got, want)
	}
	if got, want := rect1.Vertex(2), (r2.Point{X: 0.5, Y: 0.75}); got != want {
		t.Errorf("Vertex(2) = %v, want %v", got, want)
	}
	if got, want := rect1.Vertex(3), (r2.Point{X: 0, Y: 0.75}); got != want {
		t.Errorf("Vertex(3) = %v, want %v", got, want)
	}

	if pt := (r2.Point{X: 0.2, Y: 0.4}); !rect1.ContainsPoint(pt) {
		t.Errorf("ContainsPoint(%v) = false, want true", pt)
	}

	if pt := (r2.Point{X: 0.2, Y: 0.8}); rect1.ContainsPoint(pt) {
		t.Errorf("ContainsPoint(%v) = true, want false", pt)
	}
	if pt := (r2.Point{X: -0.1, Y: 0.4}); rect1.ContainsPoint(pt) {
		t.Errorf("ContainsPoint(%v) = true, want false", pt)
	}
	if pt := (r2.Point{X: 0.6, Y: 0.1}); rect1.ContainsPoint(pt) {
		t.Errorf("ContainsPoint(%v) = true, want false", pt)
	}

	if pt := sw1; !rect1.ContainsPoint(pt) {
		t.Errorf("ContainsPoint(%v) = false, want true (edge inclusive)", pt)
	}
	if pt := sw1; rect1.InteriorContainsPoint(pt) {
		t.Errorf("InteriorContainsPoint(%v) = true, want false", pt)
	}
	if pt := ne1; !rect1.ContainsPoint(pt) {
		t.Errorf("ContainsPoint(%v) = false, want true (edge inclusive)", pt)
	}
	if pt := ne1; rect1.InteriorContainsPoint(pt) {
		t.Errorf("InteriorContainsPoint(%v) = true, want false", pt)
	}

	for k := 0; k < 4; k++ {
		if !Sign(toPoint(rect1.Vertex(k-1)), toPoint(rect1.Vertex(k)), toPoint(rect1.Vertex(k+1))) {
			t.Errorf("%v.Vertex(%v), vertices were not in CCW order", rect1, k)
		}
	}
}

func TestR2RectIntervalOps(t *testing.T) {
	empty := R2Rect{r2.EmptyRect()}
	sw1 := r2.Point{X: 0, Y: 0.25}
	ne1 := r2.Point{X: 0.5, Y: 0.75}
	rect1 := R2Rect{r2.RectFromPoints(sw1, ne1)}
	rect1Mid := R2Rect{r2.RectFromPoints(r2.Point{X: 0.25, Y: 0.5}, r2.Point{X: 0.25, Y: 0.5})}
	rSW1 := R2Rect{r2.RectFromPoints(sw1, sw1)}
	rNE1 := R2Rect{r2.RectFromPoints(ne1, ne1)}

	tests := []struct {
		rect         R2Rect
		other        R2Rect
		contains     bool
		intersects   bool
		union        R2Rect
		intersection R2Rect
	}{
		{
			rect:         rect1,
			other:        rect1Mid,
			contains:     true,
			intersects:   true,
			union:        rect1,
			intersection: rect1Mid,
		},
		{
			rect:         rect1,
			other:        rSW1,
			contains:     true,
			intersects:   true,
			union:        rect1,
			intersection: rSW1,
		},
		{
			rect:         rect1,
			other:        rNE1,
			contains:     true,
			intersects:   true,
			union:        rect1,
			intersection: rNE1,
		},
		{
			rect:         rect1,
			other:        R2Rect{r2.RectFromPoints(r2.Point{X: 0.45, Y: 0.1}, r2.Point{X: 0.75, Y: 0.3})},
			contains:     false,
			intersects:   true,
			union:        R2Rect{r2.RectFromPoints(r2.Point{X: 0, Y: 0.1}, r2.Point{X: 0.75, Y: 0.75})},
			intersection: R2Rect{r2.RectFromPoints(r2.Point{X: 0.45, Y: 0.25}, r2.Point{X: 0.5, Y: 0.3})},
		},
		{
			rect:         rect1,
			other:        R2Rect{r2.RectFromPoints(r2.Point{X: 0.5, Y: 0.1}, r2.Point{X: 0.7, Y: 0.3})},
			contains:     false,
			intersects:   true,
			union:        R2Rect{r2.RectFromPoints(r2.Point{X: 0, Y: 0.1}, r2.Point{X: 0.7, Y: 0.75})},
			intersection: R2Rect{r2.RectFromPoints(r2.Point{X: 0.5, Y: 0.25}, r2.Point{X: 0.5, Y: 0.3})},
		},
		{
			rect:         rect1,
			other:        R2Rect{r2.RectFromPoints(r2.Point{X: 0.45, Y: 0.1}, r2.Point{X: 0.7, Y: 0.25})},
			contains:     false,
			intersects:   true,
			union:        R2Rect{r2.RectFromPoints(r2.Point{X: 0, Y: 0.1}, r2.Point{X: 0.7, Y: 0.75})},
			intersection: R2Rect{r2.RectFromPoints(r2.Point{X: 0.45, Y: 0.25}, r2.Point{X: 0.5, Y: 0.25})},
		},
		{
			rect:         R2Rect{r2.RectFromPoints(r2.Point{X: 0.1, Y: 0.2}, r2.Point{X: 0.1, Y: 0.3})},
			other:        R2Rect{r2.RectFromPoints(r2.Point{X: 0.15, Y: 0.7}, r2.Point{X: 0.2, Y: 0.8})},
			contains:     false,
			intersects:   false,
			union:        R2Rect{r2.RectFromPoints(r2.Point{X: 0.1, Y: 0.2}, r2.Point{X: 0.2, Y: 0.8})},
			intersection: empty,
		},
		{
			// Overlap in x but not y.
			rect:         R2Rect{r2.RectFromPoints(r2.Point{X: 0.1, Y: 0.2}, r2.Point{X: 0.4, Y: 0.5})},
			other:        R2Rect{r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 0.2, Y: 0.1})},
			contains:     false,
			intersects:   false,
			union:        R2Rect{r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 0.4, Y: 0.5})},
			intersection: empty,
		},
		{
			// Overlap in y but not x.
			rect:         R2Rect{r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 0.1, Y: 0.3})},
			other:        R2Rect{r2.RectFromPoints(r2.Point{X: 0.2, Y: 0.1}, r2.Point{X: 0.3, Y: 0.4})},
			contains:     false,
			intersects:   false,
			union:        R2Rect{r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 0.3, Y: 0.4})},
			intersection: empty,
		},
	}

	for _, test := range tests {
		if got := test.rect.Contains(test.other); got != test.contains {
			t.Errorf("%v.Contains(%v) = %t, want %t", test.rect, test.other, got, test.contains)
		}

		if got := test.rect.Intersects(test.other); got != test.intersects {
			t.Errorf("%v.Intersects(%v) = %t, want %t", test.rect, test.other, got, test.intersects)
		}

		if got := test.rect.Union(test.other) == test.rect; test.rect.Contains(test.other) != got {
			t.Errorf("%v.Union(%v) == %v = %t, want %t",
				test.rect, test.other, test.other, got, test.rect.Contains(test.other),
			)
		}

		if got := test.rect.Intersection(test.other).IsEmpty(); test.rect.Intersects(test.other) == got {
			t.Errorf("%v.Intersection(%v).IsEmpty() = %t, want %t",
				test.rect, test.other, got, test.rect.Intersects(test.other))
		}

		if got := test.rect.Union(test.other); got != test.union {
			t.Errorf("%v.Union(%v) = %v, want %v", test.rect, test.other, got, test.union)
		}

		if got := test.rect.Intersection(test.other); got != test.intersection {
			t.Errorf("%v.Intersection(%v) = %v, want %v", test.rect, test.other, got, test.intersection)
		}
	}
}

func TestR2RectAddPoint(t *testing.T) {
	// adapt below code to golang
	sw1 := r2.Point{X: 0, Y: 0.25}
	ne1 := r2.Point{X: 0.5, Y: 0.75}
	rect1 := R2Rect{r2.RectFromPoints(sw1, ne1)}

	rect2 := R2Rect{r2.EmptyRect()}
	rect2 = rect2.AddPoint(r2.Point{X: 0, Y: 0.25})
	rect2 = rect2.AddPoint(r2.Point{X: 0.5, Y: 0.25})
	rect2 = rect2.AddPoint(r2.Point{X: 0, Y: 0.75})
	rect2 = rect2.AddPoint(r2.Point{X: 0.1, Y: 0.4})
	if rect1 != rect2 {
		t.Errorf("rect2 = %v, want %v", rect2, rect1)
	}
}

func TestR2RectProject(t *testing.T) {
	rect1 := R2Rect{r2.Rect{X: r1.Interval{Lo: 0, Hi: 0.5}, Y: r1.Interval{Lo: 0.25, Hi: 0.75}}}

	// transform the code below into go and use t.Errorf with messages in line with the other tests in the module
	if got, want := rect1.Project(r2.Point{X: -0.01, Y: 0.24}), (r2.Point{X: 0, Y: 0.25}); got != want {
		t.Errorf("Project(%v) = %v, want %v", r2.Point{X: -0.01, Y: 0.24}, got, want)
	}
	if got, want := rect1.Project(r2.Point{X: -5.0, Y: 0.48}), (r2.Point{X: 0, Y: 0.48}); got != want {
		t.Errorf("Project(%v) = %v, want %v", r2.Point{X: -5.0, Y: 0.48}, got, want)
	}
	if got, want := rect1.Project(r2.Point{X: -5.0, Y: 2.48}), (r2.Point{X: 0, Y: 0.75}); got != want {
		t.Errorf("Project(%v) = %v, want %v", r2.Point{X: -5.0, Y: 2.48}, got, want)
	}
	if got, want := rect1.Project(r2.Point{X: 0.19, Y: 2.48}), (r2.Point{X: 0.19, Y: 0.75}); got != want {
		t.Errorf("Project(%v) = %v, want %v", r2.Point{X: 0.19, Y: 2.48}, got, want)
	}
	if got, want := rect1.Project(r2.Point{X: 6.19, Y: 2.48}), (r2.Point{X: 0.5, Y: 0.75}); got != want {
		t.Errorf("Project(%v) = %v, want %v", r2.Point{X: 6.19, Y: 2.48}, got, want)
	}
	if got, want := rect1.Project(r2.Point{X: 6.19, Y: 0.53}), (r2.Point{X: 0.5, Y: 0.53}); got != want {
		t.Errorf("Project(%v) = %v, want %v", r2.Point{X: 6.19, Y: 0.53}, got, want)
	}
	if got, want := rect1.Project(r2.Point{X: 6.19, Y: -2.53}), (r2.Point{X: 0.5, Y: 0.25}); got != want {
		t.Errorf("Project(%v) = %v, want %v", r2.Point{X: 6.19, Y: -2.53}, got, want)
	}
	if got, want := rect1.Project(r2.Point{X: 0.33, Y: -2.53}), (r2.Point{X: 0.33, Y: 0.25}); got != want {
		t.Errorf("Project(%v) = %v, want %v", r2.Point{X: 0.33, Y: -2.53}, got, want)
	}
	if got, want := rect1.Project(r2.Point{X: 0.33, Y: 0.37}), (r2.Point{X: 0.33, Y: 0.37}); got != want {
		t.Errorf("Project(%v) = %v, want %v", r2.Point{X: 0.33, Y: 0.37}, got, want)
	}
}

func TestR2RectExpanded(t *testing.T) {
	empty := R2Rect{r2.EmptyRect()}
	rect := R2Rect{r2.RectFromPoints(r2.Point{X: 0.2, Y: 0.4}, r2.Point{X: 0.3, Y: 0.7})}

	if !empty.Expanded(r2.Point{X: 0.1, Y: 0.3}).IsEmpty() {
		t.Errorf("%v.Expanded(%v).IsEmpty() = false, want true", empty, r2.Point{X: 0.1, Y: 0.3})
	}
	if !empty.Expanded(r2.Point{X: -0.1, Y: -0.3}).IsEmpty() {
		t.Errorf("%v.Expanded(%v).IsEmpty() = false, want true", empty, r2.Point{X: -0.1, Y: -0.3})
	}

	if !rect.Expanded(r2.Point{X: 0.1, Y: 0.3}).ApproxEqual(R2Rect{r2.RectFromPoints(r2.Point{X: 0.1, Y: 0.1}, r2.Point{X: 0.4, Y: 1.0})}) {
		t.Errorf("%v.Expanded(%v).ApproxEqual(%v) = false, want true", rect, r2.Point{X: 0.1, Y: 0.3}, R2Rect{r2.RectFromPoints(r2.Point{X: 0.1, Y: 0.1}, r2.Point{X: 0.4, Y: 1.0})})
	}

	if !rect.Expanded(r2.Point{X: -0.1, Y: 0.3}).IsEmpty() {
		t.Errorf("%v.Expanded(%v).IsEmpty() = false, want true", rect, r2.Point{X: -0.1, Y: 0.3})
	}
	if !rect.Expanded(r2.Point{X: 0.1, Y: -0.2}).IsEmpty() {
		t.Errorf("%v.Expanded(%v).IsEmpty() = false, want true", rect, r2.Point{X: 0.1, Y: -0.2})
	}

	if !rect.Expanded(r2.Point{X: 0.1, Y: -0.1}).ApproxEqual(R2Rect{r2.RectFromPoints(r2.Point{X: 0.1, Y: 0.5}, r2.Point{X: 0.4, Y: 0.6})}) {
		t.Errorf("%v.Expanded(%v).ApproxEqual(%v) = false, want true", rect, r2.Point{X: 0.1, Y: -0.1}, R2Rect{r2.RectFromPoints(r2.Point{X: 0.1, Y: 0.5}, r2.Point{X: 0.4, Y: 0.6})})
	}

	if !rect.ExpandedByMargin(0.1).ApproxEqual(R2Rect{r2.RectFromPoints(r2.Point{X: 0.1, Y: 0.3}, r2.Point{X: 0.4, Y: 0.8})}) {
		t.Errorf("%v.ExpandedByMargin(%v).ApproxEqual(%v) = false, want true", rect, 0.1, R2Rect{r2.RectFromPoints(r2.Point{X: 0.1, Y: 0.3}, r2.Point{X: 0.4, Y: 0.8})})
	}
}

func TestR2RectBounds(t *testing.T) {
	empty := R2Rect{r2.EmptyRect()}
	if !empty.CapBound().IsEmpty() {
		t.Errorf("%v.CapBound().IsEmpty() = false, want true", empty)
	}
	if !empty.RectBound().IsEmpty() {
		t.Errorf("%v.RectBound().IsEmpty() = false, want true", empty)
	}
	if CapFromPoint(PointFromCoords(1, 0, 0)) != (R2Rect{r2.RectFromPoints(r2.Point{X: 0.5, Y: 0.5}, r2.Point{X: 0.5, Y: 0.5})}).CapBound() {
		t.Errorf("CapFromPoint(PointFromCoords(1,0,0))== R2Rect{r2.RectFromPoints(r2.Point{X: 0.5, Y: 0.5}, r2.Point{X: 0.5, Y: 0.5})}.CapBound() = false, want true")
	}
	if RectFromLatLng(LatLngFromDegrees(0, 0)) != (R2Rect{r2.RectFromPoints(r2.Point{X: 0.5, Y: 0.5}, r2.Point{X: 0.5, Y: 0.5})}).RectBound() {
		t.Errorf("RectFromLatLng(LatLngFromDegrees(0, 0))== R2Rect{r2.RectFromPoints(r2.Point{X: 0.5, Y: 0.5}, r2.Point{X: 0.5, Y: 0.5})}.RectBound() = false, want true")
	}

	for i := 0; i < 10; i++ {
		rect := R2RectFromCellID(randomCellID())
		cap := rect.CapBound()
		latLngRect := rect.RectBound()

		for k := 0; k < 4; k++ {
			v := toPoint(rect.Vertex(k))
			v2 := Point{cap.Center().Vector.Add(v.Vector.Sub(cap.Center().Vector).Mul(3)).Normalize().Normalize()}
			if !cap.ContainsPoint(v) {
				t.Errorf("%v should contain %v", cap, v)
			}
			if cap.ContainsPoint(v2) {
				t.Errorf("%v should not contain %v", cap, v2)
			}
			if !latLngRect.ContainsPoint(v) {
				t.Errorf("%v should contain %v", latLngRect, v)
			}
			if latLngRect.ContainsPoint(v2) {
				t.Errorf("%v should not contain %v", latLngRect, v2)
			}
		}
	}
}

func TestR2RectCellOperations(t *testing.T) {
	empty := R2Rect{r2.EmptyRect()}
	// This rectangle includes the first quadrant of face 0.  It's expanded
	// slightly because cell bounding rectangles are slightly conservative.
	rect4 := R2Rect{r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 0.5, Y: 0.5})}
	// This rectangle intersects the first quadrant of face 0.
	rect5 := R2Rect{r2.RectFromPoints(r2.Point{X: 0, Y: 0.45}, r2.Point{X: 0.5, Y: 0.55})}

	tests := []struct {
		rect            R2Rect
		cell            Cell
		intersects      bool
		vertexContained bool
		contains        bool
	}{
		{
			rect:            empty,
			cell:            CellFromCellID(CellIDFromFace(3)),
			intersects:      false,
			vertexContained: false,
			contains:        false,
		},
		{
			rect:            rect4,
			cell:            CellFromCellID(CellIDFromFacePosLevel(0, 0, 0)),
			intersects:      true,
			vertexContained: true,
			contains:        false,
		},
		{
			rect:            rect4,
			cell:            CellFromCellID(CellIDFromFacePosLevel(0, 0, 1)),
			intersects:      true,
			vertexContained: true,
			contains:        true,
		},
		{
			rect:            rect4,
			cell:            CellFromCellID(CellIDFromFacePosLevel(1, 0, 1)),
			intersects:      false,
			vertexContained: false,
			contains:        false,
		},
		{
			rect:            rect5,
			cell:            CellFromCellID(CellIDFromFacePosLevel(0, 0, 0)),
			intersects:      true,
			vertexContained: true,
			contains:        false,
		},
		{
			rect:            rect5,
			cell:            CellFromCellID(CellIDFromFacePosLevel(0, 0, 1)),
			intersects:      true,
			vertexContained: true,
			contains:        false,
		},
		{
			rect:            rect5,
			cell:            CellFromCellID(CellIDFromFacePosLevel(1, 0, 1)),
			intersects:      false,
			vertexContained: false,
			contains:        false,
		},
		// Rectangle consisting of a single point.
		{
			rect:            R2Rect{r2.RectFromPoints(r2.Point{X: 0.51, Y: 0.51}, r2.Point{X: 0.51, Y: 0.51})},
			cell:            CellFromCellID(CellIDFromFace(0)),
			intersects:      true,
			vertexContained: true,
			contains:        false,
		},
		// Rectangle that intersects the bounding rectangle of face 0
		// but not the face itself.
		{
			rect:            R2Rect{r2.RectFromPoints(r2.Point{X: 0.01, Y: 1.001}, r2.Point{X: 0.02, Y: 1.002})},
			cell:            CellFromCellID(CellIDFromFace(0)),
			intersects:      false,
			vertexContained: false,
			contains:        false,
		},
		// Rectangle that intersects one corner of face 0.
		{
			rect:            R2Rect{r2.RectFromPoints(r2.Point{X: 0.99, Y: -0.01}, r2.Point{X: 1.01, Y: 0.01})},
			cell:            CellFromCellID(CellIDFromFacePosLevel(0, ^uint64(0)>>FaceBits, 5)),
			intersects:      true,
			vertexContained: true,
			contains:        false,
		},
	}

	for _, test := range tests {
		var vertexContained bool

		for k := 0; k < 4; k++ {
			// This would be easier to do by constructing an S2R2Rect from the cell,
			// but that would defeat the purpose of testing this code independently.
			if u, v, ok := faceXYZToUV(0, test.cell.VertexRaw(k)); ok {
				if test.rect.ContainsPoint(r2.Point{X: uvToST(u), Y: uvToST(v)}) {
					vertexContained = true
				}
			}
			if !test.rect.IsEmpty() && test.cell.ContainsPoint(toPoint(test.rect.Vertex(k))) {
				vertexContained = true
			}
		}

		if got := test.rect.MayIntersect(test.cell); got != test.intersects {
			t.Errorf("%v.MayIntersect(%v) = %t, want %t", test.rect, test.cell, got, test.intersects)
		}
		if vertexContained != test.vertexContained {
			t.Errorf("vertexContained = %t, want %t for rect %v and cell %v", vertexContained, test.contains, test.rect, test.cell)
		}
		if got := test.rect.ContainsCell(test.cell); got != test.contains {
			t.Errorf("%v.ContainsCell(%v) = %t, want %t", test.rect, test.cell, got, test.contains)
		}
	}
}
