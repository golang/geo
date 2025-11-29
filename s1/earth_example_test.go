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

package s1_test

import (
	"fmt"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/google/go-units/unit"
)

func ExampleEarthAngleFromLength() {
	length := 500 * unit.Mile
	angle := s1.EarthAngleFromLength(length)
	fmt.Printf("I would walk 500 miles (%.4f rad)", angle.Radians())
	// Output: I would walk 500 miles (0.1263 rad)
}

func ExampleEarthLengthFromAngle() {
	angle := 2 * s1.Degree
	length := s1.EarthLengthFromAngle(angle)
	fmt.Printf("2° is %.0f miles", length.Miles())
	// Output: 2° is 138 miles
}

func ExampleEarthAreaFromSteradians() {
	bermuda := s2.PointFromLatLng(s2.LatLngFromDegrees(32.361457, -64.663495))
	culebra := s2.PointFromLatLng(s2.LatLngFromDegrees(18.311199, -65.300765))
	miami := s2.PointFromLatLng(s2.LatLngFromDegrees(25.802018, -80.269892))
	triangle := s2.PolygonFromLoops(
		[]*s2.Loop{s2.LoopFromPoints([]s2.Point{bermuda, miami, culebra})})
	area := s1.EarthAreaFromSteradians(triangle.Area())
	fmt.Printf("Bermuda Triangle is %.2f square miles", area.SquareMiles())
	// Output: Bermuda Triangle is 464541.15 square miles
}

func ExampleEarthSteradiansFromArea() {
	steradians := s1.EarthSteradiansFromArea(unit.SquareCentimeter)
	fmt.Printf(
		"1 square centimeter is %g steradians, close to a level %d cell",
		steradians,
		s2.AvgAreaMetric.ClosestLevel(steradians),
	)
	// Output: 1 square centimeter is 2.4636750563592804e-18 steradians, close to a level 30 cell
}
