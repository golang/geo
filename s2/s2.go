package s2

import (
	"math"
)

// Number of bits in the mantissa of a double.
const EXPONENT_SHIFT uint = 52

// Mask to extract the exponent from a double.
const EXPONENT_MASK uint64 = 0x7ff0000000000000

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func exp(v float64) int {
	if v == 0 {
		return 0
	}
	bits := math.Float64bits(v)
	return (int)((EXPONENT_MASK&bits)>>EXPONENT_SHIFT) - 1022
}

// Defines an area or a length cell metric.
type Metric struct {
	deriv float64
	dim   uint
}

// Defines a cell metric of the given dimension (1 == length, 2 == area).
func NewMetric(dim uint, deriv float64) Metric {
	return Metric{deriv, dim}
}

// The "deriv" value of a metric is a derivative, and must be multiplied by
// a length or area in (s,t)-space to get a useful value.
func (m Metric) Deriv() float64 { return m.deriv }

// Return the value of a metric for cells at the given level.
func (m Metric) GetValue(level int) float64 {
	return math.Pow(m.deriv, float64(int(m.dim)*(1-level)))
}

/**
 * Return the level at which the metric has approximately the given value.
 * For example, S2::kAvgEdge.GetClosestLevel(0.1) returns the level at which
 * the average cell edge length is approximately 0.1. The return value is
 * always a valid level.
 */
func (m Metric) getClosestLevel(value float64) int {
	return m.getMinLevel(math.Sqrt2 * value)
}

/**
 * Return the minimum level such that the metric is at most the given value,
 * or S2CellId::kMaxLevel if there is no such level. For example,
 * S2::kMaxDiag.GetMinLevel(0.1) returns the minimum level such that all
 * cell diagonal lengths are 0.1 or smaller. The return value is always a
 * valid level.
 */
func (m Metric) getMinLevel(value float64) int {
	if value <= 0 {
		return MAX_LEVEL
	}

	// This code is equivalent to computing a floating-point "level"
	// value and rounding up.
	exponent := exp(value / (float64(int(1)<<m.dim) * m.deriv))
	level := max(0, min(MAX_LEVEL, -((exponent-1)>>(m.dim-1))))
	// assert (level == S2CellId.MAX_LEVEL || getValue(level) <= value);
	// assert (level == 0 || getValue(level - 1) > value);
	return level
}

/**
 * Return the maximum level such that the metric is at least the given
 * value, or zero if there is no such level. For example,
 * S2.kMinWidth.GetMaxLevel(0.1) returns the maximum level such that all
 * cells have a minimum width of 0.1 or larger. The return value is always a
 * valid level.
 */
func (m Metric) getMaxLevel(value float64) int {
	if value <= 0 {
		return MAX_LEVEL
	}

	// This code is equivalent to computing a floating-point "level"
	// value and rounding down.
	exponent := exp(float64(int(1)<<m.dim) * m.deriv / value)
	level := max(0, min(MAX_LEVEL, ((exponent-1)>>(m.dim-1))))
	// assert (level == 0 || getValue(level) >= value);
	// assert (level == S2CellId.MAX_LEVEL || getValue(level + 1) < value);
	return level
}
