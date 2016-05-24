/*
Copyright 2015 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package s1

import (
	"math"
	"testing"
)

func TestChordAngleBasics(t *testing.T) {
	var zeroChord ChordAngle
	tests := []struct {
		a, b     ChordAngle
		lessThan bool
		equal    bool
	}{
		{NegativeChordAngle, NegativeChordAngle, false, true},
		{NegativeChordAngle, zeroChord, true, false},
		{NegativeChordAngle, StraightChordAngle, true, false},
		{NegativeChordAngle, InfChordAngle(), true, false},

		{zeroChord, zeroChord, false, true},
		{zeroChord, StraightChordAngle, true, false},
		{zeroChord, InfChordAngle(), true, false},

		{StraightChordAngle, StraightChordAngle, false, true},
		{StraightChordAngle, InfChordAngle(), true, false},

		{InfChordAngle(), InfChordAngle(), false, true},
		{InfChordAngle(), InfChordAngle(), false, true},
	}

	for _, test := range tests {
		if got := test.a < test.b; got != test.lessThan {
			t.Errorf("%v should be less than %v", test.a, test.b)
		}
		if got := test.a == test.b; got != test.equal {
			t.Errorf("%v should be equal to %v", test.a, test.b)
		}
	}
}

func TestChordIsFunctions(t *testing.T) {
	var zeroChord ChordAngle
	tests := []struct {
		have       ChordAngle
		isNegative bool
		isZero     bool
		isInf      bool
		isSpecial  bool
	}{
		{zeroChord, false, true, false, false},
		{NegativeChordAngle, true, false, false, true},
		{zeroChord, false, true, false, false},
		{StraightChordAngle, false, false, false, false},
		{InfChordAngle(), false, false, true, true},
	}

	for _, test := range tests {
		if got := test.have < 0; got != test.isNegative {
			t.Errorf("%v.isNegative() = %t, want %t", test.have, got, test.isNegative)
		}
		if got := test.have == 0; got != test.isZero {
			t.Errorf("%v.isZero() = %t, want %t", test.have, got, test.isZero)
		}
		if got := test.have.isInf(); got != test.isInf {
			t.Errorf("%v.isInf() = %t, want %t", test.have, got, test.isInf)
		}
		if got := test.have.isSpecial(); got != test.isSpecial {
			t.Errorf("%v.isSpecial() = %t, want %t", test.have, got, test.isSpecial)
		}
	}
}

func TestChordToChordFromAngle(t *testing.T) {
	for _, angle := range []float64{0, 1, -1, math.Pi} {
		if got := ChordFromAngle(Angle(angle)).Angle().Radians(); got != angle {
			t.Errorf("ChordFromAngle(Angle(%v)) = %v, want %v", angle, got, angle)
		}
	}

	if got := ChordFromAngle(Angle(math.Pi)); got != StraightChordAngle {
		t.Errorf("a ChordAngle from an Angle of π = %v, want %v", got, StraightChordAngle)
	}

	if InfAngle() != ChordFromAngle(InfAngle()).Angle() {
		t.Errorf("converting infinite Angle to ChordAngle should yield infinite Angle")
	}
}
