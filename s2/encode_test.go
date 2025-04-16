// Copyright 2017 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s2

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"reflect"
	"testing"

	"github.com/golang/geo/r3"
)

type encodableRegion interface {
	Encode(w io.Writer) error
}

type decodableRegion interface {
	Decode(r io.Reader) error
}

const (

	// encodedCapEmpty comes from EmptyCap()
	encodedCapEmpty = "000000000000F03F00000000000000000000000000000000000000000000F0BF"
	// encodedCapFull comes from FullCap()
	encodedCapFull = "000000000000F03F000000000000000000000000000000000000000000001040"
	// Cap from Point(3, 2, 1).Normalize()
	encodedCapFromPoint = "3F36105836A8E93F2A2460E5CE1AE13F2A2460E5CE1AD13F0000000000000000"
	// Cap from Point(0, 0, 1) with height 5
	encodedCapFromCenterHeight = "00000000000000000000000000000000000000000000F03F0000000000001040"
	// Cap from Point(0, 0, 1) with height 0.5
	encodedCapFromCenterHeight2 = "00000000000000000000000000000000000000000000F03F000000000000F03F"
	// CellID from Face 0.
	encodedCellIDFace0 = "0000000000000010"
	// CellID from Face 5.
	encodedCellIDFace5 = "00000000000000B0"
	// CellID from Face 0 in the last Cell at MaxLevel.
	encodedCellIDFace0MaxLevel = "0100000000000020"
	// CellID from Face 5 in the last Cell at MaxLevel.
	encodedCellIDFace5MaxLevel = "01000000000000C0"
	// CellID FromFacePosLevel(3, 0x12345678, MaxLevel - 4)
	encodedCellIDFacePosLevel = "0057341200000060"
	// CellID from the 0 value.
	encodedCellIDInvalid = "0000000000000000"

	// Cell from Point(1, 2, 3)
	encodedCellFromPoint = "F51392E0F35DCC43"
	// Cell from (39.0, -120.0) - The Lake Tahoe border corner of CA/NV.
	encodedCellFromLatLng = "6308962A95849980"
	// Cell FromFacePosLevel(3, 0x12345678, MaxLevel - 4)
	encodedCellFromFacePosLevel = "0057341200000060"
	// Cell from Face 0.
	encodedCellFace0 = "0000000000000010"

	// An uninitialized empty CellUnion.
	encodedCellUnionEmpty = "010000000000000000"
	// CellUnion from a CellID from Face 1.
	encodedCellUnionFace1 = "0101000000000000000000000000000030"
	// CellUnion from the cells {0x33, 0x8e3748fab, 0x91230abcdef83427};
	encodedCellUnionFromCells = "0103000000000000003300000000000000AB8F74E3080000002734F8DEBC0A2391"

	// Loop
	encodedLoopEmpty = "010100000000000000000000000000000000000000000000000000F03F000000000001000000000000F03F0000000000000000182D4454FB210940182D4454FB2109C0"
	encodedLoopFull  = "010100000000000000000000000000000000000000000000000000F0BF010000000001182D4454FB21F9BF182D4454FB21F93F182D4454FB2109C0182D4454FB210940"
	// Loop from the unit test value kCross1;
	encodedLoopCross = "0108000000D44A8442C3F9EF3F7EDA2AB341DC913F27DCF7C958DEA1BFB4825F3C81FDEF3F27DCF7C958DE913F1EDD892B0BDF91BFB4825F3C81FDEF3F27DCF7C958DE913F1EDD892B0BDF913FD44A8442C3F9EF3F7EDA2AB341DC913F27DCF7C958DEA13FD44A8442C3F9EF3F7EDA2AB341DC91BF27DCF7C958DEA13FB4825F3C81FDEF3F27DCF7C958DE91BF1EDD892B0BDF913FB4825F3C81FDEF3F27DCF7C958DE91BF1EDD892B0BDF91BFD44A8442C3F9EF3F7EDA2AB341DC91BF27DCF7C958DEA1BF0000000000013EFC10E8F8DFA1BF3EFC10E8F8DFA13F389D52A246DF91BF389D52A246DF913F"
	// Loop encoded using EncodeCompressed from snapped points.
	//
	//       CellIDFromLatLng("0:178")).ToPoint(),
	//       CellIDFromLatLng("-1:180")).ToPoint(),
	//       CellIDFromLatLng("0:-179")).ToPoint(),
	//       CellIDFromLatLng("1:-180")).ToPoint()};
	// LoopFromPoints((snapped_loop_a_vertices));
	encodedLoopCompressed = "041B02222082A222A806A0C7A991DE86D905D7C3A691F2DEE40383908880A0958805000003"

	// OriginPoint()
	encodedPointOrigin = "013BED86AA997A84BF88EC8B48C53C653FACD2721A90FFEF3F"
	// Point(12.34, 56.78, 9.1011).Normalize()
	encodedPointTesting = "0109AD578332DBCA3FBC9FDB9BB4E4EE3FE67E7C2CA7CEC33F"

	// Polygon from makePolygon("").
	// This is encoded in compressed format.
	encodedPolygonEmpty = "041E00"
	// Polygon from makePolygon("full").
	// This is encoded in compressed format.
	encodedPolygonFull = "040001010B000100"
	// Loop from the unit test value cross1. This is encoded in lossless format.
	encodedPolygon1Loops = "010100010000000108000000D44A8442C3F9EF3F7EDA2AB341DC913F27DCF7C958DEA1BFB4825F3C81FDEF3F27DCF7C958DE913F1EDD892B0BDF91BFB4825F3C81FDEF3F27DCF7C958DE913F1EDD892B0BDF913FD44A8442C3F9EF3F7EDA2AB341DC913F27DCF7C958DEA13FD44A8442C3F9EF3F7EDA2AB341DC91BF27DCF7C958DEA13FB4825F3C81FDEF3F27DCF7C958DE91BF1EDD892B0BDF913FB4825F3C81FDEF3F27DCF7C958DE91BF1EDD892B0BDF91BFD44A8442C3F9EF3F7EDA2AB341DC91BF27DCF7C958DEA1BF0000000000013EFC10E8F8DFA1BF3EFC10E8F8DFA13F389D52A246DF91BF389D52A246DF913F013EFC10E8F8DFA1BF3EFC10E8F8DFA13F389D52A246DF91BF389D52A246DF913F"
	// Loop from the unit test value cross1+crossHole.
	// This is encoded in lossless format.
	encodedPolygon2Loops = "010101020000000108000000D44A8442C3F9EF3F7EDA2AB341DC913F27DCF7C958DEA1BFB4825F3C81FDEF3F27DCF7C958DE913F1EDD892B0BDF91BFB4825F3C81FDEF3F27DCF7C958DE913F1EDD892B0BDF913FD44A8442C3F9EF3F7EDA2AB341DC913F27DCF7C958DEA13FD44A8442C3F9EF3F7EDA2AB341DC91BF27DCF7C958DEA13FB4825F3C81FDEF3F27DCF7C958DE91BF1EDD892B0BDF913FB4825F3C81FDEF3F27DCF7C958DE91BF1EDD892B0BDF91BFD44A8442C3F9EF3F7EDA2AB341DC91BF27DCF7C958DEA1BF0000000000013EFC10E8F8DFA1BF3EFC10E8F8DFA13F389D52A246DF91BF389D52A246DF913F0104000000C5D7FA4B60FFEF3F1EDD892B0BDF813F214C95C437DF81BFC5D7FA4B60FFEF3F1EDD892B0BDF813F214C95C437DF813FC5D7FA4B60FFEF3F1EDD892B0BDF81BF214C95C437DF813FC5D7FA4B60FFEF3F1EDD892B0BDF81BF214C95C437DF81BF000100000001900C5E3B73DF81BF900C5E3B73DF813F399D52A246DF81BF399D52A246DF813F013EFC10E8F8DFA1BF3EFC10E8F8DFA13F389D52A246DF91BF389D52A246DF913F"
	// TODO(roberts): Create Polygons that use compressed encoding.

	// A Polyline from an empty slice.
	encodedPolylineEmpty = "0100000000"
	// A Polyline from 3 LatLngs {(0, 0),(0, 90),(0,180)};
	// TODO(roberts): The next two goldens differ from the C++ in the last few bits of the
	// IEEE 754 values.
	//
	// When converting the LatLng (0, 90) to a Point.
	//
	// Go:  cos(theta)*cos(phi) = 6.12323399573675740770266929248e-17
	// C++: cos(theta)*cos(phi) = 6.12323399573676603586882014729e-17
	//
	//   want: 005C143326A6913C
	//   got:  075C143326A6913C
	//   diff:  ^
	//
	// C++ golden: 0103000000000000000000F03F00000000000000000000000000000000075C143326A6913C000000000000F03F0000000000000000000000000000F0BF075C143326A6A13C0000000000000000
	encodedPolylineSemiEquator = "0103000000000000000000F03F00000000000000000000000000000000005C143326A6913C000000000000F03F0000000000000000000000000000F0BF005C143326A6A13C0000000000000000"

	// A Polyline from makePolyline("0:0, 0:10, 10:20, 20:30");
	// See comment above for why this golden differs from the C++ golden.
	// C++ golden: 0104000000000000000000F03F00000000000000000000000000000000171C818C8B83EF3F89730B7E1A3AC63F000000000000000061B46C3A039DED3FE2DC829F868ED53F89730B7E1A3AC63F1B995E6FA10AEA3F1B2D5242F611DE3FF50B8A74A8E3D53F
	encodedPolyline3Segments = "0104000000000000000000F03F00000000000000000000000000000000181C818C8B83EF3F89730B7E1A3AC63F000000000000000062B46C3A039DED3FE2DC829F868ED53F89730B7E1A3AC63F1B995E6FA10AEA3F1B2D5242F611DE3FF50B8A74A8E3D53F"

	// Rect from EmptyRect
	encodedRectEmpty = "01000000000000F03F0000000000000000182D4454FB210940182D4454FB2109C0"
	// Rect from FullRect
	encodedRectFull = "01182D4454FB21F9BF182D4454FB21F93F182D4454FB2109C0182D4454FB210940"
	// Rect from Center=(80,170), Size=(40,60)
	encodedRectCentersize = "0165732D3852C1F03F182D4454FB21F93FF75B8A41358C03408744E74A185706C0"

	// R2Rect - Not yet implemented.
	// RegionIntersection - Not yet implemented.
	// RegionUnion - Not yet implemented.
)

