package s2

import (
	"math"
	"sort"
)

var THICKENING float64 = 0.01
var MAX_DET_ERROR float64 = 1e-14

type EdgeIndex struct {
	cells              []uint64
	edges              []int
	minimumS2LevelUsed int
	indexComputed      bool
	queryCount         int

	getNumEdges func() int
	edgeFrom    func(int) Point
	edgeTo      func(int) Point
}

func NewEdgeIndex(getNumEdges func() int, edgeFrom func(int) Point, edgeTo func(int) Point) *EdgeIndex {
	return &EdgeIndex{
		getNumEdges: getNumEdges,
		edgeFrom:    edgeFrom,
		edgeTo:      edgeTo,
	}
}

/**
 * Empties the index in case it already contained something.
 */
func (e *EdgeIndex) Reset() {
	e.minimumS2LevelUsed = MAX_LEVEL
	e.indexComputed = false
	e.queryCount = 0
	e.cells = nil
	e.edges = nil
}

/**
 * Compares [cell1, edge1] to [cell2, edge2], by cell first and edge second.
 *
 * @return -1 if [cell1, edge1] is less than [cell2, edge2], 1 if [cell1,
 *         edge1] is greater than [cell2, edge2], 0 otherwise.
 */
func compare(cell1 uint64, edge1 int, cell2 uint64, edge2 int) int {
	if cell1 < cell2 {
		return -1
	} else if cell1 > cell2 {
		return 1
	} else if edge1 < edge2 {
		return -1
	} else if edge1 > edge2 {
		return 1
	} else {
		return 0
	}
}

/** Computes the index (if it has not been previously done). */
func (e *EdgeIndex) ComputeIndex() {
	if e.indexComputed {
		return
	}
	cellList := []uint64{}
	edgeList := []int{}
	for i := 0; i < e.getNumEdges(); i++ {
		from := e.edgeFrom(i)
		to := e.edgeTo(i)
		cover := []CellID{}
		level := e.getCovering(from, to, true, &cover)
		e.minimumS2LevelUsed = min(e.minimumS2LevelUsed, level)
		for _, cellId := range cover {
			cellList = append(cellList, uint64(cellId))
			edgeList = append(edgeList, i)
		}
	}
	e.cells = make([]uint64, len(cellList))
	e.edges = make([]int, len(edgeList))
	for i := 0; i < len(e.cells); i++ {
		e.cells[i] = cellList[i]
		e.edges[i] = edgeList[i]
	}
	e.sortIndex()
	e.indexComputed = true
}

type ByCellThenEdge struct {
	e *EdgeIndex
	a []int
}

func (s ByCellThenEdge) Len() int      { return len(s.a) }
func (s ByCellThenEdge) Swap(i, j int) { s.a[i], s.a[j] = s.a[j], s.a[i] }
func (s ByCellThenEdge) Less(i, j int) bool {
	return compare(s.e.cells[i], s.e.edges[i], s.e.cells[j], s.e.edges[j]) < 0
}

/** Sorts the parallel <code>cells</code> and <code>edges</code> arrays. */
func (e *EdgeIndex) sortIndex() {
	// create an array of indices and sort based on the values in the parallel
	// arrays at each index
	indices := make([]int, len(e.cells))
	for i := 0; i < len(indices); i++ {
		indices[i] = i
	}
	sort.Sort(ByCellThenEdge{e, indices})
	// copy the cells and edges in the order given by the sorted list of indices
	newCells := make([]uint64, len(e.cells))
	newEdges := make([]int, len(e.edges))
	for i := 0; i < len(indices); i++ {
		newCells[i] = e.cells[indices[i]]
		newEdges[i] = e.edges[indices[i]]
	}
	// replace the cells and edges with the sorted arrays
	e.cells = newCells
	e.edges = newEdges
}

func (e *EdgeIndex) IsIndexComputed() bool {
	return e.indexComputed
}

/**
 * Tell the index that we just received a new request for candidates. Useful
 * to compute when to switch to quad tree.
 */
func (e *EdgeIndex) incrementQueryCount() {
	e.queryCount++
}

