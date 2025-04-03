// Copyright 2025 Google Inc. All rights reserved.
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

// To use these, you must set the build tag with the -tags flag as:
//    go build -tags debug

//go:build debug

package s2

// dcheck is a custom debug check function.
func dcheck(condition bool, message string) {
	if !condition {
		panic("dcheck failed: " + message)
	}
}

// TODO(rsned): Also create equivalents for
// ABSL_DCHECK_EQ
// ABSL_DCHECK_GE
// ABSL_DCHECK_GT
// ABSL_DCHECK_LE
// ABSL_DCHECK_LT
// ABSL_DCHECK_NE
