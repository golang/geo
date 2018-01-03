// Copyright 2017 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s2

// Shape interface enforcement
var (
	_ Shape = (pointVectorShape)(nil)
)

// pointVectorShape is a Shape representing a set of Points. Each point
// is represented as a degenerate point with the same starting and ending
// vertices.
//
// This type is useful for adding a collection of points to an ShapeIndex.
type pointVectorShape []Point

func (p pointVectorShape) NumEdges() int                     { return len(p) }
func (p pointVectorShape) Edge(i int) Edge                   { return Edge{p[i], p[i]} }
func (p pointVectorShape) HasInterior() bool                 { return false }
func (p pointVectorShape) ReferencePoint() ReferencePoint    { return OriginReferencePoint(false) }
func (p pointVectorShape) NumChains() int                    { return len(p) }
func (p pointVectorShape) Chain(i int) Chain                 { return Chain{i, 1} }
func (p pointVectorShape) ChainEdge(i, j int) Edge           { return Edge{p[i], p[j]} }
func (p pointVectorShape) ChainPosition(e int) ChainPosition { return ChainPosition{e, 0} }
func (p pointVectorShape) dimension() dimension              { return pointGeometry }
