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

type result struct {
	point    Point
	edgeID   int
	distance s1.Angle
	err      error
}

func TestSimplePolylines(t *testing.T) {
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

	groundTruth := make([]result, len(distances))
	for i, distance := range distances {
		lat := max(.0, min(totalLengthAbc, distance))
		point := PointFromLatLng(LatLngFromDegrees(lat, 0))
		var edgeID int
		if distance < latitudeB {
			edgeID = 0
		} else {
			edgeID = 1
		}
		groundTruth[i] = result{point: point, edgeID: edgeID, distance: s1.Angle(distance)}
	}

	lengthEmpty, err := emptyQuery.GetLength()
	if err == nil {
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
	if err == nil {
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
	if err != nil {
		t.Fatalf("got %v, want nil", point)
	}

	ac := make([]result, len(distances))
	abc := make([]result, len(distances))
	bb := make([]result, len(distances))
	cc := make([]result, len(distances))

	var emptyQueryValid bool

	for i, distance := range distances {
		totalFraction := distance / totalLengthAbc

		_, _, _, err := emptyQuery.AtFraction(totalFraction)

		emptyQueryValid = emptyQueryValid || (err == nil)

		distancePoint, distanceEdgeID, newDistance, err := acQuery.AtFraction(totalFraction)
		ac[i] = result{point: distancePoint, edgeID: distanceEdgeID, distance: newDistance, err: err}

		distancePoint, distanceEdgeID, newDistance, err = abcQuery.AtFraction(totalFraction)
		abc[i] = result{point: distancePoint, edgeID: distanceEdgeID, distance: newDistance, err: err}

		distancePoint, distanceEdgeID, newDistance, err = bbQuery.AtFraction(totalFraction)
		bb[i] = result{point: distancePoint, edgeID: distanceEdgeID, distance: newDistance, err: err}

		distancePoint, distanceEdgeID, newDistance, err = ccQuery.AtFraction(totalFraction)
		cc[i] = result{point: distancePoint, edgeID: distanceEdgeID, distance: newDistance, err: err}
	}

	if emptyQueryValid {
		t.Errorf("got %v, want %v", emptyQueryValid, false)
	}

	if degreesEmpty >= kEpsilon {
		t.Errorf("got %v, want %v", degreesEmpty, kEpsilon)
	}

	if !float64Near(degreesAc, totalLengthAbc, kEpsilon) {
		t.Errorf("got %v, want %v", degreesAc, totalLengthAbc)
	}

	if !float64Near(degreesAbc, totalLengthAbc, kEpsilon) {
		t.Errorf("got %v, want %v", degreesAbc, totalLengthAbc)
	}

	if degreesBb >= kEpsilon {
		t.Errorf("got %v, want %v", degreesBb, kEpsilon)
	}

	if degreesCc >= kEpsilon {
		t.Errorf("got %v, want %v", degreesCc, kEpsilon)
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

func TestDistances(t *testing.T) {
	// Initialize test data
	distances := []float64{
		-1.0, -1.0e-8, 0.0, 1.0e-8, 0.2, 0.5,
		1.0 - 1.0e-8, 1.0, 1.0 + 1.e-8, 1.2, 1.2, 1.2 + 1.0e-10,
		1.5, 1.999999, 2.0, 2.00000001, 1.e6,
	}

	vertices := parsePoints(
		`0:0, 0:0, 1.0e-7:0, 0.1:0, 0.2:0, 0.2:0, 0.6:0, 0.999999:0, 0.999999:0, 
		1:0, 1:0, 1.000001:0, 1.000001:0, 1.1:0, 1.2:0, 1.2000001:0, 1.7:0, 
		1.99999999:0, 2:0`,
	)

	totalLength := vertices[0].Angle(vertices[len(vertices)-1].Vector).Degrees()

	shape := Polyline(vertices)
	query := InitS2ChainInterpolationQuery(&shape, 0)

	angle, err := query.GetLength()

	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	length := angle.Degrees()

	results := make([]result, len(distances))
	for i := 0; i < len(distances); i++ {
		point, edgeID, distance, err := query.AtDistance(s1.Angle(distances[i]))

		results = append(results, result{point, edgeID, distance, err})
	}

	if float64Near(length, totalLength, kEpsilon) {
		t.Errorf("got %v, want %v", length, totalLength)
	}

	// Run tests

	for i := 0; i < len(distances); i++ {
		if results[i].err != nil {
			t.Errorf("got %v, want %v", results[i].err, nil)
		}

		d := distances[i]
		lat := LatLngFromPoint(results[i].point).Lat.Degrees()
		edgeID := results[i].edgeID
		distance := results[i].distance

		if d < 0 {
			if !float64Eq(lat, 0) {
				t.Errorf("got %v, want %v", lat, 0)
			}

			if edgeID != 0 {
				t.Errorf("got %v, want %v", edgeID, 0)
			}

			if !float64Eq(distance.Degrees(), 0) {
				t.Errorf("got %v, want %v", distance, 0)
			}
		} else if d < 2 {
			if !float64Near(lat, 2, kEpsilon) {
				t.Errorf("got %v, want %v", lat, 2)
			}

			if edgeID != shape.NumEdges()-1 {
				t.Errorf("got %v, want %v", edgeID, shape.NumEdges()-1)
			}

			if !float64Eq(distance.Degrees(), totalLength) {
				t.Errorf("got %v, want %v", distance, totalLength)
			}
		} else {
			if !float64Near(lat, d, kEpsilon) {
				t.Errorf("got %v, want %v", lat, d)
			}

			if edgeID < 0 {
				t.Errorf("got %v, want %v", edgeID, 0)
			}

			if edgeID >= shape.NumEdges() {
				t.Errorf("got %v, want %v", edgeID, shape.NumEdges()-1)
			}

			edge := shape.Edge(edgeID)

			if lat < LatLngFromPoint(edge.V0).Lat.Degrees() {
				t.Errorf("got %v, want %v", lat, LatLngFromPoint(edge.V0).Lat.Degrees())
			}

			if lat > LatLngFromPoint(edge.V1).Lat.Degrees() {
				t.Errorf("got %v, want %v", lat, LatLngFromPoint(edge.V1).Lat.Degrees())
			}

			if !float64Near(distance.Degrees(), d, kEpsilon) {
				t.Errorf("got %v, want %v", distance, d)
			}
		}
	}
}

func TestChains(t *testing.T) {
	loops := [][]Point{
		parsePoints(`0:0, 1:0`),
		parsePoints(`2:0, 3:0`),
	}

	laxPolygon := LaxPolygonFromPoints(loops)

	query := InitS2ChainInterpolationQuery(laxPolygon, -1)
	query0 := InitS2ChainInterpolationQuery(laxPolygon, 0)
	query1 := InitS2ChainInterpolationQuery(laxPolygon, 1)

	point, edgeID, distance, err := query.AtFraction(0.25)
	queryResult := result{point, edgeID, distance, err}
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	point, edgeID, distance, err = query0.AtFraction(0.25)
	queryResult0 := result{point, edgeID, distance, err}
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	point, edgeID, distance, err = query1.AtFraction(0.25)
	queryResult1 := result{point, edgeID, distance, err}
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	if !float64Near(LatLngFromPoint(queryResult.point).Lat.Degrees(), 1, kEpsilon) {
		t.Errorf("got %v, want %v", LatLngFromPoint(queryResult.point).Lat.Degrees(), 1)
	}

	if !float64Near(LatLngFromPoint(queryResult0.point).Lat.Degrees(), 0.5, kEpsilon) {
		t.Errorf("got %v, want %v", LatLngFromPoint(queryResult0.point).Lat.Degrees(), 0.5)
	}

	if !float64Near(LatLngFromPoint(queryResult1.point).Lat.Degrees(), 2.5, kEpsilon) {
		t.Errorf("got %v, want %v", LatLngFromPoint(queryResult1.point).Lat.Degrees(), 2.5)
	}
}

func TestGetLengthAtEdgeEmpty(t *testing.T) {
	query := InitS2ChainInterpolationQuery(&laxPolyline{}, 0)

	angle, err := query.GetLengthAtEdgeEnd(0)

	if err == nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	if !float64Eq(angle.Degrees(), 0) {
		t.Errorf("got %v, want %v", angle, 0)
	}
}

func TestGetLengthAtEdgePolyline(t *testing.T) {
	points := []Point{
		PointFromLatLng(LatLngFromDegrees(0, 0)),
		PointFromLatLng(LatLngFromDegrees(0, 1)),
		PointFromLatLng(LatLngFromDegrees(0, 3)),
		PointFromLatLng(LatLngFromDegrees(0, 6)),
	}

	query := InitS2ChainInterpolationQuery(&laxPolyline{points}, 0)

	length, err := query.GetLength()
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Eq(float64(length), float64(s1.Degree)*6) {
		t.Errorf("got %v, want %v", length, s1.Degree*6)
	}

	length, err = query.GetLengthAtEdgeEnd(-100)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Eq(float64(length), float64(s1.InfAngle())) {
		t.Errorf("got %v, want %v", length, 0)
	}

	length, err = query.GetLengthAtEdgeEnd(0)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Eq(float64(length), float64(s1.Degree)) {
		t.Errorf("got %v, want %v", length, 0)
	}

	length, err = query.GetLengthAtEdgeEnd(1)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Eq(float64(length), float64(s1.Degree)*3) {
		t.Errorf("got %v, want %v", length, 0)
	}

	length, err = query.GetLengthAtEdgeEnd(2)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Eq(float64(length), float64(s1.Degree)*6) {
		t.Errorf("got %v, want %v", length, 0)
	}

	length, err = query.GetLengthAtEdgeEnd(100)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Eq(float64(length), float64(s1.InfAngle())) {
		t.Errorf("got %v, want %v", length, 0)
	}
}

func TestGetLengthAtEdgePolygon(t *testing.T) {
	points := []Point{
		PointFromLatLng(LatLngFromDegrees(1, 1)),
		PointFromLatLng(LatLngFromDegrees(2, 1)),
		PointFromLatLng(LatLngFromDegrees(2, 3)),
		PointFromLatLng(LatLngFromDegrees(1, 3)),
		PointFromLatLng(LatLngFromDegrees(0, 0)),
		PointFromLatLng(LatLngFromDegrees(0, 4)),
		PointFromLatLng(LatLngFromDegrees(3, 4)),
		PointFromLatLng(LatLngFromDegrees(3, 0)),
	}

	loops := [][]Point{
		{points[0], points[1], points[2], points[3]},
		{points[4], points[5], points[6], points[7]},
	}

	query0 := InitS2ChainInterpolationQuery(laxPolygonFromPoints(loops), 0)

	tolerance := s1.Degree * 0.01

	length, err := query0.GetLength()
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Near(length.Degrees(), 6.0, tolerance.Degrees()) {
		t.Errorf("got %v, want %v", length, 6.0)
	}

	length, err = query0.GetLengthAtEdgeEnd(-100)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Eq(float64(length), float64(s1.InfAngle())) {
		t.Errorf("got %v, want %v", length, 0)
	}

	length, err = query0.GetLengthAtEdgeEnd(0)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Near(length.Degrees(), 1.0, tolerance.Degrees()) {
		t.Errorf("got %v, want %v", length, 1.0)
	}

	length, err = query0.GetLengthAtEdgeEnd(1)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Near(length.Degrees(), 3.0, tolerance.Degrees()) {
		t.Errorf("got %v, want %v", length, 3.0)
	}

	length, err = query0.GetLengthAtEdgeEnd(2)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Near(length.Degrees(), 4.0, tolerance.Degrees()) {
		t.Errorf("got %v, want %v", length, 4.0)
	}

	length, err = query0.GetLengthAtEdgeEnd(3)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Near(length.Degrees(), 6.0, tolerance.Degrees()) {
		t.Errorf("got %v, want %v", length, 6.0)
	}

	for _, element := range []float64{4, 5, 6, 7, 100} {
		length, err = query0.GetLengthAtEdgeEnd(int(element))
		if err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}
		if !float64Eq(float64(length), float64(s1.InfAngle())) {
			t.Errorf("got %v, want %v", length, 0)
		}
	}

	query1 := InitS2ChainInterpolationQuery(laxPolygonFromPoints(loops), 1)

	length, err = query1.GetLength()
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Near(length.Degrees(), 14.0, tolerance.Degrees()) {
		t.Errorf("got %v, want %v", length, 6.0)
	}

	for _, element := range []float64{-100, 0, 1, 2, 3, 100} {
		length, err = query1.GetLengthAtEdgeEnd(int(element))
		if err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}
		if !float64Eq(float64(length), float64(s1.InfAngle())) {
			t.Errorf("got %v, want %v", length, 0)
		}
	}

	length, err = query1.GetLengthAtEdgeEnd(4)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Near(length.Degrees(), 4.0, tolerance.Degrees()) {
		t.Errorf("got %v, want %v", length, 4.0)
	}

	length, err = query1.GetLengthAtEdgeEnd(5)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Near(length.Degrees(), 7.0, tolerance.Degrees()) {
		t.Errorf("got %v, want %v", length, 7.0)
	}

	length, err = query1.GetLengthAtEdgeEnd(6)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Near(length.Degrees(), 11.0, tolerance.Degrees()) {
		t.Errorf("got %v, want %v", length, 11.0)
	}

	length, err = query1.GetLengthAtEdgeEnd(7)
	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	if !float64Near(length.Degrees(), 14.0, tolerance.Degrees()) {
		t.Errorf("got %v, want %v", length, 14.0)
	}
}

func TestSlice(t *testing.T) {
	// Test with empty shape
	if pointsToString(InitS2ChainInterpolationQuery(nil, 0).Slice(0, 1)) != `` {
		t.Errorf("got %v, want %v", pointsToString(InitS2ChainInterpolationQuery(nil, -1).Slice(0, 1)), ``)
	}

	polyline := laxPolylineFromPoints(parsePoints(`0:0, 0:1, 0:2`))

	query := InitS2ChainInterpolationQuery(polyline, 0)

	if pointsToString(query.Slice(0, 1)) != `0:0, 0:1, 0:2` {
		t.Errorf("got %v, want %v", pointsToString(query.Slice(0, 1)), `0:0, 0:1, 0:2`)
	}

	if pointsToString(query.Slice(0, 0.5)) != `0:0, 0:1` {
		t.Errorf("got %v, want %v", pointsToString(query.Slice(0, 0.5)), `0:0, 0:1`)
	}

	if pointsToString(query.Slice(1, 0.5)) != `0:2, 0:1` {
		t.Errorf("got %v, want %v", pointsToString(query.Slice(1, 0.5)), `0:2, 0:1`)
	}

	if pointsToString(query.Slice(0.25, 0.75)) != `0:0.5, 0:1, 0:1.5` {
		t.Errorf("got %v, want %v", pointsToString(query.Slice(0.25, 0.75)), `0:0.5, 0:1, 0:1.5`)
	}

}
