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

func TestUnion(t *testing.T) {
	tests := []struct {
		x, y Interval
		want Interval
	}{
		{Interval{99, 100}, empty, Interval{99, 100}},
		{empty, Interval{99, 100}, Interval{99, 100}},
		{Interval{5, 3}, Interval{0, -2}, empty},
		{Interval{0, -2}, Interval{5, 3}, empty},
		{unit, unit, unit},
		{unit, negunit, Interval{-1, 1}},
		{negunit, unit, Interval{-1, 1}},
		{half, unit, unit},
	}
	for _, test := range tests {
		if got := test.x.Union(test.y); !got.Equal(test.want) {
			t.Errorf("%v.Union(%v) = %v, want equal to %v", test.x, test.y, got, test.want)
		}
	}
}

func TestAddPoint(t *testing.T) {
	tests := []struct {
		interval Interval
		point    float64
		want     Interval
	}{
		{empty, 5, Interval{5, 5}},
		{Interval{5, 5}, -1, Interval{-1, 5}},
		{Interval{-1, 5}, 0, Interval{-1, 5}},
		{Interval{-1, 5}, 6, Interval{-1, 6}},
	}
	for _, test := range tests {
		if got := test.interval.AddPoint(test.point); !got.Equal(test.want) {
			t.Errorf("%v.AddPoint(%v) = %v, want equal to %v", test.interval, test.point, got, test.want)
		}
	}
}

func TestClampPoint(t *testing.T) {
	tests := []struct {
		interval Interval
		clamp    float64
		want     float64
	}{
		{Interval{0.1, 0.4}, 0.3, 0.3},
		{Interval{0.1, 0.4}, -7.0, 0.1},
		{Interval{0.1, 0.4}, 0.6, 0.4},
	}
	for _, test := range tests {
		if got := test.interval.ClampPoint(test.clamp); got != test.want {
			t.Errorf("%v.ClampPoint(%v) = %v, want equal to %v", test.interval, test.clamp, got, test.want)
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

func TestApproxEqual(t *testing.T) {
	tests := []struct {
		interval Interval
		other    Interval
		want     bool
	}{
		// Empty intervals.
		{EmptyInterval(), EmptyInterval(), true},
		{Interval{0, 0}, EmptyInterval(), true},
		{EmptyInterval(), Interval{0, 0}, true},
		{Interval{1, 1}, EmptyInterval(), true},
		{EmptyInterval(), Interval{1, 1}, true},
		{EmptyInterval(), Interval{0, 1}, false},
		{EmptyInterval(), Interval{1, 1 + 2*epsilon}, true},

		// Singleton intervals.
		{Interval{1, 1}, Interval{1, 1}, true},
		{Interval{1, 1}, Interval{1 - epsilon, 1 - epsilon}, true},
		{Interval{1, 1}, Interval{1 + epsilon, 1 + epsilon}, true},
		{Interval{1, 1}, Interval{1 - 3*epsilon, 1}, false},
		{Interval{1, 1}, Interval{1, 1 + 3*epsilon}, false},
		{Interval{1, 1}, Interval{1 - epsilon, 1 + epsilon}, true},
		{Interval{0, 0}, Interval{1, 1}, false},

		// Other intervals.
		{Interval{1 - epsilon, 2 + epsilon}, Interval{1, 2}, false},
		{Interval{1 + epsilon, 2 - epsilon}, Interval{1, 2}, true},
		{Interval{1 - 3*epsilon, 2 + epsilon}, Interval{1, 2}, false},
		{Interval{1 + 3*epsilon, 2 - epsilon}, Interval{1, 2}, false},
		{Interval{1 - epsilon, 2 + 3*epsilon}, Interval{1, 2}, false},
		{Interval{1 + epsilon, 2 - 3*epsilon}, Interval{1, 2}, false},
	}

	for _, test := range tests {
		if got := test.interval.ApproxEqual(test.other); got != test.want {
			t.Errorf("%v.ApproxEqual(%v) = %v, want %v",
				test.interval, test.other, got, test.want)
		}
	}
}
