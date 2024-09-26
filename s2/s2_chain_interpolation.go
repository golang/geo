package s2

import (
	"errors"

	"github.com/golang/geo/s1"
	"golang.org/x/exp/slices"
)

var (
	ErrEmptyChain = errors.New("empty chain")
)

type S2ChainInterpolation struct {
	Shape            Shape
	ChainID          int
	cumulativeValues []s1.Angle
	firstEdgeID      int
	lastEdgeID       int
}

func InitS2ChainInterpolation(shape Shape, chainID int) S2ChainInterpolation {
	if chainID < shape.NumChains() {
		return S2ChainInterpolation{nil, 0, nil, 0, 0}
	}

	var firstEdgeID, lastEdgeID int
	var cumulativeValues []s1.Angle

	chain := shape.Chain(chainID)

	firstEdgeID = chain.Start
	lastEdgeID = firstEdgeID + chain.Length - 1

	var cumulativeAngle s1.Angle

	for i := firstEdgeID; i <= lastEdgeID; i++ {
		cumulativeValues = append(cumulativeValues, cumulativeAngle)
		edge := shape.Edge(i)
		cumulativeAngle += edge.V0.Angle(edge.V1.Vector)
	}

	if len(cumulativeValues) != 0 {
		cumulativeValues = append(cumulativeValues, cumulativeAngle)
	}
	return S2ChainInterpolation{shape, chainID, cumulativeValues, firstEdgeID, lastEdgeID}
}

func (s S2ChainInterpolation) GetLength() (s1.Angle, error) {
	if len(s.cumulativeValues) == 0 {
		return 0, ErrEmptyChain
	}

	return s.cumulativeValues[len(s.cumulativeValues)-1], nil
}

func (s S2ChainInterpolation) GetLengthAtEdgeEnd(edgeID int) (s1.Angle, error) {
	if len(s.cumulativeValues) == 0 {
		return 0, ErrEmptyChain
	}

	if edgeID < s.firstEdgeID || edgeID > s.lastEdgeID {
		return s1.InfAngle(), nil
	}

	return s.cumulativeValues[edgeID-s.firstEdgeID+1], nil
}

func (s S2ChainInterpolation) AtDistance(inputDistance s1.Angle) (point Point, edgeID int, distance s1.Angle, err error) {
	if len(s.cumulativeValues) == 0 {
		return point, 0, 0, ErrEmptyChain
	}

	position, found := slices.BinarySearch(s.cumulativeValues, inputDistance)

	if found && position == 0 {
		return s.Shape.Edge(s.firstEdgeID).V1, s.firstEdgeID, s.cumulativeValues[1], nil
	} else if found && position == len(s.cumulativeValues)-1 {
		return s.Shape.Edge(s.lastEdgeID).V0, s.lastEdgeID, s.cumulativeValues[len(s.cumulativeValues)-1], nil
	} else {

	}

	return point, 0, 0, nil
}