/**
 * If the index hasn't been computed yet, looks at how much work has gone into
 * iterating using the brute force method, and how much more work is planned
 * as defined by 'cost'. If it were to have been cheaper to use a quad tree
 * from the beginning, then compute it now. This guarantees that we will never
 * use more than twice the time we would have used had we known in advance
 * exactly how many edges we would have wanted to test. It is the theoretical
 * best.
 *
 *  The value 'n' is the number of iterators we expect to request from this
 * edge index.
 *
 *  If we have m data edges and n query edges, then the brute force cost is m
 * * n * testCost where testCost is taken to be the cost of
 * EdgeCrosser.robustCrossing, measured to be about 30ns at the time of this
 * writing.
 *
 *  If we compute the index, the cost becomes: m * costInsert + n *
 * costFind(m)
 *
 *  - costInsert can be expected to be reasonably stable, and was measured at
 * 1200ns with the BM_QuadEdgeInsertionCost benchmark.
 *
 *  - costFind depends on the length of the edge . For m=1000 edges, we got
 * timings ranging from 1ms (edge the length of the polygon) to 40ms. The
 * latter is for very long query edges, and needs to be optimized. We will
 * assume for the rest of the discussion that costFind is roughly 3ms.
 *
 *  When doing one additional query, the differential cost is m * testCost -
 * costFind(m) With the numbers above, it is better to use the quad tree (if
 * we have it) if m >= 100.
 *
 *  If m = 100, 30 queries will give m*n*testCost = m * costInsert = 100ms,
 * while the marginal cost to find is 3ms. Thus, this is a reasonable thing to
 * do.
 */
func (e *EdgeIndex) PredictAdditionalCalls(n int) {
	if e.indexComputed {
		return
	}
	if e.getNumEdges() > 100 && (e.queryCount+n) > 30 {
		e.ComputeIndex()
	}
}

/**
 * Appends to "candidateCrossings" all edge references which may cross the
 * given edge. This is done by covering the edge and then finding all
 * references of edges whose coverings overlap this covering. Parent cells are
 * checked level by level. Child cells are checked all at once by taking
 * advantage of the natural ordering of S2CellIds.
 */
func (e *EdgeIndex) findCandidateCrossings(a, b Point, candidateCrossings *[]int) {
	// Preconditions.checkState(indexComputed);
	cover := []CellID{}
	e.getCovering(a, b, false, &cover)

	// Edge references are inserted into the map once for each covering cell, so
	// absorb duplicates here

	uniqueSet := make(map[int]bool)
	e.getEdgesInParentCells(cover, &uniqueSet)

	// TODO(user): An important optimization for long query
	// edges (Contains queries): keep a bounding cap and clip the query
	// edge to the cap before starting the descent.
	e.getEdgesInChildrenCells(a, b, &cover, &uniqueSet)

	*candidateCrossings = make([]int, len(uniqueSet))
	for k := range uniqueSet {
		*candidateCrossings = append(*candidateCrossings, k)
	}
}

/**
 * Returns the smallest cell containing all four points, or
 * {@link S2CellId#sentinel()} if they are not all on the same face. The
 * points don't need to be normalized.
 */
func (e *EdgeIndex) containingCell4(pa, pb, pc, pd Point) CellID {
	a := CellIDFromPoint(pa)
	b := CellIDFromPoint(pb)
	c := CellIDFromPoint(pc)
	d := CellIDFromPoint(pd)

	if a.Face() != b.Face() || a.Face() != c.Face() || a.Face() != d.Face() {
		return CellIDSentinel()
	}

	for a != b || a != c || a != d {
		a = a.immediateParent()
		b = b.immediateParent()
		c = c.immediateParent()
		d = d.immediateParent()
	}
	return a
}

/**
 * Returns the smallest cell containing both points, or Sentinel if they are
 * not all on the same face. The points don't need to be normalized.
 */
func (e *EdgeIndex) containingCell2(pa, pb Point) CellID {
	a := CellIDFromPoint(pa)
	b := CellIDFromPoint(pb)

	if a.Face() != b.Face() {
		return CellIDSentinel()
	}

	for a != b {
		a = a.immediateParent()
		b = b.immediateParent()
	}
	return a
}

