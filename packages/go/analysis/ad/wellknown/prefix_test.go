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

package wellknown

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeNamePrefix_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		prefix nodeNamePrefix
		want   string
	}{
		{
			name:   "empty prefix",
			prefix: nodeNamePrefix(""),
			want:   "",
		},
		{
			name:   "non-empty prefix",
			prefix: nodeNamePrefix("TEST"),
			want:   "TEST",
		},
		{
			name:   "with special characters",
			prefix: nodeNamePrefix("TEST-123"),
			want:   "TEST-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.prefix.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNodeNamePrefix_AppendSuffix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		prefix nodeNamePrefix
		suffix string
		want   string
	}{
		{
			name:   "empty prefix with empty suffix",
			prefix: nodeNamePrefix(""),
			suffix: "",
			want:   "@",
		},
		{
			name:   "non-empty prefix with empty suffix",
			prefix: nodeNamePrefix("TEST"),
			suffix: "",
			want:   "TEST@",
		},
		{
			name:   "empty prefix with non-empty suffix",
			prefix: nodeNamePrefix(""),
			suffix: "example.com",
			want:   "@example.com",
		},
		{
			name:   "non-empty prefix with non-empty suffix",
			prefix: nodeNamePrefix("TEST"),
			suffix: "example.com",
			want:   "TEST@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.prefix.AppendSuffix(tt.suffix)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewNodeNamePrefix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{
			name:   "empty prefix",
			prefix: "",
			want:   "",
		},
		{
			name:   "non-empty prefix",
			prefix: "TEST",
			want:   "TEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewNodeNamePrefix(tt.prefix)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestDefineNodeName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		prefix NodeNamePrefix
		suffix string
		want   string
	}{
		{
			name:   "with predefined prefix",
			prefix: DomainUsersNodeNamePrefix,
			suffix: "example.com",
			want:   "DOMAIN USERS@example.com",
		},
		{
			name:   "with custom prefix",
			prefix: NewNodeNamePrefix("CUSTOM"),
			suffix: "example.com",
			want:   "CUSTOM@example.com",
		},
		{
			name:   "with empty suffix",
			prefix: NewNodeNamePrefix("TEST"),
			suffix: "",
			want:   "TEST@",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DefineNodeName(tt.prefix, tt.suffix)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPredefinedNodeNamePrefixes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		prefix NodeNamePrefix
		want   string
	}{
		{
			name:   "DomainUsersNodeNamePrefix",
			prefix: DomainUsersNodeNamePrefix,
			want:   "DOMAIN USERS",
		},
		{
			name:   "AuthenticatedUsersNodeNamePrefix",
			prefix: AuthenticatedUsersNodeNamePrefix,
			want:   "AUTHENTICATED USERS",
		},
		{
			name:   "EveryoneNodeNamePrefix",
			prefix: EveryoneNodeNamePrefix,
			want:   "EVERYONE",
		},
		{
			name:   "DomainComputerNodeNamePrefix",
			prefix: DomainComputerNodeNamePrefix,
			want:   "DOMAIN COMPUTERS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.prefix.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
