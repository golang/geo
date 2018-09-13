// Copyright 2018 Google Inc. All rights reserved.
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

package s2

import (
	"math"
	"testing"

	"github.com/golang/geo/r2"
	"github.com/golang/geo/s1"
)

func TestEdgeTessellatorProjectedNoTessellation(t *testing.T) {
	proj := NewPlateCarreeProjection(180)
	tess := NewEdgeTessellator(proj, 0.01*s1.Degree)

	var vertices []r2.Point
	vertices = tess.AppendProjected(PointFromCoords(1, 0, 0), PointFromCoords(0, 1, 0), vertices)

	if len(vertices) != 2 {
		t.Errorf("2 points which don't need tessellation should only end with 2 points. got %d points", len(vertices))
	}
}

func TestEdgeTessellatorUnprojectedNoTessellation(t *testing.T) {
	proj := NewPlateCarreeProjection(180)
	tess := NewEdgeTessellator(proj, 0.01*s1.Degree)

	var vertices []Point
	vertices = tess.AppendUnprojected(r2.Point{0, 30}, r2.Point{0, 50}, vertices)

	if len(vertices) != 2 {
		t.Errorf("2 points which don't need tessellation should only end with 2 points. got %d points", len(vertices))
	}
}

func TestEdgeTessellatorUnprojectedWrapping(t *testing.T) {
	// This tests that a projected edge that crosses the 180 degree meridian
	// goes the "short way" around the sphere.
	proj := NewPlateCarreeProjection(180)
	tess := NewEdgeTessellator(proj, 0.01*s1.Degree)

	var vertices []Point
	vertices = tess.AppendUnprojected(r2.Point{-170, 0}, r2.Point{170, 80}, vertices)
	for i, v := range vertices {
		if got := math.Abs(longitude(v).Degrees()); got < 170 {
			t.Errorf("unprojected segment %d should be close to the meridian. got %v, want >= 170", i, got)
		}
	}
}

func TestEdgeTessellatorProjectedWrapping(t *testing.T) {
	// This tests projecting a geodesic edge that crosses the 180 degree
	// meridian.  This results in a set of vertices that may be non-canonical
	// (i.e., absolute longitudes greater than 180 degrees) but that don't have
	// any sudden jumps in value, which is convenient for interpolating them.
	proj := NewPlateCarreeProjection(180)
	tess := NewEdgeTessellator(proj, 0.01*s1.Degree)

	var vertices []r2.Point
	vertices = tess.AppendProjected(PointFromLatLng(LatLngFromDegrees(0, -170)), PointFromLatLng(LatLngFromDegrees(0, 170)), vertices)
	for i, v := range vertices {
		if v.X > -170 {
			t.Errorf("projected vertex %d should be close to the meridian, got %v, want <= -170 ", i, v.X)
		}
	}
}

func TestEdgeTessellatorUnprojectedWrappingMultipleCrossings(t *testing.T) {
	// Tests an edge chain that crosses the 180 degree meridian multiple times.
	// Note that due to coordinate wrapping, the last vertex of one edge may not
	// exactly match the first edge of the next edge after unprojection.
	proj := NewPlateCarreeProjection(180)
	tess := NewEdgeTessellator(proj, 0.01*s1.Degree)

	var vertices []Point
	for lat := 1.0; lat <= 60; lat += 1.0 {
		vertices = tess.AppendUnprojected(r2.Point{180 - 0.03*lat, lat},
			r2.Point{-180 + 0.07*lat, lat}, vertices)
		vertices = tess.AppendUnprojected(r2.Point{-180 + 0.07*lat, lat},
			r2.Point{180 - 0.03*(lat+1), lat + 1}, vertices)
	}

	for i, v := range vertices {
		if got := math.Abs(longitude(v).Degrees()); got < 175 {
			t.Errorf("vertex %d should be close to the meridian, got %v", i, got)
		}
	}
}

func TestEdgeTessellatorProjectedWrappingMultipleCrossings(t *testing.T) {
	// The following loop crosses the 180 degree meridian four times (twice in
	// each direction).
	loop := parsePoints("0:160, 0:-40, 0:120, 0:-80, 10:120, 10:-40, 0:160")
	proj := NewPlateCarreeProjection(180)
	tess := NewEdgeTessellator(proj, 1e-7*s1.Degree)

	var vertices []r2.Point
	for i := 0; i+1 < len(loop); i++ {
		vertices = tess.AppendProjected(loop[i], loop[i+1], vertices)
	}
	if got, want := vertices[0], vertices[len(vertices)-1]; got != want {
		t.Errorf("the first and last vertices should be the same. got %v, want %v", got, want)
	}

	// Note that the r2.Point coordinates are in (lng, lat) order.
	minLng := vertices[0].X
	maxLng := vertices[0].X
	for _, v := range vertices {
		minLng = math.Min(minLng, v.X)
		maxLng = math.Max(maxLng, v.X)
	}
	if minLng != 160 {
		t.Errorf("minLng = %v, want %v", minLng, 160)
	}
	if maxLng != 640 {
		t.Errorf("maxLng = %v, want %v", maxLng, 640)
	}
}

// TODO(roberts): Differences from C++
// The DistStats accuracy by exhaustion test cases.
