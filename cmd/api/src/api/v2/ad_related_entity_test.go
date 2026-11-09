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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	dbMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setup(t *testing.T) (*gomock.Controller, *mocks.MockGraph, *dbMocks.MockDatabase, v2.Resources) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks.NewMockGraph(mockCtrl)
		mockDB    = dbMocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph, DB: mockDB, DogTags: dogtags.NewTestService(dogtags.TestOverrides{})}
	)
	return mockCtrl, mockGraph, mockDB, resources
}

func setupCases(mockGraph *mocks.MockGraph, mockDB *dbMocks.MockDatabase) []apitest.Case {
	bheCtx := bhctx.Context{
		AuthCtx: auth.Context{
			PermissionOverrides: auth.PermissionOverrides{},
			Owner:               model.User{},
			Session:             model.UserSession{},
		},
	}

	return []apitest.Case{
		{
			Name: "RepoGetEntityQueryParamsError",
			Input: func(input *apitest.Input) {
				apitest.SetContext(input, bheCtx.ConstructGoContext())
			},
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusBadRequest)
				apitest.BodyContains(output, "no object ID found in request")
			},
		},
		{
			Name: "GraphDBGetADEntityQueryResultGraphUnsupportedError",
			Input: func(input *apitest.Input) {
				apitest.SetURLVar(input, "object_id", "1")
				apitest.SetContext(input, bheCtx.ConstructGoContext())
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, queries.ErrGraphUnsupported)
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mockDB.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
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
				apitest.SetContext(input, bheCtx.ConstructGoContext())
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, queries.ErrUnsupportedDataType)
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mockDB.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
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
				apitest.SetContext(input, bheCtx.ConstructGoContext())
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, ops.ErrGraphQueryMemoryLimit)
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mockDB.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
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
				apitest.SetContext(input, bheCtx.ConstructGoContext())
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("any other error"))
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mockDB.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
			},
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusInternalServerError)
				apitest.BodyContains(output, "an unknown error occurred during the request")
			},
		},
		{
			Name: "GraphDBGetADEntityQueryResultTypeGraphNotPaginatedResults",
			Input: func(input *apitest.Input) {
				apitest.SetURLVar(input, "object_id", "1")
				apitest.AddQueryParam(input, "type", "graph")
				apitest.SetContext(input, bheCtx.ConstructGoContext())
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(graph.PathSet{}, nil)
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mockDB.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
			},
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusOK)
				// This flat unnested shape maintains the current api contract for a type=graph query
				// Assert that the response does not contain pagination properties
				apitest.BodyNotContains(output, "data")
				apitest.BodyNotContains(output, "skip")
				apitest.BodyNotContains(output, "limit")
				apitest.BodyNotContains(output, "count")
			},
		},
		{
			Name: "Success",
			Input: func(input *apitest.Input) {
				apitest.SetURLVar(input, "object_id", "1")
				// Delete the type=graph param so that we get list results
				apitest.DeleteQueryParam(input, "type")
				apitest.SetContext(input, bheCtx.ConstructGoContext())
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(graph.NewNodeSet(), nil)
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mockDB.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
			},
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusOK)
				// List results are nested under "data" and the response contains other pagination properties
				apitest.BodyContains(output, "data")
				apitest.BodyContains(output, "skip")
				apitest.BodyContains(output, "limit")
				apitest.BodyContains(output, "count")
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

func TestResources_ListRootCAPKIHierarchy(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListRootCAPKIHierarchy).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListCAPKIHierarchy(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListCAPKIHierarchy).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListPublishedTemplates(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListPublishedTemplates).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListPublishedToCAs(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListPublishedToCAs).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListTrustedCAs(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListTrustedCAs).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADCSEscalations(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADCSEscalations).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADSiteLinkedGPOs(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADSiteLinkedGPOs).
		Run(setupCases(mockGraph, mockDB))
}

