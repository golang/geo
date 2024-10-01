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

	type reslut struct {
		point    Point
		edgeID   int
		distance s1.Angle
	}
	distances := []float64{
		-1.,
		0.,
		1.0e-8,
		latitudeB / 2,
		latitudeB - 1.0e-7,
		latitudeB,
		latitudeB + 1.0e-5,
		latitudeB + 0.5,
		latitudeC - 10.e-7,
		latitudeC,
		latitudeC + 10.e-16,
		1.e6,
	}

	groundTruth := make([]reslut, len(distances))
	for i, distance := range distances {
		lat := max(.0, min(totalLengthAbc, distance))
		point := PointFromLatLng(LatLngFromDegrees(lat, 0))
		var edgeID int
		if distance < latitudeB {
			edgeID = 0
		} else {
			edgeID = 1
		}
		groundTruth[i] = reslut{point: point, edgeID: edgeID, distance: s1.Angle(distance)}
	}

	lengthEmpty, err := emptyQuery.GetLength()
	if err != nil {
		t.Fatal(err)
	}
	lengthAc, err := acQuery.GetLength()
	if err != nil {
		t.Fatal(err)
	}
	lengthAbc, err := abcQuery.GetLength()
	if err != nil {
		t.Fatal(err)
	}
	lengthBb, err := bbQuery.GetLength()
	if err != nil {
		t.Fatal(err)
	}
	lengthCc, err := ccQuery.GetLength()
	if err != nil {
		t.Fatal(err)
	}
	degreesEmpty := lengthEmpty.Degrees()
	degreesAc := lengthAc.Degrees()
	degreesAbc := lengthAbc.Degrees()
	degreesBb := lengthBb.Degrees()
	degreesCc := lengthCc.Degrees()

	point, edgeID, distance, err := uninitializedQuery.AtFraction(0)
	if err == nil {
		t.Fatalf("got %v, want error", point)
	}
	point, edgeID, distance, err = acQuery.AtDistance(s1.InfAngle())
	if err == nil {
		t.Fatalf("got %v, want error", point)
	}

	distanceResult := make([]reslut, len(distances))

	for i, distance := range distances {
		totalFraction := distance / totalLengthAbc

		distancePoint, distanceEdgeID, newDistance, err := emptyQuery.AtFraction(totalFraction)
		if err != nil && i != len(distances)-1 {
			t.Fatal(err)
		}
		distanceResult[i] = reslut{point: distancePoint, edgeID: distanceEdgeID, distance: newDistance}
	}

	if degreesEmpty >= kEpsilon {
		t.Errorf("got %v, want %v", degreesEmpty, kEpsilon)
	}

	if !float64Near(float64(lengthAc), totalLengthAbc, kEpsilon) {
		t.Errorf("got %v, want %v", lengthAc, totalLengthAbc)
	}

	if !float64Near(float64(lengthAbc), totalLengthAbc, kEpsilon) {
		t.Errorf("got %v, want %v", lengthAbc, totalLengthAbc)
	}

	if lengthBb >= kEpsilon {
		t.Errorf("got %v, want %v", lengthBb, kEpsilon)
	}

	if lengthCc >= kEpsilon {
		t.Errorf("got %v, want %v", lengthBb, kEpsilon)
	}

	if point.Angle(c.Vector) >= kEpsilon {
		t.Errorf("got %v, want %v", point, kEpsilon)
	}

	for i := 0; i < len(groundTruth); i++ {

	}
}
