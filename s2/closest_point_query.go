// Copyright 2015 Google Inc. All rights reserved.
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
	"container/heap"
	"sort"

	"github.com/golang/geo/s1"
)

// minPointsToEnqueue is the minimum number of points in a cell required
// to enqueue it rather than process its contents directly.
const minPointsToEnqueue = 13

// closestPointResultHeap is a max-heap of ClosestPointQueryResult values,
// ordered by distance (largest distance at the top).
type closestPointResultHeap[Data comparable] []ClosestPointQueryResult[Data]

func (h closestPointResultHeap[Data]) Len() int { return len(h) }
func (h closestPointResultHeap[Data]) Less(i, j int) bool {
	return h[i].distance > h[j].distance
}
func (h closestPointResultHeap[Data]) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *closestPointResultHeap[Data]) Push(x any) {
	*h = append(*h, x.(ClosestPointQueryResult[Data]))
}
func (h *closestPointResultHeap[Data]) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}
func (h closestPointResultHeap[Data]) top() ClosestPointQueryResult[Data] { return h[0] }

// ClosestPointQueryResult holds one result from ClosestPointQuery.
type ClosestPointQueryResult[Data comparable] struct {
	distance  s1.ChordAngle
	pointData PointData[Data]
}

// Distance returns the distance from the target to this point.
func (r ClosestPointQueryResult[Data]) Distance() s1.ChordAngle { return r.distance }

// Point returns the indexed point.
func (r ClosestPointQueryResult[Data]) Point() Point { return r.pointData.Point() }

// Data returns the client data associated with this point.
func (r ClosestPointQueryResult[Data]) Data() Data { return r.pointData.Data() }

// IsEmpty reports whether this result is empty (FindClosestPoint found nothing).
func (r ClosestPointQueryResult[Data]) IsEmpty() bool {
	return r.distance == s1.InfChordAngle()
}

// ClosestPointQueryOptions controls the set of points returned by ClosestPointQuery.
// By default all points are returned, so always set MaxResults and/or DistanceLimit.
type ClosestPointQueryOptions struct {
	common *queryOptions
}

// NewClosestPointQueryOptions returns default options for closest-point queries.
func NewClosestPointQueryOptions() *ClosestPointQueryOptions {
	return &ClosestPointQueryOptions{common: newQueryOptions(minDistance(0))}
}

// MaxResults specifies that at most n points should be returned. n must be >= 1.
func (o *ClosestPointQueryOptions) MaxResults(n int) *ClosestPointQueryOptions {
	o.common = o.common.MaxResults(n)
	return o
}

// DistanceLimit specifies that only points whose distance to the target is
// strictly less than the limit should be returned.
func (o *ClosestPointQueryOptions) DistanceLimit(limit s1.ChordAngle) *ClosestPointQueryOptions {
	o.common = o.common.DistanceLimit(limit)
	return o
}

// InclusiveDistanceLimit is like DistanceLimit but also returns points
// whose distance is exactly equal to the limit.
func (o *ClosestPointQueryOptions) InclusiveDistanceLimit(limit s1.ChordAngle) *ClosestPointQueryOptions {
	o.common = o.common.ClosestInclusiveDistanceLimit(limit)
	return o
}

// ConservativeDistanceLimit expands the limit by the maximum distance
// calculation error, ensuring all points whose true distance is <= limit
// are returned (along with some slightly further ones).
func (o *ClosestPointQueryOptions) ConservativeDistanceLimit(limit s1.ChordAngle) *ClosestPointQueryOptions {
	o.common = o.common.ClosestConservativeDistanceLimit(limit)
	return o
}

// MaxError specifies that points up to this distance further than the true
// closest may be substituted in the result set, as long as they satisfy
// all other criteria. Only meaningful when MaxResults is also set.
func (o *ClosestPointQueryOptions) MaxError(dist s1.ChordAngle) *ClosestPointQueryOptions {
	o.common = o.common.MaxError(dist)
	return o
}

// Region specifies that results must be contained by the given region.
func (o *ClosestPointQueryOptions) Region(region Region) *ClosestPointQueryOptions {
	o.common.region = region
	return o
}

// UseBruteForce forces the brute-force algorithm. Useful for testing.
func (o *ClosestPointQueryOptions) UseBruteForce(x bool) *ClosestPointQueryOptions {
	o.common = o.common.UseBruteForce(x)
	return o
}