func TestEncodeDecode(t *testing.T) {
	cu := CellUnion{}
	cuFace := CellUnion([]CellID{CellIDFromFace(1)})
	cuCells := CellUnion([]CellID{
		CellID(0x33),
		CellID(0x8e3748fab),
		CellID(0x91230abcdef83427),
	})

	capPtr := func(c Cap) *Cap { return &c }
	cidPtr := func(c CellID) *CellID { return &c }
	cellPtr := func(c Cell) *Cell { return &c }
	ptPtr := func(pt Point) *Point { return &pt }
	rectPtr := func(r Rect) *Rect { return &r }

	// Polyline inputs
	// semiEquator := Polyline([]Point{
	//	PointFromLatLng(LatLngFromDegrees(0, 0)),
	// 	PointFromLatLng(LatLngFromDegrees(0, 90)),
	// 	PointFromLatLng(LatLngFromDegrees(0, 180)),
	// })
	// threeSegments := makePolyline("0:0, 0:10, 10:20, 20:30")

	const cross1 = "-2:1, -1:1, 1:1, 2:1, 2:-1, 1:-1, -1:-1, -2:-1"
	const crossCenterHole = "-0.5:0.5, 0.5:0.5, 0.5:-0.5, -0.5:-0.5;"

	emptyPolygon := func() *Polygon {
		p := &Polygon{loops: []*Loop{}, bound: EmptyRect(), subregionBound: EmptyRect()}
		p.initEdgesAndIndex()
		return p
	}

	tests := []struct {
		golden string
		reg    encodableRegion
	}{
		// Caps
		{encodedCapEmpty, capPtr(EmptyCap())},
		{encodedCapFull, capPtr(FullCap())},
		{encodedCapFromPoint, capPtr(CapFromPoint(PointFromCoords(3, 2, 1)))},
		{encodedCapFromCenterHeight, capPtr(CapFromCenterHeight(PointFromCoords(0, 0, 1), 5))},
		{encodedCapFromCenterHeight2, capPtr(CapFromCenterHeight(PointFromCoords(0, 0, 1), 0.5))},

		// CellIDs
		{encodedCellIDFace0, cidPtr(CellIDFromFace(0))},
		{encodedCellIDFace5, cidPtr(CellIDFromFace(5))},
		{encodedCellIDFace0MaxLevel, cidPtr(CellIDFromFace(0).ChildEndAtLevel(MaxLevel))},
		{encodedCellIDFace5MaxLevel, cidPtr(CellIDFromFace(5).ChildEndAtLevel(MaxLevel))},
		{encodedCellIDFacePosLevel, cidPtr(CellIDFromFacePosLevel(3, 0x12345678, MaxLevel-4))},
		{encodedCellIDInvalid, cidPtr(CellID(0))},

		// Cells
		{encodedCellFromPoint, cellPtr(CellFromPoint(Point{r3.Vector{X: 1, Y: 2, Z: 3}}))},
		// Lake Tahoe CA/NV border corner
		{encodedCellFromLatLng, cellPtr(CellFromLatLng(LatLngFromDegrees(39.0, -120.0)))},
		{encodedCellFromFacePosLevel, cellPtr(CellFromCellID(CellIDFromFacePosLevel(3, 0x12345678, MaxLevel-4)))},
		{encodedCellFace0, cellPtr(CellFromCellID(CellIDFromFace(0)))},

		// CellUnions
		{encodedCellUnionEmpty, &cu},
		{encodedCellUnionFace1, &cuFace},
		{encodedCellUnionFromCells, &cuCells},

		// Loops
		{encodedLoopEmpty, EmptyLoop()},
		{encodedLoopFull, FullLoop()},
		{encodedLoopCross, LoopFromPoints(parsePoints(cross1))},

		// Points
		{encodedPointOrigin, ptPtr(OriginPoint())},
		{encodedPointTesting, ptPtr(PointFromCoords(12.34, 56.78, 9.1011))},

		// Polygons.
		{encodedPolygonEmpty, emptyPolygon()},
		{encodedPolygonFull, FullPolygon()},
		{encodedPolygon1Loops, makePolygon(cross1, false)},
		{encodedPolygon2Loops, makePolygon(cross1+";"+crossCenterHole, false)},

		// Polylines
		{encodedPolylineEmpty, (&Polyline{})},
		// TODO(nsch): Comment these lines back in once all decoders are implemented.
		// Then, switch the test from encode->decode->deepequal to
		// decode->approxequal->encode
		// {encodedPolylineSemiEquator, semiEquator},
		// {encodedPolyline3Segments, threeSegments},

		// Rects
		{encodedRectEmpty, rectPtr(EmptyRect())},
		{encodedRectFull, rectPtr(FullRect())},
		{encodedRectCentersize, rectPtr(RectFromCenterSize(LatLngFromDegrees(80, 170), LatLngFromDegrees(40, 60)))},
	}

	for _, test := range tests {
		// Test encode.
		buf := new(bytes.Buffer)
		if err := test.reg.Encode(buf); err != nil {
			t.Errorf("error encoding %v: %v", test.reg, err)
		}

		encoded := fmt.Sprintf("%X", buf.Bytes())
		if test.golden != encoded {
			t.Errorf("%#v.Encode() = %q, want %q", test.reg, encoded, test.golden)
		}

		// Create target for decoding.
		decoded := reflect.New(reflect.TypeOf(test.reg).Elem()).Interface().(decodableRegion)

		if err := decoded.Decode(buf); err != nil {
			t.Errorf("decode(%#v): %v", test.reg, err)
			continue
		}
		if !reflect.DeepEqual(decoded, test.reg) {
			t.Errorf("decode = %#v, want %#v", decoded, test.reg)
		}
	}
}