/**
 * Computes a cell covering of an edge. Clears edgeCovering and returns the
 * level of the s2 cells used in the covering (only one level is ever used for
 * each call).
 *
 *  If thickenEdge is true, the edge is thickened and extended by 1% of its
 * length.
 *
 *  It is guaranteed that no child of a covering cell will fully contain the
 * covered edge.
 */
func (e *EdgeIndex) getCovering(a, b Point, thickenEdge bool, edgeCovering *[]CellID) int {
	*edgeCovering = []CellID{}

	// Selects the ideal s2 level at which to cover the edge, this will be the
	// level whose S2 cells have a width roughly commensurate to the length of
	// the edge. We multiply the edge length by 2*THICKENING to guarantee the
	// thickening is honored (it's not a big deal if we honor it when we don't
	// request it) when doing the covering-by-cap trick.
	edgeLength := a.Angle(b.Vector).Radians()
	idealLevel := S2_PROJECTION.MIN_WIDTH().getMaxLevel(edgeLength * (1 + 2*THICKENING))

	var containingCellId CellID
	if !thickenEdge {
		containingCellId = e.containingCell2(a, b)
	} else {
		if idealLevel == MAX_LEVEL {
			// If the edge is tiny, instabilities are more likely, so we
			// want to limit the number of operations.
			// We pretend we are in a cell much larger so as to trigger the
			// 'needs covering' case, so we won't try to thicken the edge.
			containingCellId = CellID(0xFFF0).Parent(3)
		} else {
			pq := b.Sub(a.Vector).Mul(THICKENING)
			ortho := pq.Cross(a.Vector).Normalize().Mul(edgeLength * THICKENING)
			p := a.Sub(pq)
			q := b.Add(pq)
			// If p and q were antipodal, the edge wouldn't be lengthened,
			// and it could even flip! This is not a problem because
			// idealLevel != 0 here. The farther p and q can be is roughly
			// a quarter Earth away from each other, so we remain
			// Theta(THICKENING).
			containingCellId = e.containingCell4(
				Point{p.Sub(ortho)},
				Point{p.Add(ortho)},
				Point{q.Sub(ortho)},
				Point{q.Add(ortho)},
			)
		}
	}

	// Best case: edge is fully contained in a cell that's not too big.
	if containingCellId != CellIDSentinel() && containingCellId.Level() >= idealLevel-2 {
		*edgeCovering = append(*edgeCovering, containingCellId)
		return containingCellId.Level()
	}

	if idealLevel == 0 {
		// Edge is very long, maybe even longer than a face width, so the
		// trick below doesn't work. For now, we will add the whole S2 sphere.
		// TODO(user): Do something a tad smarter (and beware of the
		// antipodal case).
		for cellid := CellIDBegin(0); cellid != CellIDEnd(0); cellid = cellid.Next() {
			*edgeCovering = append(*edgeCovering, cellid)
		}
		return 0
	}
	// TODO(user): Check trick below works even when vertex is at
	// interface
	// between three faces.

	// Use trick as in S2PolygonBuilder.PointIndex.findNearbyPoint:
	// Cover the edge by a cap centered at the edge midpoint, then cover
	// the cap by four big-enough cells around the cell vertex closest to the
	// cap center.
	middle := Point{a.Add(b.Vector).Div(2).Normalize()}
	actualLevel := min(idealLevel, MAX_LEVEL-1)
	*edgeCovering = CellIDFromPoint(middle).VertexNeighbors(actualLevel)
	return actualLevel
}

/**
 * Filters a list of entries down to the inclusive range defined by the given
 * cells, in <code>O(log N)</code> time.
 *
 * @param cell1 One side of the inclusive query range.
 * @param cell2 The other side of the inclusive query range.
 * @return An array of length 2, containing the start/end indices.
 */
