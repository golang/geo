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

func TestMetric(t *testing.T) {
	// This is not a thorough test.
	// TODO(dsymonds): Exercise this more.
	if got := MinWidthMetric.MaxLevel(0.001256); got != 9 {
		t.Errorf("MinWidthMetric.MaxLevel(0.001256) = %d, want 9", got)
	}
}
