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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/specterops/bloodhound/src/utils"

	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/gorilla/mux"
	"go.uber.org/mock/gomock"
)

func TestGetADDataQualityStats_Failure(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	domainSID := "S-1-5-21-3130019616-2776909439-2417379446"
	endpoint := "/api/v2/ad-domains/%s/data-quality-stats%s"

	mockDB := mocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetADDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, fmt.Errorf("db error"))

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
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(v2.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
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
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(v2.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
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
	mockDB.EXPECT().GetADDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.ADDataQualityStats{}, 0, nil).AnyTimes()

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
	mockDB.EXPECT().GetAzureDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, fmt.Errorf("db error"))

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
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(v2.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
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
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(v2.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
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
	mockDB.EXPECT().GetAzureDataQualityStats(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.AzureDataQualityStats{}, 0, nil).AnyTimes()

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

func TestGetPlatformAggregateStats_Failure(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	endpoint := "/api/v2/platform/%s/data-quality-stats%s"

	mockDB := mocks.NewMockDatabase(mockCtrl)
	mockDB.EXPECT().GetADDataQualityAggregations(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, fmt.Errorf("db error"))
	mockDB.EXPECT().GetAzureDataQualityAggregations(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, 0, fmt.Errorf("db error"))

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
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(v2.ErrorInvalidPlatformId, "invalidPlatform")}},
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
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(v2.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
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
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(v2.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
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
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(v2.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
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
				Errors:     []api.ErrorDetails{{Message: fmt.Sprintf(v2.ErrorInvalidRFC3339, []string{"invalidRFC3339"})}},
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
	mockDB.EXPECT().GetADDataQualityAggregations(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.ADDataQualityAggregations{}, 0, nil).AnyTimes()
	mockDB.EXPECT().GetAzureDataQualityAggregations(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.AzureDataQualityAggregations{}, 0, nil).AnyTimes()

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
