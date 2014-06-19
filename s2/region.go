package s2

// A Region represents a two-dimensional region on the unit sphere.
//
// The purpose of this interface is to allow complex regions to be
// approximated as simpler regions. The interface is restricted to methods
// that are useful for computing approximations.
type Region interface {
	// CapBound returns a bounding spherical cap. This is not guaranteed to be exact.
	CapBound() Cap

	// RectBound returns a bounding latitude-longitude rectangle that contains
	// the region. The bounds are not guaranteed to be tight.
	RectBound() Rect

	// ContainsCell reports whether the region completely contains the given region.
	// It returns false if containment could not be determined.
	ContainsCell(c Cell) bool

	// IntersectsCell reports whether the region intersects the given cell or
	// if intersection could not be determined. It returns false if the region
	// does not intersect.
	IntersectsCell(c Cell) bool
}

// Enforce interface satisfaction for a few types once they satisfy the interface.
/*
var (
	_ Region = Cap{}
	_ Region = CellUnion(nil)
	_ Region = Rect{}
)
*/
