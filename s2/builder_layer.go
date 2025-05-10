// Copyright 2025 The S2 Geometry Project Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package s2

// builderLayer defines the methods Layers must implement.
type builderLayer interface {
	// GraphOptions returns the options defined by this layer.
	GraphOptions() *graphOptions

	// Build assembles a graph of snapped edges into the geometry type
	// implemented by this layer. If an error is encountered, error is
	// set appropriately.
	//
	// Note that when there are multiple layers, the Graph object passed to all
	// layers are guaranteed to be valid until the last Build() method returns.
	// This makes it easier to write algorithms that gather the output graphs
	// from several layers and process them all at once (such as
	// closedSetNormalizer).
	Build(g *graph) (bool, error)
}
