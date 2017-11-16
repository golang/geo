/*
Copyright 2017 Google Inc. All rights reserved.

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

// laxPolyline represents a polyline. It is similar to Polyline except
// that duplicate vertices are allowed, and the representation is slightly
// more compact.
//
// Polylines may have any number of vertices, but note that polylines with
// fewer than 2 vertices do not define any edges. (To create a polyline
// consisting of a single degenerate edge, either repeat the same vertex twice
// or use laxClosedPolyline.
type laxPolyline struct {
	vertices []Point
}

func laxPolylineFromPoints(vertices []Point) *laxPolyline {
	return &laxPolyline{
		vertices: append([]Point(nil), vertices...),
	}
}

func laxPolylineFromPolyline(p Polyline) *laxPolyline {
	return laxPolylineFromPoints(p)
}

func (l *laxPolyline) NumEdges() int                     { return maxInt(0, len(l.vertices)-1) }
func (l *laxPolyline) Edge(e int) Edge                   { return Edge{l.vertices[e], l.vertices[e+1]} }
func (l *laxPolyline) HasInterior() bool                 { return false }
func (l *laxPolyline) ReferencePoint() ReferencePoint    { return OriginReferencePoint(false) }
func (l *laxPolyline) NumChains() int                    { return minInt(1, l.NumEdges()) }
func (l *laxPolyline) Chain(i int) Chain                 { return Chain{0, l.NumEdges()} }
func (l *laxPolyline) ChainEdge(i, j int) Edge           { return Edge{l.vertices[j], l.vertices[j+1]} }
func (l *laxPolyline) ChainPosition(e int) ChainPosition { return ChainPosition{0, e} }
func (l *laxPolyline) dimension() dimension              { return polylineGeometry }