func (e *EdgeIndex) getEdges(cell1, cell2 uint64) []int {
	// ensure cell1 <= cell2
	if cell1 > cell2 {
		cell1, cell2 = cell2, cell1
	}
	// The binary search returns -N-1 to indicate an insertion point at index N,
	// if an exact match cannot be found. Since the edge indices queried for are
	// not valid edge indices, we will always get -N-1, so we immediately
	// convert to N.
	return []int{
		-1 - e.binarySearch(cell1, -(int(^uint(0)>>1)-1)),
		-1 - e.binarySearch(cell2, int(^uint(0)>>1)),
	}
}

func (e *EdgeIndex) binarySearch(cell uint64, edge int) int {
	low := 0
	high := len(e.cells) - 1
	for low <= high {
		mid := (low + high) >> 1
		cmp := compare(e.cells[mid], e.edges[mid], cell, edge)
		if cmp < 0 {
			low = mid + 1
		} else if cmp > 0 {
			high = mid - 1
		} else {
			return mid
		}
	}
	return -(low + 1)
}

/**
 * Adds to candidateCrossings all the edges present in any ancestor of any
 * cell of cover, down to minimumS2LevelUsed. The cell->edge map is in the
 * variable mapping.
 */
func (e *EdgeIndex) getEdgesInParentCells(cover []CellID, candidateCrossings *map[int]bool) {
	// Find all parent cells of covering cells.
	parentCells := make(map[CellID]bool)
	for _, coverCell := range cover {
		for parentLevel := coverCell.Level() - 1; parentLevel >= e.minimumS2LevelUsed; parentLevel-- {
			if _, ok := parentCells[coverCell.Parent(parentLevel)]; ok {
				break // cell is already in => parents are too.
			}
			parentCells[coverCell.Parent(parentLevel)] = true
		}
	}

	// Put parent cell edge references into result.
	for parentCell := range parentCells {
		bounds := e.getEdges(uint64(parentCell), uint64(parentCell))
		for i := bounds[0]; i < bounds[1]; i++ {
			(*candidateCrossings)[e.edges[i]] = true
		}
	}
}

/**
 * Returns true if ab possibly crosses cd, by clipping tiny angles to zero.
 */
func lenientCrossing(a, b, c, d Point) bool {
	// assert (S2.isUnitLength(a));
	// assert (S2.isUnitLength(b));
	// assert (S2.isUnitLength(c));

	acb := a.Cross(c.Vector).Dot(b.Vector)
	bda := b.Cross(d.Vector).Dot(a.Vector)
	if math.Abs(acb) < MAX_DET_ERROR || math.Abs(bda) < MAX_DET_ERROR {
		return true
	}
	if acb*bda < 0 {
		return false
	}
	cbd := c.Cross(b.Vector).Dot(d.Vector)
	dac := c.Cross(a.Vector).Dot(c.Vector)
	if math.Abs(cbd) < MAX_DET_ERROR || math.Abs(dac) < MAX_DET_ERROR {
		return true
	}
	return (acb*cbd >= 0) && (acb*dac >= 0)
}

/**
 * Returns true if the edge and the cell (including boundary) intersect.
 */
func edgeIntersectsCellBoundary(a, b Point, cell Cell) bool {
	vertices := make([]Point, 4)
	for i := 0; i < 4; i++ {
		vertices[i] = cell.Vertex(i)
	}
	for i := 0; i < 4; i++ {
		fromPoint := vertices[i]
		toPoint := vertices[(i+1)%4]
		if lenientCrossing(a, b, fromPoint, toPoint) {
			return true
		}
	}
	return false
}

/**
 * Appends to candidateCrossings the edges that are fully contained in an S2
 * covering of edge. The covering of edge used is initially cover, but is
 * refined to eliminate quickly subcells that contain many edges but do not
 * intersect with edge.
 */
