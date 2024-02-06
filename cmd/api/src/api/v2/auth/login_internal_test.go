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

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/specterops/bloodhound/src/database/mocks"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/specterops/bloodhound/src/api"
	api_mocks "github.com/specterops/bloodhound/src/api/mocks"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
)

func TestLoginFailure(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDB := mocks.NewMockDatabase(mockCtrl)

	endpoint := "/api/v2/auth/login"
	// mfaUser := model.User{
	// 	EmailAddress: null.NewString("good@user.io", true),
	// 	AuthSecret: &model.AuthSecret{
	// 		TOTPActivated: true,
	// 	},
	// }

	goCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})

	req1 := api.LoginRequest{
		LoginMethod: auth.ProviderTypeSecret,
		Username:    "asdfghjk@specterops.io",
		Secret:      "abc1234",
	}

	req2 := api.LoginRequest{
		LoginMethod: auth.ProviderTypeSecret,
		Username:    "asdfghjk@specterops.io",
		Secret:      "imabadpassword",
	}

	req3 := api.LoginRequest{
		LoginMethod: auth.ProviderTypeSecret,
		Username:    "dberror@specterops.io",
		Secret:      "dberror",
	}

	req4 := api.LoginRequest{
		LoginMethod: auth.ProviderTypeSecret,
		Username:    "foo",
		Secret:      "bar",
	}

	mockAuthenticator := api_mocks.NewMockAuthenticator(mockCtrl)
	mockAuthenticator.EXPECT().LoginWithSecret(gomock.Any(), req1).Return(api.LoginDetails{User: model.User{EULAAccepted: true}}, auth.ErrorInvalidOTP)
	mockAuthenticator.EXPECT().LoginWithSecret(gomock.Any(), req2).Return(api.LoginDetails{User: model.User{EULAAccepted: true}}, api.ErrInvalidAuth)
	mockAuthenticator.EXPECT().LoginWithSecret(gomock.Any(), req3).Return(api.LoginDetails{User: model.User{EULAAccepted: true}}, fmt.Errorf("db error"))
	mockAuthenticator.EXPECT().LoginWithSecret(gomock.Any(), req4).Return(api.LoginDetails{User: model.User{EULAAccepted: true}}, api.ErrUserDisabled)
	mockDB.EXPECT().LookupUser(gomock.Any()).Return(model.User{EULAAccepted: false}, nil).Times(5)
	mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil).Times(5)

	resources := NewLoginResource(config.Configuration{}, mockAuthenticator, mockDB)

	type Input struct {
		Payload api.LoginRequest
	}

	var cases = []struct {
		Input    Input
		Expected map[string]any
	}{
		{
			Input{api.LoginRequest{}},
			map[string]any{
				"HTTPStatus": http.StatusBadRequest,
				"Errors": []map[string]any{
					{
						"Context": "auth",
						"Message": "Login method  is not supported.",
					},
				},
			},
		},
		{
			Input{req1},
			map[string]any{
				"HTTPStatus": http.StatusBadRequest,
				"Errors": []map[string]any{
					{
						"Context": "auth",
						"Message": api.ErrorResponseDetailsOTPInvalid,
					},
				},
			},
		},
		{
			Input{req2},
			map[string]any{
				"HTTPStatus": http.StatusUnauthorized,
				"Errors": []map[string]any{
					{
						"Context": "auth",
						"Message": api.ErrorResponseDetailsAuthenticationInvalid,
					},
				},
			},
		},
		{
			Input{req3},
			map[string]any{
				"HTTPStatus": http.StatusInternalServerError,
				"Errors": []map[string]any{
					{
						"Context": "auth",
						"Message": api.ErrorResponseDetailsInternalServerError,
					},
				},
			},
		},
		{
			Input{req4},
			map[string]any{
				"HTTPStatus": http.StatusForbidden,
				"Errors": []map[string]any{
					{
						"Context": "auth",
						"Message": api.ErrUserDisabled.Error(),
					},
				},
			},
		},
	}

	for _, tc := range cases {
		if payload, err := json.Marshal(tc.Input.Payload); err != nil {
			t.Fatal(err)
		} else if req, err := http.NewRequestWithContext(goCtx, "POST", endpoint, bytes.NewReader(payload)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
			router := mux.NewRouter()
			router.HandleFunc(endpoint, resources.Login).Methods("POST")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.Expected["HTTPStatus"] {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected["HTTPStatus"])
			}

			var body any
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatal("failed to unmarshal response body")
			}

			require.Equal(t, len(tc.Expected["Errors"].([]map[string]any)), 1)
			if !reflect.DeepEqual(body.(map[string]any)["errors"].([]any)[0].(map[string]any)["message"], tc.Expected["Errors"].([]map[string]any)[0]["Message"]) {
				t.Errorf("For input: %v, got %v, want %v", tc.Input, body, tc.Expected["Errors"].([]map[string]any)[0])
			}
		}
	}
}

func TestLoginSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocks.NewMockDatabase(mockCtrl)

	endpoint := "/api/v2/auth/login"
	goCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})

	input := api.LoginRequest{
		LoginMethod: auth.ProviderTypeSecret,
		Username:    "foo@specterops.io",
		Secret:      "bar",
	}

	mockAuthenticator := api_mocks.NewMockAuthenticator(mockCtrl)
	mockAuthenticator.EXPECT().LoginWithSecret(gomock.Any(), input).Return(api.LoginDetails{User: model.User{AuthSecret: &model.AuthSecret{}, EULAAccepted: true}, SessionToken: "imasessiontoken"}, nil)
	mockDB.EXPECT().LookupUser(gomock.Any()).Return(model.User{EULAAccepted: false}, nil)
	mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil)

	resources := NewLoginResource(config.Configuration{}, mockAuthenticator, mockDB)

	if payload, err := json.Marshal(input); err != nil {
		t.Fatal(err)
	} else if req, err := http.NewRequestWithContext(goCtx, "POST", endpoint, bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	} else {
		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.Login).Methods("POST")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		var body any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatal("failed to unmarshal response body")
		}

		require.Equal(t, rr.Code, http.StatusOK)
		require.Contains(t, rr.Body.String(), `"session_token":"imasessiontoken"`)
		require.Contains(t, rr.Body.String(), `"auth_expired":true`)
		require.NotContains(t, rr.Body.String(), "eula_accepted")
	}

}
