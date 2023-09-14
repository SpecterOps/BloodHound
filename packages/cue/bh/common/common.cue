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

package common

import "pkg.specterops.io/schemas/bh/types:types"

// Exported requirements
Properties: [...types.#StringEnum]
NodeKinds: [...types.#Kind]
RelationshipKinds: [...types.#Kind]

// Property name enumerations
ObjectID: types.#StringEnum & {
	symbol:         "ObjectID"
	schema:         "common"
	name:           "Object ID"
	representation: "objectid"
}

Name: types.#StringEnum & {
	symbol:         "Name"
	schema:         "common"
	name:           "Name"
	representation: "name"
}

DisplayName: types.#StringEnum & {
	symbol:         "DisplayName"
	schema:         "common"
	name:           "Display Name"
	representation: "displayname"
}

Description: types.#StringEnum & {
	symbol:         "Description"
	schema:         "common"
	name:           "Description"
	representation: "description"
}

OwnerObjectID: types.#StringEnum & {
	symbol:         "OwnerObjectID"
	schema:         "common"
	name:           "Owner Object ID"
	representation: "owner_objectid"
}

Collected: types.#StringEnum & {
	symbol:         "Collected"
	schema:         "common"
	name:           "Collected"
	representation: "collected"
}

OperatingSystem: types.#StringEnum & {
	symbol:         "OperatingSystem"
	schema:         "common"
	name:           "Operating System"
	representation: "operatingsystem"
}

SystemTags: types.#StringEnum & {
	symbol:         "SystemTags"
	schema:         "common"
	name:           "Node System Tags"
	representation: "system_tags"
}

UserTags: types.#StringEnum & {
	symbol:         "UserTags"
	schema:         "common"
	name:           "Node User Tags"
	representation: "user_tags"
}

LastSeen: types.#StringEnum & {
	symbol:         "LastSeen"
	schema:         "common"
	name:           "Last Collected by BloodHound"
	representation: "lastseen"
}

WhenCreated: types.#StringEnum & {
	symbol:         "WhenCreated"
	schema:         "common"
	name:           "Created"
	representation: "whencreated"
}

Enabled: types.#StringEnum & {
	symbol:         "Enabled"
	schema:         "common"
	name:           "Enabled"
	representation: "enabled"
}

PasswordLastSet: types.#StringEnum & {
	symbol:         "PasswordLastSet"
	schema:         "common"
	name:           "Password Last Set"
	representation: "pwdlastset"
}

Title: types.#StringEnum & {
	symbol:         "Title"
	schema:         "common"
	name:           "Title"
	representation: "title"
}

Email: types.#StringEnum & {
	symbol:         "Email"
	schema:         "common"
	name:           "Email"
	representation: "email"
}

IsInherited: types.#StringEnum & {
	symbol:         "IsInherited"
	schema:         "common"
	name:           "Is Inherited"
	representation: "isinherited"
}

Properties: [
	ObjectID,
	Name,
	DisplayName,
	Description,
	OwnerObjectID,
	Collected,
	OperatingSystem,
	SystemTags,
	UserTags,
	LastSeen,
	WhenCreated,
	Enabled,
	PasswordLastSet,
	Title,
	Email,
	IsInherited,
]

// Kinds
MigrationData: types.#Kind & {
	symbol:         "MigrationData"
	schema:         "common"
	representation: "MigrationData"
}

NodeKinds: [
	MigrationData,
]

RelationshipKinds: [
]