func (e *EdgeIndex) getEdgesInChildrenCells(a, b Point, cover *[]CellID, candidateCrossings *map[int]bool) {
	// Put all edge references of (covering cells + descendant cells) into
	// result.
	// This relies on the natural ordering of S2CellIds.
	for len(*cover) > 0 {
		cell := (*cover)[len(*cover)-1]
		*cover = (*cover)[0 : len(*cover)-1]

		bounds := e.getEdges(uint64(cell.RangeMin()), uint64(cell.RangeMax()))
		if bounds[1]-bounds[0] <= 16 {
			for i := bounds[0]; i < bounds[1]; i++ {
				(*candidateCrossings)[e.edges[i]] = true
			}
		} else {
			// Add cells at this level
			bounds = e.getEdges(uint64(cell), uint64(cell))
			for i := bounds[0]; i < bounds[1]; i++ {
				(*candidateCrossings)[e.edges[i]] = true
			}
			// Recurse on the children -- hopefully some will be empty.
			children := cell.Children()
			for _, child := range children {
				// TODO(user): Do the check for the four cells at once,
				// as it is enough to check the four edges between the cells. At
				// this time, we are checking 16 edges, 4 times too many.
				//
				// Note that given the guarantee of AppendCovering, it is enough
				// to check that the edge intersect with the cell boundary as it
				// cannot be fully contained in a cell.
				if edgeIntersectsCellBoundary(a, b, CellFromCellID(child)) {
					*cover = append(*cover, child)
				}
			}
		}
	}
}

/*
 * An iterator on data edges that may cross a query edge (a,b). Create the
 * iterator, call getCandidates(), then hasNext()/next() repeatedly.
 *
 * The current edge in the iteration has index index(), goes between from()
 * and to().
 */
type DataEdgeIterator struct {
	/**
	 * The structure containing the data edges.
	 */
	edgeIndex *EdgeIndex

	/**
	 * Tells whether getCandidates() obtained the candidates through brute force
	 * iteration or using the quad tree structure.
	 */
	isBruteForce bool

	/**
	 * Index of the current edge and of the edge before the last next() call.
	 */
	currentIndex int

	/**
	 * Cache of edgeIndex.getNumEdges() so that hasNext() doesn't make an extra
	 * call
	 */
	numEdges int

	/**
	 * All the candidates obtained by getCandidates() when we are using a
	 * quad-tree (i.e. isBruteForce = false).
	 */
	candidates []int

	/**
	 * Index within array above. We have: currentIndex =
	 * candidates.get(currentIndexInCandidates).
	 */
	currentIndexInCandidates int
}

func NewDataEdgeIterator(edgeIndex *EdgeIndex) *DataEdgeIterator {
	return &DataEdgeIterator{
		edgeIndex:  edgeIndex,
		candidates: []int{},
	}
}

/**
 * Initializes the iterator to iterate over a set of candidates that may
 * cross the edge (a,b).
 */
func (d *DataEdgeIterator) GetCandidates(a, b Point) {
	d.edgeIndex.PredictAdditionalCalls(1)
	d.isBruteForce = !d.edgeIndex.IsIndexComputed()
	if d.isBruteForce {
		d.edgeIndex.incrementQueryCount()
		d.currentIndex = 0
		d.numEdges = d.edgeIndex.getNumEdges()
	} else {
		d.candidates = []int{}
		d.edgeIndex.findCandidateCrossings(a, b, &d.candidates)
		d.currentIndexInCandidates = 0
		if len(d.candidates) > 0 {
			d.currentIndex = d.candidates[0]
		}
	}
}

/**
 * Index of the current edge in the iteration.
 */
func (d *DataEdgeIterator) Index() int {
	if !d.HasNext() {
		panic("No next candidate, use HasNext")
	}
	return d.currentIndex
}

/**
 * False if there are no more candidates; true otherwise.
 */
func (d *DataEdgeIterator) HasNext() bool {
	if d.isBruteForce {
		return d.currentIndex < d.numEdges
	} else {
		return d.currentIndexInCandidates < len(d.candidates)
	}
}

/**
 * Iterate to the next available candidate.
 */
func (d *DataEdgeIterator) Next() {
	if !d.HasNext() {
		panic("No next candidate, use HasNext")
	}
	if d.isBruteForce {
		d.currentIndex++
	} else {
		d.currentIndexInCandidates++
		if d.currentIndexInCandidates < len(d.candidates) {
			d.currentIndex = d.candidates[d.currentIndexInCandidates]
		}
	}
}
