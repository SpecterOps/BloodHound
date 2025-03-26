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

import "fmt"

// NodeNamePrefix represents the prefix portion of a node name in Active Directory.
// It provides an immutable, type-safe way to work with node name prefixes.
//
// The interface approach ensures that predefined node name prefixes are constant
// and cannot be modified after initialization, providing stronger guarantees than
// string constants.
type NodeNamePrefix interface {
	// String returns the string representation of the node name prefix.
	fmt.Stringer

	// AppendSuffix combines this prefix with a domain suffix to form a complete node name.
	AppendSuffix(suffix string) string
}

// nodeNamePrefix is the concrete implementation of the NodeNamePrefix interface.
type nodeNamePrefix string

// Verify that nodeNamePrefix implements the NodeNamePrefix interface.
var _ NodeNamePrefix = nodeNamePrefix("")

// String returns the string representation of the node name prefix.
func (n nodeNamePrefix) String() string {
	return string(n)
}

// AppendSuffix combines this prefix with a domain suffix to form a complete node name.
// This is used to create domain-specific node names by combining a well-known prefix
// with a domain suffix.
func (n nodeNamePrefix) AppendSuffix(suffix string) string {
	return fmt.Sprintf("%s@%s", n.String(), suffix)
}

// NewNodeNamePrefix creates a new NodeNamePrefix from a string.
// This function should be used to create custom node name prefixes when needed.
func NewNodeNamePrefix(prefix string) NodeNamePrefix {
	return nodeNamePrefix(prefix)
}

// Predefined well-known node name prefixes for common Active Directory security principals.
// These are implemented as interface values to ensure they cannot be modified
// after initialization, providing stronger guarantees than string constants.
var (
	// AuthenticatedUsersNodeNamePrefix represents the node name prefix for the "Authenticated Users" group.
	AuthenticatedUsersNodeNamePrefix NodeNamePrefix = NewNodeNamePrefix("AUTHENTICATED USERS")

	// DomainComputerNodeNamePrefix represents the node name prefix for the "Domain Computers" group.
	DomainComputerNodeNamePrefix = NewNodeNamePrefix("DOMAIN COMPUTERS")

	// DomainUsersNodeNamePrefix represents the node name prefix for the "Domain Users" group.
	DomainUsersNodeNamePrefix = NewNodeNamePrefix("DOMAIN USERS")

	// EveryoneNodeNamePrefix represents the node name prefix for the "Everyone" group.
	EveryoneNodeNamePrefix = NewNodeNamePrefix("EVERYONE")
)

// DefineNodeName creates a complete node name by combining a well-known node name prefix
// with a domain suffix.
//
// This function provides a clear, explicit way to construct complete node names while
// enforcing the use of the NodeNamePrefix interface. This design ensures that only
// properly defined node name prefixes (either predefined or created via NewNodeNamePrefix)
// can be used, reducing the risk of errors from string manipulation.
//
// The NodeNamePrefix parameter requirement makes the function's purpose explicit
// and prevents accidental misuse with arbitrary strings.
func DefineNodeName(prefix NodeNamePrefix, suffix string) string {
	return prefix.AppendSuffix(suffix)
}
