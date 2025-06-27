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

package nodeprops_test

import (
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/analysis/ad/internal/nodeprops"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorReadDomainSIDandNameAsString(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "empty message",
			message: "",
			want:    "failed to read domain SID and name: ",
		},
		{
			name:    "with message",
			message: "test error",
			want:    "failed to read domain SID and name: test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("failed to read domain SID and name: %s", tt.message)
			assert.Equal(t, tt.want, err.Error())
		})
	}
}

func TestReadDomainIDandNameAsString(t *testing.T) {
	tests := []struct {
		name          string
		node          *graph.Node
		wantSID       string
		wantName      string
		wantErrSubstr string
	}{
		{
			name:          "nil node",
			node:          nil,
			wantSID:       "",
			wantName:      "",
			wantErrSubstr: "given nodeToRead is nil",
		},
		{
			name: "missing domain SID",
			node: &graph.Node{
				Properties: graph.NewProperties(),
			},
			wantSID:       "",
			wantName:      "",
			wantErrSubstr: "read domain SID property value is nil",
		},
		{
			name: "empty domain SID",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "")
					return props
				}(),
			},
			wantSID:       "",
			wantName:      "",
			wantErrSubstr: "read domain SID is empty or blank",
		},
		{
			name: "empty domain name",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "S-1-5-21-123456789-123456789-123456789")
					props.Set(common.Name.String(), "")
					return props
				}(),
			},
			wantSID:       "",
			wantName:      "",
			wantErrSubstr: "read domain name is empty or blank",
		},
		{
			name: "successful read",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "S-1-5-21-123456789-123456789-123456789")
					props.Set(common.Name.String(), "EXAMPLE.COM")
					return props
				}(),
			},
			wantSID:       "S-1-5-21-123456789-123456789-123456789",
			wantName:      "EXAMPLE.COM",
			wantErrSubstr: "",
		},
		{
			name: "missing domain name",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "S-1-5-21-123456789-123456789-123456789")
					// No name property set
					return props
				}(),
			},
			wantSID:       "",
			wantName:      "",
			wantErrSubstr: "read domain name property value is nil",
		},
		{
			name: "domain SID with whitespace",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "  S-1-5-21-123456789-123456789-123456789  ")
					props.Set(common.Name.String(), "EXAMPLE.COM")
					return props
				}(),
			},
			wantSID:       "  S-1-5-21-123456789-123456789-123456789  ",
			wantName:      "EXAMPLE.COM",
			wantErrSubstr: "",
		},
		{
			name: "domain name with whitespace",
			node: &graph.Node{
				Properties: func() *graph.Properties {
					props := graph.NewProperties()
					props.Set(ad.DomainSID.String(), "S-1-5-21-123456789-123456789-123456789")
					props.Set(common.Name.String(), "  EXAMPLE.COM  ")
					return props
				}(),
			},
			wantSID:       "S-1-5-21-123456789-123456789-123456789",
			wantName:      "  EXAMPLE.COM  ",
			wantErrSubstr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSID, gotName, err := nodeprops.ReadDomainIDandNameAsString(tt.node)

			assert.Equal(t, tt.wantSID, gotSID, "SID value should match expected")
			assert.Equal(t, tt.wantName, gotName, "Name value should match expected")

			if tt.wantErrSubstr != "" {
				require.Error(t, err, "Expected an error but got none")
				assert.Contains(t, err.Error(), tt.wantErrSubstr, "Error message should contain expected substring")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}
		})
	}
}
