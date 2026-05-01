package s2

import "github.com/golang/geo/r2"

// R2Rect represents a closed axis-aligned rectangle in the (x,y) plane.
type R2Rect struct {
	r2.Rect
}

// ContainsR2Point reports whether the rectangle contains the given point.
// Rectangles are closed regions, i.e. they contain their boundary.
func (r R2Rect) ContainsR2Point(p r2.Point) bool {
	return r.Rect.ContainsPoint(p)
}

// Contains reports whether the rectangle contains the given rectangle.
func (r R2Rect) Contains(other R2Rect) bool {
	return r.Rect.Contains(other.Rect)
}

// InteriorContainsPoint returns true iff the given point is contained in the interior
// of the region (i.e. the region excluding its boundary).
func (r R2Rect) InteriorContainsPoint(p r2.Point) bool {
	return r.Rect.InteriorContainsPoint(p)
}

// InteriorContains reports whether the interior of this rectangle contains all of the
// points of the given other rectangle (including its boundary).
func (r R2Rect) InteriorContains(other R2Rect) bool {
	return r.Rect.InteriorContains(other.Rect)
}

// Intersects reports whether this rectangle and the other rectangle have any points in common.
func (r R2Rect) Intersects(other R2Rect) bool {
	return r.Rect.Intersects(other.Rect)
}

// InteriorIntersects reports whether the interior of this rectangle intersects
// any point (including the boundary) of the given other rectangle.
func (r R2Rect) InteriorIntersects(other R2Rect) bool {
	return r.Rect.InteriorIntersects(other.Rect)
}

// Expanded returns a rectangle that has been expanded in the x-direction
// by margin.X, and in y-direction by margin.Y. If either margin is empty,
// then shrink the interval on the corresponding sides instead. The resulting
// rectangle may be empty. Any expansion of an empty rectangle remains empty.
func (r R2Rect) Expanded(margin r2.Point) R2Rect {
	return R2Rect{r.Rect.Expanded(margin)}
}

// ExpandedByMargin returns a Rect that has been expanded by the amount on all sides.
func (r R2Rect) ExpandedByMargin(margin float64) R2Rect {
	return R2Rect{r.Rect.ExpandedByMargin(margin)}
}

// Union returns the smallest rectangle containing the union of this rectangle and
// the given rectangle.
func (r R2Rect) Union(other R2Rect) R2Rect {
	return R2Rect{r.Rect.Union(other.Rect)}
}

// Intersection returns the smallest rectangle containing the intersection of this
// rectangle and the given rectangle.
func (r R2Rect) Intersection(other R2Rect) R2Rect {
	return R2Rect{r.Rect.Intersection(other.Rect)}
}

// ApproxEqual returns true if the x- and y-intervals of the two rectangles are
// the same up to the given tolerance.
func (r R2Rect) ApproxEqual(other R2Rect) bool {
	return r.Rect.ApproxEqual(other.Rect)
}

// AddPoint expands the rectangle to include the given point. The rectangle is
// expanded by the minimum amount possible.
func (r R2Rect) AddPoint(p r2.Point) R2Rect {
	return R2Rect{r.Rect.AddPoint(p)}
}

// AddRect expands the rectangle to include the given rectangle. This is the
// same as replacing the rectangle by the union of the two rectangles, but
// is more efficient.
func (r R2Rect) AddRect(other R2Rect) R2Rect {
	return R2Rect{r.Rect.AddRect(other.Rect)}
}

// R2RectFromCell constructs a rectangle that corresponds to the boundary of the given cell
// in (s,t)-space.  Such rectangles are always a subset of [0,1]x[0,1].
func R2RectFromCell(cell Cell) R2Rect {
	size := cell.SizeST()
	return R2Rect{r2.RectFromCenterSize(cell.id.centerST(), r2.Point{X: size, Y: size})}
}

// R2RectFromCellID constructs a rectangle that corresponds to the boundary of the given cell ID
// in (s,t)-space.  Such rectangles are always a subset of [0,1]x[0,1].
func R2RectFromCellID(id CellID) R2Rect {
	size := id.sizeST(id.Level())
	return R2Rect{r2.RectFromCenterSize(id.centerST(), r2.Point{X: size, Y: size})}
}

func toPoint(p r2.Point) Point {
	return Point{faceUVToXYZ(0, stToUV(p.X), stToUV(p.Y)).Normalize()}
}

// CapBound returns a Cap that bounds this rectangle.
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

// RectBound returns a bounding latitude-longitude rectangle.
// The bounds are not guaranteed to be tight.
func (r R2Rect) RectBound() Rect {
	return r.CapBound().RectBound()
}

// CellUnionBound computes a covering of the rectangle. In general the covering
// consists of at most 4 cells except for very large rectangles, which may need
// up to 6 cells. The output is not sorted.
func (r R2Rect) CellUnionBound() []CellID {
	return r.CapBound().CellUnionBound()
}

// ContainsPoint reports whether the rectangle contains the given point.
// Rectangles are closed regions, i.e. they contain their boundary.
func (r R2Rect) ContainsPoint(p Point) bool {
	if face(p.Vector) != 0 {
		return false
	}

	u, v := validFaceXYZToUV(0, p.Vector)
	return r.Rect.ContainsPoint(r2.Point{X: u, Y: v})
}

// ContainsCell reports whether the rectangle contains the given cell.
func (r R2Rect) ContainsCell(cell Cell) bool {
	if cell.Face() != 0 {
		return false
	}

	return r.Rect.Contains(R2RectFromCell(cell).Rect)
}

// MayIntersect reports whether the rectangle may intersect the given cell.
func (r R2Rect) MayIntersect(cell Cell) bool {
	if cell.Face() != 0 {
		return false
	}

	return r.Rect.Intersects(R2RectFromCell(cell).Rect)
}
