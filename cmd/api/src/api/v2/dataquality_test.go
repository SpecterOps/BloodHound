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
	"reflect"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/cmd/api/src/utils/test"
	graphmocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetADDataQualityStats_Failure(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	domainSID := "S-1-5-21-3130019616-2776909439-2417379446"
	endpoint := "/api/v2/ad-domains/%s/data-quality-stats%s"

	mockDB := mocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetADDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, fmt.Errorf("db error"))

	resources := v2.Resources{DB: mockDB}

	type Input struct {
		DomainSid string
		Params    url.Values
	}

	var cases = []struct {
		Input    Input
		Expected api.ErrorWrapper
	}{
		{
			Input{
				domainSID,
				url.Values{
					"sort_by": []string{"invalidColumn"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsNotSortable}},
			},
		},
		{
			Input{
				domainSID,
				url.Values{
					"start": []string{"invalidRFC3339"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(api.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
			},
		},
		{
			Input{
				domainSID,
				url.Values{
					"end": []string{"invalidRFC3339"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(api.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
			},
		},
		{
			Input{
				domainSID,
				url.Values{
					"limit": []string{"invalidLimit"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(utils.ErrorInvalidLimit, []string{"invalidLimit"})}},
			},
		},
		{
			Input{
				domainSID,
				url.Values{
					"skip": []string{"invalidSkip"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(utils.ErrorInvalidSkip, []string{"invalidSkip"})}},
			},
		},
		{
			Input{
				"dbError",
				url.Values{},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusInternalServerError,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsInternalServerError}},
			},
		},
	}

	for _, tc := range cases {
		params := fmt.Sprintf("?%s", tc.Input.Params.Encode())
		if req, err := http.NewRequest("GET", fmt.Sprintf(endpoint, tc.Input.DomainSid, params), nil); err != nil {
			t.Fatal(err)
		} else {
			router := mux.NewRouter()
			router.HandleFunc("/api/v2/ad-domains/{domain_id}/data-quality-stats", resources.GetADDataQualityStats).Methods("GET")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.Expected.HTTPStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected.HTTPStatus)
			}

			var body any
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatal("failed to unmarshal response body")
			}

			require.Equal(t, len(tc.Expected.Errors), 1)
			if !reflect.DeepEqual(body.(map[string]any)["errors"].([]any)[0].(map[string]any)["message"], tc.Expected.Errors[0].Message) {
				t.Errorf("For input: %v, got %v, want %v", tc.Input, body, tc.Expected.Errors[0].Message)
			}
		}
	}
}

