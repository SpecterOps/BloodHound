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

package api_test

import (
	"testing"

	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
)

func TestAssetGroupMembers_SortBy(t *testing.T) {
	input := api.AssetGroupMembers{
		api.AssetGroupMember{
			AssetGroupID:    2,
			ObjectID:        "b",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name2",
			CustomMember:    true,
		},
		api.AssetGroupMember{
			AssetGroupID:    1,
			ObjectID:        "a",
			PrimaryKind:     ad.Computer.String(),
			Kinds:           []string{"Base", "Computer"},
			EnvironmentID:   "domainsid",
			EnvironmentKind: "Domain",
			Name:            "name1",
			CustomMember:    false,
		},
		api.AssetGroupMember{
			AssetGroupID:    3,
			ObjectID:        "c",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid2",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name3",
			CustomMember:    false,
		},
	}

	_, err := input.SortBy([]string{""})
	require.NotNil(t, err)
	require.Equal(t, api.ErrorResponseEmptySortParameter, err.Error())

	_, err = input.SortBy([]string{"foobar"})
	require.Equal(t, api.ErrorResponseDetailsNotSortable, err.Error())

	output, err := input.SortBy([]string{"-asset_group_id"})
	require.Nil(t, err)
	require.Equal(t, 3, output[0].AssetGroupID)

	output, err = input.SortBy([]string{"primary_kind"})
	require.Nil(t, err)
	require.Equal(t, azure.Group.String(), output[0].PrimaryKind)

	output, err = input.SortBy([]string{"environment_id"})
	require.Nil(t, err)
	require.Equal(t, "domainsid", output[0].EnvironmentID)

	output, err = input.SortBy([]string{"environment_kind"})
	require.Nil(t, err)
	require.Equal(t, azure.Tenant.String(), output[0].EnvironmentKind)

	output, err = input.SortBy([]string{"name"})
	require.Nil(t, err)
	require.Equal(t, "name1", output[0].Name)
}

