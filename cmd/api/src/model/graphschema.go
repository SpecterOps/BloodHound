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

// GraphSchema -
type GraphSchema struct {
	GraphSchemaExtension  GraphSchemaExtension  `json:"extension"`
	GraphSchemaProperties GraphSchemaProperties `json:"properties"`
	GraphSchemaEdgeKinds  GraphSchemaEdgeKinds  `json:"edge_kinds"`
	GraphSchemaNodeKinds  GraphSchemaNodeKinds  `json:"node_kinds"`
}

type GraphSchemaExtensions []GraphSchemaExtension

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

// GraphSchemaNodeKinds - slice of node kinds
type GraphSchemaNodeKinds []GraphSchemaNodeKind

// ToMapKeyedOnName - converts a list of graph schema node kinds to a map based on name
func (g GraphSchemaNodeKinds) ToMapKeyedOnName() map[string]GraphSchemaNodeKind {
	result := make(map[string]GraphSchemaNodeKind, 0)
	for _, kind := range g {
		result[kind.Name] = kind
	}
	return result
}

// GraphSchemaNodeKind - represents a node kind for an extension
type GraphSchemaNodeKind struct {
	Serial

	Name              string `json:"name"`
	SchemaExtensionId int32  `json:"schema_extension_id"` // indicates which extension this node kind belongs to
	DisplayName       string `json:"display_name"`        // can be different from name but usually isn't other than Base/Entity
	Description       string `json:"description"`         // human-readable description of the node kind
	IsDisplayKind     bool   `json:"is_display_kind"`     // indicates if this kind should supersede others and be displayed
	Icon              string `json:"icon"`                // font-awesome icon for the registered node kind
	IconColor         string `json:"icon_color"`          // icon hex color
}

// TableName - Retrieve table name
func (GraphSchemaNodeKind) TableName() string {
	return "schema_node_kinds"
}

// GraphSchemaProperties - slice of graph schema properties.
type GraphSchemaProperties []GraphSchemaProperty

// ToMapKeyedOnName - converts a list of graph schema properties to a map keyed on name
func (g GraphSchemaProperties) ToMapKeyedOnName() map[string]GraphSchemaProperty {
	result := make(map[string]GraphSchemaProperty, 0)
	for _, kind := range g {
		result[kind.Name] = kind
	}
	return result
}

// GraphSchemaProperty - represents a property that an edge or node kind can have. Grouped by schema extension.
type GraphSchemaProperty struct {
	Serial

	SchemaExtensionId int32  `json:"schema_extension_id"`
	Name              string `json:"name" validate:"required"`
	DisplayName       string `json:"display_name"`
	DataType          string `json:"data_type" validate:"required"`
	Description       string `json:"description"`
}

func (GraphSchemaProperty) TableName() string {
	return "schema_properties"
}

// GraphSchemaEdgeKinds - slice of GraphSchemaEdgeKind
type GraphSchemaEdgeKinds []GraphSchemaEdgeKind

// ToMapKeyedOnName - converts a list of graph schema edge kinds to a map keyed on name
func (g GraphSchemaEdgeKinds) ToMapKeyedOnName() map[string]GraphSchemaEdgeKind {
	result := make(map[string]GraphSchemaEdgeKind, 0)
	for _, kind := range g {
		result[kind.Name] = kind
	}
	return result
}

// GraphSchemaEdgeKind - represents an edge kind for an extension
type GraphSchemaEdgeKind struct {
	Serial
	SchemaExtensionId int32  `json:"schema_extension_id"` // indicates which extension this edge kind belongs to
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsTraversable     bool   `json:"isTraversable"` // indicates whether the edge-kind is a traversable path
}

func (GraphSchemaEdgeKind) TableName() string {
	return "schema_edge_kinds"
}
