// Fuzz tests for s2 decoding. Run using
// go test -fuzz=Fuzz github.com/golang/geo/s2
package s2

import (
	"bytes"
	"testing"
)

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
		// TODO: Test more invariants that should hold for all CellUnion.
	})
}
