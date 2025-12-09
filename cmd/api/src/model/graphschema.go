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

package model

type GraphSchemaExtension struct {
	Serial

	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"display_name"`
	Version     string `json:"version" validate:"required"`
	IsBuiltin   bool   `json:"is_builtin"`
}

func (GraphSchemaExtension) TableName() string {
	return "schema_extensions"
}

func (s GraphSchemaExtension) AuditData() AuditData {
	return AuditData{
		"id":           s.ID,
		"name":         s.Name,
		"display_name": s.DisplayName,
		"version":      s.Version,
		"is_builtin":   s.IsBuiltin,
	}
}

// SchemaNodeKind - represents a node kind for an extension
type SchemaNodeKind struct {
	Serial

	Name              string
	SchemaExtensionId int32  // indicates which extension this node kind belongs to
	DisplayName       string // can be different from name but usually isn't other than Base/Entity
	Description       string // human-readable description of the node kind
	IsDisplayKind     bool   // indicates if this kind should supersede others and be displayed
	Icon              string // font-awesome icon for the registered node kind
	IconColor         string // icon hex color
}

// TableName - Retrieve table name
func (SchemaNodeKind) TableName() string {
	return "schema_node_kinds"
}

type GraphSchemaProperty struct {
	Serial

	SchemaExtensionID int32  `json:"schema_extension_id"`
	Name              string `json:"name" validate:"required"`
	DisplayName       string `json:"display_name"`
	DataType          string `json:"data_type" validate:"required"`
	Description       string `json:"description"`
}

func (GraphSchemaProperty) TableName() string {
	return "schema_properties"
}

// SchemaEdgeKind - represents an edge kind for an extension
type SchemaEdgeKind struct {
	Serial
	SchemaExtensionId int32 // indicates which extension this edge kind belongs to
	Name              string
	Description       string
	IsTraversable     bool // indicates whether the edge-kind is a traversable path
}

func (SchemaEdgeKind) TableName() string {
	return "schema_edge_kinds"
}

type SchemaEnvironment struct {
	Serial
	SchemaExtensionId int32 `json:"schema_extension_id"`
	EnvironmentKindId int32 `json:"environment_kind_id"`
	SourceKindId int32 `json:"source_kind_id"`
}

func (SchemaEnvironment) TableName() string {
	return "schema_environments"
}