func TestAssetGroupMembers_Filter_Equals(t *testing.T) {
	input := api.AssetGroupMembers{
		api.AssetGroupMember{
			AssetGroupID:    2,
			ObjectID:        "b",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name2",
			CustomMember:    true,
		},
		api.AssetGroupMember{
			AssetGroupID:    1,
			ObjectID:        "a",
			PrimaryKind:     ad.Computer.String(),
			Kinds:           []string{"Base", "Computer"},
			EnvironmentID:   "domainsid",
			EnvironmentKind: "Domain",
			Name:            "name1",
			CustomMember:    false,
		},
		api.AssetGroupMember{
			AssetGroupID:    3,
			ObjectID:        "c",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid2",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name3",
			CustomMember:    false,
		},
	}

	_, err := input.Filter(model.QueryParameterFilterMap{
		"badcolumn": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "badcolumn",
				Operator: model.Equals,
				Value:    "1",
			},
		},
	})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), model.ErrorResponseDetailsColumnNotFilterable)

	_, err = input.Filter(model.QueryParameterFilterMap{
		"object_id": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "object_id",
				Operator: model.GreaterThan, // invalid operator
				Value:    "a",
			},
		},
	})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), model.ErrorResponseDetailsFilterPredicateNotSupported)

	// filter on object_id
	output, err := input.Filter(model.QueryParameterFilterMap{
		"object_id": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "object_id",
				Operator: model.Equals,
				Value:    "a",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 1, output[0].AssetGroupID)

	// filter on name
	output, err = input.Filter(model.QueryParameterFilterMap{
		"name": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "name",
				Operator: model.Equals,
				Value:    "name3",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 3, output[0].AssetGroupID)

	// filter on custom_member
	output, err = input.Filter(model.QueryParameterFilterMap{
		"custom_member": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "custom_member",
				Operator: model.Equals,
				Value:    "false",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))

	// filter on environment_id
	output, err = input.Filter(model.QueryParameterFilterMap{
		"environment_id": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "environment_id",
				Operator: model.Equals,
				Value:    "tenantid",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 2, output[0].AssetGroupID)

	// filter on environment_kind
	output, err = input.Filter(model.QueryParameterFilterMap{
		"environment_kind": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "environment_kind",
				Operator: model.Equals,
				Value:    "AZTenant",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))

	// filter on primary_kind
	output, err = input.Filter(model.QueryParameterFilterMap{
		"primary_kind": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "primary_kind",
				Operator: model.Equals,
				Value:    "Computer",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 1, output[0].AssetGroupID)
}

func TestAssetGroupMembers_Filter_NotEquals(t *testing.T) {
	input := api.AssetGroupMembers{
		api.AssetGroupMember{
			AssetGroupID:    2,
			ObjectID:        "b",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name2",
			CustomMember:    true,
		},
		api.AssetGroupMember{
			AssetGroupID:    1,
			ObjectID:        "a",
			PrimaryKind:     ad.Computer.String(),
			Kinds:           []string{"Base", "Computer"},
			EnvironmentID:   "domainsid",
			EnvironmentKind: "Domain",
			Name:            "name1",
			CustomMember:    false,
		},
		api.AssetGroupMember{
			AssetGroupID:    3,
			ObjectID:        "c",
			PrimaryKind:     azure.Group.String(),
			Kinds:           []string{"Base", "Group"},
			EnvironmentID:   "tenantid2",
			EnvironmentKind: azure.Tenant.String(),
			Name:            "name3",
			CustomMember:    false,
		},
	}

	// filter on object_id
	output, err := input.Filter(model.QueryParameterFilterMap{
		"object_id": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "object_id",
				Operator: model.NotEquals,
				Value:    "b",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))

	// filter on name
	output, err = input.Filter(model.QueryParameterFilterMap{
		"name": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "name",
				Operator: model.NotEquals,
				Value:    "name3",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))

	// filter on custom_member
	output, err = input.Filter(model.QueryParameterFilterMap{
		"custom_member": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "custom_member",
				Operator: model.NotEquals,
				Value:    "false",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 2, output[0].AssetGroupID)

	// filter on environment_id
	output, err = input.Filter(model.QueryParameterFilterMap{
		"environment_id": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "environment_id",
				Operator: model.NotEquals,
				Value:    "tenantid",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))

	// filter on environment_kind
	output, err = input.Filter(model.QueryParameterFilterMap{
		"environment_kind": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "environment_kind",
				Operator: model.NotEquals,
				Value:    "AZTenant",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 1, len(output))
	require.Equal(t, 1, output[0].AssetGroupID)

	// filter on primary_kind
	output, err = input.Filter(model.QueryParameterFilterMap{
		"primary_kind": model.QueryParameterFilters{
			model.QueryParameterFilter{
				Name:     "primary_kind",
				Operator: model.NotEquals,
				Value:    "Computer",
			},
		},
	})

	require.Nil(t, err)
	require.Equal(t, 2, len(output))
}

func TestAssetGroupMembers_BuildFilteringConditional_Error(t *testing.T) {
	input := api.AssetGroupMembers{}
	columns := input.GetFilterableColumns()

	// invalid predicates for all columns
	for _, column := range columns {
		_, err := input.BuildFilteringConditional(column, model.GreaterThan, "1")
		require.Error(t, err)
		require.Contains(t, err.Error(), api.ErrorResponseDetailsFilterPredicateNotSupported)
	}

	// invalid column
	_, err := input.BuildFilteringConditional("badcolumn", model.GreaterThan, "1")
	require.Error(t, err)
	require.Contains(t, err.Error(), api.ErrorResponseDetailsColumnNotFilterable)

	// invalid values
	_, err = input.BuildFilteringConditional("custom_member", model.Equals, "1234")
	require.Error(t, err)
	require.Contains(t, err.Error(), api.ErrorResponseDetailsBadQueryParameterFilters)

	_, err = input.BuildFilteringConditional("asset_group_id", model.Equals, "abcd")
	require.Error(t, err)
	require.Contains(t, err.Error(), api.ErrorResponseDetailsBadQueryParameterFilters)
}
