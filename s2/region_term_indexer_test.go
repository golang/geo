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
	"fmt"
	"reflect"
	"sort"
	"testing"
)

type QueryType int

const (
	POINT QueryType = iota + 1
	CAP
)

var NUM_ITERATIONS = int(100)

func testRandomCaps(t *testing.T, option Options, queryType QueryType) {
	regionsTermIndexer := NewRegionTermIndexerWithOptions(option)
	regionCoverer := NewRegionCoverer()
	regionCoverer.LevelMod = option.levelMod
	regionCoverer.MaxCells = option.maxCells
	regionCoverer.MaxLevel = option.maxLevel
	regionCoverer.MinLevel = option.minLevel

	var caps []Cap
	var coverings []CellUnion

	index := make(map[string][]int, 10)
	var indexTerms, queryTerms int

	for i := 0; i < NUM_ITERATIONS; i++ {
		var cap Cap
		var terms []string

		if option.pointsOnly {
			cap = CapFromPoint(randomPoint())
			terms = regionsTermIndexer.GetIndexTermsForPoint(cap.Center(), "")
		} else {
			cap = randomCap(.3*AvgAreaMetric.Value(option.maxLevel),
				4*AvgAreaMetric.Value(option.minLevel))
			terms = regionsTermIndexer.GetIndexTermsForRegion(cap, "")
		}
		caps = append(caps, cap)
		coverings = append(coverings, regionCoverer.Covering(cap))
		for _, term := range terms {
			index[term] = append(index[term], i)
		}
		indexTerms += len(terms)
	}

	//fmt.Printf("index : %+v", index)
	for i := 0; i < NUM_ITERATIONS; i++ {
		var cap Cap
		var terms []string

		if queryType == CAP {
			cap = CapFromPoint(randomPoint())
			terms = regionsTermIndexer.GetQueryTermsForPoint(cap.Center(), "")
		} else {
			cap = randomCap(.3*AvgAreaMetric.Value(option.maxLevel),
				4*AvgAreaMetric.Value(option.minLevel))
			terms = regionsTermIndexer.GetQueryTermsForRegion(cap, "")
		}

		covering := regionCoverer.Covering(cap)
		var expected, actual []int
		for j := 0; j < len(caps); j++ {
			if covering.Intersects(coverings[j]) {
				expected = append(expected, j)
			}
		}

		for _, term := range terms {
			docIDs := index[term]
			fmt.Printf("term: %s docIDs %+v", term, docIDs)
			for _, id := range docIDs {
				actual = append(actual, id)
			}
		}

		deduplicate(&actual)

		sort.Ints(expected)
		sort.Ints(actual)

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("expected: %+v Vs actual: %+v", expected, actual)
		}

		queryTerms += len(terms)
	}
}

func deduplicate(labels *[]int) {
	encountered := make(map[int]struct{})

	for v := range *labels {
		encountered[(*labels)[v]] = struct{}{}
	}

	(*labels) = (*labels)[:0]
	for key, _ := range encountered {
		*labels = append(*labels, key)
	}
}

func TestIndexRegionsQueryRegionsOptimizeTime(t *testing.T) {
	options := Options{maxLevel: 16,
		maxCells:   20,
		levelMod:   1,
		pointsOnly: false}
	testRandomCaps(t, options, CAP)
}

func TestIndexRegionsQueryPointsOptimizeTime(t *testing.T) {
	options := Options{maxLevel: 16,
		maxCells:   20,
		levelMod:   1,
		pointsOnly: false}
	testRandomCaps(t, options, POINT)
}

func TestIndexRegionsQueryRegionsOptimizeTimeWithLevelMod(t *testing.T) {
	options := Options{minLevel: 6,
		maxLevel:   12,
		levelMod:   3,
		pointsOnly: false}
	testRandomCaps(t, options, CAP)
}

func TestIndexRegionsQueryRegionsOptimizeSpace(t *testing.T) {
	options := Options{minLevel: 4,
		levelMod:      1,
		maxLevel:      30,
		optimizeSpace: true,
		maxCells:      8,
	}
	testRandomCaps(t, options, CAP)
}

func TestIndexPointsQueryRegionsOptimizeTime(t *testing.T) {
	options := Options{minLevel: 0,
		levelMod:      2,
		maxLevel:      30,
		optimizeSpace: false,
		pointsOnly:    true,
		maxCells:      20,
	}
	testRandomCaps(t, options, CAP)
}

func TestIndexPointsQueryRegionsOptimizeSpace(t *testing.T) {
	options := Options{minLevel: 4,
		levelMod:      1,
		maxLevel:      16,
		optimizeSpace: true,
		maxCells:      8,
	}
	testRandomCaps(t, options, CAP)
}