func TestResources_ListADIssuancePolicyLinkedCertTemplates(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraphQuery *mocks.MockGraph
		mockDatabase   *dbMocks.MockDatabase
	}
	type expected struct {
		responseBody   string
		responseCode   int
		responseHeader http.Header
	}
	type testData struct {
		name             string
		buildRequest     func() *http.Request
		setupMocks       func(t *testing.T, mock *mock)
		user             model.User
		dogTagsOverrides dogtags.TestOverrides
		expected         expected
	}

	tt := []testData{
		// Missing path parameters cannot be tested due to Gorilla Mux's strict route matching, which requires all defined path parameters to be present in the request URL for the route to match.
		{
			name: "Error: database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						RawQuery: "type=graph",
						Path:     "/api/v2/issuancepolicies/id/linkedtemplates",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{}, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an internal error has occurred that is preventing the service from servicing this request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: grapy query error queries.ErrGraphUnsupported - Bad Request",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						RawQuery: "type=graph",
						Path:     "/api/v2/issuancepolicies/id/linkedtemplates",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return(nil, queries.ErrGraphUnsupported)
			},
			expected: expected{
				responseCode:   http.StatusBadRequest,
				responseBody:   `{"errors":[{"context":"","message":"there are errors in the query parameters: type 'graph' is not supported for this endpoint"}],"http_status":400,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: graph query error op.ErrGraphQueryMemoryLimit - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						RawQuery: "type=graph",
						Path:     "/api/v2/issuancepolicies/id/linkedtemplates",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return(nil, ops.ErrGraphQueryMemoryLimit)
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"calculating the request results exceeded memory limitations due to the volume of objects involved"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Error: graph query error undefined error type - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						RawQuery: "type=graph",
						Path:     "/api/v2/issuancepolicies/id/linkedtemplates",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return(nil, errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"an unknown error occurred during the request"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: Data type graph query without pagination - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						RawQuery: "type=graph",
						Path:     "/api/v2/issuancepolicies/id/linkedtemplates",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return(graph.PathSet{}, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: Data type list query with pagination - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/issuancepolicies/id/linkedtemplates",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return(graph.NewNodeSet(adRelatedEntityTestNode(1, "related-id", "RELATED", "12345")), nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"count":1,"limit":10,"skip":0,"data":[{"objectID":"related-id","name":"RELATED","label":"User","kinds":["Base","User"]}]}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success: ETAC enabled AllEnvironments",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/issuancepolicies/id/linkedtemplates",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return(graph.NewNodeSet(adRelatedEntityTestNode(1, "related-id", "RELATED", "12345")), nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"count":1,"limit":10,"skip":0,"data":[{"objectID":"related-id","name":"RELATED","label":"User","kinds":["Base","User"]}]}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: true,
			},
		},
		{
			name: "Success: ETAC enabled For Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/issuancepolicies/id/linkedtemplates",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.Entity).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{ad.Entity},
					Properties: graph.AsProperties(map[string]any{
						"domainsid": "12345",
					}),
				}, nil)
				mock.mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return(graph.NewNodeSet(adRelatedEntityTestNode(1, "related-id", "RELATED", "12345")), nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"count":1,"limit":10,"skip":0,"data":[{"objectID":"related-id","name":"RELATED","label":"User","kinds":["Base","User"]}]}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "12345",
					},
				},
			},
		},
		{
			name: "Error: ETAC User Does Not have Access To Specific Environment",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/issuancepolicies/id/linkedtemplates",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "id", ad.Entity).Return(&graph.Node{
					ID:    graph.ID(16),
					Kinds: graph.Kinds{ad.Entity},
					Properties: graph.AsProperties(map[string]any{
						"domainsid": "12345",
					}),
				}, nil)
			},
			expected: expected{
				responseCode:   http.StatusForbidden,
				responseBody:   `{"errors":[{"context":"","message":"Forbidden"}],"http_status":403,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
			dogTagsOverrides: dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{
					dogtags.ETAC_ENABLED: true,
				},
			},
			user: model.User{
				AllEnvironments: false,
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{
						EnvironmentID: "54321",
					},
				},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockDatabase:   dbMocks.NewMockDatabase(ctrl),
				mockGraphQuery: mocks.NewMockGraph(ctrl),
			}

			request := testCase.buildRequest()
			bheCtx := bhctx.Context{
				AuthCtx: auth.Context{
					PermissionOverrides: auth.PermissionOverrides{},
					Owner:               testCase.user,
					Session:             model.UserSession{},
				},
			}
			requestWithCtx := request.WithContext(bheCtx.ConstructGoContext())

			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				DB:         mocks.mockDatabase,
				GraphQuery: mocks.mockGraphQuery,
				DogTags:    dogtags.NewTestService(testCase.dogTagsOverrides),
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/issuancepolicies/{%s}/linkedtemplates", api.URIPathVariableObjectID), resources.ListADIssuancePolicyLinkedCertTemplates).Methods(requestWithCtx.Method)
			router.ServeHTTP(response, requestWithCtx)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}

