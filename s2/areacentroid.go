package s2

/**
 * The area of an interior, i.e. the region on the left side of an odd
 * number of loops and optionally a centroid.
 * The area is between 0 and 4*Pi. If it has a centroid, it is
 * the true centroid of the interiord multiplied by the area of the shape.
 * Note that the centroid may not be contained by the shape.
 *
 * @author dbentley@google.com (Daniel Bentley)
 */
type AreaCentroid struct {
	area     float64
	centroid *Point
}

func NewAreaCentroid(area float64, centroid *Point) AreaCentroid {
	return AreaCentroid{
		area:     area,
		centroid: centroid,
	}
}

func (ac AreaCentroid) GetArea() float64 {
	return ac.area
}

func (ac AreaCentroid) GetCentroid() Point {
	if ac.centroid == nil {
		panic("no centroid")
	}
	return *ac.centroid
}
