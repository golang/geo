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

package s2

// This file implements functions for various S2 measurements.

import (
	"math"
)

// A Metric is a measure for cells.
type Metric struct {
	// Dim is either 1 or 2, for a 1D or 2D metric respectively.
	Dim int
	// Deriv is the scaling factor for the metric.
	Deriv float64
}

// Defined metrics.
// We only support the quadratic projection.
var (
	MinWidthMetric = Metric{1, 2 * math.Sqrt2 / 3}
	MaxWidthMetric = Metric{1, 1.704897179199218452}

	MinAreaMetric = Metric{2, 8 * math.Sqrt2 / 9}
	AvgAreaMetric = Metric{2, 4 * math.Pi / 6}
	MaxAreaMetric = Metric{2, 2.635799256963161491}
)

// TODO: more metrics, as needed
// TODO: port GetValue, GetClosestLevel

// Value returns the value of the metric at the given level.
func (m Metric) Value(level int) float64 {
	return math.Ldexp(m.Deriv, -m.Dim*level)
}

// MinLevel returns the minimum level such that the metric is at most
// the given value, or maxLevel (30) if there is no such level.
func (m Metric) MinLevel(val float64) int {
	if val < 0 {
		return maxLevel
	}

	level := -(math.Ilogb(val/m.Deriv) >> uint(m.Dim-1))
	if level > maxLevel {
		level = maxLevel
	}
	if level < 0 {
		level = 0
	}
	return level
}

// MaxLevel returns the maximum level such that the metric is at least
// the given value, or zero if there is no such level.
func (m Metric) MaxLevel(val float64) int {
	if val <= 0 {
		return maxLevel
	}

	level := math.Ilogb(m.Deriv/val) >> uint(m.Dim-1)
	if level > maxLevel {
		level = maxLevel
	}
	if level < 0 {
		level = 0
	}
	return level
}
