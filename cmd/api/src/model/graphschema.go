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
	"time"

	"github.com/specterops/dawgs/graph"
)

type GraphSchemaExtensions []GraphSchemaExtension

type GraphSchemaExtension struct {
	Serial

	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"display_name"`
	Version     string `json:"version" validate:"required"`
	IsBuiltin   bool   `json:"is_builtin"`
	Namespace   string `json:"namespace"`
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
		"namespace":    s.Namespace,
	}
}

// GraphSchemaNodeKinds - slice of node kinds
type GraphSchemaNodeKinds []GraphSchemaNodeKind

// GraphSchemaNodeKind - represents a node kind for an extension
type GraphSchemaNodeKind struct {
	Serial

	Name              string
	SchemaExtensionId int32  // indicates which extension this node kind belongs to
	DisplayName       string // can be different from name but usually isn't other than Base/Entity
	Description       string // human-readable description of the node kind
	IsDisplayKind     bool   // indicates if this kind should supersede others and be displayed
	Icon              string // font-awesome icon for the registered node kind
	IconColor         string // icon hex color
}

func (s GraphSchemaNodeKind) ToKind() graph.Kind {
	return graph.StringKind(s.Name)
}

// TableName - Retrieve table name
func (GraphSchemaNodeKind) TableName() string {
	return "schema_node_kinds"
}

// GraphSchemaProperties - slice of graph schema properties.
type GraphSchemaProperties []GraphSchemaProperty

// GraphSchemaProperty - represents a property that an relationship or node kind can have. Grouped by schema extension.
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

// GraphSchemaRelationshipKinds - slice of model.GraphSchemaRelationshipKind
type GraphSchemaRelationshipKinds []GraphSchemaRelationshipKind

// GraphSchemaRelationshipKind - represents an relationship kind for an extension
type GraphSchemaRelationshipKind struct {
	Serial
	SchemaExtensionId int32 // indicates which extension this relationship kind belongs to
	Name              string
	Description       string
	IsTraversable     bool // indicates whether the relationship-kind is a traversable path
}

func (s GraphSchemaRelationshipKind) ToKind() graph.Kind {
	return graph.StringKind(s.Name)
}

func (GraphSchemaRelationshipKind) TableName() string {
	return "schema_relationship_kinds"
}

type SchemaEnvironment struct {
	Serial
	SchemaExtensionId          int32  `json:"schema_extension_id"`
	SchemaExtensionDisplayName string `json:"schema_extension_display_name,omitempty"`
	EnvironmentKindId          int32  `json:"environment_kind_id"`
	EnvironmentKindName        string `json:"environment_kind_name,omitempty"`
	SourceKindId               int32  `json:"source_kind_id"`
}

func (SchemaEnvironment) TableName() string {
	return "schema_environments"
}

// SchemaRelationshipFinding represents an individual finding (e.g., T0WriteOwner, T0ADCSESC1, T0DCSync)
type SchemaRelationshipFinding struct {
	ID                 int32     `json:"id"`
	SchemaExtensionId  int32     `json:"schema_extension_id"`
	RelationshipKindId int32     `json:"relationship_kind_id"`
	EnvironmentId      int32     `json:"environment_id"`
	Name               string    `json:"name"`
	DisplayName        string    `json:"display_name"`
	CreatedAt          time.Time `json:"created_at"`
}

func (SchemaRelationshipFinding) TableName() string {
	return "schema_relationship_findings"
}

type Remediation struct {
	FindingID        int32  `json:"finding_id"`
	DisplayName      string `json:"display_name"`
	ShortDescription string `json:"short_description"`
	LongDescription  string `json:"long_description"`
	ShortRemediation string `json:"short_remediation"`
	LongRemediation  string `json:"long_remediation"`
}

func (Remediation) TableName() string {
	return "schema_remediations"
}

type SchemaEnvironmentPrincipalKinds []SchemaEnvironmentPrincipalKind

type SchemaEnvironmentPrincipalKind struct {
	EnvironmentId int32     `json:"environment_id"`
	PrincipalKind int32     `json:"principal_kind"`
	CreatedAt     time.Time `json:"created_at"`
}

func (SchemaEnvironmentPrincipalKind) TableName() string {
	return "schema_environments_principal_kinds"
}

func (GraphSchemaRelationshipKind) ValidFilters() map[string][]FilterOperator {
	return ValidFilters{
		"is_traversable": {Equals, NotEquals},
		"schema_names":   {Equals, NotEquals, ApproximatelyEquals},
	}
}

func (GraphSchemaRelationshipKind) IsStringColumn(filter string) bool {
	return filter == "schema_names"
}

type GraphSchemaRelationshipKindWithNamedSchema struct {
	ID            int32  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	IsTraversable bool   `json:"is_traversable"`
	SchemaName    string `json:"schema_name"`
}

type GraphSchemaRelationshipKindsWithNamedSchema []GraphSchemaRelationshipKindWithNamedSchema
