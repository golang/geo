package s2

import (
	"errors"
	"slices"

	"github.com/pavlov061356/geo/s1"
)

var (
	// ErrEmptyChain is returned by ChainInterpolationQuery when the query
	// contains no edges.
	ErrEmptyChain = errors.New("empty chain")

	// ErrInvalidDivisionsCount is returned by ChainInterpolationQuery when
	// divisionsCount is less than the number of edges in the shape.
	ErrInvalidDivisionsCount = errors.New("invalid divisions count")

	// ErrInvalidIndexes is returned by ChainInterpolationQuery when
	// start or end indexes are invalid.
	ErrInvalidIndexes = errors.New("invalid indexes")
)

// ChainInterpolationQuery is a helper struct for querying points on Shape's
// edges by spherical distance.  The distance is computed cumulatively along the
// edges contained in the shape, using the order in which the edges are stored
// by the Shape object.
type ChainInterpolationQuery struct {
	Shape            Shape
	ChainID          int
	cumulativeValues []s1.Angle
	firstEdgeID      int
	lastEdgeID       int
}

// InitChainInterpolationQuery initializes and conctructs a ChainInterpolationQuery.
// If a particular chain id is specified at the query initialization, then the
// distance values are computed along that single chain, which allows per-chain
// interpolation.  If no chain is specified, then the interpolated point as a
// function of distance is discontinuous at chain boundaries.  Using multiple
// chains can be used in such algorithms as quasi-random sampling along the
// total span of a multiline.
//
// Once the query object is initialized, the complexity of each subsequent query
// is O( log(number of edges) ).  The complexity of the initialization and the
// memory footprint of the query object are both O(number of edges).
func InitChainInterpolationQuery(shape Shape, chainID int) ChainInterpolationQuery {
	if shape == nil || chainID >= shape.NumChains() {
		return ChainInterpolationQuery{nil, 0, nil, 0, 0}
	}

	var firstEdgeID, lastEdgeID int
	var cumulativeValues []s1.Angle

	if chainID >= 0 {
		// If a valid chain id was provided, then the range of edge ids is defined
		// by the start and the length of the chain.
		chain := shape.Chain(chainID)
		firstEdgeID = chain.Start
		lastEdgeID = firstEdgeID + chain.Length - 1
	} else {
		// If no valid chain id was provided then we use the whole range of shape's
		// edge ids.
		firstEdgeID = 0
		lastEdgeID = shape.NumEdges() - 1
	}

	var cumulativeAngle s1.Angle

	for i := firstEdgeID; i <= lastEdgeID; i++ {
		cumulativeValues = append(cumulativeValues, cumulativeAngle)
		edge := shape.Edge(i)
		edgeAngle := edge.V0.Angle(edge.V1.Vector)
		cumulativeAngle += edgeAngle
	}

	if len(cumulativeValues) != 0 {
		cumulativeValues = append(cumulativeValues, cumulativeAngle)
	}
	return ChainInterpolationQuery{shape, chainID, cumulativeValues, firstEdgeID, lastEdgeID}
}

// Gets the total length of the chain(s), which corresponds to the distance at
// the end vertex of the last edge of the chain(s).  Returns zero length for
// shapes containing no edges.
func (s ChainInterpolationQuery) GetLength() (s1.Angle, error) {
	// The total length equals the cumulative value at the end of the last
	// edge, if there is at least one edge in the shape.
	if len(s.cumulativeValues) == 0 {
		return 0, ErrEmptyChain
	}
	return s.cumulativeValues[len(s.cumulativeValues)-1], nil
}

// Returns the cumulative length along the edges being interpolated up to the
// end of the given edge ID. Returns s1.InfAngle() if the edge
// ID does not lie within the set of edges being interpolated. Returns
// ErrEmptyChain if the ChainInterpolationQuery is empty.
func (s ChainInterpolationQuery) GetLengthAtEdgeEnd(edgeID int) (s1.Angle, error) {
	if len(s.cumulativeValues) == 0 {
		return 0, ErrEmptyChain
	}

	if edgeID < s.firstEdgeID || edgeID > s.lastEdgeID {
		return s1.InfAngle(), nil
	}

	return s.cumulativeValues[edgeID-s.firstEdgeID+1], nil
}

