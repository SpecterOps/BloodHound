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
	"errors"
	"time"

	"github.com/specterops/dawgs/graph"
)

var (
	ErrGraphExtensionBuiltIn    = errors.New("cannot modify a built-in graph extension")
	ErrGraphExtensionValidation = errors.New("graph schema validation error")
	ErrGraphDBRefreshKinds      = errors.New("error refreshing graph db kinds")

	ErrDuplicateGraphSchemaExtensionName         = errors.New("duplicate graph schema extension name")
	ErrDuplicateGraphSchemaExtensionNamespace    = errors.New("duplicate graph schema extension namespace")
	ErrDuplicateSchemaNodeKindName               = errors.New("duplicate schema node kind name")
	ErrDuplicateGraphSchemaExtensionPropertyName = errors.New("duplicate graph schema extension property name")
	ErrDuplicateSchemaRelationshipKindName       = errors.New("duplicate schema relationship kind name")
	ErrDuplicateSchemaEnvironment                = errors.New("duplicate schema environment")
	ErrDuplicateSchemaRelationshipFindingName    = errors.New("duplicate schema relationship finding name")
	ErrDuplicatePrincipalKind                    = errors.New("duplicate principal kind")
)

// ErrIsGraphSchemaDuplicateError - determines if the provided error is one of the following errors:
// ErrDuplicateGraphSchemaExtensionName
// ErrDuplicateGraphSchemaExtensionNamespace
// ErrDuplicateSchemaNodeKindName
// ErrDuplicateGraphSchemaExtensionPropertyName
// ErrDuplicateSchemaRelationshipKindName
// ErrDuplicateSchemaEnvironment
// ErrDuplicateSchemaRelationshipFindingName
// ErrDuplicatePrincipalKind
func ErrIsGraphSchemaDuplicateError(err error) bool {
	if err == nil {
		return false
	}

	var duplicateErrors = []error{
		ErrDuplicateGraphSchemaExtensionName, ErrDuplicateGraphSchemaExtensionNamespace, ErrDuplicateSchemaNodeKindName,
		ErrDuplicateGraphSchemaExtensionPropertyName, ErrDuplicateSchemaRelationshipKindName, ErrDuplicateSchemaEnvironment,
		ErrDuplicateSchemaRelationshipFindingName, ErrDuplicatePrincipalKind}
	for _, e := range duplicateErrors {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}

type GraphSchemaExtensions []GraphSchemaExtension

type GraphSchemaExtension struct {
	Serial

	Name        string
	DisplayName string
	Version     string
	IsBuiltin   bool
	Namespace   string
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

	SchemaExtensionId int32
	Name              string
	DisplayName       string
	DataType          string
	Description       string
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
	DisplayName      string
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

// Graph Extension Upsert Input

type GraphExtensionInput struct {
	ExtensionInput            ExtensionInput
	PropertiesInput           PropertiesInput
	RelationshipKindsInput    RelationshipsInput
	NodeKindsInput            NodesInput
	EnvironmentsInput         EnvironmentsInput
	RelationshipFindingsInput RelationshipFindingsInput
}

type RelationshipFindingsInput []RelationshipFindingInput
type RelationshipFindingInput struct {
	Name                 string
	DisplayName          string
	SourceKindName       string
	RelationshipKindName string // edge kind
	EnvironmentKindName  string
	RemediationInput     RemediationInput
}

type EnvironmentsInput []EnvironmentInput
type EnvironmentInput struct {
	EnvironmentKindName string
	SourceKindName      string
	PrincipalKinds      []string
}

type ExtensionInput struct {
	Name        string
	DisplayName string
	Version     string
	Namespace   string // the required extension prefix for node and edge kind names
}

type PropertiesInput []PropertyInput
type PropertyInput struct {
	Name        string
	DisplayName string
	DataType    string
	Description string
}

type NodesInput []NodeInput
type NodeInput struct {
	Name          string
	DisplayName   string // human-readable name
	Description   string // human-readable description of the node kind
	IsDisplayKind bool   // indicates if this kind should supersede others and be displayed
	Icon          string // font-awesome icon for the registered node kind
	IconColor     string // icon hex color
}

type RelationshipsInput []RelationshipInput
type RelationshipInput struct {
	Name          string
	Description   string
	IsTraversable bool // indicates whether the edge-kind is a traversable path
}
type RemediationInput struct {
	ShortDescription string
	LongDescription  string
	ShortRemediation string
	LongRemediation  string
}