// ClosestPointQuery finds the closest point(s) in a PointIndex to a given
// target (point, edge, cell, or shape index).
//
// Example:
//
//	index := &PointIndex[int]{}
//	for i, p := range indexPoints {
//	    index.Add(p, i)
//	}
//	query := NewClosestPointQuery(index, NewClosestPointQueryOptions().MaxResults(5))
//	target := NewMinDistanceToPointTarget(queryPoint)
//	for _, result := range query.FindClosestPoints(target) {
//	    // result.Distance(), result.Point(), result.Data()
//	}
//
// ClosestPointQuery is not safe for concurrent use without external synchronization.
type ClosestPointQuery[Data comparable] struct {
	index  *PointIndex[Data]
	opts   *queryOptions
	target distanceTarget

	useConservativeCellDistance bool

	// Precomputed covering of the indexed points; cleared on ReInit.
	indexCovering []CellID

	// Distance limit, tightened as results are found.
	distanceLimit distance

	// Result stores — exactly one is used per query based on opts.maxResults.
	resultSingleton ClosestPointQueryResult[Data]
	resultVector    []ClosestPointQueryResult[Data]
	resultSet       closestPointResultHeap[Data]

	// Shared iterator for the optimized algorithm; shared across processOrEnqueue calls.
	iter PointIndexIterator[Data]

	// Scratch space for direct processing of small cells. Avoids per-call allocation.
	tmpPointData [minPointsToEnqueue - 1]PointData[Data]

	// Priority queue for candidate cells.
	queue *queryQueue
}

// NewClosestPointQuery returns a ClosestPointQuery for the given index.
// Pass nil opts to use default options (returns all points).
func NewClosestPointQuery[Data comparable](index *PointIndex[Data], opts *ClosestPointQueryOptions) *ClosestPointQuery[Data] {
	if opts == nil {
		opts = NewClosestPointQueryOptions()
	}
	q := &ClosestPointQuery[Data]{queue: newQueryQueue()}
	q.Init(index, opts)
	return q
}

// Init (re)initializes the query for the given index and options.
// Must be called (or ReInit called) if the index is modified after this.
func (q *ClosestPointQuery[Data]) Init(index *PointIndex[Data], opts *ClosestPointQueryOptions) {
	q.index = index
	if opts != nil {
		q.opts = opts.common
	}
	q.ReInit()
}

// ReInit must be called whenever the underlying index is modified.
func (q *ClosestPointQuery[Data]) ReInit() {
	q.iter.Init(q.index)
	q.indexCovering = nil
}

// Options returns the current query options.
func (q *ClosestPointQuery[Data]) Options() *ClosestPointQueryOptions {
	return &ClosestPointQueryOptions{common: q.opts}
}

// FindClosestPoints returns all points satisfying the current options, sorted
// by distance (closest first). This may be called multiple times.
func (q *ClosestPointQuery[Data]) FindClosestPoints(target distanceTarget) []ClosestPointQueryResult[Data] {
	return q.findClosestPoints(target, q.opts)
}

// FindClosestPoint returns the single closest point. If no point satisfies
// the search criteria, returns a result with IsEmpty() == true.
func (q *ClosestPointQuery[Data]) FindClosestPoint(target distanceTarget) ClosestPointQueryResult[Data] {
	opts := *q.opts
	opts.maxResults = 1
	results := q.findClosestPoints(target, &opts)
	if len(results) == 0 {
		return ClosestPointQueryResult[Data]{distance: s1.InfChordAngle()}
	}
	return results[0]
}

// GetDistance returns the minimum distance to the target.
// Returns InfChordAngle if the index or target is empty.
// Use IsDistanceLess if only comparing against a threshold.
func (q *ClosestPointQuery[Data]) GetDistance(target distanceTarget) s1.ChordAngle {
	return q.FindClosestPoint(target).Distance()
}

// IsDistanceLess reports whether the distance to target is less than limit.
// This is usually faster than GetDistance since the search can stop early.
func (q *ClosestPointQuery[Data]) IsDistanceLess(target distanceTarget, limit s1.ChordAngle) bool {
	opts := *q.opts
	opts.maxResults = 1
	opts.distanceLimit = limit
	opts.maxError = s1.StraightChordAngle
	return len(q.findClosestPoints(target, &opts)) > 0
}

// IsDistanceLessOrEqual reports whether the distance to target is <= limit.
func (q *ClosestPointQuery[Data]) IsDistanceLessOrEqual(target distanceTarget, limit s1.ChordAngle) bool {
	return q.IsDistanceLess(target, limit.Successor())
}

