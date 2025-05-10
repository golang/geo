package s2

import "github.com/golang/geo/s1"

type builderOptions struct {
	// snapFunction holds the desired snap function.
	//
	// Note that if your input data includes vertices that were created using
	// Intersection(), then you should use a "snapRadius" of
	// at least intersectionMergeRadius, e.g. by calling
	//
	//  options.setSnapFunction(IdentitySnapFunction(intersectionMergeRadius));
	//
	// DEFAULT: IdentitySnapFunction(s1.Angle(0))
	// [This does no snapping and preserves all input vertices exactly.]
	snapFunction Snapper

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
	// DEFAULT: false
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
	// DEFAULT: s1.Angle(0)
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
	// DEFAULT: false
	simplifyEdgeChains bool

	// idempotent determines if snapping occurs only when the input geometry
	// does not already meet the Builder output guarantees (see the Snapper
	// type description for details). This means that if all input vertices
	// are at snapped locations, all vertex pairs are separated by at least
	// MinVertexSeparation(), and all edge-vertex pairs are separated by at
	// least MinEdgeVertexSeparation(), then no snapping is done.
	//
	// If false, then all vertex pairs and edge-vertex pairs closer than
	// "SnapRadius" will be considered for snapping. This can be useful, for
	// example, if you know that your geometry contains errors and you want to
	// make sure that features closer together than "SnapRadius" are merged.
	//
	// This option is automatically turned off when simplifyEdgeChains is true
	// since simplifying edge chains is never guaranteed to be idempotent.
	//
	// DEFAULT: true
	idempotent bool
}

// defaultBuilderOptions returns a new instance with the proper defaults.
func defaultBuilderOptions() *builderOptions {
	return &builderOptions{
		snapFunction:          NewIdentitySnapper(0),
		splitCrossingEdges:    false,
		intersectionTolerance: s1.Angle(0),
		simplifyEdgeChains:    false,
		idempotent:            true,
	}
}

// edgeSnapRadius reports the maximum distance from snapped edge vertices to
// the original edge. This is the same as SnapFunction().SnapRadius() except
// when splitCrossingEdges is true (see below), in which case the edge snap
// radius is increased by intersectionError.
func (o builderOptions) edgeSnapRadius() s1.Angle {
	return o.snapFunction.SnapRadius() + o.intersectionTolerance
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
