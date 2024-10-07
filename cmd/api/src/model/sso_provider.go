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

// SSOProviderType is the representation of the sso_provider_types enum declared in our database
// Adding a new type will require an accompanying migration
type SSOProviderType int

const (
	SSOProviderTypeSAML SSOProviderType = 0
	SSOProviderTypeOIDC SSOProviderType = 1
)

// SSOProvider is the common representation of an SSO provider that can be used to display high level information about that provider
type SSOProvider struct {
	Type SSOProviderType `json:"type"`
	Name string          `json:"name"`
	Slug string          `json:"slug"`

	Serial
}
