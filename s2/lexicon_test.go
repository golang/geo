// Copyright 2020 Google Inc. All rights reserved.
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
	"reflect"
	"testing"
)

func TestSequenceLexiconAdd(t *testing.T) {
	tests := []struct {
		have []int32
		want int32
	}{
		{have: []int32{}, want: 0},
		{have: []int32{5}, want: 1},
		{have: []int32{}, want: 0},
		{have: []int32{5, 5}, want: 2},
		{have: []int32{5, 0, -3}, want: 3},
		{have: []int32{5}, want: 1},
		{have: []int32{0x7fffffff}, want: 4},
		{have: []int32{5, 0, -3}, want: 3},
		{have: []int32{}, want: 0},
	}

	lex := newSequenceLexicon()
	for _, test := range tests {
		if got := lex.add(test.have); got != test.want {
			t.Errorf("lexicon.add(%v) = %v, want %v", test.have, got, test.want)
		}

	}

	if lex.size() != 5 {
		t.Errorf("lexicon.size() = %v, want 5", lex.size())
	}

	for _, test := range tests {
		if got := lex.sequence(test.want); !reflect.DeepEqual(got, test.have) {
			t.Errorf("lexicon.sequence(%v) = %v, want %v", test.want, got, test.have)
		}
	}
}

func TestSequenceLexiconClear(t *testing.T) {
	lex := newSequenceLexicon()

	if got, want := lex.add([]int32{1}), int32(0); got != want {
		t.Errorf("lex.add([]int32{1}) = %v, want %v", got, want)
	}
	if got, want := lex.add([]int32{2}), int32(1); got != want {
		t.Errorf("lex.add(sequence{2}) = %v, want %v", got, want)
	}
	lex.clear()
	if got, want := lex.add([]int32{2}), int32(0); got != want {
		t.Errorf("lex.add([]int32{2}) = %v, want %v", got, want)
	}
	if got, want := lex.add([]int32{1}), int32(1); got != want {
		t.Errorf("lex.add([]int32{1}) = %v, want %v", got, want)
	}
}

func TestIDSetLexiconSingletonSets(t *testing.T) {
	var m int32 = math.MaxInt32
	tests := []struct {
		have int32
		want int32
	}{
		{5, 5},
		{0, 0},
		{1, 1},
		{m, m},
	}

	lex := newIDSetLexicon()
	// Test adding
	for _, test := range tests {
		if got := lex.add(test.have); got != test.want {
			t.Errorf("lexicon.add(%v) = %v, want %v", test.have, got, test.want)
		}
	}

	// Test recall
	for _, test := range tests {
		if got := lex.idSet(test.want); !reflect.DeepEqual(got, []int32{test.have}) {
			t.Errorf("lexicon.idSet(%v) = %v, want %v", test.want, got, test.have)
		}
	}
}

func TestIDSetLexiconSetsAreSorted(t *testing.T) {
	tests := []struct {
		have []int32
		want int32
	}{
		// This test relies on order of test cases to get the expected IDs.
		{
			have: []int32{2, 5},
			want: ^0,
		},
		{
			have: []int32{3, 2, 5},
			want: ^1,
		},
		{
			have: []int32{2, 2, 2, 2, 5, 2, 5},
			want: ^0,
		},
		{
			have: []int32{2, 5},
			want: ^0,
		},
		{
			have: []int32{5, 3, 2, 5},
			want: ^1,
		},
	}

	lexicon := newIDSetLexicon()
	for _, test := range tests {
		if got := lexicon.add(test.have...); got != test.want {
			t.Errorf("lexicon.addSet(%v) = %v, want %v", test.have, got, test.want)
		}
	}

	recallTests := []struct {
		have int32
		want []int32
	}{
		{
			have: ^0,
			want: []int32{2, 5},
		},
		{
			have: ^1,
			want: []int32{2, 3, 5},
		},
	}

	for _, test := range recallTests {
		if got := lexicon.idSet(test.have); !reflect.DeepEqual(got, test.want) {
			t.Errorf("lexicon.idSet(%v) = %+v, want %+v", test.have, got, test.want)
		}
	}
}

func TestIDSetLexiconClear(t *testing.T) {
	lex := newIDSetLexicon()

	if got, want := lex.add(1, 2), int32(^0); got != want {
		t.Errorf("lex.add([]int32{1, 2}) = %v, want %v", got, want)
	}
	if got, want := lex.add(3, 4), int32(^1); got != want {
		t.Errorf("lex.add(sequence{3, 4}) = %v, want %v", got, want)
	}
	lex.clear()
	if got, want := lex.add(3, 4), int32(^0); got != want {
		t.Errorf("lex.add([]int32{3, 4}) = %v, want %v", got, want)
	}
	if got, want := lex.add(1, 2), int32(^1); got != want {
		t.Errorf("lex.add([]int32{1, 2}) = %v, want %v", got, want)
	}
}

// TODO(roberts): Differences from C++
// Benchmarking methods.
