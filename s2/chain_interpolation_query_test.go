package s2

import (
	"testing"

	"github.com/golang/geo/s1"
)

const (
	latitudeB      = 1.
	latitudeC      = 2.5
	totalLengthAbc = latitudeC
	kEpsilon       = 1e-8
	kEpsilonAngle  = s1.Angle(kEpsilon)
)

func testSimplePolylines(t *testing.T) {
	a := PointFromLatLng(LatLng{0, 0})
	b := PointFromLatLng(LatLng{latitudeB, 0})
	c := PointFromLatLng(LatLng{latitudeC, 0})

	emptyLaxPolyline := Shape(&LaxPolyline{})
	acPolyline := Shape(&LaxPolyline{vertices: []Point{a, c}})
	abcPolyline := Shape(&LaxPolyline{vertices: []Point{a, b, c}})
	bbPolyline := Shape(&LaxPolyline{vertices: []Point{b, b}})
	ccPolyline := Shape(&LaxPolyline{vertices: []Point{c}})

	uninitializedQuery := ChainInterpolationQuery{}
	emptyQuery := InitS2ChainInterpolationQuery(emptyLaxPolyline, 0)
	acQuery := InitS2ChainInterpolationQuery(acPolyline, 0)
	abcQuery := InitS2ChainInterpolationQuery(abcPolyline, 0)
	bbQuery := InitS2ChainInterpolationQuery(bbPolyline, 0)
	ccQuery := InitS2ChainInterpolationQuery(ccPolyline, 0)
}
