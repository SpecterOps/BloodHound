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
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/serde"
)

const PermissionURIScheme = "permission"

type Installation struct {
	Unique
}

type Permission struct {
	Authority string `json:"authority"`
	Name      string `json:"name"`

	Serial
}

func NewPermission(authority, name string) Permission {
	return Permission{
		Authority: authority,
		Name:      name,
	}
}

func (s Permission) Equals(other Permission) bool {
	return s.Authority == other.Authority && s.Name == other.Name
}

func (s Permission) URI() *url.URL {
	return &url.URL{
		Scheme: PermissionURIScheme,
		Host:   s.Authority,
		Path:   s.Name,
	}
}

func (s Permission) String() string {
	return s.URI().String()
}

type Permissions []Permission

func (s Permissions) IsSortable(column string) bool {
	switch column {
	case "authority",
		"name",
		"id",
		"created_at",
		"updated_at",
		"deleted_at":
		return true
	default:
		return false
	}
}

func (s Permissions) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"authority":  {Equals, NotEquals},
		"name":       {Equals, NotEquals},
		"id":         {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"updated_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"deleted_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
	}
}

func (s Permissions) IsString(column string) bool {
	return column == "authority" || column == "name"
}

func (s Permissions) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s Permissions) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

func (s Permissions) Equals(others Permissions) bool {
	if len(s) != len(others) {
		return false
	}

	for _, permission := range s {
		found := false
		for _, otherPermission := range others {
			if permission.Equals(otherPermission) {
				found = true
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func (s Permissions) Has(other Permission) bool {
	for _, permission := range s {
		if permission.Equals(other) {
			return true
		}
	}

	return false
}

type AuthToken struct {
	UserID     uuid.NullUUID `json:"user_id" gorm:"type:text"`
	ClientID   uuid.NullUUID `json:"-"  gorm:"type:text"`
	Name       null.String   `json:"name"`
	Key        string        `json:"key,omitempty"`
	HmacMethod string        `json:"hmac_method"`
	LastAccess time.Time     `json:"last_access"`

	Unique
}

func (s AuthToken) AuditData() AuditData {
	return AuditData{
		"id":          s.ID,
		"user_id":     s.UserID,
		"client_id":   s.ClientID,
		"name":        s.Name,
		"last_access": s.LastAccess,
	}
}

func (s AuthToken) StripKey() AuthToken {
	return AuthToken{
		UserID:     s.UserID,
		ClientID:   s.ClientID,
		Key:        "",
		HmacMethod: s.HmacMethod,
		LastAccess: s.LastAccess,
		Unique:     s.Unique,
		Name:       s.Name,
	}
}

type AuthTokens []AuthToken

func (s AuthTokens) IsSortable(column string) bool {
	switch column {
	case "user_id",
		"client_id",
		"name",
		"last_access",
		"id",
		"created_at",
		"updated_at",
		"deleted_at":
		return true
	default:
		return false
	}
}

func (s AuthTokens) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"user_id":     {Equals, NotEquals},
		"name":        {Equals, NotEquals},
		"key":         {Equals, NotEquals},
		"hmac_method": {Equals, NotEquals},
		"id":          {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"last_access": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_at":  {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"updated_at":  {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"deleted_at":  {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
	}
}

func (s AuthTokens) IsString(column string) bool {
	return column == "name" || column == "key" || column == "hmac_method"
}

func (s AuthTokens) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s AuthTokens) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

func (s AuthTokens) IDs() []uuid.UUID {
	ids := make([]uuid.UUID, len(s))

	for idx, token := range s {
		ids[idx] = token.ID
	}

	return ids
}

func (s AuthTokens) StripKeys() AuthTokens {
	tokens := make(AuthTokens, len(s))

	for idx, token := range s {
		tokens[idx] = token.StripKey()
	}

	return tokens
}

type AuthSecret struct {
	UserID        uuid.UUID `json:"-"`
	Digest        string    `json:"-"`
	DigestMethod  string    `json:"digest_method"`
	ExpiresAt     time.Time `json:"expires_at"`
	TOTPSecret    string    `json:"-"`
	TOTPActivated bool      `json:"totp_activated"`

	Serial
}

// Expired returns true if the auth secret has expired, false otherwise
func (s AuthSecret) Expired() bool {
	return s.ExpiresAt.Before(time.Now().UTC())
}

func (s AuthSecret) AuditData() AuditData {
	return AuditData{
		"secret_user_id":    s.ID,
		"secret_expires_at": s.ExpiresAt.UTC(),
	}
}

type SAMLProvider struct {
	Name            string `json:"name" gorm:"unique;index"`
	DisplayName     string `json:"display_name"`
	IssuerURI       string `json:"idp_issuer_uri"`
	SingleSignOnURI string `json:"idp_sso_uri"`
	MetadataXML     []byte `json:"-"`

	// PrincipalAttributeMapping is an array of OID or XML Namespace element mapping strings that can be used to map a
	// SAML assertion to a user in the database.
	//
	// For example: ["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress", "urn:oid:0.9.2342.19200300.100.1.3"]
	PrincipalAttributeMappings []string `json:"principal_attribute_mappings" gorm:"type:text[];column:ous"`

	// The below values generated values that point a client to SAML related resources hosted on the BloodHound instance
	// and should not be persisted to the database due to the fact that the URLs rely on the Host header that the user is
	// using to communicate to the API
	ServiceProviderIssuerURI     serde.URL `json:"sp_issuer_uri" gorm:"-"`
	ServiceProviderInitiationURI serde.URL `json:"sp_sso_uri" gorm:"-"`
	ServiceProviderMetadataURI   serde.URL `json:"sp_metadata_uri" gorm:"-"`
	ServiceProviderACSURI        serde.URL `json:"sp_acs_uri" gorm:"-"`

	Serial
}

func (s SAMLProvider) AuditData() AuditData {
	return AuditData{
		"saml_id":                      s.ID,
		"saml_name":                    s.Name,
		"principal_attribute_mappings": s.PrincipalAttributeMappings,
		"idp_url":                      s.IssuerURI,
	}
}

type SAMLProviders []SAMLProvider

func RoleAssociations() []string {
	return []string{
		"Permissions",
	}
}

type Role struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Permissions Permissions `json:"permissions" gorm:"many2many:roles_permissions"`

	Serial
}

func (s Role) AuditData() AuditData {
	return AuditData{
		"role_id":   s.ID,
		"role_name": s.Name,
	}
}

type Roles []Role

func (s Roles) IsSortable(column string) bool {
	switch column {
	case "name",
		"description",
		"id",
		"created_at",
		"updated_at",
		"deleted_at":
		return true
	default:
		return false
	}
}

func (s Roles) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"name":       {Equals, NotEquals},
		"id":         {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"updated_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"deleted_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
	}
}

