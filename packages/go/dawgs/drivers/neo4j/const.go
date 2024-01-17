// Copyright 2023 Specter Ops, Inc.
// 
// Licensed under the Apache License, Version 2.0
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
// 
// SPDX-License-Identifier: Apache-2.0

package neo4j

// TODO: Deprecate these

const (
	cypherDeleteNodeByID          = `match (n) where id(n) = $id detach delete n`
	cypherDeleteNodesByID         = `match (n) where id(n) in $id_list detach delete n`
	cypherDeleteRelationshipByID  = `match ()-[r]->() where id(r) = $id delete r`
	cypherDeleteRelationshipsByID = `unwind $p as rid match ()-[r]->() where id(r) = rid delete r`
	idParameterName               = "id"
	idListParameterName           = "id_list"
)
