/*
Copyright 2016 Google Inc. All rights reserved.

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

	"github.com/golang/geo/s1"
)

// Polyline represents a sequence of zero or more vertices connected by
// straight edges (geodesics). Edges of length 0 and 180 degrees are not
// allowed, i.e. adjacent vertices should not be identical or antipodal.
type Polyline []Point

// PolylineFromLatLngs creates a new Polyline from the given LatLngs.
func PolylineFromLatLngs(points []LatLng) *Polyline {
	p := make(Polyline, len(points))
	for k, v := range points {
		p[k] = PointFromLatLng(v)
	}
	return &p
}

// Reverse reverses the order of the Polyline vertices.
func (p *Polyline) Reverse() {
	for i := 0; i < len(*p)/2; i++ {
		(*p)[i], (*p)[len(*p)-i-1] = (*p)[len(*p)-i-1], (*p)[i]
	}
}

// Length returns the length of this Polyline.
func (p *Polyline) Length() s1.Angle {
	var length s1.Angle

	for i := 1; i < len(*p); i++ {
		length += (*p)[i-1].Distance((*p)[i])
	}
	return length
}

// Centroid returns the true centroid of the polyline multiplied by the length of the
// polyline. The result is not unit length, so you may wish to normalize it.
//
// Scaling by the Polyline length makes it easy to compute the centroid
// of several Polylines (by simply adding up their centroids).
func (p *Polyline) Centroid() Point {
	var centroid Point
	for i := 1; i < len(*p); i++ {
		// The centroid (multiplied by length) is a vector toward the midpoint
		// of the edge, whose length is twice the sin of half the angle between
		// the two vertices. Defining theta to be this angle, we have:
		vSum := (*p)[i-1].Add((*p)[i].Vector)  // Length == 2*cos(theta)
		vDiff := (*p)[i-1].Sub((*p)[i].Vector) // Length == 2*sin(theta)

		// Length == 2*sin(theta)
		centroid = Point{centroid.Add(vSum.Mul(math.Sqrt(vDiff.Norm2() / vSum.Norm2())))}
	}
	return centroid
}

// Equals reports whether the given Polyline is exactly the same as this one.
func (p *Polyline) Equals(b *Polyline) bool {
	if len(*p) != len(*b) {
		return false
	}
	for i, v := range *p {
		if v != (*b)[i] {
			return false
		}
	}

	return true
}

// CapBound returns the bounding Cap for this Polyline.
func (p *Polyline) CapBound() Cap {
	return p.RectBound().CapBound()
}

// RectBound returns the bounding Rect for this Polyline.
func (p *Polyline) RectBound() Rect {
	rb := NewRectBounder()
	for _, v := range *p {
		rb.AddPoint(v)
	}
	return rb.RectBound()
}

// ContainsCell reports whether this Polyline contains the given Cell. Always returns false
// because "containment" is not numerically well-defined except at the Polyline vertices.
func (p *Polyline) ContainsCell(cell Cell) bool {
	return false
}

// IntersectsCell reports whether this Polyline intersects the given Cell.
func (p *Polyline) IntersectsCell(cell Cell) bool {
	if len(*p) == 0 {
		return false
	}

	// We only need to check whether the cell contains vertex 0 for correctness,
	// but these tests are cheap compared to edge crossings so we might as well
	// check all the vertices.
	for _, v := range *p {
		if cell.ContainsPoint(v) {
			return true
		}
	}

	cellVertices := []Point{
		cell.Vertex(0),
		cell.Vertex(1),
		cell.Vertex(2),
		cell.Vertex(3),
	}

	for j := 0; j < 4; j++ {
		crosser := NewChainEdgeCrosser(cellVertices[j], cellVertices[(j+1)&3], (*p)[0])
		for i := 1; i < len(*p); i++ {
			if crosser.ChainCrossingSign((*p)[i]) != DoNotCross {
				// There is a proper crossing, or two vertices were the same.
				return true
			}
		}
	}
	return false
}

// NumEdges returns the number of edges in this shape.
func (p *Polyline) NumEdges() int {
	if len(*p) == 0 {
		return 0
	}
	return len(*p) - 1
}

// Edge returns endpoints for the given edge index.
func (p *Polyline) Edge(i int) (a, b Point) {
	return (*p)[i], (*p)[i+1]
}

// dimension returns the dimension of the geometry represented by this Polyline.
func (p *Polyline) dimension() dimension { return polylineGeometry }

// numChains reports the number of contiguous edge chains in this Polyline.
func (p *Polyline) numChains() int {
	if p.NumEdges() >= 1 {
		return 1
	}
	return 0
}

// chainStart returns the id of the first edge in the i-th edge chain in this Polyline.
func (p *Polyline) chainStart(i int) int {
	if i == 0 {
		return 0
	}

	return p.NumEdges()
}

// HasInterior returns false as Polylines are not closed.
func (p *Polyline) HasInterior() bool {
	return false
}

// ContainsOrigin returns false because there is no interior to contain s2.Origin.
func (p *Polyline) ContainsOrigin() bool {
	return false
}

// TODO(roberts): Differences from C++.
// IsValid
// Suffix
// Interpolate/UnInterpolate
// Project
// IsPointOnRight
// Intersects
// Reverse
// SubsampleVertices
// ApproxEqual
// NearlyCoversPolyline
