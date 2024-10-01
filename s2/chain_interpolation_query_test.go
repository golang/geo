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
		err      error
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

	point, _, _, err := uninitializedQuery.AtFraction(0)
	if err == nil {
		t.Fatalf("got %v, want error", point)
	}
	point, _, _, err = acQuery.AtDistance(s1.InfAngle())
	if err == nil {
		t.Fatalf("got %v, want error", point)
	}

	ac := make([]reslut, len(distances))
	abc := make([]reslut, len(distances))
	bb := make([]reslut, len(distances))
	cc := make([]reslut, len(distances))

	var emptyQueryValid bool

	for i, distance := range distances {
		totalFraction := distance / totalLengthAbc

		_, _, _, err := emptyQuery.AtFraction(totalFraction)

		emptyQueryValid = emptyQueryValid || (err == nil)

		distancePoint, distanceEdgeID, newDistance, err := acQuery.AtFraction(totalFraction)
		ac[i] = reslut{point: distancePoint, edgeID: distanceEdgeID, distance: newDistance, err: err}

		distancePoint, distanceEdgeID, newDistance, err = abcQuery.AtFraction(totalFraction)
		abc[i] = reslut{point: distancePoint, edgeID: distanceEdgeID, distance: newDistance, err: err}

		distancePoint, distanceEdgeID, newDistance, err = bbQuery.AtFraction(totalFraction)
		bb[i] = reslut{point: distancePoint, edgeID: distanceEdgeID, distance: newDistance, err: err}

		distancePoint, distanceEdgeID, newDistance, err = ccQuery.AtFraction(totalFraction)
		cc[i] = reslut{point: distancePoint, edgeID: distanceEdgeID, distance: newDistance, err: err}
	}

	if emptyQueryValid {
		t.Errorf("got %v, want %v", emptyQueryValid, false)
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

		if ac[i].err != nil {
			t.Errorf("got %v, want %v", ac[i].err, nil)
		}

		if abc[i].err != nil {
			t.Errorf("got %v, want %v", abc[i].err, nil)
		}

		if bb[i].err != nil {
			t.Errorf("got %v, want %v", bb[i].err, nil)
		}

		if cc[i].err == nil {
			t.Errorf("got %v, want %v", cc[i].err, nil)
		}

		if ac[i].point.Angle(groundTruth[i].point.Vector) >= kEpsilonAngle {
			t.Errorf("got %v, want %v", ac[i].point, kEpsilonAngle)
		}

		if abc[i].point.Angle(groundTruth[i].point.Vector) >= kEpsilonAngle {
			t.Errorf("got %v, want %v", abc[i].point, kEpsilonAngle)
		}

		if bb[i].point.Angle(bbPolyline.Edge(i).V1.Vector) >= kEpsilonAngle {
			t.Errorf("got %v, want %v", bb[i].point, kEpsilonAngle)
		}

		if ac[i].edgeID != 0 {
			t.Errorf("got %v, want %v", ac[i].edgeID, 0)
		}

		if bb[i].edgeID != 0 {
			t.Errorf("got %v, want %v", bb[i].edgeID, 0)
		}

		if abc[i].edgeID != groundTruth[i].edgeID {
			t.Errorf("got %v, want %v", abc[i].edgeID, groundTruth[i].edgeID)
		}
	}
}
