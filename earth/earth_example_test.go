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

package earth_test

import (
	"fmt"
	"math"

	"github.com/golang/geo/earth"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"github.com/google/go-units/unit"
)

func ExampleAngleFromLength() {
	length := 500 * unit.Mile
	angle := earth.AngleFromLength(length)
	fmt.Printf("I would walk 500 miles (%.4f rad)", angle.Radians())
	// Output: I would walk 500 miles (0.1263 rad)
}

func ExampleLengthFromAngle() {
	angle := 2 * s1.Degree
	length := earth.LengthFromAngle(angle)
	fmt.Printf("2° is %.0f miles", length.Miles())
	// Output: 2° is 138 miles
}

func ExampleLengthFromPoints() {
	equator := s2.PointFromCoords(1, 0, 0)
	pole := s2.PointFromCoords(0, 1, 0)
	length := earth.LengthFromPoints(equator, pole)
	fmt.Printf("Equator to pole is %.2f km, π/2*Earth radius is %.2f km",
		length.Kilometers(), math.Pi/2*earth.Radius.Kilometers())
	// Output: Equator to pole is 10007.56 km, π/2*Earth radius is 10007.56 km
}

func ExampleLengthFromLatLngs() {
	chukchi := s2.LatLngFromDegrees(66.025893, -169.699684)
	seward := s2.LatLngFromDegrees(65.609727, -168.093694)
	length := earth.LengthFromLatLngs(chukchi, seward)
	fmt.Printf("Bering Strait is %.0f feet", length.Feet())
	// Output: Bering Strait is 283979 feet
}

func ExampleAreaFromSteradians() {
	bermuda := s2.PointFromLatLng(s2.LatLngFromDegrees(32.361457, -64.663495))
	culebra := s2.PointFromLatLng(s2.LatLngFromDegrees(18.311199, -65.300765))
	miami := s2.PointFromLatLng(s2.LatLngFromDegrees(25.802018, -80.269892))
	triangle := s2.PolygonFromLoops(
		[]*s2.Loop{s2.LoopFromPoints([]s2.Point{bermuda, miami, culebra})})
	area := earth.AreaFromSteradians(triangle.Area())
	fmt.Printf("Bermuda Triangle is %.2f square miles", area.SquareMiles())
	// Output: Bermuda Triangle is 464541.15 square miles
}

func ExampleSteradiansFromArea() {
	steradians := earth.SteradiansFromArea(unit.SquareCentimeter)
	fmt.Printf("1 square centimeter is %g steradians, close to a level %d cell",
		steradians, s2.AvgAreaMetric.ClosestLevel(steradians))
	// Output: 1 square centimeter is 2.4636750563592804e-18 steradians, close to a level 30 cell
}

func ExampleInitialBearingFromLatLngs() {
	christchurch := s2.LatLngFromDegrees(-43.491402, 172.522275)
	riogrande := s2.LatLngFromDegrees(-53.777156, -67.734719)
	bearing := earth.InitialBearingFromLatLngs(christchurch, riogrande)
	fmt.Printf("Head southeast (%.2f°)", bearing.Degrees())
	// Output: Head southeast (146.90°)
}
