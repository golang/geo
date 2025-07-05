// Copyright 2025 The S2 Geometry Project Authors. All rights reserved.
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
	"sync"
	"sync/atomic"
)

// indexCellRegion represents a simple pair for defining an integer valued region.
type indexCellRegion struct {
	start int
	size  int
}

// shapeRegion is a mapping from shape id to the region of the edges array
// it's stored in.
type shapeRegion struct {
	id     int32
	region indexCellRegion
}

// edgeAndIDChain is an extension of Edge with fields for the edge id,
// chain id, and offset. It's useful to bundle these together when decoding
// ShapeIndex cells because it allows us to avoid repetitive edge and chain
// lookups in many cases.
type edgeAndIDChain struct {
	Edge         // Embed the Edge type.
	ID     int32 // ID of the edge within its shape.
	Chain  int32 // ID of the chain the edge belongs to.
	Offset int32 // Offset of the edge within the chain.
}

// indexCellData provides methods for working with ShapeIndexCell data. For
// larger queries like validation, we often look up edges multiple times,
// and sometimes need to work with the edges themselves, their edge IDs, or
// their chain and offset.
//
// ShapeIndexCell and the ClippedShape APIs fundamentally work with edge IDs
// and can't be re-worked without significant effort and loss of efficiency.
// This class provides an alternative API than repeatedly querying through the
// shapes in the index.
//
// This is meant to support larger querying and validation operations such as
// ValidationQuery that have to proceed cell-by cell through an index.
//
// To use, simply call loadCell() to decode the contents of a cell.
//
// This type promises that the edges will be looked up once when loadCell() is
// called, and the edges, edgeIDs, chain, and chain offsets are loaded into a
// contiguous chunk of memory that we can serve requests from via slices.
// Since the chain and offset are computed anyways when looking up an edge via
// the shape.Edge() API, we simply cache those values so the cost is minimal.
//
// The memory layout looks like this:
//
//	|     0D Shapes     |     1D Shapes     |     2D Shapes     |  Dimensions
//	|  5  |   1   |  3  |  2  |   7   |  0  |  6  |   4   |  8  |  Shapes
//	[ ......................... Edges ..........................]  Edges
//
// This allows us to look up individual shapes very quickly, as well as all
// shapes in a given dimension or contiguous range of dimensions:
//
//	Edges()        - Return slice over all edges.
//	ShapeEdges()   - Return slice over edges of a given shape.
//	DimEdges()     - Return slice over all edges of a given dimension.
//	DimRangeEges() - Return slice over all edges of a range of dimensions.
//
// We use a stable sort, so similarly to ShapeIndexCell, we promise that
// shapes _within a dimension_ are in the same order they are in the index
// itself, and the edges _within a shape_ are similarly in the same order.
type indexCellData struct {
	index     *ShapeIndex
	indexCell *ShapeIndexCell
	cellID    CellID

	// Computing the cell center and Cell can cost as much as looking up the
	// edges themselves, so defer doing it until needed.
	mu         sync.Mutex
	s2CellSet  atomic.Bool
	s2Cell     Cell
	centerSet  atomic.Bool
	cellCenter Point

	// Dimensions that we wish to decode, the default is all of them.
	dimWanted [3]bool

	// Storage space for edges of the current cell.
	edges []edgeAndIDChain

	// Mapping from shape id to the region of the edges array it's stored in.
	shapeRegions []shapeRegion

	// Region for each dimension we might encounter.
	dimRegions [3]indexCellRegion
}

// newIndexCellData creates a new indexCellData with the expected defaults.
func newIndexCellData() *indexCellData {
	return &indexCellData{
		dimWanted: [3]bool{true, true, true},
	}
}

// newIndexCellDataFromCell creates a new indexCellData and loads the given
// cell data.
func newIndexCellDataFromCell(index *ShapeIndex, id CellID, cell *ShapeIndexCell) *indexCellData {
	d := newIndexCellData()
	d.loadCell(index, id, cell)
	return d
}

// reset resets internal state to defaults.
// The next call to loadCell() will process the cell regardless of whether
// it was already loaded. Should also be called when processing a new index.
func (d *indexCellData) reset() {
	d.index = nil
	d.indexCell = nil
	d.edges = d.edges[:0]
	d.shapeRegions = d.shapeRegions[:0]
	d.dimWanted = [3]bool{true, true, true}
}

// setDimWanted configures whether a particular dimension of shape should be decoded.
func (d *indexCellData) setDimWanted(dim int, wanted bool) {
	if dim < 0 || dim > 2 {
		return
	}
	d.dimWanted[dim] = wanted
}

// cell returns the S2 Cell for the current index cell, loading it if it
// was not already set.
func (d *indexCellData) cell() Cell {
	if !d.s2CellSet.Load() {
		d.mu.Lock()
		if !d.s2CellSet.Load() {
			d.s2Cell = CellFromCellID(d.cellID)
			d.s2CellSet.Store(true)
		}
		d.mu.Unlock()
	}
	return d.s2Cell
}

