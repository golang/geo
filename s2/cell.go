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
	"math"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
	"github.com/golang/geo/s1"
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
		return Point{vNorm(int(c.face), c.uv.Y.Hi).Neg()} // Top
	default:
		return Point{uNorm(int(c.face), c.uv.X.Lo).Neg()} // Left
	}
}

/**
 * Return the average area for cells at the given level.
 */
func AverageArea(level int) float64 {
	return S2_PROJECTION.AVG_AREA().GetValue(level)
}

/**
 * Return the average area of cells at this level. This is accurate to within
 * a factor of 1.7 (for S2_QUADRATIC_PROJECTION) and is extremely cheap to
 * compute.
 */
func (c Cell) AverageArea() float64 {
	return AverageArea(int(c.level))
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

// We grow the bounds slightly to make sure that the bounding rectangle
// also contains the normalized versions of the vertices. Note that the
// maximum result magnitude is Pi, with a floating-point exponent of 1.
// Therefore adding or subtracting 2**-51 will always change the result.
var MAX_ERROR float64 = 1.0 / (1 << 51)

// The 4 cells around the equator extend to +/-45 degrees latitude at the
// midpoints of their top and bottom edges. The two cells covering the
// poles extend down to +/-35.26 degrees at their vertices.
// adding kMaxError (as opposed to the C version) because of asin and atan2
// roundoff errors
var POLE_MIN_LAT float64 = math.Asin(math.Sqrt(1.0/3.0)) - MAX_ERROR // 35.26 degrees

// RectBound returns a bounding latitude-longitude rectangle that contains
// the region. The bounds are not guaranteed to be tight.
func (c Cell) RectBound() Rect {
	if c.level > 0 {
		// Except for cells at level 0, the latitude and longitude extremes are
		// attained at the vertices. Furthermore, the latitude range is
		// determined by one pair of diagonally opposite vertices and the
		// longitude range is determined by the other pair.
		//
		// We first determine which corner (i,j) of the cell has the largest
		// absolute latitude. To maximize latitude, we want to find the point in
		// the cell that has the largest absolute z-coordinate and the smallest
		// absolute x- and y-coordinates. To do this we look at each coordinate
		// (u and v), and determine whether we want to minimize or maximize that
		// coordinate based on the axis direction and the cell's (u,v) quadrant.
		u := c.uv.X.Lo + c.uv.X.Hi
		v := c.uv.Y.Lo + c.uv.Y.Hi
		var i, j int
		if uAxis(int(c.face)).Z == 0 {
			if u < 0 {
				i = 1
			}
		} else {
			if u > 0 {
				i = 1
			}
		}
		if vAxis(int(c.face)).Z == 0 {
			if v < 0 {
				j = 1
			}
		} else {
			if v > 0 {
				j = 1
			}
		}

		lat := r1.IntervalFromPointPair(c.latitude(i, j), c.latitude(1-i, 1-j))
		lat = lat.Expanded(MAX_ERROR).Intersection(validRectLatRange)
		if lat.Lo == validRectLatRange.Lo || lat.Hi == validRectLatRange.Hi {
			return Rect{lat, s1.FullInterval()}
		}
		lng := s1.IntervalFromPointPair(c.longitude(i, 1-j), c.longitude(1-i, j))
		return Rect{lat, lng.Expanded(MAX_ERROR)}
	}

	switch c.face {
	case 0:
		return Rect{r1.Interval{-math.Pi / 4, math.Pi / 4}, s1.Interval{-math.Pi / 4, math.Pi / 4}}
	case 1:
		return Rect{r1.Interval{-math.Pi / 4, math.Pi / 4}, s1.Interval{math.Pi / 4, 3 * math.Pi / 4}}
	case 2:
		return Rect{r1.Interval{POLE_MIN_LAT, math.Pi / 2}, s1.Interval{-math.Pi, math.Pi}}
	case 3:
		return Rect{r1.Interval{-math.Pi / 4, math.Pi / 4}, s1.Interval{3 * math.Pi / 4, -3 * math.Pi / 4}}
	case 4:
		return Rect{r1.Interval{-math.Pi / 4, math.Pi / 4}, s1.Interval{-3 * math.Pi / 4, -math.Pi / 4}}
	default:
		return Rect{r1.Interval{-math.Pi / 2, -POLE_MIN_LAT}, s1.Interval{-math.Pi, math.Pi}}
	}
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

func (c Cell) latitude(i, j int) float64 {
	u := c.uv.X.Lo
	if i == 1 {
		u = c.uv.X.Hi
	}
	v := c.uv.Y.Lo
	if j == 1 {
		v = c.uv.Y.Hi
	}
	p := Point{faceUVToXYZ(int(c.face), u, v)}
	return latitude(p).Radians()
}

func (c Cell) longitude(i, j int) float64 {
	u := c.uv.X.Lo
	if i == 1 {
		u = c.uv.X.Hi
	}
	v := c.uv.Y.Lo
	if j == 1 {
		v = c.uv.Y.Hi
	}
	p := Point{faceUVToXYZ(int(c.face), u, v)}
	return longitude(p).Radians()
}
