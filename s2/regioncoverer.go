package s2

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

	candidateQueue []queueEntry
}

type candidate struct {
	cell        Cell
	isTerminal  bool
	numChildren int
	children    []*candidate
}

type queueEntry struct {
	id        int
	candidate *candidate
}

func newQueueEntry(id int, candidate *candidate) *queueEntry {
	return &queueEntry{id, candidate}
}

func NewRegionCoverer() *RegionCoverer {
	return &RegionCoverer{
		minLevel: 0,
		maxLevel: MAX_LEVEL,
		levelMod: 1,
		maxCells: DEFAULT_MAX_CELLS,
	}
}

func (rc *RegionCoverer) SetMinLevel(minLevel int) {
	rc.minLevel = minLevel
}

func (rc *RegionCoverer) SetMaxLevel(maxLevel int) {
	rc.maxLevel = maxLevel
}

func (rc *RegionCoverer) SetLevelMod(levelMod int) {
	rc.levelMod = levelMod
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

func (rc *RegionCoverer) GetCovering(region Region) *CellUnion {
	rc.interiorCovering = false
	rc.GetCoveringInternal(region)
	return (*CellUnion)(&rc.result)
}

func (rc *RegionCoverer) GetCoveringInternal(region Region) {
	if len(rc.result) > 0 || len(rc.candidateQueue) > 0 {
		panic("Preconditions not met")
	}

	rc.region = region
	rc.candidatesCreatedCounter = 0

	rc.getInitialCandidates()
	for len(rc.candidateQueue) > 0 && (!rc.interiorCovering || len(rc.result) < rc.maxCells) {
		candidate := rc.candidateQueue[0].candidate // TODO: rc.candidateQueue.poll().candidate
		sz := len(rc.result) + candidate.numChildren
		if rc.interiorCovering {
			sz = sz + len(rc.candidateQueue)
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
	// TODO
	rc.result = []CellID{
		0x9000000000000000,
	}

}

func (rc *RegionCoverer) addCandidate(candidate *candidate) {
	if candidate == nil {
		return
	}

	if candidate.isTerminal {
		rc.result = append(rc.result, candidate.cell.Id())
		return
	}

	numLevels := rc.levelMod
	if int(candidate.cell.Level()) < rc.minLevel {
		numLevels = 1
	}
	numTerminals := rc.expandChildren(candidate, candidate.cell, numLevels)

	if candidate.numChildren == 0 {
		// do nothing
	} else if !rc.interiorCovering && numTerminals == 1<<rc.maxChildrenShift() &&
		int(candidate.cell.Level()) >= rc.minLevel {
		// Optimization: add the parent cell rather than all of its children.
		// We can't do this for interior coverings, since the children just
		// intersect the region, but may not be contained by it - we need to
		// subdivide them further.
		candidate.isTerminal = true
		rc.addCandidate(candidate)
	} else {
		// We negate the priority so that smaller absolute priorities are returned
		// first. The heuristic is designed to refine the largest cells first,
		// since those are where we have the largest potential gain. Among cells
		// at the same level, we prefer the cells with the smallest number of
		// intersecting children. Finally, we prefer cells that have the smallest
		// number of children that cannot be refined any further.
		// priority := -((((int(candidate.cell.Level()) << rc.maxChildrenShift()) + candidate.numChildren) << rc.maxChildrenShift()) + numTerminals)
		// TODO: candidateQueue.add(new QueueEntry(priority, candidate));
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
	if rc.region.IntersectsCell(cell) {
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