// Computes the Point located at the given distance along the edges from the
// first vertex of the first edge. Also computes the edge id and the actual
// distance corresponding to the resulting point.
//
// This method returns a valid result if the query has been initialized with
// at least one edge.
//
// If the input distance exceeds the total length, then the resulting point is
// the end vertex of the last edge, and the resulting distance is set to the
// total length.
//
// If there are one or more degenerate (zero-length) edges corresponding to
// the given distance, then the resulting point is located on the first of
// these edges.
func (s ChainInterpolationQuery) AtDistance(inputDistance s1.Angle) (point Point, edgeID int, distance s1.Angle, err error) {
	if len(s.cumulativeValues) == 0 {
		return point, 0, 0, ErrEmptyChain
	}

	distance = inputDistance

	position, found := slices.BinarySearch(s.cumulativeValues, inputDistance)

	if position <= 0 {
		// Corner case: the first vertex of the shape at distance = 0.
		return s.Shape.Edge(s.firstEdgeID).V0, s.firstEdgeID, s.cumulativeValues[0], nil
	} else if (found && position == len(s.cumulativeValues)-1) || (!found && position >= len(s.cumulativeValues)) {
		// Corner case: the input distance is greater than the total length, hence
		// we snap the result to the last vertex of the shape at distance = total
		// length.
		return s.Shape.Edge(s.lastEdgeID).V1, s.lastEdgeID, s.cumulativeValues[len(s.cumulativeValues)-1], nil
	} else {
		// Obtain the edge index and compute the interpolated result from the edge
		// vertices.
		edgeID = max(position+s.firstEdgeID-1, 0)
		edge := s.Shape.Edge(edgeID)
		point = GetPointOnLine(edge.V0, edge.V1, inputDistance-s.cumulativeValues[max(0, position-1)])
	}

	return point, edgeID, distance, nil
}

// Similar to the above function, but takes the normalized fraction of the
// distance as input, with inputFraction = 0 corresponding to the beginning of the
// shape or chain and inputFraction = 1 to the end.  Forwards the call to
// AtDistance().  A small precision loss may occur due to converting the
// fraction to distance by multiplying it by the total length.
func (s ChainInterpolationQuery) AtFraction(inputFraction float64) (point Point, edgeID int, distance s1.Angle, err error) {
	length, error := s.GetLength()
	if error != nil {
		return point, 0, 0, error
	}

	return s.AtDistance(s1.Angle(inputFraction * float64(length)))
}

// Returns the vector of points that is a slice of the chain from
// beginFraction to endFraction. If beginFraction is greater than
// endFraction, then the points are returned in reverse order.
//
// For example, Slice(0,1) returns the entire chain, Slice(0, 0.5) returns the
// first half of the chain, and Slice(1, 0.5) returns the second half of the
// chain in reverse.
//
// The endpoints of the slice are interpolated (except when coinciding with an
// existing vertex of the chain), and all the internal points are copied from
// the chain as is.
//
// If the query is either uninitialized, or initialized with a shape
// containing no edges, then an empty vector is returned.
func (s ChainInterpolationQuery) Slice(beginFraction, endFraction float64) []Point {
	var points []Point
	s.AddSlice(beginFraction, endFraction, &points)
	return points
}

// Returns the vector of points that is a slice of the chain from
// beginFraction to endFraction. If beginFraction is greater than
// endFraction, then the points are returned in reverse order.
//
// For example, Slice(0,1) returns the entire chain, Slice(0, 0.5) returns the
// first half of the chain, and Slice(1, 0.5) returns the second half of the
// chain in reverse.
//
// The endpoints of the slice are interpolated (except when coinciding with an
// existing vertex of the chain), and all the internal points are copied from
// the chain as is.
//
// divisions is the number of segments to divide the polyline into.
// divisions must be >= len(Slice(beginFraction, endFraction)).
//
// If the query is either uninitialized, or initialized with a shape
// containing no edges, then an empty vector is returned.
func (s ChainInterpolationQuery) SliceDivided(beginFraction, endFraction float64, divisions int) []Point {
	var points []Point
	s.AddDividedSlice(beginFraction, endFraction, &points, divisions)
	return points
}

