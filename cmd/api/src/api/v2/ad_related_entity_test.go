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

package v2_test

import (
	"net/http"
	"testing"

	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	dbMocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/queries"
	graphMocks "github.com/specterops/bloodhound/src/queries/mocks"
	"go.uber.org/mock/gomock"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/errors"
)

func setup(t *testing.T) (*gomock.Controller, *graphMocks.MockGraph, *dbMocks.MockDatabase, v2.Resources) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = graphMocks.NewMockGraph(mockCtrl)
		mockDB    = dbMocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph, DB: mockDB}
	)
	return mockCtrl, mockGraph, mockDB, resources
}

func setupCases(mockGraph *graphMocks.MockGraph, mockDB *dbMocks.MockDatabase) []apitest.Case {
	return []apitest.Case{
		{
			Name: "RepoGetEntityQueryParamsError",
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusBadRequest)
				apitest.BodyContains(output, "no object ID found in request")
			},
		},
		{
			Name: "GraphDBGetADEntityQueryResultGraphUnsupportedError",
			Input: func(input *apitest.Input) {
				apitest.SetURLVar(input, "object_id", "1")
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, queries.ErrGraphUnsupported)
				mockDB.EXPECT().
					GetFlagByKey("entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
			},
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusBadRequest)
				apitest.BodyContains(output, "type 'graph' is not supported for this endpoint")
			},
		},
		{
			Name: "GraphDBGetADEntityQueryResultUnsupportedDataTypeError",
			Input: func(input *apitest.Input) {
				apitest.SetURLVar(input, "object_id", "1")
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, queries.ErrUnsupportedDataType)
				mockDB.EXPECT().
					GetFlagByKey("entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
			},
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusBadRequest)
				apitest.BodyContains(output, "unsupported result type for this query")
			},
		},
		{
			Name: "GraphDBGetADEntityQueryResultMemoryLimitError",
			Input: func(input *apitest.Input) {
				apitest.SetURLVar(input, "object_id", "1")
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, ops.ErrTraversalMemoryLimit)
				mockDB.EXPECT().
					GetFlagByKey("entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
			},
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusInternalServerError)
				apitest.BodyContains(output, "calculating the request results exceeded memory limitations due to the volume of objects involved")
			},
		},
		{
			Name: "GraphDBGetADEntityQueryResultUnexpectedError",
			Input: func(input *apitest.Input) {
				apitest.SetURLVar(input, "object_id", "1")
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("any other error"))
				mockDB.EXPECT().
					GetFlagByKey("entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
			},
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusInternalServerError)
				apitest.BodyContains(output, "an unknown error occurred during the request")
			},
		},
		{
			Name: "Success",
			Input: func(input *apitest.Input) {
				apitest.SetURLVar(input, "object_id", "1")
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)
				mockDB.EXPECT().
					GetFlagByKey("entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
			},
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusOK)
			},
		},
	}
}

func TestResources_ListADUserSessions(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADUserSessions).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADUserSQLAdminRights(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADUserSQLAdminRights).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADGroupSessions(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADGroupSessions).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADComputerSessions(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADComputerSessions).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADComputerAdmins(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADComputerAdmins).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADComputerPSRemoteUsers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADComputerPSRemoteUsers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADComputerRDPUsers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADComputerRDPUsers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADComputerDCOMUsers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADComputerDCOMUsers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADGroupMembership(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADGroupMembership).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADGroupMembers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADGroupMembers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADComputerSQLAdmins(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADComputerSQLAdmins).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADComputerConstrainedDelegationUsers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADComputerConstrainedDelegationUsers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADEntityConstrainedDelegationRights(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADEntityConstrainedDelegationRights).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADEntityAdminRights(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADEntityAdminRights).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADEntityRDPRights(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADEntityRDPRights).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADEntityPSRemoteRights(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADEntityPSRemoteRights).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADEntityDCOMRights(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADEntityDCOMRights).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADEntityControllers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADEntityControllers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADEntityControllables(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADEntityControllables).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADEntityLinkedGPOs(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADEntityLinkedGPOs).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainContainedUsers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainContainedUsers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainContainedComputers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainContainedComputers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainContainedGroups(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainContainedGroups).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainContainedOUs(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainContainedOUs).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainContainedGPOs(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainContainedGPOs).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainForeignGroups(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainForeignGroups).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainForeignUsers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainForeignUsers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainForeignAdmins(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainForeignAdmins).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainForeignGPOControllers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainForeignGPOControllers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainOutboundTrusts(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainOutboundTrusts).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainInboundTrusts(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainInboundTrusts).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADDomainDCSyncers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADDomainDCSyncers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADOUContainedUsers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADOUContainedUsers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADOUContainedGroups(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADOUContainedGroups).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADOUContainedComputers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADOUContainedComputers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADGPOAffectedContainers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADGPOAffectedContainers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADGPOAffectedUsers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADGPOAffectedUsers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADGPOAffectedComputers(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADGPOAffectedComputers).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADGPOAffectedTierZero(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADGPOAffectedTierZero).
		Run(setupCases(mockGraph, mockDB))
}
