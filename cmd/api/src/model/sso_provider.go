// Copyright 2024 Specter Ops, Inc.
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
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type SSOProviderAutoProvisionConfig struct {
	Enabled       bool  `json:"enabled"`
	DefaultRole   int32 `json:"default_role"`
	RoleProvision bool  `json:"role_provision"`
}

type SSOProviderConfig struct {
	AutoProvision SSOProviderAutoProvisionConfig `json:"auto_provision"`
}

// SSOProvider is the common representation of an SSO provider that can be used to display high level information about that provider
type SSOProvider struct {
	Type SessionAuthProvider `json:"type" gorm:"column:type"`
	Name string              `json:"name"`
	Slug string              `json:"slug"`

	OIDCProvider *OIDCProvider `json:"oidc_provider,omitempty" gorm:"foreignKey:SSOProviderID"`
	SAMLProvider *SAMLProvider `json:"saml_provider,omitempty" gorm:"foreignKey:SSOProviderID"`

	Config SSOProviderConfig `json:"config" gorm:"type:jsonb column:config"`

	Serial
}

// Implement the sql.Scanner interface so that GORM can scan the jsonb column from the database into a golang struct
func (cfg *SSOProviderConfig) Scan(value interface{}) error {
	// Handle null values from the database
	if value == nil {
		*cfg = SSOProviderConfig{}
		return nil
	}

	// Convert the database value to []byte
	if bytes, ok := value.([]byte); !ok {
		return errors.New("type assertion to []byte failed for SSOProviderConfig")
	} else {
		// Unmarshal JSON into the struct
		return json.Unmarshal(bytes, cfg)
	}
}

// Value returns the json-marshaled value of the receiver
func (cfg SSOProviderConfig) Value() (driver.Value, error) {
	return json.Marshal(cfg)
}

// AuditData returns the fields to log in the audit log
func (s SSOProvider) AuditData() AuditData {
	var (
		details any
	)

	switch s.Type {
	case SessionAuthProviderSAML:
		details = s.SAMLProvider
		break
	case SessionAuthProviderOIDC:
		details = s.OIDCProvider
		break
	}

	return AuditData{
		"id":      s.ID,
		"name":    s.Name,
		"slug":    s.Slug,
		"type":    s.Type,
		"details": details,
	}
}

// Define sortable fields
func SSOProviderSortableFields(field string) bool {
	switch field {
	case "id", "name", "slug", "type", "created_at", "updated_at":
		return true
	default:
		return false
	}
}

// Define valid filter predicates for each field
func SSOProviderValidFilterPredicates(field string) ([]string, error) {
	switch field {
	case "id", "type":
		return []string{string(Equals), string(NotEquals), string(GreaterThan), string(GreaterThanOrEquals), string(LessThan), string(LessThanOrEquals)}, nil
	case "name", "slug":
		return []string{string(Equals), string(NotEquals), string(ApproximatelyEquals)}, nil
	case "created_at", "updated_at":
		return []string{string(Equals), string(NotEquals), string(GreaterThan), string(GreaterThanOrEquals), string(LessThan), string(LessThanOrEquals)}, nil
	default:
		return nil, fmt.Errorf("the specified column cannot be filtered: %s", field)
	}
}

// Define which fields are string type
func SSOProviderIsStringField(field string) bool {
	switch field {
	case "name", "slug":
		return true
	default:
		return false
	}
}