func adRelatedEntityTestNode(nodeID graph.ID, objectID, name, domainSID string) *graph.Node {
	var properties = graph.AsProperties(map[string]any{
		"objectid": objectID,
		"name":     name,
	})

	if domainSID != "" {
		properties.Set(ad.DomainSID.String(), domainSID)
	}

	return graph.NewNode(nodeID, properties, ad.Entity, ad.User)
}

func adRelatedEntityTestRequest(user model.User, rawQuery string) *http.Request {
	var (
		request        = httptest.NewRequest(http.MethodGet, "/api/v2/users/anchor-id/adminrights?"+rawQuery, nil)
		requestContext = bhctx.Context{
			AuthCtx: auth.Context{Owner: user},
		}
	)

	return request.WithContext(requestContext.ConstructGoContext())
}

func TestResources_ListADEntityAdminRights_ETACListPagination(t *testing.T) {
	t.Parallel()

	var (
		allowedUser = model.User{
			EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
				{EnvironmentID: "allowed-domain"},
			},
		}
		anchorNode    = adRelatedEntityTestNode(10, "anchor-id", "ANCHOR", "allowed-domain")
		firstAllowed  = adRelatedEntityTestNode(4, "allowed-first", "ALLOWED FIRST", "allowed-domain")
		foreignNode   = adRelatedEntityTestNode(3, "foreign-secret-id", "FOREIGN SECRET NAME", "foreign-domain")
		lastAllowed   = adRelatedEntityTestNode(2, "allowed-last", "ALLOWED LAST", "allowed-domain")
		missingDomain = adRelatedEntityTestNode(1, "missing-domain-secret-id", "MISSING DOMAIN SECRET NAME", "")
		mixedResults  = graph.NewNodeSet(firstAllowed, foreignNode, lastAllowed, missingDomain)
		deniedResults = graph.NewNodeSet(foreignNode, missingDomain)
		testCases     = []struct {
			name            string
			user            model.User
			etacEnabled     bool
			cacheEnabled    bool
			skip            int
			limit           int
			results         graph.NodeSet
			expectedStatus  int
			expectedCount   int
			expectedNodeIDs []string
		}{
			{
				name: "filter before first page fills the page with authorized nodes",
				user: allowedUser, etacEnabled: true, cacheEnabled: true,
				limit: 2, results: mixedResults,
				expectedStatus: http.StatusOK, expectedCount: 2,
				expectedNodeIDs: []string{"allowed-first", "allowed-last"},
			},
			{
				name: "skip applies to authorized ordering with caching disabled",
				user: allowedUser, etacEnabled: true,
				skip: 1, limit: 1, results: mixedResults,
				expectedStatus: http.StatusOK, expectedCount: 2,
				expectedNodeIDs: []string{"allowed-last"},
			},
			{
				name: "offset at authorized count returns an empty page",
				user: allowedUser, etacEnabled: true, cacheEnabled: true,
				skip: 2, limit: 2, results: mixedResults,
				expectedStatus: http.StatusOK, expectedCount: 2, expectedNodeIDs: []string{},
			},
			{
				name: "zero limit retains authorized count",
				user: allowedUser, etacEnabled: true, cacheEnabled: true,
				limit: 0, results: mixedResults,
				expectedStatus: http.StatusOK, expectedCount: 2, expectedNodeIDs: []string{},
			},
			{
				name: "offset beyond authorized count is rejected",
				user: allowedUser, etacEnabled: true, cacheEnabled: true,
				skip: 3, limit: 2, results: mixedResults,
				expectedStatus: http.StatusBadRequest,
			},
			{
				name: "foreign and missing domain identities leave no authorized results",
				user: allowedUser, etacEnabled: true, cacheEnabled: true,
				limit: 2, results: deniedResults,
				expectedStatus: http.StatusOK, expectedCount: 0, expectedNodeIDs: []string{},
			},
			{
				name: "disabled ETAC preserves all related results",
				user: allowedUser, cacheEnabled: true,
				limit: 10, results: mixedResults,
				expectedStatus: http.StatusOK, expectedCount: 4,
				expectedNodeIDs: []string{"allowed-first", "foreign-secret-id", "allowed-last", "missing-domain-secret-id"},
			},
			{
				name: "all environments preserves all related results",
				user: model.User{AllEnvironments: true}, etacEnabled: true, cacheEnabled: true,
				limit: 10, results: mixedResults,
				expectedStatus: http.StatusOK, expectedCount: 4,
				expectedNodeIDs: []string{"allowed-first", "foreign-secret-id", "allowed-last", "missing-domain-secret-id"},
			},
		}
	)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				controller      = gomock.NewController(t)
				mockDatabase    = dbMocks.NewMockDatabase(controller)
				mockGraphQuery  = mocks.NewMockGraph(controller)
				response        = httptest.NewRecorder()
				request         = adRelatedEntityTestRequest(testCase.user, fmt.Sprintf("type=list&skip=%d&limit=%d", testCase.skip, testCase.limit))
				router          = mux.NewRouter()
				expectedNodeIDs = testCase.expectedNodeIDs
				actualNodeIDs   = make([]string, 0)
				responseBody    struct {
					Count int                        `json:"count"`
					Skip  int                        `json:"skip"`
					Limit int                        `json:"limit"`
					Data  []model.PagedNodeListEntry `json:"data"`
				}
				resources = v2.Resources{
					DB: mockDatabase, GraphQuery: mockGraphQuery,
					DogTags: dogtags.NewTestService(dogtags.TestOverrides{
						Bools: map[dogtags.BoolDogTag]bool{dogtags.ETAC_ENABLED: testCase.etacEnabled},
					}),
				}
			)

			if testCase.etacEnabled && !testCase.user.AllEnvironments {
				mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "anchor-id", ad.Entity).Return(anchorNode, nil)
			}
			mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureEntityPanelCaching).Return(appcfg.FeatureFlag{Enabled: testCase.cacheEnabled}, nil)
			mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
			mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), testCase.cacheEnabled).Return(testCase.results, nil)

			router.HandleFunc("/api/v2/users/{object_id}/adminrights", resources.ListADEntityAdminRights)
			router.ServeHTTP(response, request)

			require.Equal(t, testCase.expectedStatus, response.Code, response.Body.String())
			if testCase.expectedStatus != http.StatusOK {
				return
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &responseBody))
			assert.Equal(t, testCase.expectedCount, responseBody.Count)
			assert.Equal(t, testCase.skip, responseBody.Skip)
			assert.Equal(t, testCase.limit, responseBody.Limit)
			assert.NotNil(t, responseBody.Data, "empty results retain the list response shape")
			for _, entry := range responseBody.Data {
				actualNodeIDs = append(actualNodeIDs, entry.ObjectID)
			}
			assert.Equal(t, expectedNodeIDs, actualNodeIDs)
			if testCase.etacEnabled && !testCase.user.AllEnvironments {
				assert.NotContains(t, response.Body.String(), "foreign-secret-id")
				assert.NotContains(t, response.Body.String(), "FOREIGN SECRET NAME")
				assert.NotContains(t, response.Body.String(), "missing-domain-secret-id")
				assert.NotContains(t, response.Body.String(), "MISSING DOMAIN SECRET NAME")
			}
		})
	}
}

