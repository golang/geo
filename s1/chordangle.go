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
)

// ChordAngle represents the angle subtended by a chord (i.e., the straight
// line segment connecting two points on the sphere). Its representation
// makes it very efficient for computing and comparing distances, but unlike
// Angle it is only capable of representing angles between 0 and π radians.
// Generally, ChordAngle should only be used in loops where many angles need
// to be calculated and compared. Otherwise it is simpler to use Angle.
//
// ChordAngles are represented by the squared chord length, which can
// range from 0 to 4. Positive infinity represents an infinite squared length.
type ChordAngle float64

const (
	// NegativeChordAngle represents a chord angle smaller than the zero angle.
	// The only valid operations on a NegativeChordAngle are comparisons and
	// Angle conversions.
	NegativeChordAngle = ChordAngle(-1)

	// StraightChordAngle represents a chord angle of 180 degrees (a "straight angle").
	// This is the maximum finite chord angle.
	StraightChordAngle = ChordAngle(4)
)

// InfChordAngle returns a chord angle larger than any finite chord angle.
// The only valid operations on an InfChordAngle are comparisons and Angle conversions.
func InfChordAngle() ChordAngle {
	return ChordAngle(math.Inf(1))
}

// isInf reports whether this ChordAngle is infinite.
func (c ChordAngle) isInf() bool {
	return math.IsInf(float64(c), 1)
}

// isSpecial reports whether this ChordAngle is one of the special cases.
func (c ChordAngle) isSpecial() bool {
	return c < 0 || c.isInf()
}