func TestDecodeCompressedLoop(t *testing.T) {
	dat, err := hex.DecodeString(encodedLoopCompressed)
	if err != nil {
		t.Fatal(err)
	}
	d := &decoder{r: bytes.NewReader(dat)}
	gotDecoded := new(Loop)
	gotDecoded.decodeCompressed(d, MaxLevel)
	if d.err != nil {
		t.Fatalf("loop.decodeCompressed: %v", d.err)
	}
	wantDecoded := []LatLng{LatLngFromDegrees(0, 178), LatLngFromDegrees(-1, 180), LatLngFromDegrees(0, -179), LatLngFromDegrees(1, -180)}
	for i, v := range gotDecoded.Vertices() {
		got := LatLngFromPoint(v)
		want := wantDecoded[i]
		const margin = 1e-9
		if math.Abs((got.Lat-want.Lat).Radians()) >= margin || math.Abs((got.Lng-want.Lng).Radians()) >= margin {
			t.Errorf("decoding golden at %d = %v, want %v", i, got, want)
		}
	}
	var buf bytes.Buffer
	e := &encoder{w: &buf}
	gotDecoded.encodeCompressed(e, MaxLevel, gotDecoded.xyzFaceSiTiVertices())
	if e.err != nil {
		t.Fatalf("encodeCompressed(decodeCompressed(loop)): %v", err)
	}
	gotReencoded := fmt.Sprintf("%X", buf.Bytes())
	if gotReencoded != encodedLoopCompressed {
		t.Errorf("encodeCompressed(decodeCompressed(loop)) = %q, want %q", gotReencoded, encodedLoopCompressed)
	}
}

