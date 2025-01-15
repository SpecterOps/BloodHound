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

package schema

import (
	"pkg.specterops.io/schemas/bh/ad:ad"
	"pkg.specterops.io/schemas/bh/types:types"
	"pkg.specterops.io/schemas/bh/azure:azure"
	"pkg.specterops.io/schemas/bh/common:common"
)

// Schema
#Common: {
	Properties: [...types.#StringEnum]
	NodeKinds: [...types.#Kind]
	RelationshipKinds: [...types.#Kind]
}

#Azure: {
	Properties: [...types.#StringEnum]
	NodeKinds: [...types.#Kind]
	RelationshipKinds: [...types.#Kind]
	AppRoleTransitRelationshipKinds: [...types.#Kind]
	AbusableAppRoleRelationshipKinds: [... types.#Kind]
	ControlRelationshipKinds: [...types.#Kind]
	ExecutionPrivilegeKinds: [...types.#Kind]
	PathfindingRelationships: [...types.#Kind]
}

#ActiveDirectory: {
	Properties: [...types.#StringEnum]
	NodeKinds: [...types.#Kind]
	RelationshipKinds: [...types.#Kind]
	ACLRelationships: [...types.#Kind]
	PathfindingRelationships: [...types.#Kind]
	EdgeCompositionRelationships: [...types.#Kind]
}

// Definitons
Common: #Common & {
	Properties:        common.Properties
	NodeKinds:         common.NodeKinds
	RelationshipKinds: common.RelationshipKinds
}

Azure: #Azure & {
	Properties:                       azure.Properties
	NodeKinds:                        azure.NodeKinds
	RelationshipKinds:                azure.RelationshipKinds
	AppRoleTransitRelationshipKinds:  azure.AppRoleTransitRelationshipKinds
	AbusableAppRoleRelationshipKinds: azure.AbusableAppRoleRelationshipKinds
	ControlRelationshipKinds:         azure.ControlRelationshipKinds
	ExecutionPrivilegeKinds:          azure.ExecutionPrivilegeKinds
	PathfindingRelationships:         azure.PathfindingRelationships
}

ActiveDirectory: #ActiveDirectory & {
	Properties:               		ad.Properties
	NodeKinds:                		ad.NodeKinds
	RelationshipKinds:        		ad.RelationshipKinds
	ACLRelationships:         		ad.ACLRelationships
	PathfindingRelationships: 		ad.PathfindingRelationships
	EdgeCompositionRelationships: 	ad.EdgeCompositionRelationships

}