// IsConservativeDistanceLessOrEqual reports whether the true distance to the
// target is likely <= limit. It accounts for rounding error: all points whose
// true distance is <= limit are guaranteed to be found.
func (q *ClosestPointQuery[Data]) IsConservativeDistanceLessOrEqual(target distanceTarget, limit s1.ChordAngle) bool {
	opts := *q.opts
	opts.maxResults = 1
	opts.distanceLimit = limit.Expanded(minUpdateDistanceMaxError(limit)).Successor()
	opts.maxError = s1.StraightChordAngle
	return len(q.findClosestPoints(target, &opts)) > 0
}

func (q *ClosestPointQuery[Data]) findClosestPoints(target distanceTarget, opts *queryOptions) []ClosestPointQueryResult[Data] {
	q.findClosestPointsInternal(target, opts)

	if opts.maxResults == 1 {
		if q.resultSingleton.IsEmpty() {
			return nil
		}
		return []ClosestPointQueryResult[Data]{q.resultSingleton}
	}

	if opts.maxResults == maxQueryResults {
		sort.Slice(q.resultVector, func(i, j int) bool {
			return q.resultVector[i].distance < q.resultVector[j].distance
		})
		results := q.resultVector
		q.resultVector = nil
		return results
	}

	// Drain the max-heap (largest first) then reverse to get ascending order.
	results := make([]ClosestPointQueryResult[Data], 0, q.resultSet.Len())
	for q.resultSet.Len() > 0 {
		results = append(results, heap.Pop(&q.resultSet).(ClosestPointQueryResult[Data]))
	}
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}
	return results
}

func (q *ClosestPointQuery[Data]) findClosestPointsInternal(target distanceTarget, opts *queryOptions) {
	q.target = target
	q.opts = opts

	q.distanceLimit = minDistance(opts.distanceLimit)
	q.resultSingleton = ClosestPointQueryResult[Data]{distance: s1.InfChordAngle()}
	q.resultVector = nil
	q.resultSet = closestPointResultHeap[Data]{}

	if q.distanceLimit == minDistance(0) {
		return
	}

	targetUsesMaxError := opts.maxError != 0 && target.setMaxError(opts.maxError)
	q.useConservativeCellDistance = targetUsesMaxError &&
		(q.distanceLimit == minDistance(0).infinity() ||
			minDistance(0).less(q.distanceLimit.sub(minDistance(opts.maxError))))

	if opts.useBruteForce || q.index.NumPoints() <= target.maxBruteForceIndexSize() {
		q.findClosestPointsBruteForce()
	} else {
		q.findClosestPointsOptimized()
	}
}

func (q *ClosestPointQuery[Data]) findClosestPointsBruteForce() {
	for it := NewPointIndexIterator(q.index); !it.Done(); it.Next() {
		q.maybeAddResult(it.PointData())
	}
}

func (q *ClosestPointQuery[Data]) findClosestPointsOptimized() {
	q.initQueue()
	for q.queue.size() > 0 {
		entry := q.queue.pop()
		if !entry.distance.less(q.distanceLimit) {
			q.queue.reset()
			break
		}
		child := entry.id.ChildBegin()
		seek := true
		for i := 0; i < 4; i++ {
			seek = q.processOrEnqueue(child, seek)
			child = child.Next()
		}
	}
}

func (q *ClosestPointQuery[Data]) initQueue() {
	cb := q.target.capBound()
	if cb.IsEmpty() {
		return
	}

	if q.opts.maxResults == 1 {
		// Optimization: seek near the target cap center to get an early upper
		// bound on the search radius. The two adjacent index points (in CellID
		// order) often yield a tight bound.
		q.iter.Seek(cellIDFromPoint(cb.Center()))
		if !q.iter.Done() {
			q.maybeAddResult(q.iter.PointData())
		}
		if q.iter.Prev() {
			q.maybeAddResult(q.iter.PointData())
		}
		if q.distanceLimit == minDistance(0) {
			return
		}
	}

	if q.indexCovering == nil {
		q.initCovering()
	}

	initialCells := []CellID(q.indexCovering)

	if q.opts.region != nil {
		coverer := &RegionCoverer{MaxCells: 4, LevelMod: 1, MaxLevel: MaxLevel}
		regionCover := coverer.Covering(q.opts.region)
		initialCells = CellUnionFromIntersection(CellUnion(q.indexCovering), regionCover)
	}

	if q.distanceLimit != minDistance(0).infinity() {
		coverer := &RegionCoverer{MaxCells: 4, LevelMod: 1, MaxLevel: MaxLevel}
		radius := cb.Radius() + q.distanceLimit.chordAngleBound().Angle()
		searchCap := CapFromCenterAngle(cb.Center(), radius)
		maxDistCover := coverer.FastCovering(searchCap)
		initialCells = CellUnionFromIntersection(CellUnion(initialCells), maxDistCover)
	}

	q.iter.Begin()
	for _, id := range initialCells {
		if q.iter.Done() {
			break
		}
		q.processOrEnqueue(id, id.RangeMin() > q.iter.CellID())
	}
}

