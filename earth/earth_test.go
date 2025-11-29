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

package earth

import (
	"testing"
)

func TestRadius(t *testing.T) {
	// Verify the Earth radius constant is approximately correct.
	if got, want := Radius.Kilometers(), 6371.01; got != want {
		t.Errorf("Radius.Kilometers() = %v, want %v", got, want)
	}
}

func TestLowestAltitude(t *testing.T) {
	// Mariana Trench depth
	if got, want := LowestAltitude.Meters(), -10898.0; got != want {
		t.Errorf("LowestAltitude.Meters() = %v, want %v", got, want)
	}
}

func TestHighestAltitude(t *testing.T) {
	// Mount Everest height
	if got, want := HighestAltitude.Meters(), 8848.0; got != want {
		t.Errorf("HighestAltitude.Meters() = %v, want %v", got, want)
	}
}