// Appends the chain slice from beginFraction to endFraction to the given
// slice. If beginFraction is greater than endFraction, then the points are
// appended in reverse order. If the query is either uninitialized, or
// initialized with a shape containing no edges, then no points are appended.
func (s ChainInterpolationQuery) AddSlice(beginFraction, endFraction float64, points *[]Point) {
	if len(s.cumulativeValues) == 0 {
		return
	}

	reverse := beginFraction > endFraction
	if reverse {
		// Swap the begin and end fractions so that we can iterate in ascending order.
		beginFraction, endFraction = endFraction, beginFraction
	}

	atBegin, beginEdgeID, _, err := s.AtFraction(beginFraction)
	if err != nil {
		return
	}
	*points = append(*points, atBegin)
	lastPoint := atBegin

	atEnd, endEdgeID, _, err := s.AtFraction(endFraction)
	if err != nil {
		return
	}

	// Copy the internal points from the chain.
	for edgeID := beginEdgeID; edgeID < endEdgeID; edgeID++ {
		edge := s.Shape.Edge(edgeID)
		if lastPoint != edge.V1 {
			lastPoint = edge.V1
			*points = append(*points, lastPoint)
		}
	}
	*points = append(*points, atEnd)

	// Reverse the slice if necessary.
	if reverse {
		slices.Reverse(*points)
	}
}

// Appends the slice from beginFraction to endFraction to the given
// slice. If beginFraction is greater than endFraction, then the points are
// appended in reverse order. If the query is either uninitialized, or
// initialized with a shape containing no edges, then no points are appended.
// divisions is the number of segments to divide the polyline into.
// divisions must be greater or equal of NumEdges of Shape.
// A polyline is divided into segments of equal length, and then edges are added to the slice.
func (s ChainInterpolationQuery) AddDividedSlice(beginFraction, endFraction float64, points *[]Point, pointsNum int) {
	if len(s.cumulativeValues) == 0 {
		return
	}

	slice := s.Slice(beginFraction, endFraction)

	if len(slice) > pointsNum {
		return
	} else if len(slice) == pointsNum {
		*points = append(*points, slice...)
		return
	}

	reverse := beginFraction > endFraction
	if reverse {
		// Swap the begin and end fractions so that we can iterate in ascending order.
		beginFraction, endFraction = endFraction, beginFraction
	}

	atBegin, currentEdgeID, _, err := s.AtFraction(beginFraction)
	if err != nil {
		return
	}

	atEnd, _, _, err := s.AtFraction(endFraction)
	if err != nil {
		return
	}

	// divisionsExcludingEdges := pointsNum - len(slice)

	*points = append(*points, atBegin)

	// // Copy the internal points from the chain.
	for fraction := beginFraction + (endFraction-beginFraction)/float64(pointsNum-1); fraction < endFraction; fraction += (endFraction - beginFraction) / float64(pointsNum-1) {
		atFraction, edgeID, _, err := s.AtFraction(fraction)
		if err != nil {
			return
		}

		// If the current edge is the same as the previous edge, then skip it.
		// Otherwise, append all edges in between.
		if currentEdgeID != edgeID {
			for i := currentEdgeID; i < edgeID; i++ {
				edge := s.Shape.Edge(i)
				if edge.V1 != atFraction {
					if len(*points) == pointsNum-1 {
						break
					}
					*points = append(*points, edge.V1)
				}
			}
			currentEdgeID = edgeID
			continue
		} else if edge := s.Shape.Edge(edgeID); edge.V1.approxEqual(atFraction, epsilon) {
			*points = append(*points, edge.V1)
			currentEdgeID++
			continue
		}
		if len(*points) == pointsNum-1 {
			break
		}
		*points = append(*points, atFraction)
	}
	// Append last edge
	*points = append(*points, atEnd)

	// Reverse the slice if necessary.
	if reverse {
		slices.Reverse(*points)
	}
}
