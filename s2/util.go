// Copyright 2017 Google Inc. All rights reserved.
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
	"cmp"

	"github.com/golang/geo/s1"
)

// roundAngle returns the value rounded to nearest as an int32.
// This does not match C++ exactly for the case of x.5.
func roundAngle(val s1.Angle) int32 {
	if val < 0 {
		return int32(val - 0.5)
	}
	return int32(val + 0.5)
}

// clamp restricts a value to be within the range [lo, hi].
func clamp[T cmp.Ordered](x, lo, hi T) T {
	return min(max(x, lo), hi)
}