func TestResources_ListADEntityAdminRights_ETACSharedRawResults(t *testing.T) {
	t.Parallel()

	var (
		controller     = gomock.NewController(t)
		mockDatabase   = dbMocks.NewMockDatabase(controller)
		mockGraphQuery = mocks.NewMockGraph(controller)
		anchorNode     = adRelatedEntityTestNode(10, "anchor-id", "ANCHOR", "anchor-domain")
		firstNode      = adRelatedEntityTestNode(2, "first-domain-object", "FIRST DOMAIN", "first-domain")
		secondNode     = adRelatedEntityTestNode(1, "second-domain-object", "SECOND DOMAIN", "second-domain")
		sharedResults  = graph.NewNodeSet(firstNode, secondNode)
		resources      = v2.Resources{
			DB: mockDatabase, GraphQuery: mockGraphQuery,
			DogTags: dogtags.NewTestService(dogtags.TestOverrides{
				Bools: map[dogtags.BoolDogTag]bool{dogtags.ETAC_ENABLED: true},
			}),
		}
		testCases = []struct {
			domainSID        string
			expectedObjectID string
		}{
			{domainSID: "first-domain", expectedObjectID: "first-domain-object"},
			{domainSID: "second-domain", expectedObjectID: "second-domain-object"},
		}
	)

	mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "anchor-id", ad.Entity).Return(anchorNode, nil).Times(2)
	mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureEntityPanelCaching).Return(appcfg.FeatureFlag{Enabled: true}, nil).Times(2)
	mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any()).Times(2)
	mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return(sharedResults, nil).Times(2)

	for _, testCase := range testCases {
		var (
			user = model.User{
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{EnvironmentID: "anchor-domain"},
					{EnvironmentID: testCase.domainSID},
				},
			}
			request      = adRelatedEntityTestRequest(user, "type=list&limit=10")
			response     = httptest.NewRecorder()
			router       = mux.NewRouter()
			responseBody struct {
				Count int                        `json:"count"`
				Data  []model.PagedNodeListEntry `json:"data"`
			}
		)

		router.HandleFunc("/api/v2/users/{object_id}/adminrights", resources.ListADEntityAdminRights)
		router.ServeHTTP(response, request)

		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &responseBody))
		assert.Equal(t, 1, responseBody.Count)
		require.Len(t, responseBody.Data, 1)
		assert.Equal(t, testCase.expectedObjectID, responseBody.Data[0].ObjectID)
	}

	assert.Len(t, sharedResults, 2, "filtering one response must preserve the raw set for another user")
	assert.Contains(t, sharedResults, firstNode.ID)
	assert.Contains(t, sharedResults, secondNode.ID)
}