func TestGetADDataQualityStats_Success(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	domainSID := "S-1-5-21-3130019616-2776909439-2417379446"
	endpoint := "/api/v2/ad-domains/%s/data-quality-stats%s"

	mockDB := mocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetADDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.ADDataQualityStats{}, 0, nil).AnyTimes()

	resources := v2.Resources{DB: mockDB}

	type Input struct {
		DomainSid string
		Params    url.Values
	}

	type Expected struct {
		Code int
	}

	var cases = []struct {
		Input    Input
		Expected Expected
	}{
		{
			Input{
				domainSID,
				url.Values{
					"sort_by": []string{"-updated_at"},
					"limit":   []string{"1"},
					"start":   []string{"2022-03-23T07:20:50.52Z"},
					"end":     []string{"2022-04-23T07:20:50.52Z"},
					"skip":    []string{"0"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				domainSID,
				url.Values{
					"sort_by": []string{"updated_at"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				domainSID,
				url.Values{
					"start": []string{"2022-03-23T07:20:50.52Z"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				domainSID,
				url.Values{
					"end": []string{"2022-04-23T07:20:50.52Z"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				domainSID,
				url.Values{
					"limit": []string{"2"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				domainSID,
				url.Values{
					"skip": []string{"1"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		params := fmt.Sprintf("?%s", tc.Input.Params.Encode())
		if req, err := http.NewRequest("GET", fmt.Sprintf(endpoint, tc.Input.DomainSid, params), nil); err != nil {
			t.Fatal(err)
		} else {
			router := mux.NewRouter()
			router.HandleFunc("/api/v2/ad-domains/{domain_id}/data-quality-stats", resources.GetADDataQualityStats).Methods("GET")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.Expected.Code {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected.Code)
			}
		}
	}
}

func TestGetAzureDataQualityStats_Failure(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tenantID := "7ac5533e-9881-4e0f-b51c-000000000000"
	endpoint := "/api/v2/azure-tenants/%s/data-quality-stats%s"

	mockDB := mocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAzureDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, fmt.Errorf("db error"))

	resources := v2.Resources{DB: mockDB}

	type Input struct {
		TenantID string
		Params   url.Values
	}

	var cases = []struct {
		Input    Input
		Expected api.ErrorWrapper
	}{
		{
			Input{
				tenantID,
				url.Values{
					"sort_by": []string{"invalidColumn"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsNotSortable}},
			},
		},
		{
			Input{
				tenantID,
				url.Values{
					"start": []string{"invalidRFC3339"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(api.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
			},
		},
		{
			Input{
				tenantID,
				url.Values{
					"end": []string{"invalidRFC3339"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(api.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
			},
		},
		{
			Input{
				tenantID,
				url.Values{
					"limit": []string{"invalidLimit"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(utils.ErrorInvalidLimit, []string{"invalidLimit"})}},
			},
		},
		{
			Input{
				tenantID,
				url.Values{
					"skip": []string{"invalidSkip"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(utils.ErrorInvalidSkip, []string{"invalidSkip"})}},
			},
		},
		{
			Input{
				"dbError",
				url.Values{},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusInternalServerError,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsInternalServerError}},
			},
		},
	}

	for _, tc := range cases {
		params := fmt.Sprintf("?%s", tc.Input.Params.Encode())
		if req, err := http.NewRequest("GET", fmt.Sprintf(endpoint, tc.Input.TenantID, params), nil); err != nil {
			t.Fatal(err)
		} else {
			router := mux.NewRouter()
			router.HandleFunc("/api/v2/azure-tenants/{tenant_id}/data-quality-stats", resources.GetAzureDataQualityStats).Methods("GET")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.Expected.HTTPStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected.HTTPStatus)
			}

			var body any
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatal("failed to unmarshal response body")
			}

			require.Equal(t, len(tc.Expected.Errors), 1)
			if !reflect.DeepEqual(body.(map[string]any)["errors"].([]any)[0].(map[string]any)["message"], tc.Expected.Errors[0].Message) {
				t.Errorf("For input: %v, got %v, want %v", tc.Input, body, tc.Expected.Errors[0].Message)
			}
		}
	}
}

func TestGetAzureDataQualityStats_Success(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tenantID := "S-1-5-21-3130019616-2776909439-2417379446"
	endpoint := "/api/v2/azure-tenants/%s/data-quality-stats%s"

	mockDB := mocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetAzureDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.AzureDataQualityStats{}, 0, nil).AnyTimes()

	resources := v2.Resources{DB: mockDB}

	type Input struct {
		TenantID string
		Params   url.Values
	}

	type Expected struct {
		Code int
	}

	var cases = []struct {
		Input    Input
		Expected Expected
	}{
		{
			Input{
				tenantID,
				url.Values{
					"sort_by": []string{"-updated_at"},
					"limit":   []string{"1"},
					"start":   []string{"2022-03-23T07:20:50.52Z"},
					"end":     []string{"2022-04-23T07:20:50.52Z"},
					"skip":    []string{"0"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				tenantID,
				url.Values{
					"sort_by": []string{"updated_at"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				tenantID,
				url.Values{
					"start": []string{"2022-03-23T07:20:50.52Z"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				tenantID,
				url.Values{
					"end": []string{"2022-04-23T07:20:50.52Z"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				tenantID,
				url.Values{
					"limit": []string{"2"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				tenantID,
				url.Values{
					"skip": []string{"1"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		params := fmt.Sprintf("?%s", tc.Input.Params.Encode())
		if req, err := http.NewRequest("GET", fmt.Sprintf(endpoint, tc.Input.TenantID, params), nil); err != nil {
			t.Fatal(err)
		} else {
			router := mux.NewRouter()
			router.HandleFunc("/api/v2/azure-tenants/{tenant_id}/data-quality-stats", resources.GetAzureDataQualityStats).Methods("GET")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.Expected.Code {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected.Code)
			}
		}
	}
}

func TestGetDataQualityStats_Failure(t *testing.T) {
	environmentID := "environment-1"
	endpoint := "/api/v2/data-quality-stats%s"
	user := setupUser()
	userCtx := setupUserCtx(user)
	etacError := errors.New("etac error")

	type Input struct {
		Params  url.Values
		Context context.Context
	}

	type mockResources struct {
		Database *mocks.MockDatabase
	}

	defaultResources := func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources {
		return v2.Resources{
			DB:      mock.Database,
			DogTags: dogtags.NewTestService(dogtags.TestOverrides{}),
		}
	}

	var cases = []struct {
		Name     string
		Input    Input
		Setup    func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources
		Expected api.ErrorWrapper
	}{
		{
			Name: "NoEnvironmentID",
			Input: Input{
				Params:  url.Values{},
				Context: userCtx,
			},
			Setup: defaultResources,
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: v2.ErrNoEnvironmentId}},
			},
		},
		{
			Name: "UnknownUser",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
				},
				Context: context.Background(),
			},
			Setup: defaultResources,
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusInternalServerError,
				Errors:     []api.ErrorDetails{{Message: v2.ErrUnknownUser}},
			},
		},
		{
			Name: "CheckUserETACAccessError",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
				},
				Context: userCtx,
			},
			Setup: func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources {
				mock.Database.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), user).Return(nil, etacError)

				return v2.Resources{
					DB: mock.Database,
					DogTags: dogtags.NewTestService(dogtags.TestOverrides{
						Bools: map[dogtags.BoolDogTag]bool{
							dogtags.ETAC_ENABLED: true,
						},
					}),
				}
			},
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusInternalServerError,
				Errors:     []api.ErrorDetails{{Message: etacError.Error()}},
			},
		},
		{
			Name: "NoEnvironmentAccess",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
				},
				Context: userCtx,
			},
			Setup: func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources {
				mock.Database.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), user).Return([]model.EnvironmentTargetedAccessControl{
					{EnvironmentID: "other-environment"},
				}, nil)

				return v2.Resources{
					DB: mock.Database,
					DogTags: dogtags.NewTestService(dogtags.TestOverrides{
						Bools: map[dogtags.BoolDogTag]bool{
							dogtags.ETAC_ENABLED: true,
						},
					}),
				}
			},
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusForbidden,
				Errors:     []api.ErrorDetails{{Message: v2.ErrNoAccess}},
			},
		},
		{
			Name: "EnvironmentIDNotFound",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
				},
				Context: userCtx,
			},
			Setup: func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources {
				mock.Database.EXPECT().GetDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.DataQualityStats{}, 0, database.ErrNotFound)

				return defaultResources(mockCtrl, mock)
			},
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusNotFound,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsResourceNotFound}},
			},
		},
		{
			Name: "NonSortableColumn",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"sort_by":                       []string{"metric_name"},
				},
				Context: userCtx,
			},
			Setup: func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources {
				return defaultResources(mockCtrl, mock)
			},
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsNotSortable}},
			},
		},
		{
			Name: "EmptySortByParameter",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"sort_by":                       []string{""},
				},
				Context: userCtx,
			},
			Setup: func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources {
				return defaultResources(mockCtrl, mock)
			},
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseEmptySortParameter}},
			},
		},
		{
			Name: "InvalidStartTime",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"start":                         []string{"invalidRFC3339"},
				},
				Context: userCtx,
			},
			Setup: func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources {
				return defaultResources(mockCtrl, mock)
			},
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(api.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
			},
		},
		{
			Name: "InvalidEndTime",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"end":                           []string{"invalidRFC3339"},
				},
				Context: userCtx,
			},
			Setup: func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources {
				return defaultResources(mockCtrl, mock)
			},
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(api.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
			},
		},
		{
			Name: "InvalidLimit",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"limit":                         []string{"invalidLimit"},
				},
				Context: userCtx,
			},
			Setup: func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources {
				return defaultResources(mockCtrl, mock)
			},
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(utils.ErrorInvalidLimit, []string{"invalidLimit"})}},
			},
		},
		{
			Name: "InvalidSkip",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"skip":                          []string{"invalidSkip"},
				},
				Context: userCtx,
			},
			Setup: func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources {
				return defaultResources(mockCtrl, mock)
			},
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(utils.ErrorInvalidSkip, []string{"invalidSkip"})}},
			},
		},
		{
			Name: "DatabaseError",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
				},
				Context: userCtx,
			},
			Setup: func(mockCtrl *gomock.Controller, mock mockResources) v2.Resources {
				mock.Database.EXPECT().GetDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.DataQualityStats{}, 0, fmt.Errorf("db error"))

				return defaultResources(mockCtrl, mock)
			},
			Expected: api.ErrorWrapper{
				HTTPStatus: http.StatusInternalServerError,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsInternalServerError}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mock := mockResources{
				Database: mocks.NewMockDatabase(mockCtrl),
			}
			resources := tc.Setup(mockCtrl, mock)
			params := fmt.Sprintf("?%s", tc.Input.Params.Encode())
			req, err := http.NewRequest("GET", fmt.Sprintf(endpoint, params), nil)
			require.NoError(t, err)
			req = req.WithContext(tc.Input.Context)

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/data-quality-stats", resources.GetDataQualityStats).Methods("GET")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.Expected.HTTPStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected.HTTPStatus)
			}

			var body any
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatal("failed to unmarshal response body")
			}

			require.Equal(t, len(tc.Expected.Errors), 1)
			if !reflect.DeepEqual(body.(map[string]any)["errors"].([]any)[0].(map[string]any)["message"], tc.Expected.Errors[0].Message) {
				t.Errorf("For input: %v, got %v, want %v", tc.Input, body, tc.Expected.Errors[0].Message)
			}
		})
	}

}

