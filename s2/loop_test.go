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

import (
	"testing"
)

func TestEmptyFullLoops(t *testing.T) {
	if !emptyLoop.IsEmpty() {
		t.Errorf("empty loop should be empty")
	}
	if emptyLoop.IsFull() {
		t.Errorf("empty loop should not be full")
	}
	if !emptyLoop.isEmptyOrFull() {
		t.Errorf("empty loop should pass IsEmptyOrFull")
	}

	if fullLoop.IsEmpty() {
		t.Errorf("full loop should not be empty")
	}
	if !fullLoop.IsFull() {
		t.Errorf("full loop should be full")
	}
	if !fullLoop.isEmptyOrFull() {
		t.Errorf("full loop should pass IsEmptyOrFull")
	}

	if !empty.RectBound().IsEmpty() {
		t.Errorf("empty loops RectBound should be empty")
	}

	if !full.RectBound().IsFull() {
		t.Errorf("full loops RectBound should be full")
	}

}
