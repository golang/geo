/*
Copyright 2016 Google Inc. All rights reserved.

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
	"testing"
)

// testShape is a minimal implementation of the Shape interface for use in testing
// until such time as there are other s2 types that implement it.
type testShape struct {
	a, b  Point
	edges int
}

func newTestShape() *testShape {
	return &testShape{}
}

func (s *testShape) NumEdges() int {
	return s.edges
}

func (s *testShape) Edge(id int) (a, b Point) {
	return s.a, s.b
}

func (s *testShape) HasInterior() bool {
	return false
}

func (s *testShape) ContainsOrigin() bool {
	return false
}

func TestShapeIndexBasics(t *testing.T) {
	si := NewShapeIndex()
	s := newTestShape()

	if si.Len() != 0 {
		t.Errorf("initial index should be empty after creation")
	}
	si.Add(s)

	if si.Len() == 0 {
		t.Errorf("index should not be empty after adding shape")
	}

	si.Reset()
	if si.Len() != 0 {
		t.Errorf("index should be empty after reset")
	}
}