func TestGetDataQualityStats_Success(t *testing.T) {
	environmentID := "environment-1"
	endpoint := "/api/v2/data-quality-stats%s"
	etacEnabled := dogtags.TestOverrides{
		Bools: map[dogtags.BoolDogTag]bool{
			dogtags.ETAC_ENABLED: true,
		},
	}

	type Input struct {
		Params           url.Values
		User             model.User
		DogTagsOverrides dogtags.TestOverrides
	}

	type Expected struct {
		Code int
	}

	var cases = []struct {
		Name     string
		Input    Input
		Setup    func(mockDB *mocks.MockDatabase, user model.User)
		Expected Expected
	}{
		{
			Name: "AllQueryParams",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"sort_by":                       []string{"-updated_at"},
					"limit":                         []string{"1"},
					"start":                         []string{"2022-03-23T07:20:50.52Z"},
					"end":                           []string{"2022-04-23T07:20:50.52Z"},
					"skip":                          []string{"0"},
				},
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
		{
			Name: "SortByUpdatedAtAscending",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"sort_by":                       []string{"updated_at"},
				},
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
		{
			Name: "SortByCreatedAtAscending",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"sort_by":                       []string{"created_at"},
				},
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
		{
			Name: "SortByCreatedAtDescending",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"sort_by":                       []string{"-created_at"},
				},
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
		{
			Name: "StartTime",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"start":                         []string{"2022-03-23T07:20:50.52Z"},
				},
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
		{
			Name: "EndTime",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"end":                           []string{"2022-04-23T07:20:50.52Z"},
				},
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
		{
			Name: "Limit",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"limit":                         []string{"2"},
				},
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
		{
			Name: "Skip",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
					"skip":                          []string{"1"},
				},
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
		{
			Name: "DefaultOptionalParams",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
				},
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
		{
			Name: "ETACDisabledSkipsAccessCheck",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
				},
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
		{
			Name: "ETACEnabledWithAllEnvironmentsSkipsAccessCheck",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
				},
				User: func() model.User {
					user := setupUser()
					user.AllEnvironments = true
					return user
				}(),
				DogTagsOverrides: etacEnabled,
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
		{
			Name: "ETACEnabledWithEnvironmentAccessChecksAccessThenFetchesStats",
			Input: Input{
				Params: url.Values{
					api.QueryParameterEnvironmentId: []string{environmentID},
				},
				DogTagsOverrides: etacEnabled,
			},
			Setup: func(mockDB *mocks.MockDatabase, user model.User) {
				successfulAccessCheck := mockDB.EXPECT().GetEnvironmentTargetedAccessControlForUser(gomock.Any(), user).Return([]model.EnvironmentTargetedAccessControl{
					{EnvironmentID: environmentID},
				}, nil).Times(1)
				successfulStatsFetch := mockDB.EXPECT().GetDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.DataQualityStats{}, 0, nil).Times(1)
				gomock.InOrder(successfulAccessCheck, successfulStatsFetch)
			},
			Expected: Expected{
				Code: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			user := tc.Input.User
			if user.PrincipalName == "" {
				user = setupUser()
			}

			mockDB := mocks.NewMockDatabase(mockCtrl)
			if tc.Setup == nil {
				mockDB.EXPECT().GetDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.DataQualityStats{}, 0, nil).Times(1)
			} else {
				tc.Setup(mockDB, user)
			}

			resources := v2.Resources{
				DB:      mockDB,
				DogTags: dogtags.NewTestService(tc.Input.DogTagsOverrides),
			}
			params := fmt.Sprintf("?%s", tc.Input.Params.Encode())
			req, err := http.NewRequest("GET", fmt.Sprintf(endpoint, params), nil)
			require.NoError(t, err)
			req = req.WithContext(setupUserCtx(user))

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/data-quality-stats", resources.GetDataQualityStats).Methods("GET")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.Expected.Code {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected.Code)
			}
		})
	}

}

