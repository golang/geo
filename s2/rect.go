/*
Copyright 2014 Google Inc. All rights reserved.

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

import (
	"fmt"
	"math"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/s1"
)

// Rect represents a closed latitude-longitude rectangle.
type Rect struct {
	Lat r1.Interval
	Lng s1.Interval
}

var (
	validRectLatRange = r1.Interval{-math.Pi / 2, math.Pi / 2}
	validRectLngRange = s1.FullInterval()
)

// EmptyRect returns the empty rectangle.
func EmptyRect() Rect { return Rect{r1.EmptyInterval(), s1.EmptyInterval()} }

// FullRect returns the full rectangle.
func FullRect() Rect { return Rect{validRectLatRange, validRectLngRange} }

// RectFromLatLng constructs a rectangle containing a single point p.
func RectFromLatLng(p LatLng) Rect {
	return Rect{
		Lat: r1.Interval{p.Lat.Radians(), p.Lat.Radians()},
		Lng: s1.Interval{p.Lng.Radians(), p.Lng.Radians()},
	}
}

// RectFromCenterSize constructs a rectangle with the given size and center.
// center needs to be normalized, but size does not. The latitude
// interval of the result is clamped to [-90,90] degrees, and the longitude
// interval of the result is FullRect() if and only if the longitude size is
// 360 degrees or more.
//
// Examples of clamping (in degrees):
//   center=(80,170),  size=(40,60)   -> lat=[60,90],   lng=[140,-160]
//   center=(10,40),   size=(210,400) -> lat=[-90,90],  lng=[-180,180]
//   center=(-90,180), size=(20,50)   -> lat=[-90,-80], lng=[155,-155]
func RectFromCenterSize(center, size LatLng) Rect {
	half := LatLng{size.Lat / 2, size.Lng / 2}
	return RectFromLatLng(center).expanded(half)
}

// IsValid returns true iff the rectangle is valid.
// This requires Lat ⊆ [-π/2,π/2] and Lng ⊆ [-π,π], and Lat = ∅ iff Lng = ∅
func (r Rect) IsValid() bool {
	return math.Abs(r.Lat.Lo) <= math.Pi/2 &&
		math.Abs(r.Lat.Hi) <= math.Pi/2 &&
		r.Lng.IsValid() &&
		r.Lat.IsEmpty() == r.Lng.IsEmpty()
}

// IsEmpty reports whether the rectangle is empty.
func (r Rect) IsEmpty() bool { return r.Lat.IsEmpty() }

// IsFull reports whether the rectangle is full.
func (r Rect) IsFull() bool { return r.Lat.Equal(validRectLatRange) && r.Lng.IsFull() }

// IsPoint reports whether the rectangle is a single point.
func (r Rect) IsPoint() bool { return r.Lat.Lo == r.Lat.Hi && r.Lng.Lo == r.Lng.Hi }

// Lo returns one corner of the rectangle.
func (r Rect) Lo() LatLng {
	return LatLng{s1.Angle(r.Lat.Lo) * s1.Radian, s1.Angle(r.Lng.Lo) * s1.Radian}
}

// Hi returns the other corner of the rectangle.
func (r Rect) Hi() LatLng {
	return LatLng{s1.Angle(r.Lat.Hi) * s1.Radian, s1.Angle(r.Lng.Hi) * s1.Radian}
}

// Center returns the center of the rectangle.
func (r Rect) Center() LatLng {
	return LatLng{s1.Angle(r.Lat.Center()) * s1.Radian, s1.Angle(r.Lng.Center()) * s1.Radian}
}

// Size returns the size of the Rect.
func (r Rect) Size() LatLng {
	return LatLng{s1.Angle(r.Lat.Length()) * s1.Radian, s1.Angle(r.Lng.Length()) * s1.Radian}
}

// Area returns the surface area of the Rect.
func (r Rect) Area() float64 {
	if r.IsEmpty() {
		return 0
	}
	capDiff := math.Abs(math.Sin(r.Lat.Hi) - math.Sin(r.Lat.Lo))
	return r.Lng.Length() * capDiff
}

// ContainsLatLng reports whether the given LatLng is within the Rect.
func (r Rect) ContainsLatLng(ll LatLng) bool {
	if !ll.IsValid() {
		return false
	}
	return r.Lat.Contains(ll.Lat.Radians()) && r.Lng.Contains(ll.Lng.Radians())
}

// AddPoint increases the size of the rectangle to include the given point.
func (r Rect) AddPoint(ll LatLng) Rect {
	if !ll.IsValid() {
		return r
	}
	return Rect{
		Lat: r.Lat.AddPoint(ll.Lat.Radians()),
		Lng: r.Lng.AddPoint(ll.Lng.Radians()),
	}
}

// expanded returns a rectangle that contains all points whose latitude distance from
// this rectangle is at most margin.Lat, and whose longitude distance from
// this rectangle is at most margin.Lng. In particular, latitudes are
// clamped while longitudes are wrapped. Any expansion of an empty rectangle
// remains empty. Both components of margin must be non-negative.
//
// Note that if an expanded rectangle contains a pole, it may not contain
// all possible lat/lng representations of that pole, e.g., both points [π/2,0]
// and [π/2,1] represent the same pole, but they might not be contained by the
// same Rect.
//
// If you are trying to grow a rectangle by a certain distance on the
// sphere (e.g. 5km), refer to the ConvolveWithCap() C++ method implementation
// instead.
func (r Rect) expanded(margin LatLng) Rect {
	return Rect{
		Lat: r.Lat.Expanded(margin.Lat.Radians()).Intersection(validRectLatRange),
		Lng: r.Lng.Expanded(margin.Lng.Radians()),
	}
}

func (r Rect) String() string { return fmt.Sprintf("[Lo%v, Hi%v]", r.Lo(), r.Hi()) }

// BUG(dsymonds): The major differences from the C++ version are:
//   - almost everything
