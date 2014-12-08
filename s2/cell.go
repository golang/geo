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
	return CellFromCellID(cellIDFromPoint(p))
}

// CellFromLatLng constructs a cell for the given LatLng.
func CellFromLatLng(ll LatLng) Cell {
	return CellFromCellID(CellIDFromLatLng(ll))
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
	switch k {
	case 0:
		return Point{vNorm(int(c.face), c.uv.Y.Lo).Normalize()} // Bottom
	case 1:
		return Point{uNorm(int(c.face), c.uv.X.Hi).Normalize()} // Right
	case 2:
		return Point{vNorm(int(c.face), c.uv.Y.Hi).Mul(-1.0).Normalize()} // Top
	default:
		return Point{uNorm(int(c.face), c.uv.X.Lo).Mul(-1.0).Normalize()} // Left
	}
}

// ExactArea return the area of this cell as accurately as possible.
func (c Cell) ExactArea() float64 {
	v0, v1, v2, v3 := c.Vertex(0), c.Vertex(1), c.Vertex(2), c.Vertex(3)
	return PointArea(v0, v1, v2) + PointArea(v0, v2, v3)
}

// TODO(roberts, or $SOMEONE): Differences from C++, almost everything else still.
// Implement the accessor methods on the internal fields.