func (s Roles) IsString(column string) bool {
	return column == "name"
}

func (s Roles) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s Roles) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

func (s Roles) IDs() []int32 {
	ids := make([]int32, len(s))

	for idx, role := range s {
		ids[idx] = role.ID
	}

	return ids
}

func (s Roles) Permissions() Permissions {
	var permissions Permissions

	for _, role := range s {
		permissions = append(permissions, role.Permissions...)
	}

	return permissions
}

func (s Roles) Has(other Role) bool {
	for _, role := range s {
		if role.Name == other.Name {
			return true
		}
	}

	return false
}

func (s Roles) RemoveByName(name string) Roles {
	for idx, role := range s {
		if role.Name == name {
			return append(s[:idx], s[idx+1:]...)
		}
	}

	return s
}

func (s Roles) FindByName(name string) (Role, bool) {
	for _, role := range s {
		if role.Name == name {
			return role, true
		}
	}

	return Role{}, false
}

func (s Roles) FindByPermissions(permissions Permissions) (Role, bool) {
	for _, role := range s {
		if role.Permissions.Equals(permissions) {
			return role, true
		}
	}

	return Role{}, false
}

func UserAssociations() []string {
	return []string{
		"SAMLProvider",
		"AuthSecret",
		"AuthTokens",
		"Roles.Permissions",
	}
}

type User struct {
	SAMLProviderID null.Int32    `json:"saml_provider_id,omitempty"`
	SAMLProvider   *SAMLProvider `json:"-" `
	AuthSecret     *AuthSecret   `gorm:"constraint:OnDelete:CASCADE;"`
	AuthTokens     AuthTokens    `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	Roles          Roles         `json:"roles" gorm:"many2many:users_roles"`
	FirstName      null.String   `json:"first_name"`
	LastName       null.String   `json:"last_name"`
	EmailAddress   null.String   `json:"email_address"`
	PrincipalName  string        `json:"principal_name" gorm:"unique;index"`
	LastLogin      time.Time     `json:"last_login"`
	IsDisabled     bool          `json:"is_disabled"`
	// EULA Acceptance does not pertain to Bloodhound Community Edition; this flag is used for Bloodhound Enterprise users.
	// This value is automatically set to true for Bloodhound Community Edition in the patchEULAAcceptance and CreateUser functions.
	EULAAccepted bool `json:"eula_accepted"`

	Unique
}

func (s *User) AuditData() AuditData {
	return AuditData{
		"id":               s.ID,
		"principal_name":   s.PrincipalName,
		"first_name":       s.FirstName,
		"last_name":        s.LastName,
		"email_address":    s.EmailAddress,
		"roles":            s.Roles.IDs(),
		"saml_provider_id": s.SAMLProviderID.ValueOrZero(),
		"is_disabled":      s.IsDisabled,
		"eula_accepted":    s.EULAAccepted,
	}
}

func (s *User) RemoveRole(role Role) {
	s.Roles = s.Roles.RemoveByName(role.Name)
}

type Users []User

func (s Users) IsSortable(column string) bool {
	switch column {
	case "first_name",
		"last_name",
		"email_address",
		"principal_name",
		"last_login",
		"created_at",
		"updated_at",
		"deleted_at":
		return true
	default:
		return false
	}
}

func (s Users) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"first_name":     {Equals, NotEquals},
		"last_name":      {Equals, NotEquals},
		"email_address":  {Equals, NotEquals},
		"principal_name": {Equals, NotEquals},
		"id":             {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"last_login":     {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_at":     {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"updated_at":     {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"deleted_at":     {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
	}
}

func (s Users) IsString(column string) bool {
	switch column {
	case "first_name",
		"last_name",
		"email_address",
		"principal_name":
		return true
	default:
		return false
	}
}

func (s Users) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s Users) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

func UserSessionAssociations() []string {
	return []string{
		"User.SAMLProvider",
		"User.AuthSecret",
		"User.AuthTokens",
		"User.Roles.Permissions",
	}
}

type SessionAuthProvider int

const (
	SessionAuthProviderSecret SessionAuthProvider = 0
	SessionAuthProviderSAML   SessionAuthProvider = 1
)

type UserSession struct {
	User             User `gorm:"constraint:OnDelete:CASCADE;"`
	UserID           uuid.UUID
	AuthProviderType SessionAuthProvider
	AuthProviderID   int32
	ExpiresAt        time.Time

	BigSerial
}

// Expired returns true if the user session has expired, false otherwise
func (s UserSession) Expired() bool {
	return s.ExpiresAt.Before(time.Now().UTC())
}
