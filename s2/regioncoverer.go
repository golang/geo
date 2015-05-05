package s2

import (
	"container/heap"
)

var (
	DEFAULT_MAX_CELLS int = 8

	FACE_CELLS = []Cell{
		CellFromCellID(CellIDFromFace(0)),
		CellFromCellID(CellIDFromFace(1)),
		CellFromCellID(CellIDFromFace(2)),
		CellFromCellID(CellIDFromFace(3)),
		CellFromCellID(CellIDFromFace(4)),
		CellFromCellID(CellIDFromFace(5)),
	}
)

type RegionCoverer struct {
	minLevel int
	maxLevel int
	levelMod int
	maxCells int

	interiorCovering bool

	candidatesCreatedCounter int

	region Region

	result []CellID

	candidateQueue PriorityQueue
}

type candidate struct {
	cell        Cell
	isTerminal  bool
	numChildren int
	children    []*candidate
}

type queueEntry struct {
	priority  int
	candidate *candidate
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

func newQueueEntry(id int, candidate_ *candidate) *queueEntry {
	return &queueEntry{priority: id, candidate: candidate_}
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*queueEntry

func newPriorityQueue(space int) PriorityQueue {
	pq := PriorityQueue(make([]*queueEntry, 0, space))
	heap.Init(&pq)
	return pq
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority >= pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*queueEntry)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *queueEntry, candidate *candidate, priority int) {
	item.candidate = candidate
	item.priority = priority
	heap.Fix(pq, item.index)
}

func NewRegionCoverer() *RegionCoverer {
	return &RegionCoverer{
		minLevel:       0,
		maxLevel:       MAX_LEVEL,
		levelMod:       1,
		maxCells:       DEFAULT_MAX_CELLS,
		candidateQueue: newPriorityQueue(10),
	}
}

func (rc *RegionCoverer) SetMinLevel(minLevel int) {
	rc.minLevel = max(0, min(MAX_LEVEL, minLevel))
}

func (rc *RegionCoverer) SetMaxLevel(maxLevel int) {
	rc.maxLevel = max(0, min(MAX_LEVEL, maxLevel))
}

func (rc *RegionCoverer) SetLevelMod(levelMod int) {
	rc.levelMod = max(1, min(3, levelMod))
}

func (rc *RegionCoverer) SetMaxCells(maxCells int) {
	rc.maxCells = maxCells
}

func (rc *RegionCoverer) MinLevel() int {
	return rc.minLevel
}

func (rc *RegionCoverer) MaxLevel() int {
	return rc.maxLevel
}

func (rc *RegionCoverer) LevelMod() int {
	return rc.levelMod
}

func (rc *RegionCoverer) MaxCells() int {
	return rc.maxCells
}

/**
 * Computes a list of cell ids that covers the given region and satisfies the
 * various restrictions specified above.
 *
 * @param region The region to cover
 * @param covering The list filled in by this method
 */
func (rc *RegionCoverer) GetCovering(region Region, covering *[]CellID) {
	// Rather than just returning the raw list of cell ids generated by
	// GetCoveringInternal(), we construct a cell union and then denormalize it.
	// This has the effect of replacing four child cells with their parent
	// whenever this does not violate the covering parameters specified
	// (min_level, level_mod, etc). This strategy significantly reduces the
	// number of cells returned in many cases, and it is cheap compared to
	// computing the covering in the first place.
	tmp := rc.GetCoveringAsUnion(region)
	tmp.DeNormalize(rc.minLevel, rc.levelMod, covering)
}

/**
 * Return a normalized cell union that covers the given region and satisfies
 * the restrictions *EXCEPT* for min_level() and level_mod(). These criteria
 * cannot be satisfied using a cell union because cell unions are
 * automatically normalized by replacing four child cells with their parent
 * whenever possible. (Note that the list of cell ids passed to the cell union
 * constructor does in fact satisfy all the given restrictions.)
 */
func (rc *RegionCoverer) GetCoveringAsUnion(region Region) *CellUnion {
	rc.interiorCovering = false
	rc.GetCoveringInternal(region)
	union := CellUnionFromArrayAndSwap(&rc.result)
	return union
}

/** Generates a covering and stores it in result. */
func (rc *RegionCoverer) GetCoveringInternal(region Region) {
	if len(rc.result) > 0 || rc.candidateQueue.Len() > 0 {
		panic("Preconditions not met")
	}

	rc.region = region
	rc.candidatesCreatedCounter = 0

	rc.getInitialCandidates()
	for rc.candidateQueue.Len() > 0 && (!rc.interiorCovering || len(rc.result) < rc.maxCells) {
		candidate := heap.Pop(&rc.candidateQueue).(*queueEntry).candidate
		sz := len(rc.result) + candidate.numChildren
		if !rc.interiorCovering {
			sz = sz + rc.candidateQueue.Len()
		}
		if int(candidate.cell.Level()) < rc.minLevel || candidate.numChildren == 1 || sz <= rc.maxCells {
			for i := 0; i < candidate.numChildren; i++ {
				rc.addCandidate(candidate.children[i])
			}

		} else if rc.interiorCovering {
			// do nothing
		} else {
			candidate.isTerminal = true
			rc.addCandidate(candidate)
		}

	}

	rc.candidateQueue = nil
	rc.region = nil
}

func (rc *RegionCoverer) getInitialCandidates() {
	// Optimization: if at least 4 cells are desired (the normal case),
	// start with a 4-cell covering of the region's bounding cap. This
	// lets us skip quite a few levels of refinement when the region to
	// be covered is relatively small.
	if rc.maxCells >= 4 {
		// Find the maximum level such that the bounding cap contains at most one
		// cell vertex at that level.
		cap := rc.region.CapBound()
		level := min(S2_PROJECTION.MIN_WIDTH().getMaxLevel(2*cap.Radius().Radians()), min(rc.maxLevel, MAX_LEVEL-1))
		if rc.levelMod > 1 && level > rc.minLevel {
			level -= (level - rc.minLevel) % rc.levelMod
		}
		// We don't bother trying to optimize the level == 0 case, since more than
		// four face cells may be required.
		if level > 0 {
			// Find the leaf cell containing the cap axis, and determine which
			// subcell of the parent cell contains it.
			id := CellIDFromPoint(cap.Center())
			base := id.VertexNeighbors(level)
			for i := 0; i < len(base); i++ {
				rc.addCandidate(rc.newCandidate(CellFromCellID(base[i])))
			}
			return
		}
	}
	// Default: start with all six cube faces.
	for face := 0; face < 6; face++ {
		rc.addCandidate(rc.newCandidate(FACE_CELLS[face]))
	}
}

func (rc *RegionCoverer) addCandidate(candidate_ *candidate) {
	if candidate_ == nil {
		return
	}

	if candidate_.isTerminal {
		rc.result = append(rc.result, candidate_.cell.Id())
		return
	}

	numLevels := rc.levelMod
	if int(candidate_.cell.Level()) < rc.minLevel {
		numLevels = 1
	}
	numTerminals := rc.expandChildren(candidate_, candidate_.cell, numLevels)

	if candidate_.numChildren == 0 {
		// do nothing
	} else if !rc.interiorCovering && numTerminals == 1<<rc.maxChildrenShift() &&
		int(candidate_.cell.Level()) >= rc.minLevel {
		// Optimization: add the parent cell rather than all of its children.
		// We can't do this for interior coverings, since the children just
		// intersect the region, but may not be contained by it - we need to
		// subdivide them further.
		candidate_.isTerminal = true
		rc.addCandidate(candidate_)
	} else {
		// We negate the priority so that smaller absolute priorities are returned
		// first. The heuristic is designed to refine the largest cells first,
		// since those are where we have the largest potential gain. Among cells
		// at the same level, we prefer the cells with the smallest number of
		// intersecting children. Finally, we prefer cells that have the smallest
		// number of children that cannot be refined any further.
		priority := -((((int(candidate_.cell.Level()) << rc.maxChildrenShift()) + candidate_.numChildren) << rc.maxChildrenShift()) + numTerminals)
		heap.Push(&rc.candidateQueue, newQueueEntry(priority, candidate_))
	}
}

func (rc *RegionCoverer) maxChildrenShift() uint {
	return 2 * uint(rc.levelMod)
}

func (rc *RegionCoverer) expandChildren(candidate *candidate, cell Cell, numLevels int) int {
	numLevels--
	childCellIds := cell.Id().Children()
	numTerminals := 0
	for i := 0; i < 4; i++ {
		childCell := CellFromCellID(childCellIds[i])
		if numLevels > 0 {
			if rc.region.IntersectsCell(childCell) {
				numTerminals = numTerminals + rc.expandChildren(candidate, childCell, numLevels)
			}
			continue
		}
		child := rc.newCandidate(childCell)
		if child != nil {
			candidate.children[candidate.numChildren] = child
			candidate.numChildren++
			if child.isTerminal {
				numTerminals++
			}
		}
	}
	return numTerminals
}

func (rc *RegionCoverer) newCandidate(cell Cell) *candidate {
	if !rc.region.IntersectsCell(cell) {
		return nil
	}
	isTerminal := false
	if int(cell.Level()) >= rc.minLevel {
		if rc.interiorCovering {
			if rc.region.ContainsCell(cell) {
				isTerminal = true
			} else if int(cell.Level())+rc.levelMod > rc.maxLevel {
				return nil
			}
		} else {
			if int(cell.Level())+rc.levelMod > rc.maxLevel || rc.region.ContainsCell(cell) {
				isTerminal = true
			}
		}

	}
	candidate_ := &candidate{
		cell:       cell,
		isTerminal: isTerminal,
	}
	if !isTerminal {
		candidate_.children = make([]*candidate, 1<<rc.maxChildrenShift())
	}
	rc.candidatesCreatedCounter++
	return candidate_
}