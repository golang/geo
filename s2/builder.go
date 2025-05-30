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

import (
	"math"

	"github.com/golang/geo/s1"
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
//  3. Simplifying geometry (e.g. for indexing, display, or storage).
//
//  4. Importing geometry from other formats, including repairing geometry
//     that has errors.
//
//  5. As a tool for implementing more complex operations such as polygon
//     intersections and unions.
//
// The implementation is based on the framework of "snap rounding". Unlike
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
//     accuracy, separation, and topology guarantees ("simplifyEdgeChains").
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
//     empty and full polygons, you must implement logic outside of this type.
//
// Example showing how to snap a polygon to E7 coordinates:
//
//	builder := NewBuilder(NewBuilderOptions(IntLatLngSnapFunction(7)));
//	var output *Polygon
//	builder.StartLayer(NewPolygonLayer(output))
//	builder.AddPolygon(input);
//	if err := builder.Build(); err != nil {
//	    fmt.Printf("error building: %v\n"), err)
//	    // ...
//	}
//
// TODO(rsned): Make this type public when Builder is ready.
type builder struct {
	opts builderOptions

	// siteSnapRadiusChordAngle is the maximum distance (inclusive) that a
	// vertex can move when snapped, equal to opts.snapper.SnapRadius().
	siteSnapRadiusChordAngle s1.ChordAngle

	// edgeSnapRadiusChordAngle is the maximum distance (inclusive) that an
	// edge can move when snapping to a snap site. It can be slightly larger
	// than the site snap radius when edges are being split at crossings.
	edgeSnapRadiusChordAngle s1.ChordAngle

	// checkAllSiteCrossings reports if we need to check that snapping has not
	// changed the input topology around any vertex (i.e. Voronoi site).
	// Normally this is only necessary for forced vertices, but if the snap
	// radius is very small (e.g., zero) and opts.splitCrossingEdges is true
	// then we need to do this for all vertices.
	// In all other situations, any snapped edge that crosses a vertex will also
	// be closer than opts.minEdgeVertexSeparation() to that vertex, which will
	// cause us to add a separation site anyway.
	checkAllSiteCrossings bool

	// maxEdgeDeviation is the maximum distance that a vertex can be separated
	// from an edge while still affecting how that edge is snapped.
	maxEdgeDeviation s1.Angle

	// edgeSiteQueryRadiusChordAngle is the distance from an edge in which candidates for
	// snapping and/or avoidance are considered.
	edgeSiteQueryRadiusChordAngle s1.ChordAngle

	// minEdgeLengthToSplitChordAngle is the maximum edge length such that
	// even if both endpoints move by the maximum distance allowed (i.e.
	// edgeSnapRadius), the center of the edge will still move by less than
	// maxEdgeDeviation.
	minEdgeLengthToSplitChordAngle s1.ChordAngle

	// minSiteSeparation comes from the snapper.
	minSiteSeparation s1.Angle

	// minSiteSeparationChordAngle is the ChordAngle of the minSiteSeparation.
	minSiteSeparationChordAngle s1.ChordAngle

	// minEdgeSiteSeparationChordAngle is the minimum separation between edges
	// and sites as a ChordAngle.
	minEdgeSiteSeparationChordAngle s1.ChordAngle

	// minEdgeSiteSeparationChordAngleLimit is the upper bound on the distance
	// from ClosestEdgeQuery as a ChordAngle.
	minEdgeSiteSeparationChordAngleLimit s1.ChordAngle

	// maxAdjacentSiteSeparationChordAngle is the maximum possible distance
	// between two sites whose Voronoi regions touch, increased to account
	// for errors. (The maximum radius of each Voronoi region is
	// edgeSnapRadius.)
	maxAdjacentSiteSeparationChordAngle s1.ChordAngle

	// edgeSnapRadiusSin2 is the squared sine of the edge snap radius.
	// This is equivalent to the snap radius (squared) for distances
	// measured through the interior of the sphere to the plane containing
	// an edge. This value is used only when interpolating new points along
	// edges (see getSeparationSite()).
	edgeSnapRadiusSin2 float64

	// snappingRequested reports if snapping was requested.
	// This is true if either opts.snapper.SnapRadius() is positive, or
	// opts.splitCrossingEdges is true (which implicitly requests snapping to
	// ensure that both crossing edges are snapped to the intersection point).
	snappingRequested bool

	// snappingNeeded will be set to to true when it is discovered that at
	// least one input vertex or edge does not meet the output guarantees
	// (e.g., that vertices are separated by at least snapper.minVertexSeparation).
	snappingNeeded bool

	//  labelSetModified indicates whether labelSet has been modified since the
	//  last time labelSetId was computed.
	labelSetModified bool

	// inputVertices is a slice of input vertices.
	inputVertices []Point

	// inputEdges is a slice of all Edges for all Layers.
	///// inputEdges []builderEdge

	// layers is a slice of layers for this Builder.  The last layer is the
	// current layer.  All edges are assigned to the current layer when the
	// edge is added.
	layers []*builderLayer

	// layerOptions is a slice of graphOptions corresponding to the layers slice.
	// TODO(rsned): pull this struct out or replace with a generic.
	layerOptions []struct {
		first, second int32
	}

	// layerBegins is a slice of int32s, each indicating the index into
	// inputEdges corresponding to each layer.
	layerBegins []int32

	// layerIsFullPolygonPredicates is a slice of isFullPolygonPredicate,
	// each indicating the predicate for the corresponding layer.
	layerIsFullPolygonPredicates []isFullPolygonPredicate

	// Each input edge has a labelSetID (an int32) representing the set of
	// labels attached to that edge. This slice is populated only if at least
	// one label is used.
	labelSetIDs []int32
	// labelSetLexicon stores labels assigned to each labelSet.
	labelSetLexicon *idSetLexicon

	// labelSet is the current set of labels (represented as a stack).
	labelSet []int32

	// The labelSetID corresponding to the current label set, computed on demand
	// (by adding it to labelSetLexicon).
	labelSetID int32

	// The remaining fields are used for snapping and simplifying.

	// numForcedSites is the number of sites specified using forceVertex().
	// These sites are always at the beginning of the sites vector.
	numForcedSites int32

	// sites is the set of snapped vertex locations ("sites").
	sites []Point

	// edgeSites is a map from each input edge to the set of sites "nearby"
	// that edge, defined as the set of sites that are candidates for snapping
	// and/or avoidance. Sites are kept by increasing distance from the origin
	// of the input edge.
	//
	// Once snapping is finished, this field is discarded unless edge chain
	// simplification was requested, in which case instead the sites are
	// filtered by removing the ones that each edge was snapped to, leaving only
	// the "sites to avoid" (needed for simplification).
	edgeSites [][]int32

	// TODO(rsned): Add memoryTracker if it becomes available.
}

// init initializes this instance with the given options.
func (b *builder) init(opts builderOptions) {
	b.opts = opts

	snapper := opts.snapper
	sr := snapper.SnapRadius()

	// Cap the snap radius to the limit.
	sr = min(sr, maxSnapRadius)

	// Convert the snap radius to an ChordAngle. This is the "true snap
	// radius" used when evaluating exact predicates.
	b.siteSnapRadiusChordAngle = s1.ChordAngleFromAngle(sr)

	// When intersectionTolerance is non-zero we need to use a larger snap
	// radius for edges than for vertices to ensure that both edges are snapped
	// to the edge intersection location.  This is because the computed
	// intersection point is not exact; it may be up to intersectionTolerance
	// away from its true position. The computed intersection point might then
	// be snapped to some other vertex up to SnapRadius away.  So to ensure
	// that both edges are snapped to a common vertex, we need to increase the
	// snap radius for edges to at least the sum of these two values (calculated
	// conservatively).
	edgeSnapRadius := opts.edgeSnapRadius()
	b.edgeSnapRadiusChordAngle = roundChordAngleUp(edgeSnapRadius)
	b.snappingRequested = (edgeSnapRadius > 0)

	// Compute the maximum distance that a vertex can be separated from an
	// edge while still affecting how that edge is snapped.
	b.maxEdgeDeviation = opts.maxEdgeDeviation()
	b.edgeSiteQueryRadiusChordAngle = s1.ChordAngleFromAngle(b.maxEdgeDeviation +
		snapper.MinEdgeVertexSeparation())

	// Compute the maximum edge length such that even if both endpoints move by
	// the maximum distance allowed (i.e., edgeSnapRadius), the center of the
	// edge will still move by less than maxEdgeDeviation. This saves us a
	// lot of work since then we don't need to check the actual deviation.
	if !b.snappingRequested {
		b.minEdgeLengthToSplitChordAngle = s1.InfChordAngle()
	} else {
		// This value varies between 30 and 50 degrees depending on
		// the snap radius.
		b.minEdgeLengthToSplitChordAngle = s1.ChordAngleFromAngle(s1.Angle(2 *
			math.Acos(math.Sin(edgeSnapRadius.Radians())/
				math.Sin(b.maxEdgeDeviation.Radians()))))
	}

	// In rare cases we may need to explicitly check that the input topology is
	// preserved, i.e. that edges do not cross vertices when snapped.  This is
	// only necessary (1) for vertices added using forceVertex, and (2) when the
	// snap radius is smaller than intersectionTolerance (which is typically
	// either zero or intersectionError, about 9e-16 radians). This
	// condition arises because when a geodesic edge is snapped, the edge center
	// can move further than its endpoints. This can cause an edge to pass on the
	// wrong side of an input vertex. (Note that this could not happen in a
	// planar version of this algorithm.) Usually we don't need to consider this
	// possibility explicitly, because if the snapped edge passes on the wrong
	// side of a vertex then it is also closer than minEdgeVertexSeparation
	// to that vertex, which will cause a separation site to be added.
	//
	// If the condition below is true then we need to check all sites (i.e.,
	// snapped input vertices) for topology changes.  However this is almost never
	// the case because
	//
	//            maxEdgeDeviation() == 1.1 * edgeSnapRadius
	//      and   minEdgeVertexSeparation() >= 0.219 * SnapRadius
	//
	// for all currently implemented snap functions. The condition below is
	// only true when intersectionTolerance() is non-zero (which causes
	// edgeSnapRadius() to exceed SnapRadius() by intersectionError) and
	// SnapRadius() is very small (at most intersectionError / 1.19).
	b.checkAllSiteCrossings = (opts.maxEdgeDeviation() >
		opts.edgeSnapRadius()+snapper.MinEdgeVertexSeparation())

	// TODO(rsned): need to add check that b.checkAllSiteCrossings is false when tolerance is <= 0.
	// if opts.intersectionTolerance <= 0 {
	// }

	// To implement idempotency, we check whether the input geometry could
	// possibly be the output of a previous Builder invocation. This involves
	// testing whether any site/site or edge/site pairs are too close together.
	// This is done using exact predicates, which require converting the minimum
	// separation values to a ChordAngle.
	b.minSiteSeparation = snapper.MinVertexSeparation()
	b.minSiteSeparationChordAngle = s1.ChordAngleFromAngle(b.minSiteSeparation)
	b.minEdgeSiteSeparationChordAngle = s1.ChordAngleFromAngle(snapper.MinEdgeVertexSeparation())

	// This is an upper bound on the distance computed by ClosestPointQuery
	// where the true distance might be less than minEdgeSiteSeparationChordAngle.
	b.minEdgeSiteSeparationChordAngleLimit = addPointToEdgeError(b.minEdgeSiteSeparationChordAngle)

	// Compute the maximum possible distance between two sites whose Voronoi
	// regions touch. (The maximum radius of each Voronoi region is
	// edgeSnapRadius.) Then increase this bound to account for errors.
	b.maxAdjacentSiteSeparationChordAngle = addPointToPointError(roundChordAngleUp(2 * opts.edgeSnapRadius()))

	// Finally, we also precompute sin^2(edgeSnapRadius), which is simply the
	// squared distance between a vertex and an edge measured perpendicular to
	// the plane containing the edge, and increase this value by the maximum
	// error in the calculation to compare this distance against the bound.
	d := math.Sin(opts.edgeSnapRadius().Radians())
	b.edgeSnapRadiusSin2 = d * d
	b.edgeSnapRadiusSin2 += ((9.5*d+2.5+2*sqrt3)*d + 9*dblEpsilon) * dblEpsilon

	// Initialize the current label set.
	b.labelSetID = emptySetID
	b.labelSetModified = false

	// If snapping was requested, we try to determine whether the input geometry
	// already meets the output requirements. This is necessary for
	// idempotency, and can also save work. If we discover any reason that the
	// input geometry needs to be modified, snappingNeeded is set to true.
	b.snappingNeeded = false

	// TODO(rsned): Memory tracker init.
}

// roundUp rounds the given angle up by the max error and returns it as a chord angle.
func roundChordAngleUp(a s1.Angle) s1.ChordAngle {
	ca := s1.ChordAngleFromAngle(a)
	return ca.Expanded(ca.MaxAngleError())
}

func addPointToPointError(ca s1.ChordAngle) s1.ChordAngle {
	return ca.Expanded(ca.MaxPointError())
}

func addPointToEdgeError(ca s1.ChordAngle) s1.ChordAngle {
	return ca.Expanded(minUpdateDistanceMaxError(ca))
}

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

// isFullPolygonPredicate is an interface for determining if Polygons are
// full or not. For output layers that represent polygons, there is an ambiguity
// inherent in spherical geometry that does not exist in planar geometry.
// Namely, if a polygon has no edges, does it represent the empty polygon
// (containing no points) or the full polygon (containing all points)? This
// ambiguity also occurs for polygons that consist only of degeneracies, e.g.
// a degenerate loop with only two edges could be either a degenerate shell in
// the empty polygon or a degenerate hole in the full polygon.
//
// To resolve this ambiguity, an IsFullPolygonPredicate may be specified for
// each output layer (see AddIsFullPolygonPredicate below). If the output
// after snapping consists only of degenerate edges and/or sibling pairs
// (including the case where there are no edges at all), then the layer
// implementation calls the given predicate to determine whether the polygon
// is empty or full except for those degeneracies. The predicate is given
// a Graph containing the output edges, but note that in general
// the predicate must also have knowledge of the input geometry in order to
// determine the correct result.
//
// This predicate is only needed by layers that are assembled into polygons.
// It is not used by other layer types.
type isFullPolygonPredicate func(g *graph) (bool, error)

// polylineType indicates whether polylines should be "paths" (which don't
// allow duplicate vertices, except possibly the first and last vertex) or
// "walks" (which allow duplicate vertices and edges).
type polylineType uint8

const (
	polylineTypePath polylineType = iota
	polylineTypeWalk
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

// builderOptions holds the options for the Builder.
//
// TODO(rsned): Add public setters.
type builderOptions struct {
	// snapper holds the desired snap function.
	//
	// Note that if your input data includes vertices that were created using
	// Intersection(), then you should use a "snapRadius" of
	// at least intersectionMergeRadius, e.g. by calling
	//
	//  options.setSnapper(IdentitySnapFunction(intersectionMergeRadius));
	//
	// The default for this should be the IdentitySnapFunction(s1.Angle(0))
	// [This does no snapping and preserves all input vertices exactly.]
	snapper Snapper

	// splitCrossingEdges determines how crossing edges are handled by Builder.
	// If true, then detect all pairs of crossing edges and eliminate them by
	// adding a new vertex at their intersection point. See also the
	// AddIntersection() method which allows intersection points to be added
	// selectively.
	//
	// When this option if true, intersectionTolerance is automatically set
	// to a minimum of intersectionError (see intersectionTolerance
	// for why this is necessary). Note that this means that edges can move
	// by up to intersectionError even when the specified snap radius is
	// zero. The exact distance that edges can move is always given by
	// MaxEdgeDeviation().
	//
	// Undirected edges should always be used when the output is a polygon,
	// since splitting a directed loop at a self-intersection converts it into
	// two loops that don't define a consistent interior according to the
	// "interior is on the left" rule. (On the other hand, it is fine to use
	// directed edges when defining a polygon *mesh* because in that case the
	// input consists of sibling edge pairs.)
	//
	// Self-intersections can also arise when importing data from a 2D
	// projection. You can minimize this problem by subdividing the input
	// edges so that the S2 edges (which are geodesics) stay close to the
	// original projected edges (which are curves on the sphere). This can
	// be done using EdgeTessellator, for example.
	//
	// The default for this is false.
	splitCrossingEdges bool

	// intersectionTolerance specifies the maximum allowable distance between
	// a vertex added by AddIntersection() and the edge(s) that it is intended
	// to snap to. This method must be called before AddIntersection() can be
	// used. It has the effect of increasing the snap radius for edges (but not
	// vertices) by the given distance.
	//
	// The intersection tolerance should be set to the maximum error in the
	// intersection calculation used. For example, if Intersection()
	// is used then the error should be set to intersectionError. If
	// PointOnLine is used then the error should be set to PointOnLineError.
	// If Project is used then the error should be set to
	// projectPerpendicularError. If more than one method is used then the
	// intersection tolerance should be set to the maximum such error.
	//
	// The reason this option is necessary is that computed intersection
	// points are not exact. For example, Intersection(a, b, c, d)
	// returns a point up to intersectionError away from the true
	// mathematical intersection of the edges AB and CD. Furthermore such
	// intersection points are subject to further snapping in order to ensure
	// that no pair of vertices is closer than the specified snap radius. For
	// example, suppose the computed intersection point X of edges AB and CD
	// is 1 nanonmeter away from both edges, and the snap radius is 1 meter.
	// In that case X might snap to another vertex Y exactly 1 meter away,
	// which would leave us with a vertex Y that could be up to 1.000000001
	// meters from the edges AB and/or CD. This means that AB and/or CD might
	// not snap to Y leaving us with two edges that still cross each other.
	//
	// However if the intersection tolerance is set to 1 nanometer then the
	// snap radius for edges is increased to 1.000000001 meters ensuring that
	// both edges snap to a common vertex even in this worst case. (Tthis
	// technique does not work if the vertex snap radius is increased as well;
	// it requires edges and vertices to be handled differently.)
	//
	// Note that this option allows edges to move by up to the given
	// intersection tolerance even when the snap radius is zero. The exact
	// distance that edges can move is always given by maxEdgeDeviation()
	// defined above.
	//
	// When splitCrossingEdges is true, the intersection tolerance is
	// automatically set to a minimum of intersectionError. A larger
	// value can be specified by calling this method explicitly.
	//
	// The default tolerance should be 0.
	intersectionTolerance s1.Angle

	// simplifyEdgeChains determines if the output geometry should be simplified
	// by replacing nearly straight chains of short edges with a single long edge.
	//
	// The combined effect of snapping and simplifying will not change the
	// input by more than the guaranteed tolerances (see the list documented
	// with the SnapFunction class). For example, simplified edges are
	// guaranteed to pass within snapRadius() of the *original* positions of
	// all vertices that were removed from that edge. This is a much tighter
	// guarantee than can be achieved by snapping and simplifying separately.
	//
	// However, note that this option does not guarantee idempotency. In
	// other words, simplifying geometry that has already been simplified once
	// may simplify it further. (This is unavoidable, since tolerances are
	// measured with respect to the original geometry, which is no longer
	// available when the geometry is simplified a second time.)
	//
	// When the output consists of multiple layers, simplification is
	// guaranteed to be consistent: for example, edge chains are simplified in
	// the same way across layers, and simplification preserves topological
	// relationships between layers (e.g., no crossing edges will be created).
	// Note that edge chains in different layers do not need to be identical
	// (or even have the same number of vertices, etc) in order to be
	// simplified together. All that is required is that they are close
	// enough together so that the same simplified edge can meet all of their
	// individual snapping guarantees.
	//
	// Note that edge chains are approximated as parametric curves rather than
	// point sets. This means that if an edge chain backtracks on itself (for
	// example, ABCDEFEDCDEFGH) then such backtracking will be preserved to
	// within snapRadius() (for example, if the preceding point were all in a
	// straight line then the edge chain would be simplified to ACFCFH, noting
	// that C and F have degree > 2 and therefore can't be simplified away).
	//
	// Simplified edges are assigned all labels associated with the edges of
	// the simplified chain.
	//
	// For this option to have any effect, a SnapFunction with a non-zero
	// snapRadius() must be specified. Also note that vertices specified
	// using ForceVertex are never simplified away.
	//
	// The default for this is false.
	simplifyEdgeChains bool

	// idempotent determines if snapping occurs only when the input geometry
	// does not already meet the Builder output guarantees (see the Snapper
	// type description for details). This means that if all input vertices
	// are at snapped locations, all vertex pairs are separated by at least
	// MinVertexSeparation(), and all edge-vertex pairs are separated by at
	// least MinEdgeVertexSeparation(), then no snapping is done.
	//
	// If false, then all vertex pairs and edge-vertex pairs closer than
	// "edgeSnapRadius" will be considered for snapping. This can be useful, for
	// example, if you know that your geometry contains errors and you want to
	// make sure that features closer together than "edgeSnapRadius" are merged.
	//
	// This option is automatically turned off when simplifyEdgeChains is true
	// since simplifying edge chains is never guaranteed to be idempotent.
	//
	// The default for this is false. (meaning it IS idempotent)
	nonIdempotent bool
}

// defaultBuilderOptions returns a new instance with the proper defaults.
func defaultBuilderOptions() builderOptions {
	return builderOptions{
		snapper:               NewIdentitySnapper(0),
		splitCrossingEdges:    false,
		intersectionTolerance: s1.Angle(0),
		simplifyEdgeChains:    false,
		nonIdempotent:         false,
	}
}

// edgeSnapRadius reports the maximum distance from snapped edge vertices to
// the original edge. This is the same as SnapFunction().SnapRadius() except
// when splitCrossingEdges is true (see below), in which case the edge snap
// radius is increased by intersectionError.
func (o builderOptions) edgeSnapRadius() s1.Angle {
	return o.snapper.SnapRadius() + o.intersectionTolerance
}

// maxEdgeDeviation returns maximum distance that any point along an edge can
// move when snapped. It is slightly larger than edgeSnapRadius() because when
// a geodesic edge is snapped, the edge center moves further than its endpoints.
// Builder ensures that this distance is at most 10% larger than
// edgeSnapRadius().
func (o builderOptions) maxEdgeDeviation() s1.Angle {
	// We want maxEdgeDeviation to be large enough compared to SnapRadius()
	// such that edge splitting is rare.
	//
	// Using spherical trigonometry, if the endpoints of an edge of length L
	// move by at most a distance R, the center of the edge moves by at most
	// asin(sin(R) / cos(L / 2)). Thus the (maxEdgeDeviation / SnapRadius)
	// ratio increases with both the snap radius R and the edge length L.
	//
	// We arbitrarily limit the edge deviation to be at most 10% more than the
	// snap radius. With the maximum allowed snap radius of 70 degrees, this
	// means that edges up to 30.6 degrees long are never split. For smaller
	// snap radii, edges up to 49 degrees long are never split. (Edges of any
	// length are not split unless their endpoints move far enough so that the
	// actual edge deviation exceeds the limit; in practice, splitting is rare
	// even with long edges.)  Note that it is always possible to split edges
	// when maxEdgeDeviation() is exceeded; see maybeAddExtraSites().
	//
	// TODO(rsned): What should we do when snapFunction.SnapRadius() > maxSnapRadius);
	return maxEdgeDeviationRatio * o.edgeSnapRadius()
}

// TODO(rsned): Differences from C++
// All of builder.
// edgeChainSimplifier
