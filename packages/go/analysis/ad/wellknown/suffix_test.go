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

func TestSIDSuffix_String(t *testing.T) {
	tests := []struct {
		name   string
		suffix sidSuffix
		want   string
	}{
		{
			name:   "empty suffix",
			suffix: sidSuffix(""),
			want:   "",
		},
		{
			name:   "non-empty suffix",
			suffix: sidSuffix("-123"),
			want:   "-123",
		},
		{
			name:   "with special characters",
			suffix: sidSuffix("-S-1-5-11"),
			want:   "-S-1-5-11",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.suffix.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSIDSuffix_PrependPrefix(t *testing.T) {
	tests := []struct {
		name   string
		suffix sidSuffix
		prefix string
		want   string
	}{
		{
			name:   "empty suffix with empty prefix",
			suffix: sidSuffix(""),
			prefix: "",
			want:   "",
		},
		{
			name:   "non-empty suffix with empty prefix",
			suffix: sidSuffix("-123"),
			prefix: "",
			want:   "-123",
		},
		{
			name:   "empty suffix with non-empty prefix",
			suffix: sidSuffix(""),
			prefix: "S-1-5",
			want:   "S-1-5",
		},
		{
			name:   "non-empty suffix with non-empty prefix",
			suffix: sidSuffix("-123"),
			prefix: "S-1-5",
			want:   "S-1-5-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.suffix.PrependPrefix(tt.prefix)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewSIDSuffix(t *testing.T) {
	tests := []struct {
		name   string
		suffix string
		want   string
	}{
		{
			name:   "empty suffix",
			suffix: "",
			want:   "",
		},
		{
			name:   "non-empty suffix",
			suffix: "-123",
			want:   "-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSIDSuffix(tt.suffix)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestDefineSID(t *testing.T) {
	tests := []struct {
		name      string
		sidPrefix string
		sidSuffix SIDSuffix
		want      string
	}{
		{
			name:      "with predefined suffix",
			sidPrefix: "S-1-5-21-3130019616-2776909439-2417379446",
			sidSuffix: DomainUsersSIDSuffix,
			want:      "S-1-5-21-3130019616-2776909439-2417379446-513",
		},
		{
			name:      "with custom suffix",
			sidPrefix: "S-1-5-21-3130019616-2776909439-2417379446",
			sidSuffix: NewSIDSuffix("-999"),
			want:      "S-1-5-21-3130019616-2776909439-2417379446-999",
		},
		{
			name:      "with empty prefix",
			sidPrefix: "",
			sidSuffix: NewSIDSuffix("-123"),
			want:      "-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefineSID(tt.sidPrefix, tt.sidSuffix)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPredefinedSIDSuffixes(t *testing.T) {
	tests := []struct {
		name   string
		suffix SIDSuffix
		want   string
	}{
		{
			name:   "DomainUsersSIDSuffix",
			suffix: DomainUsersSIDSuffix,
			want:   "-513",
		},
		{
			name:   "AuthenticatedUsersSIDSuffix",
			suffix: AuthenticatedUsersSIDSuffix,
			want:   "-S-1-5-11",
		},
		{
			name:   "EveryoneSIDSuffix",
			suffix: EveryoneSIDSuffix,
			want:   "-S-1-1-0",
		},
		{
			name:   "DomainComputersSIDSuffix",
			suffix: DomainComputersSIDSuffix,
			want:   "-515",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.suffix.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
