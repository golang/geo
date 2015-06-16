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
	"fmt"

	"github.com/golang/geo/s1"
)

/**
 * A Polyline represents a sequence of zero or more vertices connected by
 * straight edges (geodesics). Edges of length 0 and 180 degrees are not
 * allowed, i.e. adjacent vertices should not be identical or antipodal.
 *
 * <p>Note: Polylines do not have a Contains(S2Point) method, because
 * "containment" is not numerically well-defined except at the polyline
 * vertices.
 */
type Polyline struct {
	Vertices []Point
}

func PolylineFromPoints(points []Point) Polyline {
	return Polyline{points}
}

// Return true if the given vertices form a valid polyline.
func (p Polyline) IsValid() bool {
	// All vertices must be unit length.
	n := len(p.Vertices)
	for i := 0; i < n; i++ {
		if !p.Vertices[i].IsUnit() {
			fmt.Printf("Vertex %d is not unit length", i)
			return false
		}
	}
	// Adjacent vertices must not be identical or antipodal.
	for i := 0; i < n; i++ {
		if p.Vertices[i-1].ApproxEquals(p.Vertices[i], EPSILON) || p.Vertices[i-1].ApproxEquals(Point{p.Vertices[i].Neg()}, EPSILON) {
			fmt.Printf("Vertices %d and %d are identical or antipodal", (i - 1), i)
			return false
		}
	}
	return true
}

func (p Polyline) NumVertices() int { return len(p.Vertices) }

func (p Polyline) Vertex(k int) Point { return p.Vertices[k] }

// Return the angle corresponding to the total arclength of the polyline on a unit sphere.
func (p Polyline) GetArclengthAngle() s1.Angle {
	var lengthSum s1.Angle = 0
	for i := 1; i < p.NumVertices(); i++ {
		lengthSum += p.Vertices[i-1].Angle(p.Vertices[i].Vector)
	}
	return lengthSum
}

// CapBound returns a bounding spherical cap. This is not guaranteed to be exact.
func (p Polyline) CapBound() Cap {
	return p.RectBound().CapBound()
}

// RectBound returns a bounding latitude-longitude rectangle that contains
// the region. The bounds are not guaranteed to be tight.
func (p Polyline) RectBound() Rect {
	rb := NewRectBounder()
	for i := 0; i < p.NumVertices(); i++ {
		rb.AddPoint(p.Vertex(i))
	}
	return rb.GetBound()
}

// ContainsCell reports whether the region completely contains the given region.
// It returns false if containment could not be determined.
func (p Polyline) ContainsCell(cell Cell) bool {
	return false
}

// IntersectsCell reports whether the region intersects the given cell or
// if intersection could not be determined. It returns false if the region
// does not intersect.
func (p Polyline) IntersectsCell(cell Cell) bool {
	if p.NumVertices() == 0 {
		return false
	}

	// We only need to check whether the cell contains vertex 0 for correctness,
	// but these tests are cheap compared to edge crossings so we might as well
	// check all the vertices.
	for i := 0; i < p.NumVertices(); i++ {
		if cell.ContainsPoint(p.Vertex(i)) {
			return true
		}
	}

	cellVertices := make([]Point, 4)
	for i := 0; i < 4; i++ {
		cellVertices[i] = cell.Vertex(i)
	}
	for j := 0; j < 4; j++ {
		ec := NewEdgeCrosser(cellVertices[j], cellVertices[(j+1)&3], p.Vertex(0))
		for i := 1; i < p.NumVertices(); i++ {
			if ec.RobustCrossing(p.Vertex(i)) >= 0 {
				// There is a proper crossing, or two vertices were the same.
				return true
			}
		}
	}
	return false
}
