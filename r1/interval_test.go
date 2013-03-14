package r1

import (
	"testing"
)

// Some standard intervals for use throughout the tests.
var (
	unit    = Interval{0, 1}
	negunit = Interval{-1, 0}
	half    = Interval{0.5, 0.5}
	empty   = EmptyInterval()
)

func TestIsEmpty(t *testing.T) {
	var zero Interval
	if unit.IsEmpty() {
		t.Errorf("%v should not be empty", unit)
	}
	if half.IsEmpty() {
		t.Errorf("%v should not be empty", half)
	}
	if !empty.IsEmpty() {
		t.Errorf("%v should be empty", empty)
	}
	if zero.IsEmpty() {
		t.Errorf("zero Interval %v should not be empty", zero)
	}
}

func TestCenter(t *testing.T) {
	tests := []struct {
		interval Interval
		want     float64
	}{
		{unit, 0.5},
		{negunit, -0.5},
		{half, 0.5},
	}
	for _, test := range tests {
		got := test.interval.Center()
		if got != test.want {
			t.Errorf("%v.Center() = %v, want %v", test.interval, got, test.want)
		}
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		interval Interval
		want     float64
	}{
		{unit, 1},
		{negunit, 1},
		{half, 0},
	}
	for _, test := range tests {
		if l := test.interval.Length(); l != test.want {
			t.Errorf("%v.Length() = %v, want %v", test.interval, l, test.want)
		}
	}
	if l := empty.Length(); l >= 0 {
		t.Errorf("empty interval has non-negative length")
	}
}

// TODO(dsymonds): Tests for Contains, InteriorContains, ContainsInterval, InteriorContainsInterval, Intersects, InteriorIntersects

func TestIntersection(t *testing.T) {
	tests := []struct {
		x, y Interval
		want Interval
	}{
		{unit, half, half},
		{unit, negunit, Interval{0, 0}},
		{negunit, half, empty},
		{unit, empty, empty},
		{empty, unit, empty},
	}
	for _, test := range tests {
		if got := test.x.Intersection(test.y); !got.Equal(test.want) {
			t.Errorf("%v.Intersection(%v) = %v, want equal to %v", test.x, test.y, got, test.want)
		}
	}
}

func TestExpanded(t *testing.T) {
	tests := []struct {
		interval Interval
		margin   float64
		want     Interval
	}{
		{empty, 0.45, empty},
		{unit, 0.5, Interval{-0.5, 1.5}},
		{unit, -0.5, Interval{0.5, 0.5}},
		{unit, -0.51, empty},
	}
	for _, test := range tests {
		if got := test.interval.Expanded(test.margin); !got.Equal(test.want) {
			t.Errorf("%v.Expanded(%v) = %v, want equal to %v", test.interval, test.margin, got, test.want)
		}
	}
}

func TestIntervalString(t *testing.T) {
	i := Interval{2, 4.5}
	if s, exp := i.String(), "[2.0000000, 4.5000000]"; s != exp {
		t.Errorf("i.String() = %q, want %q", s, exp)
	}
}
