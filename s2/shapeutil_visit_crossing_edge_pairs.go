package s2

import "fmt"

type ShapeEdgeVector []ShapeEdge

type EdgePairVisitor func(a, b ShapeEdge, isInterior bool) bool

// getShapeEdges returns all edges in the given S2ShapeIndexCell.
func getShapeEdges(index *ShapeIndex, cell *ShapeIndexCell) ShapeEdgeVector {
	var shapeEdges ShapeEdgeVector
	for _, clipped := range cell.shapes {
		shape := index.Shape(clipped.shapeID)
		for _, edgeID := range clipped.edges {
			shapeEdges = append(shapeEdges, ShapeEdge{
				ID:   ShapeEdgeID{
					ShapeID: clipped.shapeID,
					EdgeID: int32(edgeID),
				},
				Edge: shape.Edge(edgeID),
			})
		}
	}
	return shapeEdges
}

// VisitCrossings finds and processes all crossing edge pairs.
func visitCrossings(shapeEdges ShapeEdgeVector, crossingType CrossingType, needAdjacent bool, visitor EdgePairVisitor) bool {
	minCrossingSign := MaybeCross
	if crossingType == CrossingTypeInterior {
		minCrossingSign = Cross
	}
	for i := 0; i < len(shapeEdges) - 1; i++ {
		a := shapeEdges[i]
		j := i + 1
		// A common situation is that an edge AB is followed by an edge BC.  We
		// only need to visit such crossings if "needAdjacent" is true (even if
		// AB and BC belong to different edge chains).
		if !needAdjacent && a.Edge.V1 == shapeEdges[j].Edge.V0 {
			j++
			if j >= len(shapeEdges) {
				break
			}
		}
		crosser := NewEdgeCrosser(a.Edge.V0, a.Edge.V1)
		for ; j < len(shapeEdges); j++ {
			b := shapeEdges[j]
			if crosser.c != b.Edge.V0 {
				crosser.RestartAt(b.Edge.V0)
			}
			sign := crosser.ChainCrossingSign(b.Edge.V1)
			// missinglink: enum ordering is reversed compared to C++
			if sign <= minCrossingSign {
				if !visitor(a, b, sign == Cross) {
					return false
				}
			}
		}
	}
	return true
}

// Visits all pairs of crossing edges in the given S2ShapeIndex, terminating
// early if the given EdgePairVisitor function returns false (in which case
// VisitCrossings returns false as well).  "type" indicates whether all
// crossings should be visited, or only interior crossings.
//
// If "needAdjacent" is false, then edge pairs of the form (AB, BC) may
// optionally be ignored (even if the two edges belong to different edge
// chains).  This option exists for the benefit of FindSelfIntersection(),
// which does not need such edge pairs (see below).
func VisitCrossings(index *ShapeIndex, crossingType CrossingType, needAdjacent bool, visitor EdgePairVisitor) bool {
	// TODO(b/262264880): Use brute force if the total number of edges is small
	// enough (using a larger threshold if the S2ShapeIndex is not constructed
	// yet).
	for it := index.Iterator(); !it.Done(); it.Next() {
		shapeEdges := getShapeEdges(index, it.cell)
		if !visitCrossings(shapeEdges, crossingType, needAdjacent, visitor) {
			return false
		}
	}
	return true
}

// VisitCrossingEdgePairs finds all crossing edge pairs in an index.
func VisitCrossingEdgePairs(index *ShapeIndex, crossingType CrossingType, visitor EdgePairVisitor) bool {
	needAdjacent := crossingType == CrossingTypeAll
	for it := index.Iterator(); !it.Done(); it.Next() {
		shapeEdges := getShapeEdges(index, it.cell)
		if !visitCrossings(shapeEdges, crossingType, needAdjacent, visitor) {
			return false
		}
	}
	return true
}

func FindCrossingError(shape Shape, a, b ShapeEdge, isInterior bool) error {
	ap := shape.ChainPosition(int(a.ID.EdgeID))
	bp := shape.ChainPosition(int(b.ID.EdgeID))

	if isInterior {
		if ap.ChainID != bp.ChainID {
			return fmt.Errorf(
				"Loop %d edge %d crosses loop %d edge %d",
				ap.ChainID, ap.Offset, bp.ChainID, bp.Offset)
		}
		return fmt.Errorf("Edge %d crosses edge %d", ap, bp)
	}

	// Loops are not allowed to have duplicate vertices, and separate loops
  // are not allowed to share edges or cross at vertices.  We only need to
  // check a given vertex once, so we also require that the two edges have
  // the same end vertex
	if a.Edge.V1 != b.Edge.V1 {
		return nil
	}

	if ap.ChainID == bp.ChainID {
		return fmt.Errorf("Edge %d has duplicate vertex with edge %d", ap, bp)
	}

	aLen := shape.Chain(ap.ChainID).Length
	bLen := shape.Chain(bp.ChainID).Length
	aNext := ap.Offset + 1
	if aNext == aLen {
		aNext = 0
	}

	bNext := bp.Offset + 1
	if bNext == bLen {
		bNext = 0
	}

	a2 := shape.ChainEdge(ap.ChainID, aNext).V1
	b2 := shape.ChainEdge(bp.ChainID, bNext).V1

	if a.Edge.V0 == b.Edge.V0 || a.Edge.V0 == b2 {
		// The second edge index is sometimes off by one, hence "near".
		return fmt.Errorf(
			"Loop %d edge %d has duplicate near loop %d edge %d",
			ap.ChainID, ap.Offset, bp.ChainID, bp.Offset)
	}

	// Since S2ShapeIndex loops are oriented such that the polygon interior is
  // always on the left, we need to handle the case where one wedge contains
  // the complement of the other wedge.  This is not specifically detected by
  // GetWedgeRelation, so there are two cases to check for.
  //
  // Note that we don't need to maintain any state regarding loop crossings
  // because duplicate edges are detected and rejected above.
	if WedgeRelation(a.Edge.V0, a.Edge.V1, a2, b.Edge.V0, b2) == WedgeProperlyOverlaps &&
		WedgeRelation(a.Edge.V0, a.Edge.V1, a2, b2, b.Edge.V0) == WedgeProperlyOverlaps {
		return fmt.Errorf(
			"Loop %d edge %d crosses loop %d edge %d",
			ap.ChainID, ap.Offset, bp.ChainID, bp.Offset)
	}

	return nil
}

func FindSelfIntersection(index *ShapeIndex) bool {
	if len(index.shapes) == 0 {
		return false
	}
	shape := index.Shape(0)

	// Visit all crossing pairs except possibly for ones of the form (AB, BC),
	// since such pairs are very common and FindCrossingError() only needs pairs
	// of the form (AB, AC).
	return !VisitCrossings(
		index, CrossingTypeAll, false,
		func(a, b ShapeEdge, isInterior bool) bool {
			return FindCrossingError(shape, a, b, isInterior) == nil
		},
	)
}