func TestResources_ListADEntityAdminRights_ETACGraph(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		name           string
		etacEnabled    bool
		user           model.User
		expectRedacted bool
	}{
		{
			name:        "restricted user receives hidden placeholders for foreign objects and edges",
			etacEnabled: true, expectRedacted: true,
			user: model.User{
				EnvironmentTargetedAccessControl: []model.EnvironmentTargetedAccessControl{
					{EnvironmentID: "allowed-domain"},
				},
			},
		},
		{
			name: "disabled ETAC retains full graph", user: model.User{},
		},
		{
			name: "all environments retains full graph", etacEnabled: true,
			user: model.User{AllEnvironments: true},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				controller     = gomock.NewController(t)
				mockDatabase   = dbMocks.NewMockDatabase(controller)
				mockGraphQuery = mocks.NewMockGraph(controller)
				anchorNode     = adRelatedEntityTestNode(10, "anchor-id", "ANCHOR", "allowed-domain")
				allowedNode    = adRelatedEntityTestNode(11, "allowed-object", "ALLOWED NAME", "allowed-domain")
				foreignNode    = adRelatedEntityTestNode(12, "foreign-secret-id", "FOREIGN SECRET NAME", "foreign-secret-domain")
				allowedEdge    = &graph.Relationship{
					ID: 31, StartID: anchorNode.ID, EndID: allowedNode.ID, Kind: ad.AdminTo,
					Properties: graph.AsProperties(map[string]any{"allowed_edge_property": "ALLOWED EDGE VALUE"}),
				}
				foreignEdge = &graph.Relationship{
					ID: 32, StartID: allowedNode.ID, EndID: foreignNode.ID, Kind: ad.GenericAll,
					Properties: graph.AsProperties(map[string]any{"foreign_secret_edge_property": "FOREIGN SECRET EDGE VALUE"}),
				}
				paths = graph.NewPathSet(graph.Path{
					Nodes: []*graph.Node{anchorNode, allowedNode, foreignNode},
					Edges: []*graph.Relationship{allowedEdge, foreignEdge},
				})
				request   = adRelatedEntityTestRequest(testCase.user, "type=graph")
				response  = httptest.NewRecorder()
				router    = mux.NewRouter()
				resources = v2.Resources{
					DB: mockDatabase, GraphQuery: mockGraphQuery,
					DogTags: dogtags.NewTestService(dogtags.TestOverrides{
						Bools: map[dogtags.BoolDogTag]bool{dogtags.ETAC_ENABLED: testCase.etacEnabled},
					}),
				}
				graphResponse map[string]json.RawMessage
			)

			if testCase.expectRedacted {
				mockGraphQuery.EXPECT().GetEntityByObjectId(gomock.Any(), "anchor-id", ad.Entity).Return(anchorNode, nil)
			}
			mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureEntityPanelCaching).Return(appcfg.FeatureFlag{Enabled: true}, nil)
			mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
			mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return(paths, nil)

			router.HandleFunc("/api/v2/users/{object_id}/adminrights", resources.ListADEntityAdminRights)
			router.ServeHTTP(response, request)

			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &graphResponse))
			assert.Len(t, graphResponse, 5, "graph keeps its flat node/edge response shape")
			assert.Contains(t, response.Body.String(), "allowed-object")
			assert.Contains(t, response.Body.String(), "ALLOWED NAME")
			assert.Contains(t, response.Body.String(), "ALLOWED EDGE VALUE")
			if testCase.expectRedacted {
				assert.Contains(t, string(graphResponse["12"]), "** Hidden Object **")
				assert.Contains(t, string(graphResponse["rel_32"]), "** Hidden Edge **")
				assert.NotContains(t, response.Body.String(), "foreign-secret-id")
				assert.NotContains(t, response.Body.String(), "FOREIGN SECRET NAME")
				assert.NotContains(t, response.Body.String(), "foreign-secret-domain")
				assert.NotContains(t, response.Body.String(), "foreign_secret_edge_property")
				assert.NotContains(t, response.Body.String(), "FOREIGN SECRET EDGE VALUE")
				assert.NotContains(t, string(graphResponse["rel_32"]), "GenericAll")
			} else {
				assert.Contains(t, response.Body.String(), "foreign-secret-id")
				assert.Contains(t, response.Body.String(), "FOREIGN SECRET NAME")
				assert.Contains(t, response.Body.String(), "foreign-secret-domain")
				assert.Contains(t, response.Body.String(), "FOREIGN SECRET EDGE VALUE")
				assert.NotContains(t, response.Body.String(), "** Hidden Object **")
				assert.NotContains(t, response.Body.String(), "** Hidden Edge **")
			}
		})
	}
}

