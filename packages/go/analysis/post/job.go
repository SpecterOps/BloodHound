// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
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
//
// SPDX-License-Identifier: Apache-2.0
package post

import "github.com/specterops/dawgs/graph"

// EnsureRelationshipJob is an asynchronous graph assertion. If the edge does not
// exist in the graph between the from and to node IDs with the given kind then
// the edge added to a batch creation process to be pushed down to the database.
type EnsureRelationshipJob struct {
	FromID        graph.ID
	ToID          graph.ID
	Kind          graph.Kind
	RelProperties map[string]any
}
