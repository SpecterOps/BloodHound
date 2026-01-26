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

import (
	"fmt"
	"time"
)

var (
	ErrGraphExtensionBuiltIn    = fmt.Errorf("cannot modify a built-in graph extension")
	ErrGraphExtensionValidation = fmt.Errorf("graph schema validation error")
	ErrGraphDBRefreshKinds      = fmt.Errorf("error refreshing graph db kinds")
)

type GraphExtension struct {
	GraphSchemaExtension  GraphSchemaExtension
	GraphSchemaProperties GraphSchemaProperties
	GraphSchemaEdgeKinds  GraphSchemaEdgeKinds
	GraphSchemaNodeKinds  GraphSchemaNodeKinds
	GraphEnvironments     GraphEnvironments
	GraphFindings         GraphFindings
}

type GraphFindings []GraphFinding

type GraphFinding struct {
	ID                int32
	Name              string
	SchemaExtensionId int32
	DisplayName       string
	SourceKind        string
	RelationshipKind  string // edge kind
	EnvironmentKind   string
	Remediation       Remediation
}

type GraphEnvironments []GraphEnvironment

// GraphEnvironment - represents an Environment for a given extension. Serial is pulled from the schema_environments table
type GraphEnvironment struct {
	Serial
	SchemaExtensionId int32
	EnvironmentKind   string
	SourceKind        string
	PrincipalKinds    []string
}

type GraphSchemaExtensions []GraphSchemaExtension

type GraphSchemaExtension struct {
	Serial

	Name        string
	DisplayName string
	Version     string
	IsBuiltin   bool
	Namespace   string // the required extension prefix for node and edge kind names
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

	Name              string
	SchemaExtensionId int32  // indicates which extension this node kind belongs to
	DisplayName       string // human-readable name
	Description       string // human-readable description of the node kind
	IsDisplayKind     bool   // indicates if this kind should supersede others and be displayed
	Icon              string // font-awesome icon for the registered node kind
	IconColor         string // icon hex color
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

	SchemaExtensionId int32
	Name              string
	DisplayName       string
	DataType          string
	Description       string
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

	SchemaExtensionId int32 // indicates which extension this edge kind belongs to
	Name              string
	Description       string
	IsTraversable     bool // indicates whether the edge-kind is a traversable path
}

func (GraphSchemaEdgeKind) TableName() string {
	return "schema_edge_kinds"
}

type SchemaEnvironment struct {
	Serial
	SchemaExtensionId          int32
	SchemaExtensionDisplayName string
	EnvironmentKindId          int32
	EnvironmentKindName        string
	SourceKindId               int32
}

func (SchemaEnvironment) TableName() string {
	return "schema_environments"
}

// SchemaRelationshipFinding represents an individual finding (e.g., T0WriteOwner, T0ADCSESC1, T0DCSync)
type SchemaRelationshipFinding struct {
	ID                 int32
	SchemaExtensionId  int32
	RelationshipKindId int32
	EnvironmentId      int32
	Name               string
	DisplayName        string
	CreatedAt          time.Time
}

func (SchemaRelationshipFinding) TableName() string {
	return "schema_relationship_findings"
}

type Remediation struct {
	FindingID        int32
	ShortDescription string
	LongDescription  string
	ShortRemediation string
	LongRemediation  string
}

func (Remediation) TableName() string {
	return "schema_remediations"
}

type SchemaEnvironmentPrincipalKinds []SchemaEnvironmentPrincipalKind

type SchemaEnvironmentPrincipalKind struct {
	EnvironmentId int32
	PrincipalKind int32
	CreatedAt     time.Time
}

func (SchemaEnvironmentPrincipalKind) TableName() string {
	return "schema_environments_principal_kinds"
}

func (GraphSchemaEdgeKind) ValidFilters() map[string][]FilterOperator {
	return ValidFilters{
		"is_traversable": {Equals, NotEquals},
		"schema_names":   {Equals, NotEquals, ApproximatelyEquals},
	}
}

func (GraphSchemaEdgeKind) IsStringColumn(filter string) bool {
	return filter == "schema_names"
}

type GraphSchemaEdgeKindWithNamedSchema struct {
	ID            int32  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	IsTraversable bool   `json:"is_traversable"`
	SchemaName    string `json:"schema_name"`
}

type GraphSchemaEdgeKindsWithNamedSchema []GraphSchemaEdgeKindWithNamedSchema
