# S2 geometry library in Go

[![Go Build and Test](https://github.com/golang/geo/actions/workflows/go.yml/badge.svg)](https://github.com/golang/geo/actions/workflows/go.yml) [![CodeQL](https://github.com/golang/geo/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/golang/geo/actions/workflows/github-code-scanning/codeql) [![golangci-lint](https://github.com/golang/geo/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/golang/geo/actions/workflows/golangci-lint.yml) [![OpenSSF Scorecard](https://img.shields.io/ossf-scorecard/github.com/golang/geo?label=OpenSSF%20Scorecard&style=flat)](https://scorecard.dev/viewer/?uri=github.com/golang/geo)

S2 is a library for spherical geometry that aims to have the same robustness,
flexibility, and performance as the best planar geometry libraries.

This is a library for manipulating geometric shapes. Unlike many geometry
libraries, S2 is primarily designed to work with *spherical geometry*, i.e.,
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

See https://pkg.go.dev/github.com/golang/geo for specific package documentation.

For an analogous library in C++, see https://github.com/google/s2geometry, in
Java, see https://github.com/google/s2-geometry-library-java, and Python, see
https://github.com/google/s2geometry/tree/master/src/python

# Status of the Go Library

This library is principally a port of the
[C++ S2 library](https://github.com/google/s2geometry), adapting to Go idioms
where it makes sense. We detail the progress of this port below relative to that
C++ library.

Legend:

*   ✅ - Feature Complete
*   🟡 - Mostly Complete
*   ❌ - Not available

## [ℝ¹](https://pkg.go.dev/github.com/golang/geo/r1) - One-dimensional Cartesian coordinates

C++ Type   | Go
:--------- | ---
R1Interval | ✅

## [ℝ²](https://pkg.go.dev/github.com/golang/geo/r2) - Two-dimensional Cartesian coordinates

C++ Type | Go
:------- | ---
R2Point  | ✅
R2Rect   | ✅

## [ℝ³](https://pkg.go.dev/github.com/golang/geo/r3) - Three-dimensional Cartesian coordinates

C++ Type      | Go
:------------ | ---
R3Vector      | ✅
R3ExactVector | ✅
Matrix3x3     | ✅

## [S¹](https://pkg.go.dev/github.com/golang/geo/s1) - Circular Geometry

C++ Type     | Go
:----------- | ---
S1Angle      | ✅
S1ChordAngle | ✅
S1Interval   | ✅

## [S²](https://pkg.go.dev/github.com/golang/geo/s2) - Spherical Geometry

### Basic Types

C++ Type             | Go
:------------------- | ---
S2Cap                | ✅
S2Cell               | ✅
S2CellId             | ✅
S2CellIdVector       | ❌
S2CellIndex          | 🟡
S2CellUnion          | ✅
S2Coords             | ✅
S2DensityTree        | ❌
S2DistanceTarget     | ✅
S2EdgeVector         | ✅
S2LatLng             | ✅
S2LatLngRect         | ✅
S2LaxLoop            | 🟡
S2LaxPolygon         | 🟡
S2LaxPolyline        | 🟡
S2Loop               | ✅
S2PaddedCell         | ✅
S2Point              | ✅
S2PointIndex         | ❌
S2PointSpan          | ❌
S2PointRegion        | ❌
S2PointVector        | ✅
S2Polygon            | 🟡
S2Polyline           | ✅
S2R2Rect             | ❌
S2Region             | ✅
S2RegionCoverer      | ✅
S2RegionIntersection | ❌
S2RegionUnion        | ✅
S2Shape              | ✅
S2ShapeIndex         | ✅
S2ShapeIndexRegion   | ❌
EncodedLaxPolygon    | ❌
EncodedLaxPolyline   | ❌
EncodedShapeIndex    | ❌
EncodedStringVector  | ❌
EncodedUintVector    | ❌
IdSetLexicon         | ❌
ValueSetLexicon      | ❌
SequenceLexicon      | ❌
LaxClosedPolyline    | ❌
VertexIDLaxLoop      | ❌

### Query Types

C++ Type             | Go
:------------------- | ---
S2ChainInterpolation | ✅
S2ClosestCell        | ❌
S2FurthestCell       | ❌
S2ClosestEdge        | ✅
S2FurthestEdge       | ✅
S2ClosestPoint       | ❌
S2FurthestPoint      | ❌
S2ContainsPoint      | ✅
S2ContainsVertex     | ✅
S2ConvexHull         | ✅
S2CrossingEdge       | ✅
S2HausdorffDistance  | ❌
S2ShapeNesting       | ❌
S2ValidationQuery    | ❌

### Supporting Types

C++ Type                         | Go
:------------------------------- | ---
S2BooleanOperation               | ❌
S2BufferOperation                | ❌
S2Builder                        | ❌
S2BuilderGraph                   | ❌
S2BuilderLayer                   | ❌
S2BuilderUtil_\*                 | ❌
S2CellIterator                   | ❌
S2CellIteratorJoin               | ❌
S2CellRangeIterator              | ❌
S2Coder                          | ❌
S2Earth                          | ❌
S2EdgeClipping                   | ✅
S2EdgeCrosser                    | ✅
S2EdgeCrossings                  | ✅
S2EdgeDistances                  | ✅
S2EdgeTessellator                | ✅
S2Fractal                        | ❌
S2LoopMeasures                   | ❌
S2Measures                       | ✅
S2MemoryTracker                  | ❌
S2Metrics                        | ❌
S2PointUtil                      | 🟡
S2PointCompression               | 🟡
S2PolygonBuilder                 | ❌
S2PolylineAlignment              | ❌
S2PolylineMeasures               | ✅
S2PolylineSimplifier             | ❌
S2Predicates                     | ✅
S2Projections                    | ❌
S2Random                         | ❌
S2RectBounder                    | ❌
S2RegionSharder                  | ❌
S2RegionTermIndexer              | ❌
S2ShapeIndexBufferedRegion       | ❌
S2ShapeIndexMeasures             | ❌
S2ShapeIndexUtil\*               | 🟡
S2ShapeMeasures                  | ❌
S2ShapeUtil\*                    | 🟡
S2Stats                          | ❌
S2Testing                        | ✅
S2TextFormat                     | ✅
S2WedgeRelations                 | ✅
S2WindingOperation               | ❌


### Encode/Decode

Encoding and decoding of S2 types is fully implemented and interoperable with
C++ and Java.


## Disclaimer

This is not an official Google product.
