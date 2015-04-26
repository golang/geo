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
	"github.com/golang/geo/r2"
)

// Cell is an S2 region object that represents a cell. Unlike CellIDs,
// it supports efficient containment and intersection tests. However, it is
// also a more expensive representation.
type Cell struct {
	face        int8
	level       int8
	orientation int8
	id          CellID
	uv          r2.Rect
}

// CellFromCellID constructs a Cell corresponding to the given CellID.
func CellFromCellID(id CellID) Cell {
	c := Cell{}
	c.id = id
	f, i, j, o := c.id.faceIJOrientation()
	c.face = int8(f)
	c.level = int8(c.id.Level())
	c.orientation = int8(o)
	c.uv = ijLevelToBoundUV(i, j, int(c.level))
	return c
}

// CellFromPoint constructs a cell for the given Point.
func CellFromPoint(p Point) Cell {
	return CellFromCellID(CellIDFromPoint(p))
}

// CellFromLatLng constructs a cell for the given LatLng.
func CellFromLatLng(ll LatLng) Cell {
	return CellFromCellID(CellIDFromLatLng(ll))
}

func (c Cell) Id() CellID {
	return c.id
}

func (c Cell) Face() int8 {
	return c.face
}

func (c Cell) Level() int8 {
	return c.level
}

func (c Cell) Orientation() int8 {
	return c.orientation
}

// IsLeaf returns whether this Cell is a leaf or not.
func (c Cell) IsLeaf() bool {
	return c.level == maxLevel
}

// SizeIJ returns the CellID value for the cells level.
func (c Cell) SizeIJ() int {
	return sizeIJ(int(c.level))
}

// Vertex returns the k-th vertex of the cell (k = [0,3]) in CCW order
// (lower left, lower right, upper right, upper left in the UV plane).
func (c Cell) Vertex(k int) Point {
	return Point{faceUVToXYZ(int(c.face), c.uv.Vertices()[k].X, c.uv.Vertices()[k].Y).Normalize()}
}

// Edge returns the inward-facing normal of the great circle passing through
// the CCW ordered edge from vertex k to vertex k+1 (mod 4).
func (c Cell) Edge(k int) Point {
	return Point{c.EdgeRaw(k).Normalize()}
}

func (c Cell) EdgeRaw(k int) Point {
	switch k {
	case 0:
		return Point{vNorm(int(c.face), c.uv.Y.Lo)} // Bottom
	case 1:
		return Point{uNorm(int(c.face), c.uv.X.Hi)} // Right
	case 2:
		return Point{vNorm(int(c.face), c.uv.Y.Hi).Mul(-1.0)} // Top
	default:
		return Point{uNorm(int(c.face), c.uv.X.Lo).Mul(-1.0)} // Left
	}
}

// ExactArea return the area of this cell as accurately as possible.
func (c Cell) ExactArea() float64 {
	v0, v1, v2, v3 := c.Vertex(0), c.Vertex(1), c.Vertex(2), c.Vertex(3)
	return PointArea(v0, v1, v2) + PointArea(v0, v2, v3)
}

// CapBound returns a bounding spherical cap. This is not guaranteed to be exact.
func (c Cell) CapBound() Cap {
	// Use the cell center in (u,v)-space as the cap axis. This vector is
	// very close to GetCenter() and faster to compute. Neither one of these
	// vectors yields the bounding cap with minimal surface area, but they
	// are both pretty close.
	//
	// It's possible to show that the two vertices that are furthest from
	// the (u,v)-origin never determine the maximum cap size (this is a
	// possible future optimization).
	u := c.uv.Center().X
	v := c.uv.Center().Y
	cap := CapFromCenterHeight(Point{faceUVToXYZ(int(c.face), u, v).Normalize()}, 0)
	for k := 0; k < 4; k++ {
		cap = cap.AddPoint(c.Vertex(k))
	}
	return cap
}

func (c Cell) ContainsPoint(point Point) bool {
	// We can't just call XYZtoFaceUV, because for points that lie on the
	// boundary between two faces (i.e. u or v is +1/-1) we need to return
	// true for both adjacent cells.
	u, v, ok := faceXYZToUV(int(c.face), point)
	if !ok {
		return false
	}
	return u >= c.uv.X.Lo && u <= c.uv.X.Hi && v >= c.uv.Y.Lo && v <= c.uv.Y.Hi
}

// RectBound returns a bounding latitude-longitude rectangle that contains
// the region. The bounds are not guaranteed to be tight.
func (c Cell) RectBound() Rect {
	// TODO: Implement
	return EmptyRect()
}

// ContainsCell reports whether the region completely contains the given region.
// It returns false if containment could not be determined.
func (c Cell) ContainsCell(other Cell) bool {
	return c.Id().Contains(other.Id())
}

// IntersectsCell reports whether the region intersects the given cell or
// if intersection could not be determined. It returns false if the region
// does not intersect.
func (c Cell) IntersectsCell(other Cell) bool {
	return c.Id().Intersects(other.Id())
}
