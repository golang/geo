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

// This file is for bit manipulation code.  It will soon be deleted.

import "math/bits"

// findMSBSetNonZero64 returns the index (between 0 and 63) of the most
// significant set bit. Passing zero to this function returns -1.
func findMSBSetNonZero64(x uint64) int {
	return bits.Len64(x) - 1
}

// findLSBSetNonZero64 returns the index (between 0 and 63) of the least
// significant set bit. Passing zero to this function returns 64.
func findLSBSetNonZero64(x uint64) int {
	return bits.TrailingZeros64(x)
}
