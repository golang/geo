package s2

type RegionCoverer struct {
	maxCells int
}

func NewRegionCoverer() *RegionCoverer {
	return &RegionCoverer{}
}

func (rc *RegionCoverer) SetMaxCells(maxCells int) {
	rc.maxCells = maxCells

}

func (rc *RegionCoverer) GetCovering(region Region) *CellUnion {
	return &CellUnion{
		0x9000000000000000,
	}
}