// center returns the center point of the current index cell, loading it
// if it was not already set.
func (d *indexCellData) center() Point {
	if !d.centerSet.Load() {
		d.mu.Lock()
		if !d.centerSet.Load() {
			d.cellCenter = d.cellID.Point()
			d.centerSet.Store(true)
		}
		d.mu.Unlock()
	}
	return d.cellCenter
}

// loadCell loads the data from the given cell, previous cell data is cleared.
// Both the index and cell lifetimes must span the lifetime until this
// indexCellData is destroyed or loadCell() is called again.
//
// If the index, id and cell pointer are the same as in the previous call to
// loadCell, loading is not performed since we already have the data decoded.
func (d *indexCellData) loadCell(index *ShapeIndex, id CellID, cell *ShapeIndexCell) {
	// If this is still the same ShapeIndexCell as last time, then we are good.
	if d.index == index && d.cellID == id {
		return
	}

	d.index = index

	// Cache cell information.
	d.indexCell = cell
	d.cellID = id

	// Reset atomic flags so we'll recompute cached values.
	d.s2CellSet.Store(false)
	d.centerSet.Store(false)

	// Clear previous edges
	d.edges = d.edges[:0]
	d.shapeRegions = d.shapeRegions[:0]

	// Reset per-dimension region information.
	for i := range d.dimRegions {
		d.dimRegions[i] = indexCellRegion{}
	}

	minDim := 0
	for minDim <= 2 && !d.dimWanted[minDim] {
		minDim++
	}

	maxDim := 2
	for maxDim >= 0 && !d.dimWanted[maxDim] {
		maxDim--
	}

	// No dimensions wanted, we're done.
	if minDim > 2 || maxDim < 0 {
		return
	}

	// For each dimension, load the edges from all shapes of that dimension
	for dim := minDim; dim <= maxDim; dim++ {
		dimStart := len(d.edges)

		for _, clipped := range cell.shapes {
			shapeID := clipped.shapeID
			shape := index.Shape(shapeID)

			// Only process the current dimension.
			if shape.Dimension() != dim {
				continue
			}

			// In the event we wanted dimensions 0 and 2, but not 1.
			//
			// TODO(rsned): Should this be earlier in the loop?
			// Why not skip this dim altogether if it's not wanted?
			// Same question for C++.
			if !d.dimWanted[dim] {
				continue
			}

			// Materialize clipped shape edges into the edges
			// slice. Track where we start so we can add
			// information about the region for this shape.
			shapeStart := len(d.edges)
			for _, edgeID := range clipped.edges {
				// Looking up an edge requires looking up
				// which chain it's in, which is often a binary
				// search. So let's manually lookup the chain
				// information and use that to find the edge,
				// so we only have to do that search once.
				position := shape.ChainPosition(edgeID)
				edge := shape.ChainEdge(position.ChainID, position.Offset)
				d.edges = append(d.edges, edgeAndIDChain{
					Edge:   edge,
					ID:     int32(edgeID),
					Chain:  int32(position.ChainID),
					Offset: int32(position.Offset),
				})
			}

			// Note which block of edges belongs to the shape.
			d.shapeRegions = append(d.shapeRegions, shapeRegion{
				id: shapeID,
				region: indexCellRegion{
					start: shapeStart,
					size:  len(d.edges) - shapeStart,
				},
			})
		}

		// Save region information for the current dimension.
		d.dimRegions[dim] = indexCellRegion{
			start: dimStart,
			size:  len(d.edges) - dimStart,
		}
	}
}

// shapeEdges returns a slice of the edges in the current cell for a given shape.
func (d *indexCellData) shapeEdges(shapeID int32) []edgeAndIDChain {
	for _, sr := range d.shapeRegions {
		if sr.id == shapeID {
			region := sr.region
			if region.start < len(d.edges) {
				return d.edges[region.start : region.start+region.size]
			}
			return nil
		}
	}
	return nil
}

// dimEdges returns a slice of the edges in the current cell for the given
// dimension.
func (d *indexCellData) dimEdges(dim int) []edgeAndIDChain {
	if dim < 0 || dim > 2 {
		return nil
	}

	region := d.dimRegions[dim]
	if region.start < len(d.edges) {
		return d.edges[region.start : region.start+region.size]
	}
	return nil
}

// dimRangeEdges returns a slice of the edges in the current cell for a
// range of dimensions.
func (d *indexCellData) dimRangeEdges(dim0, dim1 int) []edgeAndIDChain {
	if dim0 > dim1 || dim0 < 0 || dim0 > 2 || dim1 < 0 || dim1 > 2 {
		return nil
	}

	start := d.dimRegions[dim0].start
	size := 0

	for dim := dim0; dim <= dim1; dim++ {
		start = minInt(start, d.dimRegions[dim].start)
		size += d.dimRegions[dim].size
	}

	if start < len(d.edges) {
		return d.edges[start:size]
	}
	return nil
}

// TODO(rsned): Differences from C++
// shapeContains
