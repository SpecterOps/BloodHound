// Copyright 2025 Specter Ops, Inc.
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

// Package wellknown provides constants and utilities for working with well-known
// Active Directory security identifiers (SIDs) and node names.
//
// This package helps maintain consistency when referencing common AD entities
// across the codebase by providing immutable, type-safe references to well-known
// values.
package wellknown

import (
	"fmt"
)

// SIDSuffix represents the suffix portion of a Security Identifier (SID) in Active Directory.
// It provides an immutable, type-safe way to work with SID suffixes.
//
// The interface approach ensures that predefined SID suffixes are constant and cannot
// be modified after initialization, providing stronger guarantees than string constants.
type SIDSuffix interface {
	// String returns the string representation of the SID suffix.
	fmt.Stringer

	// PrependPrefix combines a SID prefix with this suffix to form a complete SID.
	PrependPrefix(prefix string) string
}

// sidSuffix is the concrete implementation of the SIDSuffix interface.
type sidSuffix string

// Verify that sidSuffix implements the SIDSuffix interface.
var _ SIDSuffix = sidSuffix("")

// String returns the string representation of the SID suffix.
func (s sidSuffix) String() string {
	return string(s)
}

// PrependPrefix combines a SID prefix with this suffix to form a complete SID.
// This is used to create domain-specific SIDs by combining a domain's unique
// prefix with a well-known suffix.
func (s sidSuffix) PrependPrefix(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, s.String())
}

// NewSIDSuffix creates a new SIDSuffix from a string.
// This function should be used to create custom SID suffixes when needed.
func NewSIDSuffix(suffix string) SIDSuffix {
	return sidSuffix(suffix)
}

// Predefined well-known SID suffixes for common Active Directory security principals.
// These are implemented as interface values to ensure they cannot be modified
// after initialization, providing stronger guarantees than string constants.
var (
	AdministratorAccountSIDSuffix                    = NewSIDSuffix("-500")
	GuestSIDSuffix                                   = NewSIDSuffix("-501")
	DomainAdminsGroupSIDSuffix                       = NewSIDSuffix("-512")
	DomainUsersSIDSuffix                             = NewSIDSuffix("-513")
	DomainComputersSIDSuffix                         = NewSIDSuffix("-515")
	DomainControllersGroupSIDSuffix                  = NewSIDSuffix("-516")
	SchemaAdminsGroupSIDSuffix                       = NewSIDSuffix("-518")
	EnterpriseAdminsGroupSIDSuffix                   = NewSIDSuffix("-519")
	ProtectedUsersSIDSuffix                          = NewSIDSuffix("-525")
	KeyAdminsGroupSIDSuffix                          = NewSIDSuffix("-526")
	EnterpriseKeyAdminsGroupSIDSuffix                = NewSIDSuffix("-527")
	AdministratorsSIDSuffix                          = NewSIDSuffix("-544")
	BackupOperatorsGroupSIDSuffix                    = NewSIDSuffix("-551")
	RemoteDesktopUsersSIDSuffix                      = NewSIDSuffix("-555")
	EnterpriseDomainControllersGroupSIDSuffix        = NewSIDSuffix("1-5-9")
	EveryoneSIDSuffix                                = NewSIDSuffix("-S-1-1-0")
	NetworkSIDSuffix                                 = NewSIDSuffix("-S-1-5-2")
	AuthenticatedUsersSIDSuffix                      = NewSIDSuffix("-S-1-5-11")
	ThisOrganizationSIDSuffix                        = NewSIDSuffix("-S-1-5-15")
	AuthenticationAuthorityAssertedIdentitySIDSuffix = NewSIDSuffix("-S-1-18-1")
	KeyTrustSIDSuffix                                = NewSIDSuffix("-S-1-18-4")
	MFAKeyPropertySIDSuffix                          = NewSIDSuffix("-S-1-18-5")
	NTLMAuthenticationSIDSuffix                      = NewSIDSuffix("-S-1-64-10")
	SchannelAuthenticationSIDSuffix                  = NewSIDSuffix("-S-1-64-14")
	ThisOrganizationCertificateSIDSuffix             = NewSIDSuffix("-S-1-5-65-1")
)

// DefineSID creates a complete SID by combining a domain-specific SID prefix with
// a well-known SID suffix.
//
// This function provides a clear, explicit way to construct complete SIDs while
// enforcing the use of the SIDSuffix interface. This design ensures that only
// properly defined SID suffixes (either predefined or created via NewSIDSuffix)
// can be used, reducing the risk of errors from string manipulation.
//
// The SIDSuffix parameter requirement makes the function's purpose explicit
// and prevents accidental misuse with arbitrary strings.
func DefineSID(sidPrefix string, sidSuffix SIDSuffix) string {
	return sidSuffix.PrependPrefix(sidPrefix)
}
