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

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/version"
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

// reservedGraphKindNamespaces lists namespaces that cannot be used in custom
// graph extensions or ingest payloads. These are owned by internal subsystems
// (e.g., "tag" is reserved for asset tagging). Comparisons are case-insensitive.
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

type EnvironmentKindsToEnvironment map[string]SchemaEnvironment

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
	// PZ Variant Display Title
	PZDisplayName null.String
	CreatedAt     time.Time

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

func (s SchemaFinding) GetPZDisplayName() string {
	return s.PZDisplayName.String
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
		"pz_display_name",
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
		"name":            {Equals, NotEquals, ApproximatelyEquals},
		"display_name":    {Equals, NotEquals, ApproximatelyEquals},
		"pz_display_name": {Equals, NotEquals, ApproximatelyEquals},
		"id":              {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_at":      {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"extension_name":  {Equals, NotEquals, ApproximatelyEquals},
		"extension_id":    {Equals, NotEquals},
		"is_builtin":      {Equals, NotEquals},
		"kind":            {Equals, NotEquals},
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

// Validate performs comprehensive validation on a GraphExtensionInput
func (s GraphExtensionInput) Validate() error {
	var (
		nodeKinds         = make(map[string]any, 0)
		relationshipKinds = make(map[string]any, 0)
		properties        = make(map[string]any, 0)
		environments      = make(map[string]any, 0)
		findings          = make(map[string]any, 0)
	)
	if strings.TrimSpace(s.ExtensionInput.Name) == "" {
		return errors.New("graph schema extension name is required")
	} else if strings.TrimSpace(s.ExtensionInput.Version) == "" {
		return errors.New("graph schema extension version is required")
	} else if _, err := version.Parse(s.ExtensionInput.Version); err != nil {
		return fmt.Errorf("graph schema extension version is not valid semver: %w", err)
	} else if strings.TrimSpace(s.ExtensionInput.Namespace) == "" {
		return errors.New("graph schema extension namespace is required")
	} else if reservedNamespace, isReserved := MatchReservedGraphKindNamespace(s.ExtensionInput.Namespace); isReserved {
		return fmt.Errorf("graph schema extension namespace '%s' uses reserved namespace '%s'", s.ExtensionInput.Namespace, reservedNamespace)
	} else if len(s.NodeKindsInput) == 0 {
		return errors.New("graph schema node kinds are required")
	}

	for _, kind := range s.NodeKindsInput {
		if kindName, found := strings.CutPrefix(kind.Name, fmt.Sprintf("%s_", s.ExtensionInput.Namespace)); !found {
			return fmt.Errorf("graph schema node kind %s is missing extension namespace prefix", kind.Name)
		} else if strings.TrimSpace(kindName) == "" {
			return errors.New("graph schema node kind cannot be empty after the namespace prefix")
		}
		if _, ok := nodeKinds[kind.Name]; ok {
			return fmt.Errorf("duplicate graph kinds: %s", kind.Name)
		}
		if kind.IconColor != "" && !IsValidIconColor(kind.IconColor) {
			return fmt.Errorf("invalid hex color string %s for node kind %s", kind.IconColor, kind.Name)
		}
		nodeKinds[kind.Name] = struct{}{}
	}

	for _, kind := range s.RelationshipKindsInput {
		if kindName, found := strings.CutPrefix(kind.Name, fmt.Sprintf("%s_", s.ExtensionInput.Namespace)); !found {
			return fmt.Errorf("graph schema edge kind %s is missing extension namespace prefix", kind.Name)
		} else if strings.TrimSpace(kindName) == "" {
			return errors.New("graph schema edge kind cannot be empty after the namespace prefix")
		}
		if _, ok := relationshipKinds[kind.Name]; ok {
			return fmt.Errorf("duplicate graph kinds: %s", kind.Name)
		}
		if _, ok := nodeKinds[kind.Name]; ok {
			return fmt.Errorf("duplicate graph kinds: %s", kind.Name)
		}
		relationshipKinds[kind.Name] = struct{}{}
	}

	for _, property := range s.PropertiesInput {
		if _, ok := properties[property.Name]; ok {
			return fmt.Errorf("duplicate graph properties: %s", property.Name)
		}
		properties[property.Name] = struct{}{}
	}

	for _, environment := range s.EnvironmentsInput {
		if environmentKindName, found := strings.CutPrefix(environment.EnvironmentKindName, fmt.Sprintf("%s_", s.ExtensionInput.Namespace)); !found {
			return fmt.Errorf("graph schema environment kind %s is missing extension namespace prefix", environment.EnvironmentKindName)
		} else if strings.TrimSpace(environmentKindName) == "" {
			return errors.New("graph schema environment kind cannot be empty after the namespace prefix")
		}
		if _, ok := nodeKinds[environment.EnvironmentKindName]; !ok {
			return fmt.Errorf("graph schema environment %s not declared as a node kind", environment.EnvironmentKindName)
		}
		if _, ok := environments[environment.EnvironmentKindName]; ok {
			return fmt.Errorf("duplicate graph environments: %s", environment.EnvironmentKindName)
		}
		if strings.TrimSpace(environment.SourceKindName) == "" {
			return fmt.Errorf("graph schema environment source kind cannot be empty")
		}
		if _, ok := nodeKinds[environment.SourceKindName]; ok {
			return fmt.Errorf("graph schema environment source kind name %s conflicts with existing node kind", environment.SourceKindName)
		}
		if _, ok := relationshipKinds[environment.SourceKindName]; ok {
			return fmt.Errorf("graph schema environment source kind name %s conflicts with existing relationship kind", environment.SourceKindName)
		}
		for _, principalKind := range environment.PrincipalKinds {
			if principalKindName, found := strings.CutPrefix(principalKind, fmt.Sprintf("%s_", s.ExtensionInput.Namespace)); !found {
				return fmt.Errorf("graph schema environment principal kind %s is missing extension namespace prefix", principalKind)
			} else if strings.TrimSpace(principalKindName) == "" {
				return errors.New("graph schema environment principal kind cannot be empty after the namespace prefix")
			}
			if _, ok := nodeKinds[principalKind]; !ok {
				return fmt.Errorf("graph schema environment principal kind %s not declared node kind", principalKind)
			}
		}
		environments[environment.EnvironmentKindName] = struct{}{}
	}

	for _, relationshipFindingInput := range s.RelationshipFindingsInput {
		if findingName, found := strings.CutPrefix(relationshipFindingInput.Name, fmt.Sprintf("%s_", s.ExtensionInput.Namespace)); !found {
			return fmt.Errorf("graph schema relationship finding %s is missing extension namespace prefix", relationshipFindingInput.Name)
		} else if strings.TrimSpace(findingName) == "" {
			return errors.New("graph schema relationship finding cannot be empty after the namespace prefix")
		}
		if _, ok := findings[relationshipFindingInput.Name]; ok {
			return fmt.Errorf("duplicate graph schema relationship finding: %s", relationshipFindingInput.Name)
		}
		if !strings.HasPrefix(relationshipFindingInput.RelationshipKindName, fmt.Sprintf("%s_", s.ExtensionInput.Namespace)) {
			return fmt.Errorf("graph schema relationship finding relationship kind %s is missing extension namespace prefix", relationshipFindingInput.RelationshipKindName)
		}
		if _, ok := relationshipKinds[relationshipFindingInput.RelationshipKindName]; !ok {
			return fmt.Errorf("graph schema relationship finding relationship kind %s not declared as a relationship kind", relationshipFindingInput.RelationshipKindName)
		}
		findings[relationshipFindingInput.Name] = struct{}{}
	}
	return nil
}

type RelationshipFindingsInput []RelationshipFindingInput
type RelationshipFindingInput struct {
	Name                 string
	DisplayName          string
	PZDisplayName        string
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