func TestResources_ListADEntityAdminRights_PaginationValidation(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		name          string
		query         string
		expectedError string
	}{
		{name: "negative skip", query: "skip=-1", expectedError: "invalid skip: -1"},
		{name: "malformed skip", query: "skip=invalid", expectedError: "invalid syntax"},
		{name: "negative limit", query: "limit=-1", expectedError: "invalid limit: -1"},
		{name: "malformed limit", query: "limit=invalid", expectedError: "invalid syntax"},
	}

	for _, responseType := range []string{"list", "graph"} {
		for _, testCase := range testCases {
			t.Run(responseType+"/"+testCase.name, func(t *testing.T) {
				t.Parallel()

				var (
					controller     = gomock.NewController(t)
					mockDatabase   = dbMocks.NewMockDatabase(controller)
					mockGraphQuery = mocks.NewMockGraph(controller)
					request        = adRelatedEntityTestRequest(model.User{}, "type="+responseType+"&"+testCase.query)
					response       = httptest.NewRecorder()
					router         = mux.NewRouter()
					resources      = v2.Resources{
						DB: mockDatabase, GraphQuery: mockGraphQuery,
						DogTags: dogtags.NewTestService(dogtags.TestOverrides{
							Bools: map[dogtags.BoolDogTag]bool{dogtags.ETAC_ENABLED: true},
						}),
					}
				)

				// No database or graph calls are expected before rejecting invalid pagination.
				router.HandleFunc("/api/v2/users/{object_id}/adminrights", resources.ListADEntityAdminRights)
				router.ServeHTTP(response, request)

				require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
				assert.Contains(t, response.Body.String(), "error getting paging parameters")
				assert.Contains(t, response.Body.String(), testCase.expectedError)
			})
		}
	}
}

