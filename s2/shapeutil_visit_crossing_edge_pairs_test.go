package s2

import (
	"slices"
	"sort"
	"testing"
)

type EdgePair struct {
	A, B ShapeEdgeID
}

// A set of edge pairs within an S2ShapeIndex.
type EdgePairVector []EdgePair

// Get crossings in one index.
func getCrossings(index *ShapeIndex, crossingType CrossingType) EdgePairVector {
	edgePairs := EdgePairVector{}
	VisitCrossingEdgePairs(index, crossingType, func(a, b ShapeEdge, _ bool) bool {
		edgePairs = append(edgePairs, EdgePair{a.ID, b.ID})
		return true // Continue visiting.
	})
	if len(edgePairs) > 1 {
		sort.Slice(edgePairs, func(i, j int) bool {
			return edgePairs[i].A.Cmp(edgePairs[j].A) == -1 || edgePairs[i].B.Cmp(edgePairs[j].B) == -1
		})
		slices.Compact(edgePairs)
	}
	return edgePairs
}

// Brute force crossings in one index.
func getCrossingEdgePairsBruteForce(index *ShapeIndex, crossingType CrossingType) EdgePairVector {
	var result EdgePairVector
	minSign := Cross
	if crossingType == CrossingTypeAll {
		minSign = MaybeCross
	}

	for aIter := NewEdgeIterator(index); !aIter.Done(); aIter.Next() {
		a := aIter.Edge()
		bIter := EdgeIterator{
			index:    aIter.index,
			shapeID:  aIter.shapeID,
			numEdges: aIter.numEdges,
			edgeID:   aIter.edgeID,
		}
		for bIter.Next(); !bIter.Done(); bIter.Next() {
			b := bIter.Edge()
			// missinglink: enum ordering is reversed compared to C++
			if CrossingSign(a.V0, a.V1, b.V0, b.V1) <= minSign {
				result = append(result, EdgePair{
					aIter.ShapeEdgeID(),
					bIter.ShapeEdgeID(),
				})
			}
		}
	}
	return result
}

func TestGetCrossingEdgePairs(t *testing.T) {
	var index ShapeIndex
	if len(getCrossings(&index, CrossingTypeAll)) != 0 {
		t.Error("Expected 0 crossings in empty index")
	}
	if len(getCrossings(&index, CrossingTypeInterior)) != 0 {
		t.Error("Expected 0 interior crossings in empty index")
	}
}

func TestGetCrossingEdgePairsGrid(t *testing.T) {
	kGridSize := 10.0
  epsilon := 1e-10

	// There are 11 horizontal and 11 vertical lines. The expected number of
  // interior crossings is 9x9, plus 9 "touching" intersections along each of
  // the left, right, and bottom edges. "epsilon" is used to make the interior
  // lines slightly longer so the "touches" actually cross, otherwise 3 of the
  // 27 touches are not considered intersecting.
  // However, the vertical lines do not reach the top line as it curves on the
  // surface of the sphere: despite "epsilon" those 9 are not even very close
  // to intersecting. Thus 9 * 12 = 108 interior and four more at the corners
  // when CrossingType::ALL is used.

	index :=  NewShapeIndex()
	shape := edgeVectorShape{}

  for i := 0.0; i <= kGridSize; i++ {
    var e = epsilon
		if i == 0 || i == kGridSize {
			e = 0
		}

    shape.Add(PointFromLatLng(LatLngFromDegrees(-e, i)), PointFromLatLng(LatLngFromDegrees(kGridSize + e, i)));
    shape.Add(PointFromLatLng(LatLngFromDegrees(i, -e)), PointFromLatLng(LatLngFromDegrees(i, kGridSize + e)));
  }

	index.Add(&shape)
	if len(getCrossingEdgePairsBruteForce(index, CrossingTypeAll)) != 112 {
		t.Errorf("Fail")
	}
	if len(getCrossingEdgePairsBruteForce(index, CrossingTypeInterior)) != 108 {
		t.Errorf("Fail")
	}
}

func testHasCrossingPermutations(t *testing.T, loops []*Loop, i int, hasCrossing bool) {
	if i == len(loops) {
			index := NewShapeIndex()
			polygon := PolygonFromLoops(loops)
			index.Add(polygon)
			
			if hasCrossing != FindSelfIntersection(index) {
					t.Error("Test failed: expected and actual crossing results do not match")
			}
			return
	}

	origLoop := loops[i]
	for j := 0; j < origLoop.NumVertices(); j++ {
			vertices := make([]Point, origLoop.NumVertices())
			for k := 0; k < origLoop.NumVertices(); k++ {
					vertices[k] = origLoop.Vertex((j + k) % origLoop.NumVertices())
			}
			
			loops[i] = LoopFromPoints(vertices)
			testHasCrossingPermutations(t, loops, i+1, hasCrossing)
	}
	loops[i] = origLoop
}

func TestHasCrossing(t *testing.T) {
	// Coordinates are (lat,lng), which can be visualized as (y,x).
	cases := []struct {
		polygonStr   string
		hasCrossing  bool
	}{
		{"0:0, 0:1, 0:2, 1:2, 1:1, 1:0", false},
		{"0:0, 0:1, 0:2, 1:2, 0:1, 1:0", true}, // duplicate vertex
		{"0:0, 0:1, 1:0, 1:1", true}, // edge crossing
		{"0:0, 1:1, 0:1; 0:0, 1:1, 1:0", true}, // duplicate edge
		{"0:0, 1:1, 0:1; 1:1, 0:0, 1:0", true}, // reversed edge
		{"0:0, 0:2, 2:2, 2:0; 1:1, 0:2, 3:1, 2:0", true}, // vertex crossing
	}
	for _, tc := range cases {
		polygon := makePolygon(tc.polygonStr, true)
		testHasCrossingPermutations(t, polygon.loops, 0, tc.hasCrossing)
	}
}