// Captures the uncompressed path.
func TestLoopEncodeDecode(t *testing.T) {
	pts := parsePoints("30:20, 40:20, 39:43, 33:35")
	loops := []*Loop{LoopFromPoints(pts), EmptyLoop(), FullLoop()}
	for i, l := range loops {
		var buf bytes.Buffer
		l.Encode(&buf)
		ll := new(Loop)
		if err := ll.Decode(&buf); err != nil {
			t.Errorf("Decode %d: %v", i, err)
			continue
		}
		if !reflect.DeepEqual(l, ll) {
			t.Errorf("encoding roundtrip failed")
		}
	}
}

func TestLoopEncodeDecodeFuzzed(t *testing.T) {
	for i := 3; i < 100; i++ {
		var pts []Point
		for j := 0; j < i; j++ {
			pts = append(pts, randomPoint())
		}
		loop := LoopFromPoints(pts)
		if err := loop.Validate(); err != nil {
			t.Fatalf("loop(%v).Validate: %v", loop, err)
		}
		polygon := PolygonFromLoops([]*Loop{loop})
		var buf bytes.Buffer
		if err := polygon.Encode(&buf); err != nil {
			t.Fatal(err)
		}
		got := new(Loop)
		if err := got.Decode(&buf); err != nil {
			// TODO(nsch): Uncomment the next line as soon as decoding of all encoded loops works.
			// t.Fatalf("decode(encode(%v)): %v", loop, err)
			continue
		}

		if !reflect.DeepEqual(got, polygon) {
			t.Errorf("decode(encode()) = %v, want %v", got, polygon)
		}
	}
}

func BenchmarkRectDecode(b *testing.B) {
	rect := RectFromCenterSize(LatLngFromDegrees(80, 170), LatLngFromDegrees(40, 60))
	var buf bytes.Buffer
	if err := rect.Encode(&buf); err != nil {
		b.Fatal(err)
	}
	encoded := buf.Bytes()
	b.ReportAllocs()
	b.SetBytes(int64(len(encoded)))
	b.ResetTimer()
	var out Rect
	for i := 0; i < b.N; i++ {
		if err := out.Decode(bytes.NewReader(encoded)); err != nil {
			b.Fatal(err)
		}
	}
}
