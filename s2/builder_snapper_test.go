// Copyright 2023 Google Inc. All rights reserved.
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
	"testing"

	"github.com/pavlov061356/geo/s1"
)

func TestIdentitySnapper(t *testing.T) {
	rad := s1.Angle(1.0)
	i := NewIdentitySnapper(rad)

	if i.SnapRadius() != rad {
		t.Errorf("identSnap.SnapRadius() = %v, want %v", i.SnapRadius(), rad)
	}

	if i.MinVertexSeparation() != rad {
		t.Errorf("identSnap.MinVertexSeparation() = %v, want %v", i.MinVertexSeparation(), rad)
	}

	if i.MinEdgeVertexSeparation() != 0.5*rad {
		t.Errorf("identSnap.MinEdgeVertexSeparation() = %v, want %v", i.MinEdgeVertexSeparation(), 0.5*rad)
	}

	if i.MaxEdgeDeviation() != maxEdgeDeviationRatio {
		t.Errorf("identSnap.SnapRadius() = %v, want %v", i.MaxEdgeDeviation(), maxEdgeDeviationRatio)
	}

	p := randomPoint()
	if got := i.SnapPoint(p); !p.ApproxEqual(got) {
		t.Errorf("identSnap.SnapPoint(%v) = %v, want %v", p, got, p)
	}

}

func TestCellIDSnapperLevelToFromSnapRadius(t *testing.T) {
	f := NewCellIDSnapper()
	for level := 0; level <= MaxLevel; level++ {
		radius := f.minSnapRadiusForLevel(level)
		if got := f.levelForMaxSnapRadius(radius); got != level {
			t.Errorf("levelForMaxSnapRadius(%v) = %v, want %v", radius, got, level)
		}
		if got, want := f.levelForMaxSnapRadius(0.999*radius), minInt(level+1, MaxLevel); got != want {
			t.Errorf("levelForMaxSnapRadius(0.999*%v) = %v, want %v (level %d)", radius, got, want, level)
		}
	}

	if got := f.levelForMaxSnapRadius(5); got != 0 {
		t.Errorf("levelForMaxSnapRadius(5) = %v, want 0", got)
	}
	if got := f.levelForMaxSnapRadius(1e-30); got != MaxLevel {
		t.Errorf("levelForMaxSnapRadius(1e-30) = %v, want 0", got)
	}
}

func TestCellIDSnapperSnapPoint(t *testing.T) {
	for iter := 0; iter < 1; iter++ {
		for level := 0; level <= MaxLevel; level++ {
			// This checks that points are snapped to the correct level, since
			// CellID centers at different levels are always different.
			f := CellIDSnapperForLevel(level)
			p := randomCellIDForLevel(level).Point()
			if got, want := f.SnapPoint(p), p; !got.ApproxEqual(want) {
				t.Errorf("%v.SnapPoint(%v) = %v, want %v", f, p, got, want)
			}
		}
	}
}

func TestIntLatLngSnapperExponentToFromSnapRadius(t *testing.T) {
	for exp := minIntSnappingExponent; exp <= maxIntSnappingExponent; exp++ {
		sf := NewIntLatLngSnapper(exp)
		radius := sf.minSnapRadiusForExponent(exp)
		if got := sf.exponentForMaxSnapRadius(radius); got != exp {
			t.Errorf("exponentForMaxSnapRadius(%v) = %v, want %v", radius, got, exp)
		}
		if got, want := sf.exponentForMaxSnapRadius(0.999*radius), minInt(exp+1, maxIntSnappingExponent); got != want {
			t.Errorf("exponentForMaxSnapRadius(%v) = %v, want %v", 0.999*radius, got, want)
		}
	}
	sf := IntLatLngSnapper{}
	if got := sf.exponentForMaxSnapRadius(5); got != minIntSnappingExponent {
		t.Errorf("exponentForMaxSnapRadius(5) = %v, want %v", got, minIntSnappingExponent)
	}
	if got := sf.exponentForMaxSnapRadius(1e-30); got != maxIntSnappingExponent {
		t.Errorf("exponentForMaxSnapRadius(1e-30) = %v, want %v", got, maxIntSnappingExponent)
	}
}

/*
TODO(roberts): Uncomment when LatLng helpers are incorporated.
func TestIntLatLngSnapperSnapPoint(t *testing.T) {
	for iter := 0; iter < 1000; iter++ {
		// Test that IntLatLngSnapper does not modify points that were
		// generated using the LatLngFrom{E5,E6,E7} methods. This ensures
		// that both functions are using bitwise-compatible conversion methods.
		p := randomPoint()
		ll := LatLngFromPoint(p)

		p5 := LatLngFromE5(ll.Lat.E5(), ll.Lng.E5()).ToPoint()
		if got := NewIntLatLngSnapper(5).SnapPoint(p5); !p5.ApproxEqual(got) {
			t.Errorf("NewIntLatLngSnapper(5).SnapPoint(%v) = %v, want %v", p5, got, p5)
		}
		p6 := LatLngFromE6(ll.Lat.E6(), ll.Lng.E6()).ToPoint()
		if got := NewIntLatLngSnapper(6).SnapPoint(p6); !p6.ApproxEqual(got) {
			t.Errorf("NewIntLatLngSnapper(6).SnapPoint(%v) = %v, want %v", p6, got, p6)
		}
		p7 := LatLngFromE7(ll.Lat.E7(), ll.Lng.E7()).ToPoint()
		if got := NewIntLatLngSnapper(7).SnapPoint(p7); !p7.ApproxEqual(got) {
			t.Errorf("NewIntLatLngSnapper(7).SnapPoint(%v) = %v, want %v", p7, got, p7)
		}

		// Make sure that we're not snapping using some lower exponent.
		p7not6 := LatLngFromE7(10*ll.Lat.E6()+1, 10*ll.Lng.E6()+1).ToPoint()
		if got := NewIntLatLngSnapper(6).SnapPoint(p7not6); p7not6 == got {
			t.Errorf("NewIntLatLngSnapper(6).SnapPoint(%v) = %v, want %v", p7not6, got, p7not6)
		}
	}
}
*/

// TODO(roberts): Differences from C++:
// bunch of helper methods for these tests:
// func TestCellIdSnapFunctionMinVertexSeparationSnapRadiusRatio(t *testing.T) {
// func TestCellIdSnapFunctionMinEdgeVertexSeparationForLevel(t *testing.T) {
// func TestCellIdSnapFunctionMinEdgeVertexSeparationAtMinSnapRadius(t *testing.T) {
// func TestCellIdSnapFunctionMinEdgeVertexSeparationSnapRadiusRatio(t *testing.T) {
// func TestIntLatLngSnapFunctionMinVertexSeparationSnapRadiusRatio(t *testing.T) {
// func TestIntLatLngSnapFunctionMinEdgeVertexSeparationForLevel(t *testing.T) {
// func TestIntLatLngSnapFunctionMinEdgeVertexSeparationSnapRadiusRatio(t *testing.T) {
