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
	dbMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/queries/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	"github.com/specterops/dawgs/ops"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setup(t *testing.T) (*gomock.Controller, *mocks.MockGraph, *dbMocks.MockDatabase, v2.Resources) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockGraph = mocks.NewMockGraph(mockCtrl)
		mockDB    = dbMocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{GraphQuery: mockGraph, DB: mockDB}
	)
	return mockCtrl, mockGraph, mockDB, resources
}

func setupCases(mockGraph *mocks.MockGraph, mockDB *dbMocks.MockDatabase) []apitest.Case {
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
					Return(nil, 0, queries.ErrGraphUnsupported)
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
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
					Return(nil, 0, queries.ErrUnsupportedDataType)
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
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
					Return(nil, 0, ops.ErrGraphQueryMemoryLimit)
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
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
					Return(nil, 0, errors.New("any other error"))
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
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
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, 0, nil)
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
			},
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusOK)
				//This flat unnested shape maintains the current api contract for a type=graph query
				//Assert that the response does not contain pagination properties
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
				//Delete the type=graph param so that we get list results
				apitest.DeleteQueryParam(input, "type")
			},
			Setup: func() {
				mockGraph.EXPECT().
					GetADEntityQueryResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, 0, nil)
				mockDB.EXPECT().
					GetFlagByKey(gomock.Any(), "entity_panel_cache").
					Return(appcfg.FeatureFlag{Enabled: true}, nil)
			},
			Test: func(output apitest.Output) {
				apitest.StatusCode(output, http.StatusOK)
				//List results are nested under "data" and the response contains other pagination properties
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

func TestResources_ListADComputerBackupOperators(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADComputerBackupOperators).
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

func TestResources_ListADEntityBackupRights(t *testing.T) {
	var mockCtrl, mockGraph, mockDB, resources = setup(t)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.ListADEntityBackupRights).
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
		name         string
		buildRequest func() *http.Request
		setupMocks   func(t *testing.T, mock *mock)
		expected     expected
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
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return("", 0, queries.ErrGraphUnsupported)
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
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return("", 0, ops.ErrGraphQueryMemoryLimit)
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
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return("", 0, errors.New("error"))
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
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return("results", 1, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `"results"`,
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
				mock.mockDatabase.EXPECT().GetFlagByKey(gomock.Any(), "entity_panel_cache").Return(appcfg.FeatureFlag{Enabled: true}, nil)
				mock.mockGraphQuery.EXPECT().GetADEntityQueryResult(gomock.Any(), gomock.Any(), true).Return("", 1, nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"count":1,"limit":10,"skip":0,"data":""}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
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
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				DB:         mocks.mockDatabase,
				GraphQuery: mocks.mockGraphQuery,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc(fmt.Sprintf("/api/v2/issuancepolicies/{%s}/linkedtemplates", api.URIPathVariableObjectID), resources.ListADIssuancePolicyLinkedCertTemplates).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
