// Fuzz tests for s2 decoding.
package s2

import (
	"bytes"
	"testing"
)

// go test -fuzz=FuzzDecodeCellUnion github.com/golang/geo/s2
func FuzzDecodeCellUnion(f *testing.F) {
	cu := CellUnion([]CellID{
		CellID(0x33),
		CellID(0x8e3748fab),
		CellID(0x91230abcdef83427),
	})
	buf := new(bytes.Buffer)
	if err := cu.Encode(buf); err != nil {
		f.Errorf("error encoding %v: ", err)
	}
	f.Add(buf.Bytes())

	f.Fuzz(func(t *testing.T, encoded []byte) {
		var c CellUnion
		if err := c.Decode(bytes.NewReader(encoded)); err != nil {
			// Construction failed, no need to test further.
			return
		}
		if got := c.ApproxArea(); got < 0 {
			t.Errorf("ApproxArea() = %v, want >= 0. CellUnion: %v", got, c)
		}
		buf := new(bytes.Buffer)
		if err := c.Encode(buf); err != nil {
			// Re-encoding the cell union does not necessarily produce the same bytes, as there could be additional bytes following after n cells were read.
			t.Errorf("encode() = %v. got %v, want %v. CellUnion: %v", err, buf.Bytes(), encoded, c)
		}
		if c.IsValid() {
			c.Normalize()
			if !c.IsNormalized() {
				t.Errorf("IsNormalized() = false, want true. CellUnion: %v", c)
			}
		}
	})
}

// go test -fuzz=FuzzDecodePolygon github.com/golang/geo/s2
func FuzzDecodePolygon(f *testing.F) {
	for _, p := range []*Polygon{near0Polygon, near01Polygon, near30Polygon, near23Polygon, far01Polygon, far21Polygon, south0abPolygon} {
		buf := new(bytes.Buffer)
		if err := p.Encode(buf); err != nil {
			f.Errorf("error encoding %v: ", err)
		}
		f.Add(buf.Bytes())
	}

	f.Fuzz(func(t *testing.T, encoded []byte) {
		p := &Polygon{}
		if err := p.Decode(bytes.NewReader(encoded)); err != nil {
			// Construction failed, no need to test further.
			return
		}
		if got := p.Area(); got < 0 {
			t.Errorf("Area() = %v, want >= 0. Polygon: %v", got, p)
		}
		// TODO: Test more methods on Polygon.
	})
}
