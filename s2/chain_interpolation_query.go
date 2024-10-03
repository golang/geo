package s2

import (
	"errors"
	"slices"

	"github.com/golang/geo/s1"
)

var (
	ErrEmptyChain = errors.New("empty chain")
)

type ChainInterpolationQuery struct {
	Shape            Shape
	ChainID          int
	cumulativeValues []s1.Angle
	firstEdgeID      int
	lastEdgeID       int
}

func InitS2ChainInterpolationQuery(shape Shape, chainID int) ChainInterpolationQuery {
	if shape == nil || chainID >= shape.NumChains() {
		return ChainInterpolationQuery{nil, 0, nil, 0, 0}
	}

	var firstEdgeID, lastEdgeID int
	var cumulativeValues []s1.Angle

	if chainID >= 0 {
		chain := shape.Chain(chainID)
		firstEdgeID = chain.Start
		lastEdgeID = firstEdgeID + chain.Length - 1
	} else {
		firstEdgeID = 0
		lastEdgeID = shape.NumEdges() - 1
	}

	var cumulativeAngle s1.Angle

	for i := firstEdgeID; i <= lastEdgeID; i++ {
		cumulativeValues = append(cumulativeValues, cumulativeAngle)
		edge := shape.Edge(i)
		cumulativeAngle += edge.V0.Angle(edge.V1.Vector)
	}

	if len(cumulativeValues) != 0 {
		cumulativeValues = append(cumulativeValues, cumulativeAngle)
	}
	return ChainInterpolationQuery{shape, chainID, cumulativeValues, firstEdgeID, lastEdgeID}
}

func (s ChainInterpolationQuery) GetLength() (s1.Angle, error) {
	if len(s.cumulativeValues) == 0 {
		return 0, ErrEmptyChain
	}

	return s.cumulativeValues[len(s.cumulativeValues)-1], nil
}

func (s ChainInterpolationQuery) GetLengthAtEdgeEnd(edgeID int) (s1.Angle, error) {
	if len(s.cumulativeValues) == 0 {
		return 0, ErrEmptyChain
	}

	if edgeID < s.firstEdgeID || edgeID > s.lastEdgeID {
		return s1.InfAngle(), nil
	}

	return s.cumulativeValues[edgeID-s.firstEdgeID+1], nil
}

func (s ChainInterpolationQuery) AtDistance(inputDistance s1.Angle) (point Point, edgeID int, distance s1.Angle, err error) {
	if len(s.cumulativeValues) == 0 {
		return point, 0, 0, ErrEmptyChain
	}

	position, found := slices.BinarySearch(s.cumulativeValues, inputDistance)

	if position <= 0 {
		return s.Shape.Edge(s.firstEdgeID).V0, s.firstEdgeID, s.cumulativeValues[0], nil
	} else if (found && position == len(s.cumulativeValues)-1) || (!found && position >= len(s.cumulativeValues)) {
		return s.Shape.Edge(s.lastEdgeID).V1, s.lastEdgeID, s.cumulativeValues[len(s.cumulativeValues)-1], nil
	} else {
		edgeID = max(position+s.firstEdgeID-1, 0)
		edge := s.Shape.Edge(edgeID)
		distance = inputDistance - s.cumulativeValues[max(0, position-1)]
		point = GetPointOnLine(edge.V0, edge.V1, distance)
	}

	return point, edgeID, distance, nil
}

func (s ChainInterpolationQuery) AtFraction(inputFraction float64) (point Point, edgeID int, distance s1.Angle, err error) {
	length, error := s.GetLength()
	if error != nil {
		return point, 0, 0, error
	}

	return s.AtDistance(s1.Angle(inputFraction * float64(length)))
}

func (s ChainInterpolationQuery) Slice(beginFraction, endFraction float64) []Point {
	var points []Point
	s.AddSlice(beginFraction, endFraction, &points)
	return points
}

func (s ChainInterpolationQuery) AddSlice(beginFraction, endFraction float64, points *[]Point) {
	if len(s.cumulativeValues) == 0 {
		return
	}

	reverse := beginFraction > endFraction
	if reverse {
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

	for edgeID := beginEdgeID; edgeID < endEdgeID; edgeID++ {
		edge := s.Shape.Edge(edgeID)
		if lastPoint != edge.V1 {
			lastPoint = edge.V1
			*points = append(*points, lastPoint)
		}
	}
	*points = append(*points, atEnd)

	if reverse {
		slices.Reverse(*points)
	}
}