func TestGetPlatformAggregateStats_Failure(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/platform/%s/data-quality-stats%s"

	mockDB := mocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetADDataQualityAggregations(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, fmt.Errorf("db error"))
	mockDB.EXPECT().GetAzureDataQualityAggregations(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, fmt.Errorf("db error"))

	resources := v2.Resources{DB: mockDB}

	type Input struct {
		TenantID string
		Params   url.Values
	}

	var cases = []struct {
		Input    Input
		Expected api.ErrorWrapper
	}{
		{
			Input{
				"invalidPlatform",
				url.Values{},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(v2.ErrInvalidPlatformId, "invalidPlatform")}},
			},
		},
		// AD
		{
			Input{
				"ad",
				url.Values{
					"sort_by": []string{"invalidColumn"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsNotSortable}},
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"start": []string{"invalidRFC3339"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(api.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"end": []string{"invalidRFC3339"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(api.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"limit": []string{"invalidLimit"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(utils.ErrorInvalidLimit, []string{"invalidLimit"})}},
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"skip": []string{"invalidSkip"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(utils.ErrorInvalidSkip, []string{"invalidSkip"})}},
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"db_error": []string{"dbError"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusInternalServerError,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsInternalServerError}},
			},
		},
		// Azure
		{
			Input{
				"azure",
				url.Values{
					"sort_by": []string{"invalidColumn"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsNotSortable}},
			},
		},
		{
			Input{
				"azure",
				url.Values{
					"start": []string{"invalidRFC3339"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(api.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
			},
		},
		{
			Input{
				"azure",
				url.Values{
					"end": []string{"invalidRFC3339"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(api.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
			},
		},
		{
			Input{
				"azure",
				url.Values{
					"limit": []string{"invalidLimit"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(utils.ErrorInvalidLimit, []string{"invalidLimit"})}},
			},
		},
		{
			Input{
				"azure",
				url.Values{
					"skip": []string{"invalidSkip"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(utils.ErrorInvalidSkip, []string{"invalidSkip"})}},
			},
		},
		{
			Input{
				"azure",
				url.Values{
					"db_error": []string{"dbError"},
				},
			},
			api.ErrorWrapper{
				HTTPStatus: http.StatusInternalServerError,
				Errors:     []api.ErrorDetails{{Message: api.ErrorResponseDetailsInternalServerError}},
			},
		},
	}

	for _, tc := range cases {
		params := fmt.Sprintf("?%s", tc.Input.Params.Encode())
		if req, err := http.NewRequest("GET", fmt.Sprintf(endpoint, tc.Input.TenantID, params), nil); err != nil {
			t.Fatal(err)
		} else {
			router := mux.NewRouter()
			router.HandleFunc("/api/v2/platform/{platform_id}/data-quality-stats", resources.GetPlatformAggregateStats).Methods("GET")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.Expected.HTTPStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected.HTTPStatus)
			}

			var body any
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatal("failed to unmarshal response body")
			}

			require.Equal(t, len(tc.Expected.Errors), 1)
			if !reflect.DeepEqual(body.(map[string]any)["errors"].([]any)[0].(map[string]any)["message"], tc.Expected.Errors[0].Message) {
				t.Errorf("For input: %v, got %v, want %v", tc.Input, body, tc.Expected.Errors[0].Message)
			}
		}
	}
}

func TestGetPlatformAggregateStats_Success(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/platform/%s/data-quality-stats%s"

	mockDB := mocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetADDataQualityAggregations(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.ADDataQualityAggregations{}, 0, nil).AnyTimes()
	mockDB.EXPECT().GetAzureDataQualityAggregations(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.AzureDataQualityAggregations{}, 0, nil).AnyTimes()

	resources := v2.Resources{DB: mockDB}

	type Input struct {
		TenantID string
		Params   url.Values
	}

	type Expected struct {
		Code int
	}

	var cases = []struct {
		Input    Input
		Expected Expected
	}{
		// AD
		{
			Input{
				"ad",
				url.Values{
					"sort_by": []string{"-updated_at"},
					"limit":   []string{"1"},
					"start":   []string{"2022-03-23T07:20:50.52Z"},
					"end":     []string{"2022-04-23T07:20:50.52Z"},
					"skip":    []string{"0"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"sort_by": []string{"updated_at"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"start": []string{"2022-03-23T07:20:50.52Z"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"end": []string{"2022-04-23T07:20:50.52Z"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"limit": []string{"2"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"skip": []string{"1"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		// Azure
		{
			Input{
				"ad",
				url.Values{
					"sort_by": []string{"-updated_at"},
					"limit":   []string{"1"},
					"start":   []string{"2022-03-23T07:20:50.52Z"},
					"end":     []string{"2022-04-23T07:20:50.52Z"},
					"skip":    []string{"0"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"sort_by": []string{"updated_at"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"start": []string{"2022-03-23T07:20:50.52Z"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"end": []string{"2022-04-23T07:20:50.52Z"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"limit": []string{"2"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
		{
			Input{
				"ad",
				url.Values{
					"skip": []string{"1"},
				},
			},
			Expected{
				Code: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		params := fmt.Sprintf("?%s", tc.Input.Params.Encode())
		if req, err := http.NewRequest("GET", fmt.Sprintf(endpoint, tc.Input.TenantID, params), nil); err != nil {
			t.Fatal(err)
		} else {
			router := mux.NewRouter()
			router.HandleFunc("/api/v2/platform/{platform_id}/data-quality-stats", resources.GetPlatformAggregateStats).Methods("GET")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.Expected.Code {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected.Code)
			}
		}
	}
}

func TestResources_GetDatabaseCompleteness(t *testing.T) {
	t.Parallel()

	type mock struct {
		mockGraph *graphmocks.MockDatabase
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
		{
			name: "Error: database error - Internal Server Error",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/completeness",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				responseCode:   http.StatusInternalServerError,
				responseBody:   `{"errors":[{"context":"","message":"Error getting quality stat: error"}],"http_status":500,"request_id":"","timestamp":"0001-01-01T00:00:00Z"}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
		{
			name: "Success - OK",
			buildRequest: func() *http.Request {
				return &http.Request{
					URL: &url.URL{
						Path: "/api/v2/completeness",
					},
					Method: http.MethodGet,
				}
			},
			setupMocks: func(t *testing.T, mock *mock) {
				t.Helper()
				mock.mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil)
			},
			expected: expected{
				responseCode:   http.StatusOK,
				responseBody:   `{"data":{}}`,
				responseHeader: http.Header{"Content-Type": []string{"application/json"}},
			},
		},
	}
	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mock{
				mockGraph: graphmocks.NewMockDatabase(ctrl),
			}

			request := testCase.buildRequest()
			testCase.setupMocks(t, mocks)

			resources := v2.Resources{
				Graph: mocks.mockGraph,
			}

			response := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/api/v2/completeness", resources.GetDatabaseCompleteness).Methods(request.Method)
			router.ServeHTTP(response, request)

			status, header, body := test.ProcessResponse(t, response)

			assert.Equal(t, testCase.expected.responseCode, status)
			assert.Equal(t, testCase.expected.responseHeader, header)
			assert.JSONEq(t, testCase.expected.responseBody, body)
		})
	}
}