func TestResources_ListADEntityAdminRights_DefaultPagination(t *testing.T) {
	t.Parallel()

	var (
		controller     = gomock.NewController(t)
		mockDatabase   = dbMocks.NewMockDatabase(controller)
		mockGraphQuery = mocks.NewMockGraph(controller)
		rawResults     = graph.NewNodeSet()
		request        = adRelatedEntityTestRequest(model.User{AllEnvironments: true}, "")
		response       = httptest.NewRecorder()
		router         = mux.NewRouter()
		resources      = v2.Resources{
			DB: mockDatabase, GraphQuery: mockGraphQuery,
			DogTags: dogtags.NewTestService(dogtags.TestOverrides{}),
		}
		responseBody struct {
			Count int                        `json:"count"`
			Skip  int                        `json:"skip"`
			Limit int                        `json:"limit"`
			Data  []model.PagedNodeListEntry `json:"data"`
		}
	)

	for nodeID := graph.ID(1); nodeID <= 12; nodeID++ {
		rawResults.Add(adRelatedEntityTestNode(nodeID, fmt.Sprintf("object-%d", nodeID), "RELATED", "allowed-domain"))
	}

	mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureEntityPanelCaching).Return(appcfg.FeatureFlag{Enabled: true}, nil)
	mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
	mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).DoAndReturn(
		func(_ context.Context, params queries.EntityQueryParameters, _ bool) (any, error) {
			assert.Equal(t, model.DataTypeList, params.RequestedType)
			return rawResults, nil
		})

	router.HandleFunc("/api/v2/users/{object_id}/adminrights", resources.ListADEntityAdminRights)
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &responseBody))
	assert.Equal(t, 12, responseBody.Count)
	assert.Zero(t, responseBody.Skip)
	assert.Equal(t, 10, responseBody.Limit)
	require.Len(t, responseBody.Data, 10)
	assert.Equal(t, "object-12", responseBody.Data[0].ObjectID)
	assert.Equal(t, "object-3", responseBody.Data[9].ObjectID)
}

func TestResources_ListADEntityAdminRights_CountUnsupported(t *testing.T) {
	t.Parallel()

	var (
		controller     = gomock.NewController(t)
		mockDatabase   = dbMocks.NewMockDatabase(controller)
		mockGraphQuery = mocks.NewMockGraph(controller)
		request        = adRelatedEntityTestRequest(model.User{AllEnvironments: true}, "type=count")
		response       = httptest.NewRecorder()
		router         = mux.NewRouter()
		resources      = v2.Resources{
			DB: mockDatabase, GraphQuery: mockGraphQuery,
			DogTags: dogtags.NewTestService(dogtags.TestOverrides{}),
		}
	)

	mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), appcfg.FeatureEntityPanelCaching).Return(appcfg.FeatureFlag{Enabled: true}, nil)
	mockDatabase.EXPECT().GetPrimaryDisplayKinds(gomock.Any())
	mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).DoAndReturn(
		func(_ context.Context, params queries.EntityQueryParameters, _ bool) (any, error) {
			assert.Equal(t, model.DataTypeCount, params.RequestedType)
			return nil, queries.ErrUnsupportedDataType
		})

	router.HandleFunc("/api/v2/users/{object_id}/adminrights", resources.ListADEntityAdminRights)
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	assert.Contains(t, response.Body.String(), queries.ErrUnsupportedDataType.Error())
}
