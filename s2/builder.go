// Copyright 2023 Google Inc. All rights reserved.
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

const (
	// maxEdgeDeviationRatio is set so that MaxEdgeDeviation will be large enough
	// compared to snapRadius such that edge splitting is rare.
	//
	// Using spherical trigonometry, if the endpoints of an edge of length L
	// move by at most a distance R, the center of the edge moves by at most
	// asin(sin(R) / cos(L / 2)). Thus the (MaxEdgeDeviation / SnapRadius)
	// ratio increases with both the snap radius R and the edge length L.
	//
	// We arbitrarily limit the edge deviation to be at most 10% more than the
	// snap radius. With the maximum allowed snap radius of 70 degrees, this
	// means that edges up to 30.6 degrees long are never split. For smaller
	// snap radii, edges up to 49 degrees long are never split. (Edges of any
	// length are not split unless their endpoints move far enough so that the
	// actual edge deviation exceeds the limit; in practice, splitting is rare
	// even with long edges.) Note that it is always possible to split edges
	// when MaxEdgeDeviation is exceeded.
	maxEdgeDeviationRatio = 1.1
)

// edgeType indicates whether the input edges are undirected. Typically this is
// specified for each output layer (e.g., PolygonBuilderLayer).
//
// Directed edges are preferred, since otherwise the output is ambiguous.
// For example, output polygons may be the *inverse* of the intended result
// (e.g., a polygon intended to represent the world's oceans may instead
// represent the world's land masses). Directed edges are also somewhat
// more efficient.
//
// However even with undirected edges, most Builder layer types try to
// preserve the input edge direction whenever possible. Generally, edges
// are reversed only when it would yield a simpler output. For example,
// PolygonLayer assumes that polygons created from undirected edges should
// cover at most half of the sphere. Similarly, PolylineVectorBuilderLayer
// assembles edges into as few polylines as possible, even if this means
// reversing some of the "undirected" input edges.
//
// For shapes with interiors, directed edges should be oriented so that the
// interior is to the left of all edges. This means that for a polygon with
// holes, the outer loops ("shells") should be directed counter-clockwise
// while the inner loops ("holes") should be directed clockwise. Note that
// AddPolygon() follows this convention automatically.
type edgeType uint8

const (
	edgeTypeDirected edgeType = iota
	edgeTypeUndirected
)

// builder is a tool for assembling polygonal geometry from edges. Here are
// some of the things it is designed for:
//
//  1. Building polygons, polylines, and polygon meshes from unsorted
//     collections of edges.
//
//  2. Snapping geometry to discrete representations (such as CellID centers
//     or E7 lat/lng coordinates) while preserving the input topology and with
//     guaranteed error bounds.
//
// 3. Simplifying geometry (e.g. for indexing, display, or storage).
//
//  4. Importing geometry from other formats, including repairing geometry
//     that has errors.
//
//  5. As a tool for implementing more complex operations such as polygon
//     intersections and unions.
//
// The implementation is based on the framework of "snap rounding".  Unlike
// most snap rounding implementations, Builder defines edges as geodesics on
// the sphere (straight lines) and uses the topology of the sphere (i.e.,
// there are no "seams" at the poles or 180th meridian). The algorithm is
// designed to be 100% robust for arbitrary input geometry. It offers the
// following properties:
//
//   - Guaranteed bounds on how far input vertices and edges can move during
//     the snapping process (i.e., at most the given "snapRadius").
//
//   - Guaranteed minimum separation between edges and vertices other than
//     their endpoints (similar to the goals of Iterated Snap Rounding). In
//     other words, edges that do not intersect in the output are guaranteed
//     to have a minimum separation between them.
//
//   - Idempotency (similar to the goals of Stable Snap Rounding), i.e. if the
//     input already meets the output criteria then it will not be modified.
//
//   - Preservation of the input topology (up to the creation of
//     degeneracies). This means that there exists a continuous deformation
//     from the input to the output such that no vertex crosses an edge.  In
//     other words, self-intersections won't be created, loops won't change
//     orientation, etc.
//
//   - The ability to snap to arbitrary discrete point sets (such as CellID
//     centers, E7 lat/lng points on the sphere, or simply a subset of the
//     input vertices), rather than being limited to an integer grid.
//
// Here are some of its other features:
//
//   - It can handle both directed and undirected edges. Undirected edges can
//     be useful for importing data from other formats, e.g. where loops have
//     unspecified orientations.
//
//   - It can eliminate self-intersections by finding all edge pairs that cross
//     and adding a new vertex at each intersection point.
//
//   - It can simplify polygons to within a specified tolerance. For example,
//     if two vertices are close enough they will be merged, and if an edge
//     passes nearby a vertex then it will be rerouted through that vertex.
//     Optionally, it can also detect nearly straight chains of short edges and
//     replace them with a single long edge, while maintaining the same
//     accuracy, separation, and topology guarantees ("simplify_edge_chains").
//
//   - It supports many different output types through the concept of "layers"
//     (polylines, polygons, polygon meshes, etc). You can build multiple
//     layers at once in order to ensure that snapping does not create
//     intersections between different objects (for example, you can simplify a
//     set of contour lines without the risk of having them cross each other).
//
//   - It supports edge labels, which allow you to attach arbitrary information
//     to edges and have it preserved during the snapping process. (This can
//     also be achieved using layers, at a coarser level of granularity.)
//
// Caveats:
//
//   - Because Builder only works with edges, it cannot distinguish between
//     the empty and full polygons. If your application can generate both the
//     empty and full polygons, you must implement logic outside of this class.
//
// Example showing how to snap a polygon to E7 coordinates:
//
//	builder := NewBuilder(BuilderOptions(IntLatLngSnapFunction(7)));
//	var output *Polygon
//	builder.StartLayer(NewPolygonLayer(output))
//	builder.AddPolygon(input);
//	if err := builder.Build(); err != nil {
//	  fmt.Printf("error building: %v\n"), err
//	  ...
//	}
//
// TODO(rsned): Make the type public when Builder is ready.
type builder struct {
	opts *builderOptions
}

// init initializes this instance with the given options.
func (b *builder) init(opts *builderOptions) {
	b.opts = opts
}
