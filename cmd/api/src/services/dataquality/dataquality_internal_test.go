// Copyright 2026 Specter Ops, Inc.
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

package dataquality

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
)

func TestSelectDataQualityObjectNodeKind(t *testing.T) {
	var (
		oktaBaseKind    = graph.StringKind("OktaBase")
		oktaUserKind    = graph.StringKind("OktaUser")
		oktaGroupKind   = graph.StringKind("OktaGroup")
		githubBaseKind  = graph.StringKind("GitHubBase")
		sourceKindNames = map[string]struct{}{
			oktaBaseKind.String():   {},
			githubBaseKind.String(): {},
		}
		primaryDisplayKinds = graphschema.PrimaryDisplayKinds{
			oktaGroupKind: {},
		}
	)

	testCases := []struct {
		name         string
		sourceKind   graph.Kind
		nodeKinds    graph.Kinds
		expectedKind graph.Kind
	}{
		{
			name:         "source only node falls back to source kind",
			sourceKind:   oktaBaseKind,
			nodeKinds:    graph.Kinds{oktaBaseKind},
			expectedKind: oktaBaseKind,
		},
		{
			name:         "known active directory kind is selected",
			sourceKind:   ad.Entity,
			nodeKinds:    graph.Kinds{ad.Entity, ad.User},
			expectedKind: ad.User,
		},
		{
			name:         "registered source kind labels are ignored as object kinds",
			sourceKind:   oktaBaseKind,
			nodeKinds:    graph.Kinds{oktaBaseKind, githubBaseKind, oktaUserKind},
			expectedKind: oktaUserKind,
		},
		{
			name:         "primary display kind wins over earlier non display kind",
			sourceKind:   oktaBaseKind,
			nodeKinds:    graph.Kinds{oktaBaseKind, oktaUserKind, oktaGroupKind},
			expectedKind: oktaGroupKind,
		},
		{
			name:         "extended node kinds are ignored",
			sourceKind:   oktaBaseKind,
			nodeKinds:    graph.Kinds{oktaBaseKind, graph.StringKind("Tag_Tier_Zero"), oktaUserKind},
			expectedKind: oktaUserKind,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualKind := selectDataQualityObjectNodeKind(
				mergeDataQualityPrimaryDisplayKinds(primaryDisplayKinds),
				testCase.sourceKind,
				sourceKindNames,
				testCase.nodeKinds,
			)

			assert.Equal(t, testCase.expectedKind, actualKind)
		})
	}
}

func TestDataQualityEnvironmentContexts(t *testing.T) {
	var (
		oktaEnvironmentKind = graph.StringKind("OktaOrg")
		sourceKindKinds     = []model.Kind{
			{ID: 11, Name: ad.Entity.String()},
			{ID: 22, Name: azure.Entity.String()},
			{ID: 33, Name: "OktaBase"},
		}
		environments = []model.SchemaEnvironment{
			{
				EnvironmentKindName: oktaEnvironmentKind.String(),
				SourceKindId:        33,
			},
			{
				EnvironmentKindName: ad.Domain.String(),
				SourceKindId:        11,
			},
			{
				EnvironmentKindName: azure.Tenant.String(),
				SourceKindId:        22,
			},
			{
				EnvironmentKindName: "UnknownEnvironment",
				SourceKindId:        99,
			},
		}
		expectedContexts = []dataQualityEnvironmentContext{
			{
				SourceKind:            azure.Entity.String(),
				EnvironmentKind:       azure.Tenant.String(),
				EnvironmentGraphKind:  azure.Tenant,
				EnvironmentIDProperty: azure.TenantID.String(),
			},
			{
				SourceKind:            ad.Entity.String(),
				EnvironmentKind:       ad.Domain.String(),
				EnvironmentGraphKind:  ad.Domain,
				EnvironmentIDProperty: ad.DomainSID.String(),
			},
			{
				SourceKind:            "OktaBase",
				EnvironmentKind:       oktaEnvironmentKind.String(),
				EnvironmentGraphKind:  oktaEnvironmentKind,
				EnvironmentIDProperty: graphschema.EnvironmentIDKey,
			},
		}
	)

	actualContexts := dataQualityEnvironmentContexts(environments, dataQualityKindNameByID(sourceKindKinds))

	assert.Equal(t, expectedContexts, actualContexts)
}

func TestDataQualityEnvironmentSourceKindIDs(t *testing.T) {
	var (
		environments = []model.SchemaEnvironment{
			{SourceKindId: 33},
			{SourceKindId: 11},
			{SourceKindId: 33},
			{SourceKindId: 0},
		}
		expectedSourceKindIDs = []int32{11, 33}
	)

	actualSourceKindIDs := dataQualityEnvironmentSourceKindIDs(environments)

	assert.Equal(t, expectedSourceKindIDs, actualSourceKindIDs)
}
