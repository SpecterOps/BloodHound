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
	"fmt"
	"slices"
	"strconv"
	"strings"
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
	ErrDuplicateSchemaFindingName                = errors.New("duplicate schema finding name")
	ErrDuplicatePrincipalKind                    = errors.New("duplicate principal kind")
)

// ErrIsGraphSchemaDuplicateError - determines if the provided error is one of the following errors:
// ErrDuplicateGraphSchemaExtensionName
// ErrDuplicateGraphSchemaExtensionNamespace
// ErrDuplicateSchemaNodeKindName
// ErrDuplicateGraphSchemaExtensionPropertyName
// ErrDuplicateSchemaRelationshipKindName
// ErrDuplicateSchemaEnvironment
// ErrDuplicateSchemaFindingName
// ErrDuplicatePrincipalKind
func ErrIsGraphSchemaDuplicateError(err error) bool {
	switch {
	case errors.Is(err, ErrDuplicateGraphSchemaExtensionName),
		errors.Is(err, ErrDuplicateGraphSchemaExtensionNamespace),
		errors.Is(err, ErrDuplicateSchemaNodeKindName),
		errors.Is(err, ErrDuplicateGraphSchemaExtensionPropertyName),
		errors.Is(err, ErrDuplicateSchemaRelationshipKindName),
		errors.Is(err, ErrDuplicateSchemaEnvironment),
		errors.Is(err, ErrDuplicateSchemaFindingName),
		errors.Is(err, ErrDuplicatePrincipalKind):
		return true
	default:
		return false
	}
}

// reservedGraphKindNamespaces enumerates graph kind namespaces that may not be
// registered as an opengraph schema extension namespace or appear as the
// namespace of any node/edge kind in an ingest payload. These namespaces are
// owned by internal subsystems (ex: "tag" is reserved for the asset
// tagging subsystem). Comparisons against this list are case-insensitive.
var reservedGraphKindNamespaces = []string{"tag"}

// MatchReservedGraphKindNamespace reports whether a kind either exactly
// matches a reserved namespace or uses one as its "namespace_" prefix. The
// matched namespace is returned (in its canonical lowercase form) for use in
// error messages. Comparisons are case-insensitive.
//
// This function is used to validate kind names at ingest time and at
// extension upload time.
func MatchReservedGraphKindNamespace(candidate string) (string, bool) {
	for _, reserved := range reservedGraphKindNamespaces {
		if strings.EqualFold(candidate, reserved) {
			return reserved, true
		}
		if len(candidate) > len(reserved) && candidate[len(reserved)] == '_' && strings.EqualFold(candidate[:len(reserved)], reserved) {
			return reserved, true
		}
	}
	return "", false
}

// ReservedKindError indicates that a node or edge kind uses a namespace that
// is reserved for internal use. Callers can render it as a string via Error()
// or type-assert to inspect the offending kind and namespace.
type ReservedKindError struct {
	KindName  string
	Namespace string
}

func (s *ReservedKindError) Error() string {
	return fmt.Sprintf("kind '%s' uses reserved namespace '%s'", s.KindName, s.Namespace)
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
	KindId            int32
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

type SchemaFindingType int

const (
	SchemaFindingTypeRelationship SchemaFindingType = 1
	SchemaFindingTypeList         SchemaFindingType = 2
)

func (s SchemaFindingType) String() string {
	switch s {
	case SchemaFindingTypeRelationship:
		return "relationship"
	case SchemaFindingTypeList:
		return "list"
	default:
		return "invalid enumeration case: " + strconv.Itoa(int(s))
	}
}

// SchemaFinding represents an individual finding (e.g., T0WriteOwner, T0ADCSESC1, T0DCSync)
type SchemaFinding struct {
	ID                int32
	Type              SchemaFindingType
	SchemaExtensionId int32
	EnvironmentId     int32
	KindId            int32
	Name              string
	DisplayName       string
	CreatedAt         time.Time

	// This is the kind that the finding is associated with based on the kind_id, it is enriched by db getters
	Kind graph.Kind `gorm:"-"`
	// This is the subtypes a finding is associated with, it is enriched by the db getters
	Subtypes []string `gorm:"-"`
	// This is the extension a finding is associated with, it is enriched by the db getters
	Extension GraphSchemaExtension `gorm:"-"`
}

func (s SchemaFinding) GetType() SchemaFindingType {
	return s.Type
}

func (s SchemaFinding) IsType(findingType SchemaFindingType) bool {
	return s.Type == findingType
}

func (s SchemaFinding) String() string {
	return s.Name
}

func (s SchemaFinding) FindingKind() graph.Kind {
	return s.Kind
}

func (s SchemaFinding) GetDisplayName() string {
	return s.DisplayName
}

func (s SchemaFinding) GetExtensionName() string {
	return s.Extension.Name
}

func (s SchemaFinding) GetSubtypes() []string {
	return s.Subtypes
}

func (s SchemaFinding) Is(others ...graph.Kind) bool {
	for _, other := range others {
		if other.String() == s.String() {
			return true
		}
	}
	return false
}

func (s SchemaFinding) IsSubtype(subtype string) bool {
	return slices.Contains(s.Subtypes, subtype)
}

func (SchemaFinding) TableName() string {
	return "schema_findings"
}

func (s SchemaFinding) IsSortable(column string) bool {
	switch column {
	case "name",
		"display_name",
		"type",
		"id",
		"created_at":
		return true
	default:
		return false
	}
}

func (SchemaFinding) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"name":           {Equals, NotEquals, ApproximatelyEquals},
		"display_name":   {Equals, NotEquals, ApproximatelyEquals},
		"id":             {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_at":     {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"extension_name": {Equals, NotEquals, ApproximatelyEquals},
		"extension_id":   {Equals, NotEquals},
		"is_builtin":     {Equals, NotEquals},
		"kind":           {Equals, NotEquals},
	}
}

type SchemaFindingsSubtype struct {
	SchemaFindingId int32
	Subtype         string
}

func (SchemaFindingsSubtype) TableName() string {
	return "schema_findings_subtypes"
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
	IsBuiltin     bool   `json:"is_builtin"`
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

func (s ExtensionInput) GetDisplayName() string {
	if s.DisplayName != "" {
		return s.DisplayName
	}
	return s.Name
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
