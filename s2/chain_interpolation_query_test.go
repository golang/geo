package s2

import (
	"testing"

	"github.com/pavlov061356/geo/s1"
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
	a := PointFromLatLng(LatLngFromDegrees(0, 0))
	b := PointFromLatLng(LatLngFromDegrees(latitudeB, 0))
	c := PointFromLatLng(LatLngFromDegrees(latitudeC, 0))

	emptyLaxPolyline := Shape(&LaxPolyline{})
	acPolyline := Shape(&LaxPolyline{vertices: []Point{a, c}})
	abcPolyline := Shape(&LaxPolyline{vertices: []Point{a, b, c}})
	bbPolyline := Shape(&LaxPolyline{vertices: []Point{b, b}})
	ccPolyline := Shape(&LaxPolyline{vertices: []Point{c}})

	uninitializedQuery := ChainInterpolationQuery{}
	emptyQuery := InitChainInterpolationQuery(emptyLaxPolyline, 0)
	acQuery := InitChainInterpolationQuery(acPolyline, 0)
	abcQuery := InitChainInterpolationQuery(abcPolyline, 0)
	bbQuery := InitChainInterpolationQuery(bbPolyline, 0)
	ccQuery := InitChainInterpolationQuery(ccPolyline, 0)

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

	for _, args := range []struct {
		query   ChainInterpolationQuery
		want    float64
		wantErr bool
	}{
		{query: uninitializedQuery, want: 0, wantErr: true},
		{query: emptyQuery, want: 0, wantErr: true},
		{query: acQuery, want: totalLengthAbc, wantErr: false},
		{query: abcQuery, want: totalLengthAbc, wantErr: false},
		{query: bbQuery, want: 0, wantErr: false},
		{query: ccQuery, want: 0, wantErr: true},
	} {
		length, err := args.query.GetLength()
		if args.wantErr != (err != nil) {
			t.Fatalf("got %v, want %v", err, args.wantErr)
		}
		if !float64Near(length.Degrees(), args.want, kEpsilon) {
			t.Errorf("got %v, want %v", length.Degrees(), args.want)
		}
	}

	for _, args := range []struct {
		query         ChainInterpolationQuery
		totalFraction float64
		wantPoint     Point
		wantEdgeID    int
		wantDistance  s1.Angle
		wantErr       bool
	}{
		{query: uninitializedQuery, totalFraction: 0, wantPoint: Point{}, wantEdgeID: 0, wantDistance: 0, wantErr: true},
		{query: emptyQuery, totalFraction: 0, wantPoint: Point{}, wantEdgeID: 0, wantDistance: 0, wantErr: true},
		{query: acQuery, totalFraction: 0, wantPoint: a, wantEdgeID: 0, wantDistance: s1.Angle(0), wantErr: false},
		{query: abcQuery, totalFraction: 0, wantPoint: a, wantEdgeID: 0, wantDistance: s1.Angle(0), wantErr: false},
		{query: bbQuery, totalFraction: 0, wantPoint: b, wantEdgeID: 0, wantDistance: s1.Angle(0), wantErr: false},
		{query: ccQuery, totalFraction: 0, wantPoint: c, wantEdgeID: 0, wantDistance: 0, wantErr: true},
	} {
		distancePoint, distanceEdgeID, newDistance, err := args.query.AtFraction(args.totalFraction)
		if args.wantErr != (err != nil) {
			t.Fatalf("got %v, want %v", err, args.wantErr)
		}
		if distancePoint.Angle(args.wantPoint.Vector) >= kEpsilonAngle {
			t.Errorf("got %v, want %v", distancePoint, args.wantPoint.Vector)
		}
		if distanceEdgeID != args.wantEdgeID {
			t.Errorf("got %v, want %v", distanceEdgeID, args.wantEdgeID)
		}
		if !float64Near(newDistance.Radians(), args.wantDistance.Radians(), kEpsilon) {
			t.Errorf("got %v, want %v", newDistance, args.wantDistance)
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
	query := InitChainInterpolationQuery(&shape, 0)

	angle, err := query.GetLength()

	if err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
	length := angle.Degrees()

	results := make([]result, len(distances))
	for i := 0; i < len(distances); i++ {
		point, edgeID, distance, err := query.AtDistance(s1.Angle(distances[i] * float64(s1.Degree)))

		results[i] = result{point, edgeID, distance, err}
	}

	if !float64Near(length, totalLength, kEpsilon) {
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
			if !float64Eq(lat, LatLngFromPoint(shape.Edge(0).V0).Lat.Degrees()) {
				t.Errorf("got %v, want %v", lat, 0)
			}

			if edgeID != 0 {
				t.Errorf("got %v, want %v", edgeID, 0)
			}

			if !float64Eq(distance.Degrees(), 0) {
				t.Errorf("got %v, want %v", distance, 0)
			}
		} else if d > 2 {
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

	tests := []struct {
		query ChainInterpolationQuery
		want  result
		args  float64
	}{
		{
			query: InitChainInterpolationQuery(laxPolygon, -1),
			want: result{
				point:    PointFromLatLng(LatLngFromDegrees(1, 0)),
				edgeID:   1,
				distance: s1.Angle(1 * s1.Degree),
				err:      nil,
			},
			args: 0.25,
		},
		{
			query: InitChainInterpolationQuery(laxPolygon, 0),
			want: result{
				point:    PointFromLatLng(LatLngFromDegrees(0.5, 0)),
				edgeID:   0,
				distance: s1.Angle(0.5 * s1.Degree),
				err:      nil,
			},
			args: 0.25,
		},
		{
			query: InitChainInterpolationQuery(laxPolygon, 1),
			want: result{
				point:    PointFromLatLng(LatLngFromDegrees(2.5, 0)),
				edgeID:   2,
				distance: s1.Angle(2.5 * s1.Degree),
				err:      nil,
			},
			args: 0.25,
		},
	}

	for i, tt := range tests {
		point, edgeID, distance, err := tt.query.AtFraction(tt.args)

		got := result{
			point:    point,
			edgeID:   edgeID,
			distance: distance,
			err:      err,
		}

		if !float64Near(LatLngFromPoint(got.point).Lat.Degrees(), LatLngFromPoint(tt.want.point).Lat.Degrees(), kEpsilon) {
			t.Errorf("%d. got %v, want %v", i, LatLngFromPoint(got.point).Lat.Degrees(), LatLngFromPoint(tt.want.point).Lat.Degrees())
		}

		if got.edgeID != tt.want.edgeID {
			t.Errorf("%d. got %v, want %v", i, got.edgeID, tt.want.edgeID)
		}
		if got.err != tt.want.err {
			t.Errorf("%d. got %v, want %v", i, got.err, tt.want.err)
		}
	}
}

func TestGetLengthAtEdgeEmpty(t *testing.T) {
	query := InitChainInterpolationQuery(&laxPolyline{}, 0)

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

	polyline := laxPolyline{points}

	query := InitChainInterpolationQuery(&polyline, 0)

	tests := []struct {
		edgeID int
		want   s1.Angle
	}{
		{-100, s1.InfAngle()},
		{0, s1.Degree},
		{1, s1.Degree * 3},
		{2, s1.Degree * 6},
		{100, s1.InfAngle()},
	}

	for _, tt := range tests {
		got, err := query.GetLengthAtEdgeEnd(tt.edgeID)

		if err != nil {
			t.Errorf("edgeID %d: got %v, want %v", tt.edgeID, err, nil)
		}

		if tt.edgeID <= polyline.NumEdges() && tt.edgeID >= 0 {
			if !float64Near(got.Degrees(), tt.want.Degrees(), kEpsilon) {
				t.Errorf("edgeID %d: got %v, want %v", tt.edgeID, got, tt.want)
			}
		} else if got != tt.want {
			t.Errorf("edgeID %d: got %v, want %v", tt.edgeID, got, tt.want)
		}
	}
}
func TestGetLengthAtEdgePolygon(t *testing.T) {
	polygon := laxPolygonFromPoints([][]Point{
		{
			PointFromLatLng(LatLngFromDegrees(1, 1)),
			PointFromLatLng(LatLngFromDegrees(2, 1)),
			PointFromLatLng(LatLngFromDegrees(2, 3)),
			PointFromLatLng(LatLngFromDegrees(1, 3)),
		},
		{
			PointFromLatLng(LatLngFromDegrees(0, 0)),
			PointFromLatLng(LatLngFromDegrees(0, 4)),
			PointFromLatLng(LatLngFromDegrees(3, 4)),
			PointFromLatLng(LatLngFromDegrees(3, 0)),
		}})

	tolerance := .01

	tests := []struct {
		name string
		args struct {
			shape   Shape
			edge    int
			chainID int
		}
		want s1.Angle
	}{
		{
			name: "edge before first edge of first loop",
			args: struct {
				shape   Shape
				edge    int
				chainID int
			}{
				shape:   polygon,
				edge:    -100,
				chainID: 0,
			},
			want: s1.InfAngle(),
		},
		{
			name: "first edge of first loop",
			args: struct {
				shape   Shape
				edge    int
				chainID int
			}{
				shape:   polygon,
				edge:    0,
				chainID: 0,
			},
			want: s1.Degree,
		},
		{
			name: "second edge of first loop",
			args: struct {
				shape   Shape
				edge    int
				chainID int
			}{
				shape:   polygon,
				edge:    1,
				chainID: 0,
			},
			want: s1.Degree * 3,
		},
		{
			name: "last edge of first loop",
			args: struct {
				shape   Shape
				edge    int
				chainID int
			}{
				shape:   polygon,
				edge:    3,
				chainID: 0,
			},
			want: s1.Degree * 6,
		},
		{
			name: "edge after last edge of first loop",
			args: struct {
				shape   Shape
				edge    int
				chainID int
			}{
				shape:   polygon,
				edge:    4,
				chainID: 0,
			},
			want: s1.InfAngle(),
		},
		{
			name: "edge before first edge of second loop",
			args: struct {
				shape   Shape
				edge    int
				chainID int
			}{
				shape:   polygon,
				edge:    3,
				chainID: 1,
			},
			want: s1.InfAngle(),
		},
		{
			name: "first edge of second loop",
			args: struct {
				shape   Shape
				edge    int
				chainID int
			}{
				shape:   polygon,
				edge:    4,
				chainID: 1,
			},
			want: s1.Degree * 4,
		},
		{
			name: "second edge of second loop",
			args: struct {
				shape   Shape
				edge    int
				chainID int
			}{
				shape:   polygon,
				edge:    5,
				chainID: 1,
			},
			want: s1.Degree * 7,
		},
		{
			name: "midlle edge of second loop",
			args: struct {
				shape   Shape
				edge    int
				chainID int
			}{
				shape:   polygon,
				edge:    6,
				chainID: 1,
			},
			want: s1.Degree * 11,
		},
		{
			name: "last edge of second loop",
			args: struct {
				shape   Shape
				edge    int
				chainID int
			}{
				shape:   polygon,
				edge:    7,
				chainID: 1,
			},
			want: s1.Degree * 14,
		},
		{
			name: "edge after last edge of second loop",
			args: struct {
				shape   Shape
				edge    int
				chainID int
			}{
				shape:   polygon,
				edge:    8,
				chainID: 1,
			},
			want: s1.InfAngle(),
		},
	}

	for _, tt := range tests {
		query := InitChainInterpolationQuery(tt.args.shape, tt.args.chainID)

		got, err := query.GetLengthAtEdgeEnd(tt.args.edge)

		if err != nil {
			t.Errorf("%d. got %v, want %v", tt.args.edge, err, nil)
		}

		if tt.want == s1.InfAngle() {
			if got != tt.want {
				t.Errorf("%d. got %v, want %v", tt.args.edge, got, tt.want)
			}
		} else if !float64Near(got.Degrees(), tt.want.Degrees(), float64(tolerance)) {
			t.Errorf("%d. got %v, want %v", tt.args.edge, got.Degrees(), tt.want.Degrees())
		}
	}
}

func TestSlice(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			shape              Shape
			startSliceFraction float64
			endSliceFraction   float64
		}
		want string
	}{
		{
			name: "empty shape",
			args: struct {
				shape              Shape
				startSliceFraction float64
				endSliceFraction   float64
			}{nil, 0, 1},
			want: ``,
		},
		{
			name: "full polyline",
			args: struct {
				shape              Shape
				startSliceFraction float64
				endSliceFraction   float64
			}{laxPolylineFromPoints(parsePoints(`0:0, 0:1, 0:2`)), 0, 1},
			want: `0:0, 0:1, 0:2`,
		},
		{
			name: "first half of polyline",
			args: struct {
				shape              Shape
				startSliceFraction float64
				endSliceFraction   float64
			}{laxPolylineFromPoints(parsePoints(`0:0, 0:1, 0:2`)), 0, 0.5},
			want: `0:0, 0:1`,
		},
		{
			name: "second half of polyline",
			args: struct {
				shape              Shape
				startSliceFraction float64
				endSliceFraction   float64
			}{laxPolylineFromPoints(parsePoints(`0:0, 0:1, 0:2`)), 1, 0.5},
			want: `0:2, 0:1`,
		},
		{
			name: "middle of polyline",
			args: struct {
				shape              Shape
				startSliceFraction float64
				endSliceFraction   float64
			}{laxPolylineFromPoints(parsePoints(`0:0, 0:1, 0:2`)), 0.25, 0.75},
			want: `0:0.5, 0:1, 0:1.5`,
		},
	}

	for _, test := range tests {
		query := InitChainInterpolationQuery(test.args.shape, 0)
		if got := pointsToString(query.Slice(test.args.startSliceFraction, test.args.endSliceFraction)); got != test.want {
			t.Errorf("%v: got %v, want %v", test.name, got, test.want)
		}
	}
}

func TestSliceDivided(t *testing.T) {
	type args struct {
		shape              Shape
		startSliceFraction float64
		endSliceFraction   float64
		divisions          int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty shape",
			args: args{nil, 0, 1., 1},
			want: ``,
		},
		{name: "full polyline", args: args{
			shape: laxPolylineFromPoints([]Point{
				PointFromLatLng(LatLngFromDegrees(0, 0)),
				PointFromLatLng(LatLngFromDegrees(0, 1)),
				PointFromLatLng(LatLngFromDegrees(0, 2)),
			},
			),
			startSliceFraction: 0,
			endSliceFraction:   1,
			divisions:          3,
		}, want: `0:0, 0:1, 0:2`},
		{
			name: "first half of polyline",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
				},
				),
				startSliceFraction: 0,
				endSliceFraction:   0.5,
				divisions:          2,
			},
			want: `0:0, 0:1`,
		},
		{
			name: "second half of polyline",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
				},
				),
				startSliceFraction: 1,
				endSliceFraction:   0.5,
				divisions:          2,
			},
			want: `0:2, 0:1`,
		},
		{
			name: "middle of polyline",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
				},
				),
				startSliceFraction: 0.25,
				endSliceFraction:   0.75,
				divisions:          3,
			},
			want: `0:0.5, 0:1, 0:1.5`,
		},
		{
			name: "middle of polyline; divisions = 5",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
				},
				),
				startSliceFraction: 0.25,
				endSliceFraction:   0.75,
				divisions:          5,
			},
			want: `0:0.5, 0:0.75, 0:1, 0:1.25, 0:1.5`,
		},
		{
			name: "middle of polyline; divisions = 11",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
				},
				),
				startSliceFraction: 0.25,
				endSliceFraction:   0.75,
				divisions:          11,
			},
			want: `0:0.5, 0:0.6, 0:0.7, 0:0.8, 0:0.9, 0:1, 0:1.1, 0:1.2, 0:1.3, 0:1.4, 0:1.5`,
		},
		{
			name: "corner case: divisions = s.NumEdges()+1",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
				},
				),
				startSliceFraction: 0.3,
				endSliceFraction:   0.6,
				divisions:          4,
			},
			want: `0:0.6, 0:0.8, 0:1, 0:1.2`,
		},
		{
			name: "divisions = s.NumEdges()+2",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
				},
				),
				startSliceFraction: 0.25,
				endSliceFraction:   0.75,
				divisions:          5,
			},
			want: `0:0.5, 0:0.75, 0:1, 0:1.25, 0:1.5`,
		},
		{
			name: "divisions = s.NumEdges()+3",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
				},
				),
				startSliceFraction: 0.45,
				endSliceFraction:   0.75,
				divisions:          6,
			},
			want: `0:0.9, 0:1, 0:1.14, 0:1.26, 0:1.38, 0:1.5`,
		},
		{
			name: "divisions = s.NumEdges()+10",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
				},
				),
				startSliceFraction: 0.105,
				endSliceFraction:   0.605,
				divisions:          11,
			},
			want: `0:0.21, 0:0.31, 0:0.41, 0:0.51, 0:0.61, 0:0.71, 0:0.81, 0:0.91, 0:1, 0:1.11, 0:1.21`,
		},
		{
			name: "divisions = 10, 0 edges inside resulting points",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
				},
				),
				startSliceFraction: 0.05,
				endSliceFraction:   0.1,
				divisions:          11,
			},
			want: `0:0.1, 0:0.11, 0:0.12, 0:0.13, 0:0.14, 0:0.15, 0:0.16, 0:0.17, 0:0.18, 0:0.19, 0:0.2`,
		},
		{
			name: "divisions = s.NumEdges()+1",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
					PointFromLatLng(LatLngFromDegrees(0, 3)),
					PointFromLatLng(LatLngFromDegrees(0, 4)),
					PointFromLatLng(LatLngFromDegrees(0, 5)),
				},
				),
				startSliceFraction: 0.3,
				endSliceFraction:   0.84,
				divisions:          5,
			},
			want: `0:1.5, 0:2, 0:3, 0:4, 0:4.2`,
		},
		{
			name: "divisions = s.NumEdges()+1",
			args: args{
				shape: laxPolylineFromPoints([]Point{
					PointFromLatLng(LatLngFromDegrees(0, 0)),
					PointFromLatLng(LatLngFromDegrees(0, 1)),
					PointFromLatLng(LatLngFromDegrees(0, 2)),
				},
				),
				startSliceFraction: 0.3,
				endSliceFraction:   0.99999999999995,
				divisions:          3,
			},
			want: `0:0.6, 0:1, 0:1.9999999999999`,
		},
	}

	for _, test := range tests {
		query := InitChainInterpolationQuery(test.args.shape, 0)
		got := query.SliceDivided(
			test.args.startSliceFraction,
			test.args.endSliceFraction,
			test.args.divisions,
		)
		want := parsePoints(test.want)
		if len(got) != test.args.divisions && len(got) != len(want) {
			t.Errorf("length mismatch: got %d, want %d", len(got), test.args.divisions)
		}
		if !pointSlicesApproxEqual(got, want, kEpsilon) {
			t.Errorf("%v: got %v, want %v", test.name, got, want)
		}
	}
}

func Benchmark_SliceDivided(b *testing.B) {
	chainInterpolationQuery := InitChainInterpolationQuery(
		laxPolylineFromPoints(
			[]Point{
				PointFromLatLng(LatLngFromDegrees(0, 0)),
				PointFromLatLng(LatLngFromDegrees(0, 1)),
				PointFromLatLng(LatLngFromDegrees(0, 2)),
			},
		),
		0,
	)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		slice := chainInterpolationQuery.SliceDivided(0.3, 0.84, 500)
		if len(slice) != 500 {
			b.Errorf("length mismatch: got %d, want %d", len(slice), 500)
		}
	}

	b.StopTimer()

	points := make([]Point, 500)

	for i := 0; i < 100; i++ {
		points[i] = PointFromLatLng(LatLngFromDegrees(0, float64(i)))
	}

	chainInterpolationQuery = InitChainInterpolationQuery(
		laxPolylineFromPoints(
			points,
		),
		0,
	)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		slice := chainInterpolationQuery.SliceDivided(0.3, 0.84, 500)
		if len(slice) != 500 {
			b.Errorf("length mismatch: got %d, want %d", len(slice), 500)
		}
	}

}
