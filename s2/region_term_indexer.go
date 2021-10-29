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
	"github.com/golang/geo/s1"
)

type TermType int

var marker = string('$')

const (
	ANCESTOR TermType = iota + 1
	COVERING
)

var defaultMaxCells = int(8)

type Options struct {
	maxCells      int
	minLevel      int
	maxLevel      int
	levelMod      int
	pointsOnly    bool
	optimizeSpace bool
}

func (o *Options) MaxCells() int {
	return o.maxCells
}

func (o *Options) SetMaxCells(mc int) {
	o.maxCells = mc
}

func (o *Options) MinLevel() int {
	return o.minLevel
}

func (o *Options) SetMinLevel(ml int) {
	o.minLevel = ml
}

func (o *Options) MaxLevel() int {
	return o.maxLevel
}

func (o *Options) SetMaxLevel(ml int) {
	o.maxLevel = ml
}

func (o *Options) LevelMod() int {
	return o.levelMod
}

func (o *Options) SetLevelMod(lm int) {
	o.levelMod = lm
}

func (o *Options) SetPointsOnly(v bool) {
	o.pointsOnly = v
}

func (o *Options) SetOptimizeSpace(v bool) {
	o.optimizeSpace = v
}

func (o *Options) trueMaxLevel() int {
	trueMax := o.maxLevel
	if o.levelMod != 1 {
		trueMax = o.maxLevel - (o.maxLevel-o.minLevel)%o.levelMod
	}
	return trueMax
}

type RegionTermIndexer struct {
	options       Options
	regionCoverer RegionCoverer
}

func NewRegionTermIndexer() *RegionTermIndexer {
	rv := &RegionTermIndexer{
		options: Options{
			maxCells: 8,
			minLevel: 4,
			maxLevel: 16,
			levelMod: 1,
		},
	}
	return rv
}

func NewRegionTermIndexerWithOptions(option Options) *RegionTermIndexer {
	return &RegionTermIndexer{options: option}
}

func (rti *RegionTermIndexer) GetTerm(termTyp TermType, id CellID,
	prefix string) string {
	return prefix + id.ToToken()
	/*
		    TODO - revisit this if needed.
			if termTyp == ANCESTOR {
				return prefix + id.ToToken()
			}
			return prefix + marker + id.ToToken()
	*/
}

func (rti *RegionTermIndexer) GetIndexTermsForPoint(p Point, prefix string) []string {
	cellID := cellIDFromPoint(p)
	var rv []string
	for l := rti.options.minLevel; l <= rti.options.maxLevel; l += rti.options.levelMod {
		rv = append(rv, rti.GetTerm(ANCESTOR, cellID.Parent(l), prefix))
	}
	return rv
}

func (rti *RegionTermIndexer) GetIndexTermsForRegion(region Region,
	prefix string) []string {
	rti.regionCoverer.LevelMod = rti.options.levelMod
	rti.regionCoverer.MaxLevel = rti.options.maxLevel
	rti.regionCoverer.MinLevel = rti.options.minLevel
	rti.regionCoverer.MaxCells = rti.options.maxCells

	covering := rti.regionCoverer.Covering(region)
	return rti.GetIndexTermsForCanonicalCovering(covering, prefix)
}

func (rti *RegionTermIndexer) GetIndexTermsForCanonicalCovering(
	covering CellUnion, prefix string) []string {
	var rv []string
	prevID := CellID(0)
	tml := rti.options.trueMaxLevel()

	for _, cellID := range covering {
		level := cellID.Level()
		if level < tml {
			rv = append(rv, rti.GetTerm(COVERING, cellID, prefix))
		}

		if level == tml || !rti.options.optimizeSpace {
			rv = append(rv, rti.GetTerm(ANCESTOR, cellID.Parent(level), prefix))
		}

		for (level - rti.options.levelMod) >= rti.options.minLevel {
			level -= rti.options.levelMod
			ancestorID := cellID.Parent(level)
			if prevID != CellID(0) && prevID.Level() > level &&
				prevID.Parent(level) == ancestorID {
				break
			}
			rv = append(rv, rti.GetTerm(ANCESTOR, ancestorID, prefix))
		}
		prevID = cellID
	}

	return rv
}

func (rti *RegionTermIndexer) GetQueryTermsForPoint(p Point, prefix string) []string {
	cellID := cellIDFromPoint(p)
	var rv []string

	level := rti.options.trueMaxLevel()
	rv = append(rv, rti.GetTerm(ANCESTOR, cellID.Parent(level), prefix))
	if rti.options.pointsOnly {
		return rv
	}

	for level >= rti.options.minLevel {
		rv = append(rv, rti.GetTerm(COVERING, cellID.Parent(level), prefix))
		level -= rti.options.levelMod
	}

	return rv
}

func (rti *RegionTermIndexer) GetQueryTermsForRegion(region Region,
	prefix string) []string {
	rti.regionCoverer.LevelMod = rti.options.levelMod
	rti.regionCoverer.MaxLevel = rti.options.maxLevel
	rti.regionCoverer.MinLevel = rti.options.minLevel
	rti.regionCoverer.MaxCells = rti.options.maxCells

	covering := rti.regionCoverer.Covering(region)
	return rti.GetQueryTermsForCanonicalCovering(covering, prefix)

}

func (rti *RegionTermIndexer) GetQueryTermsForCanonicalCovering(
	covering CellUnion, prefix string) []string {
	var rv []string
	prevID := CellID(0)
	tml := rti.options.trueMaxLevel()
	for _, cellID := range covering {
		level := cellID.Level()
		rv = append(rv, rti.GetTerm(ANCESTOR, cellID, prefix))

		if rti.options.pointsOnly {
			continue
		}

		if rti.options.optimizeSpace && level < tml {
			rv = append(rv, rti.GetTerm(COVERING, cellID, prefix))
		}

		for level-rti.options.levelMod >= rti.options.minLevel {
			level -= rti.options.levelMod
			ancestorID := cellID.Parent(level)
			if prevID != CellID(0) && prevID.Level() > level &&
				prevID.Parent(level) == ancestorID {
				break
			}
			rv = append(rv, rti.GetTerm(COVERING, ancestorID, prefix))
		}

		prevID = cellID
	}

	return rv
}

func CapFromCenterAndRadius(centerLat, centerLon, dist float64) Cap {
	return CapFromCenterAngle(PointFromLatLng(
		LatLngFromDegrees(centerLat, centerLon)), s1.Angle((dist/1000)/6378))
}
