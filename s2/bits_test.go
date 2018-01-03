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
	"testing"
)

func TestFindMSBSetNonZero64(t *testing.T) {
	testOne := uint64(0x8000000000000000)
	testAll := uint64(0xFFFFFFFFFFFFFFFF)
	testSome := uint64(0xFEDCBA9876543210)
	for i := 63; i >= 0; i-- {
		if got := findMSBSetNonZero64(testOne); got != i {
			t.Errorf("findMSBSetNonZero64(%x) = %d, want = %d", testOne, got, i)
		}
		if got := findMSBSetNonZero64(testAll); got != i {
			t.Errorf("findMSBSetNonZero64(%x) = %d, want = %d", testAll, got, i)
		}
		if got := findMSBSetNonZero64(testSome); got != i {
			t.Errorf("findMSBSetNonZero64(%x) = %d, want = %d", testSome, got, i)
		}
		testOne >>= 1
		testAll >>= 1
		testSome >>= 1
	}

	if got := findMSBSetNonZero64(1); got != 0 {
		t.Errorf("findMSBSetNonZero64(1) = %v, want 0", got)
	}

	if got := findMSBSetNonZero64(0); got != 0 {
		t.Errorf("findMSBSetNonZero64(0) = %v, want 0", got)
	}
}

func TestFindLSBSetNonZero64(t *testing.T) {
	testOne := uint64(0x0000000000000001)
	testAll := uint64(0xFFFFFFFFFFFFFFFF)
	testSome := uint64(0x0123456789ABCDEF)
	for i := 0; i < 64; i++ {
		if got := findLSBSetNonZero64(testOne); got != i {
			t.Errorf("findLSBSetNonZero64(%x) = %d, want = %d", testOne, got, i)
		}
		if got := findLSBSetNonZero64(testAll); got != i {
			t.Errorf("findLSBSetNonZero64(%x) = %d, want = %d", testAll, got, i)
		}
		if got := findLSBSetNonZero64(testSome); got != i {
			t.Errorf("findLSBSetNonZero64(%x) = %d, want = %d", testSome, got, i)
		}
		testOne <<= 1
		testAll <<= 1
		testSome <<= 1
	}

	if got := findLSBSetNonZero64(0); got != 0 {
		t.Errorf("findLSBSetNonZero64(0) = %v, want 0", got)
	}
}
