package s2

import "github.com/golang/geo/r2"

type R2Rect struct {
	r2.Rect
}

func (r R2Rect) ContainsR2Point(p r2.Point) bool {
	return r.Rect.ContainsPoint(p)
}

func (r R2Rect) Contains(other R2Rect) bool {
	return r.Rect.Contains(other.Rect)
}

func (r R2Rect) InteriorContainsPoint(p r2.Point) bool {
	return r.Rect.InteriorContainsPoint(p)
}

func (r R2Rect) InteriorContains(other R2Rect) bool {
	return r.Rect.InteriorContains(other.Rect)
}

func (r R2Rect) Intersects(other R2Rect) bool {
	return r.Rect.Intersects(other.Rect)
}

func (r R2Rect) InteriorIntersects(other R2Rect) bool {
	return r.Rect.InteriorIntersects(other.Rect)
}

func (r R2Rect) Expanded(margin r2.Point) R2Rect {
	return R2Rect{r.Rect.Expanded(margin)}
}

func (r R2Rect) ExpandedByMargin(margin float64) R2Rect {
	return R2Rect{r.Rect.ExpandedByMargin(margin)}
}

func (r R2Rect) Union(other R2Rect) R2Rect {
	return R2Rect{r.Rect.Union(other.Rect)}
}

func (r R2Rect) Intersection(other R2Rect) R2Rect {
	return R2Rect{r.Rect.Intersection(other.Rect)}
}

func (r R2Rect) ApproxEqual(other R2Rect) bool {
	return r.Rect.ApproxEqual(other.Rect)
}

func (r R2Rect) AddPoint(p r2.Point) R2Rect {
	return R2Rect{r.Rect.AddPoint(p)}
}

func (r R2Rect) AddRect(other R2Rect) R2Rect {
	return R2Rect{r.Rect.AddRect(other.Rect)}
}

func R2RectFromCell(cell Cell) R2Rect {
	size := cell.SizeST()
	return R2Rect{r2.RectFromCenterSize(cell.id.centerST(), r2.Point{X: size, Y: size})}
}

func R2RectFromCellID(id CellID) R2Rect {
	size := id.sizeST(id.Level())
	return R2Rect{r2.RectFromCenterSize(id.centerST(), r2.Point{X: size, Y: size})}
}

func toPoint(p r2.Point) Point {
	return Point{faceUVToXYZ(0, stToUV(p.X), stToUV(p.Y)).Normalize()}
}

func (r R2Rect) CapBound() Cap {
	if r.IsEmpty() {
		return EmptyCap()
	}

	// The rectangle is a convex polygon on the sphere, since it is a subset of
	// one cube face.  Its bounding cap is also a convex region on the sphere,
	// and therefore we can bound the rectangle by just bounding its vertices.
	// We use the rectangle's center in (s,t)-space as the cap axis.  This
	// doesn't yield the minimal cap but it's pretty close.
	cap := CapFromPoint(toPoint(r.Center()))
	for k := 0; k < 4; k++ {
		cap = cap.AddPoint(toPoint(r.Vertex(k)))
	}

	return cap
}

func (r R2Rect) RectBound() Rect {
	return r.CapBound().RectBound()
}

func (r R2Rect) CellUnionBound() []CellID {
	return r.CapBound().CellUnionBound()
}

func (r R2Rect) ContainsPoint(p Point) bool {
	if face(p.Vector) != 0 {
		return false
	}

	u, v := validFaceXYZToUV(0, p.Vector)
	return r.Rect.ContainsPoint(r2.Point{X: u, Y: v})
}

func (r R2Rect) ContainsCell(cell Cell) bool {
	if cell.Face() != 0 {
		return false
	}

	return r.Rect.Contains(R2RectFromCell(cell).Rect)
}

func (r R2Rect) MayIntersect(cell Cell) bool {
	if cell.Face() != 0 {
		return false
	}

	return r.Rect.Intersects(R2RectFromCell(cell).Rect)
}