func (q *ClosestPointQuery[Data]) initCovering() {
	// Compute a minimal covering (at most 6 cells) of all indexed points.
	// See the equivalent method in EdgeQuery for a detailed explanation.
	q.indexCovering = make([]CellID, 0, 6)
	q.iter.Finish()
	if !q.iter.Prev() {
		return // Empty index.
	}
	indexLastID := q.iter.CellID()
	q.iter.Begin()
	if q.iter.CellID() != indexLastID {
		level, ok := q.iter.CellID().CommonAncestorLevel(indexLastID)
		if !ok {
			level = 0
		} else {
			level++
		}
		lastID := indexLastID.Parent(level)
		for id := q.iter.CellID().Parent(level); id != lastID; id = id.Next() {
			if id.RangeMax() < q.iter.CellID() {
				continue
			}
			cellFirstID := q.iter.CellID()
			q.iter.Seek(id.RangeMax().Next())
			q.iter.Prev()
			cellLastID := q.iter.CellID()
			q.iter.Next()
			q.addInitialRange(cellFirstID, cellLastID)
		}
	}
	q.addInitialRange(q.iter.CellID(), indexLastID)
}

// addInitialRange appends to indexCovering the lowest common ancestor of firstID and lastID.
func (q *ClosestPointQuery[Data]) addInitialRange(firstID, lastID CellID) {
	level, _ := firstID.CommonAncestorLevel(lastID)
	q.indexCovering = append(q.indexCovering, firstID.Parent(level))
}

func (q *ClosestPointQuery[Data]) maybeAddResult(pd PointData[Data]) {
	dist := q.distanceLimit
	updated, ok := q.target.updateDistanceToPoint(pd.Point(), dist)
	if !ok {
		return
	}
	if q.opts.region != nil && !q.opts.region.ContainsPoint(pd.Point()) {
		return
	}
	result := ClosestPointQueryResult[Data]{
		distance:  updated.chordAngle(),
		pointData: pd,
	}
	switch {
	case q.opts.maxResults == 1:
		q.resultSingleton = result
		q.distanceLimit = updated.sub(minDistance(q.opts.maxError))
	case q.opts.maxResults == maxQueryResults:
		q.resultVector = append(q.resultVector, result)
	default:
		if q.resultSet.Len() >= q.opts.maxResults {
			heap.Pop(&q.resultSet)
		}
		heap.Push(&q.resultSet, result)
		if q.resultSet.Len() >= q.opts.maxResults {
			q.distanceLimit = minDistance(q.resultSet.top().distance).sub(minDistance(q.opts.maxError))
		}
	}
}

// processOrEnqueue either processes the points in id directly (if few enough)
// or enqueues id for later subdivision.
//
// If seek is false, q.iter must already be positioned at the first indexed
// point within or after id. Returns true if the cell was enqueued (caller
// must seek for the next sibling), false if it was processed (q.iter is now
// positioned at the next cell in CellID order).
func (q *ClosestPointQuery[Data]) processOrEnqueue(id CellID, seek bool) bool {
	if seek {
		q.iter.Seek(id.RangeMin())
	}
	if id.IsLeaf() {
		for !q.iter.Done() && q.iter.CellID() == id {
			q.maybeAddResult(q.iter.PointData())
			q.iter.Next()
		}
		return false
	}
	last := id.RangeMax()
	numPoints := 0
	for !q.iter.Done() && q.iter.CellID() <= last {
		if numPoints == minPointsToEnqueue-1 {
			// Cell has at least minPointsToEnqueue points; enqueue for subdivision.
			cell := CellFromCellID(id)
			dist := q.distanceLimit
			if updated, ok := q.target.updateDistanceToCell(cell, dist); ok {
				if q.opts.region == nil || q.opts.region.IntersectsCell(cell) {
					if q.useConservativeCellDistance {
						updated = updated.sub(minDistance(q.opts.maxError))
					}
					q.queue.push(&queryQueueEntry{distance: updated, id: id})
				}
			}
			return true
		}
		q.tmpPointData[numPoints] = q.iter.PointData()
		numPoints++
		q.iter.Next()
	}
	for i := 0; i < numPoints; i++ {
		q.maybeAddResult(q.tmpPointData[i])
	}
	return false
}
