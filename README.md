# Overview

S2 is a library for spherical geometry that aims to have the same robustness,
flexibility, and performance as the best planar geometry libraries.

This is a library for manipulating geometric shapes. Unlike many geometry
libraries, S2 is primarily designed to work with _spherical geometry_, i.e.,
shapes drawn on a sphere rather than on a planar 2D map. (In fact, the name S2
is derived from the mathematical notation for the unit sphere *S²*.) This makes
it especially suitable for working with geographic data.

More details about S2 in general are available on the S2 Geometry Website
[s2geometry.io](https://s2geometry.io/).

## Scope

The library provides the following:

*   Representations of angles, intervals, latitude-longitude points, unit
    vectors, and so on, and various operations on these types.

*   Geometric shapes over the unit sphere, such as spherical caps ("discs"),
    latitude-longitude rectangles, polylines, and polygons. These are
    collectively known as "regions".

*   A hierarchical decomposition of the sphere into regions called "cells". The
    hierarchy starts with the six faces of a projected cube and recursively
    subdivides them in a quadtree-like fashion.

*   Robust constructive operations (e.g., union) and boolean predicates (e.g.,
    containment) for arbitrary collections of points, polylines, and polygons.

*   Fast in-memory indexing of collections of points, polylines, and polygons.

*   Algorithms for measuring distances and finding nearby objects.

*   Robust algorithms for snapping and simplifying geometry (with accuracy and
    topology guarantees).

*   A collection of efficient yet exact mathematical predicates for testing
    relationships among geometric objects.

*   Support for spatial indexing, including the ability to approximate regions
    as collections of discrete "S2 cells". This feature makes it easy to build
    large distributed spatial indexes.

On the other hand, the following are outside the scope of S2:

*   Planar geometry.

*   Conversions to/from common GIS formats.

### Robustness

What do we mean by "robust"?

In the S2 library, the core operations are designed to be 100% robust. This
means that each operation makes strict mathematical guarantees about its output,
and is implemented in such a way that it meets those guarantees for all possible
valid inputs. For example, if you compute the intersection of two polygons, not
only is the output guaranteed to be topologically correct (up to the creation of
degeneracies), but it is also guaranteed that the boundary of the output stays
within a user-specified tolerance of true, mathematically exact result.

Robustness is very important when building higher-level algorithms, since
unexpected results from low-level operations can be very difficult to handle. S2
achieves this goal using a combination of techniques from computational
geometry, including *conservative error bounds*, *exact geometric predicates*,
and *snap rounding*.

The implementation attempts to be precise both in terms of mathematical
definitions (e.g. whether regions include their boundaries, and how degeneracies
are handled) and numerical accuracy (e.g. minimizing cancellation error).

Note that the intent of this library is to represent geometry as a mathematical
abstraction. For example, although the unit sphere is obviously a useful
approximation for the Earth's surface, functions that are specifically related
to geography are not part of the core library (e.g. easting/northing
conversions, ellipsoid approximations, geodetic vs. geocentric coordinates,
etc).

See http://godoc.org/github.com/golang/geo for specific package documentation.

For an analogous library in C++, see https://github.com/google/s2geometry, in
Java, see https://github.com/google/s2-geometry-library-java, and Python, see
https://github.com/google/s2geometry/tree/master/src/python

# Status of the Go Library

This library is principally a port of the
[C++ S2 library](https://github.com/google/s2geometry), adapting to Go idioms
where it makes sense. We detail the progress of this port below relative to that
C++ library.

## [ℝ¹](https://godoc.org/github.com/golang/geo/r1) - One-dimensional Cartesian coordinates

Full parity with C++.

## [ℝ²](https://godoc.org/github.com/golang/geo/r2) - Two-dimensional Cartesian coordinates

Full parity with C++.

## [ℝ³](https://godoc.org/github.com/golang/geo/r3) - Three-dimensional Cartesian coordinates

Full parity with C++.

## [S¹](https://godoc.org/github.com/golang/geo/s1) - Circular Geometry

Full parity with C++.

## [S²](https://godoc.org/github.com/golang/geo/s2) - Spherical Geometry

Approximately ~40% complete.

**Complete** These files have full parity with the C++ implementation.

*   Cap
*   Cell
*   CellID
*   CellUnion
*   ContainsVertexQuery
*   ConvexHullQuery
*   CrossingEdgeQuery
*   LatLng
*   matrix3x3
*   Metric
*   PaddedCell
*   Point
*   PointCompression
*   Region
*   RegionCoverer
*   s2edge_clipping
*   s2edge_crosser
*   s2edge_crossings
*   s2edge_distances
*   edgeVectorShape
*   laxLoop
*   laxPolyline
*   s2projections - Helpers for projecting points between R2 and S2.
*   s2rect_bounder
*   s2stuv.go (s2coords.h in C++) - This file is a collection of helper and
    conversion methods to and from ST-space, UV-space, and XYZ-space.
*   s2wedge_relations
*   ShapeIndex
*   idSetLexicon,sequenceLexicon

**Mostly Complete** Files that have almost all of the features of the original
C++ code, and are reasonably complete enough to use in live code. Up to date
listing of the incomplete methods are documented at the end of each file.

*   EdgeQuery/Closest/Furthest - missing Project, GetEdge
*   ContainsPointQuery - missing visit edges
*   laxPolygon
*   Loop - Loop is mostly complete now. Missing Project, Distance, Union, etc.
*   Polyline - Missing InitTo... methods, NearlyCoversPolyline
*   Rect (AKA s2latlngrect in C++) - Missing Centroid, InteriorContains.
*   s2_test.go (AKA s2testing and s2textformat in C++) - Missing Fractal test
    shape generation. This file is a collection of testing helper methods.
*   s2edge_distances - Missing Intersection

**In Progress** Files that have some work done, but are probably not complete
enough for general use in production code.

*   CellIndex - A queryable index of CellIDs.
*   Polygon - Polygons with multiple loops are supported. It fully implements
    Shape and Region, but it's missing most other methods. (Area, Centroid,
    Distance, Projection, Intersection, Union, Contains, Normalized, etc.)
*   PolylineSimplifier - Initial work has begun on this.
*   s2predicates.go - This file is a collection of helper methods used by other
    parts of the library.
*   s2shapeutil - Initial elements added. Missing VisitCrossings.

**Not Started Yet.** These files (and their associated unit tests) have
dependencies on most of the In Progress files before they can begin to be
started.

*   BooleanOperation - used when assembling polygons and loops.
*   Builder - This is a robust tool for creating the various Shape types from
    collection of simpler S2 types.
*   BuilderClosedSetNormalizer
*   BuilderFindPolygonDegneracies
*   BuilderGraph
*   BuilderLayers
*   BuilderSnapFunctions
*   BuilderTesting
*   Centroids
*   ClosestPointQuery
*   EdgeTesselator
*   LoopMeasures
*   PointIndex
*   PointRegion
*   PointUtil
*   PolygonMeasures
*   RegionIntersection
*   RegionTermIndexer
*   RegionUnion
*   ShapeIndexRegion - Allows ShapeIndexes to be used as Regions for things like

### Encode/Decode

Encoding and decoding of S2 types is fully implemented and interoperable with
C++ and Java.
