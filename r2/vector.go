/*
 * Copyright 2005 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package r2

import (
	"fmt"
	"math"
)

/**
 * r2.Vector represents a vector in the two-dimensional space. It defines the
 * basic geometrical operations for 2D vectors, e.g. cross product, addition,
 * norm, comparison etc.
 *
 */
type Vector struct {
	X, Y float64
}

func (v Vector) String() string { return fmt.Sprintf("(%v, %v)", v.X, v.Y) }

// Norm returns the vector's norm.
func (v Vector) Norm() float64 { return math.Sqrt(v.Dot(v)) }

// Norm2 returns the square of the norm.
func (v Vector) Norm2() float64 { return v.Dot(v) }

// Normalize returns a unit vector in the same direction as v.
func (v Vector) Normalize() Vector {
	if v == (Vector{0, 0}) {
		return v
	}
	return v.Mul(1 / v.Norm())
}

// Abs returns the vector with nonnegative components.
func (v Vector) Abs() Vector { return Vector{math.Abs(v.X), math.Abs(v.Y)} }

// Neg returns the negated vector
func (v Vector) Neg() Vector { return Vector{-v.X, -v.Y} }

// Add returns the standard vector sum of v and ov.
func (v Vector) Add(ov Vector) Vector { return Vector{v.X + ov.X, v.Y + ov.Y} }

// Sub returns the standard vector difference of v and ov.
func (v Vector) Sub(ov Vector) Vector { return Vector{v.X - ov.X, v.Y - ov.Y} }

// Mul returns the standard scalar product of v and m.
func (v Vector) Mul(m float64) Vector { return Vector{v.X * m, v.Y * m} }

// Mul returns the standard scalar product of v and m.
func (v Vector) Div(m float64) Vector { return Vector{v.X / m, v.Y / m} }

// Dot returns the standard dot product of v and ov.
func (v Vector) Dot(ov Vector) float64 { return v.X*ov.X + v.Y*ov.Y }

// Cross returns the standard cross product of v and ov.
func (v Vector) Cross(ov Vector) float64 {
	return v.X*ov.Y - v.Y*ov.X
}

func (v Vector) Equals(other Vector) bool {
	return v.X == other.X && v.Y == other.Y
}

func (v Vector) LessThan(vb Vector) bool {
	if v.X < vb.X {
		return true
	}
	if vb.X < v.X {
		return false
	}
	if v.Y < vb.Y {
		return true
	}
	if vb.Y < v.Y {
		return false
	}
	return false
}

func (v Vector) CompareTo(other Vector) int {
	if v.LessThan(other) {
		return -1
	} else {
		if v.Equals(other) {
			return 0
		}
		return 1
	}
}
