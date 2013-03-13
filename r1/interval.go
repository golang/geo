package r1

import (
	"fmt"
	"math"
)

// Interval represents a closed interval on ℝ.
// Zero-length intervals (where Lo == Hi) represent single points.
// If Lo > Hi then the interval is empty.
type Interval struct {
	Lo, Hi float64
}

// EmptyInterval returns an empty interval.
func EmptyInterval() Interval { return Interval{1, 0} }

// IntervalFromPoint returns an interval representing a single point.
func IntervalFromPoint(p float64) Interval { return Interval{p, p} }

// IsEmpty reports whether the interval is empty.
func (i Interval) IsEmpty() bool { return i.Lo > i.Hi }

// Equal returns true iff the interval contains the same points as oi.
func (i Interval) Equal(oi Interval) bool {
	return i == oi || i.IsEmpty() && oi.IsEmpty()
}

// Center returns the midpoint of the interval.
// It is undefined for empty intervals.
func (i Interval) Center() float64 { return 0.5 * (i.Lo + i.Hi) }

// Length returns the length of the interval.
// The length of an empty interval is negative.
func (i Interval) Length() float64 { return i.Hi - i.Lo }

// Contains returns true iff the interval contains p.
func (i Interval) Contains(p float64) bool { return i.Lo <= p && p <= i.Hi }

// ContainsInterval returns true iff the interval contains oi.
func (i Interval) ContainsInterval(oi Interval) bool {
	if oi.IsEmpty() {
		return true
	}
	return i.Lo <= oi.Lo && oi.Hi <= i.Hi
}

// InteriorContains returns true iff the the interval strictly contains p.
func (i Interval) InteriorContains(p float64) bool {
	return i.Lo < p && p < i.Hi
}

// InteriorContainsInterval returns true iff the interval strictly contains oi.
func (i Interval) InteriorContainsInterval(oi Interval) bool {
	if oi.IsEmpty() {
		return true
	}
	return i.Lo < oi.Lo && oi.Hi < i.Hi
}

// Intersects returns true iff the interval contains any points in common with oi.
func (i Interval) Intersects(oi Interval) bool {
	if i.Lo <= oi.Lo {
		return oi.Lo <= i.Hi && oi.Lo <= oi.Hi // oi.Lo ∈ i and oi is not empty
	}
	return i.Lo <= oi.Hi && i.Lo <= i.Hi // i.Lo ∈ oi and i is not empty
}

// InteriorIntersects returns true iff the interior of the interval contains any points in common with oi, including the latter's boundary.
func (i Interval) InteriorIntersects(oi Interval) bool {
	return oi.Lo < i.Hi && i.Lo < oi.Hi && i.Lo < i.Hi && oi.Lo <= i.Hi
}

// Intersection returns the interval containing all points common to i and j.
func (i Interval) Intersection(j Interval) Interval {
	// Empty intervals do not need to be special-cased.
	return Interval{
		Lo: math.Max(i.Lo, j.Lo),
		Hi: math.Min(i.Hi, j.Hi),
	}
}

// Expanded returns an interval that has been expanded on each side by margin.
// If margin is negative, then the function shrinks the interval on
// each side by margin instead. The resulting interval may be empty. Any
// expansion of an empty interval remains empty.
func (i Interval) Expanded(margin float64) Interval {
	if i.IsEmpty() {
		return i
	}
	return Interval{i.Lo - margin, i.Hi + margin}
}

func (i Interval) String() string { return fmt.Sprintf("[%.7f, %.7f]", i.Lo, i.Hi) }

// BUG(dsymonds): The major differences from the C++ version are:
//   - Union, ApproxEquals
//   - a few other miscellaneous operations
