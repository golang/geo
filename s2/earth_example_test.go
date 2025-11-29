// Copyright 2025 Google LLC
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

package s2_test

import (
	"fmt"
	"math"

	"github.com/golang/geo/earth"
	"github.com/golang/geo/s2"
)

func ExampleEarthLengthFromPoints() {
	equator := s2.PointFromCoords(1, 0, 0)
	pole := s2.PointFromCoords(0, 1, 0)
	length := s2.EarthLengthFromPoints(equator, pole)
	fmt.Printf(
		"Equator to pole is %.2f km, π/2*Earth radius is %.2f km",
		length.Kilometers(),
		math.Pi/2*earth.Radius.Kilometers(),
	)
	// Output: Equator to pole is 10007.56 km, π/2*Earth radius is 10007.56 km
}

func ExampleEarthLengthFromLatLngs() {
	chukchi := s2.LatLngFromDegrees(66.025893, -169.699684)
	seward := s2.LatLngFromDegrees(65.609727, -168.093694)
	length := s2.EarthLengthFromLatLngs(chukchi, seward)
	fmt.Printf("Bering Strait is %.0f feet", length.Feet())
	// Output: Bering Strait is 283979 feet
}

func ExampleEarthInitialBearingFromLatLngs() {
	christchurch := s2.LatLngFromDegrees(-43.491402, 172.522275)
	riogrande := s2.LatLngFromDegrees(-53.777156, -67.734719)
	bearing := s2.EarthInitialBearingFromLatLngs(christchurch, riogrande)
	fmt.Printf("Head southeast (%.2f°)", bearing.Degrees())
	// Output: Head southeast (146.90°)
}
