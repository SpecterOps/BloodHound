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
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/database/mocks"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"

	"github.com/gorilla/mux"
	"go.uber.org/mock/gomock"

	"github.com/specterops/bloodhound/src/api"
	api_mocks "github.com/specterops/bloodhound/src/api/mocks"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
)

func TestLoginExpiry(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mocks.NewMockDatabase(mockCtrl)

	endpoint := "/api/v2/auth/login"
	goCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})

	req1 := api.LoginRequest{
		LoginMethod: auth.ProviderTypeSecret,
		Username:    "irshad@specterops.io",
		Secret:      "rules",
	}

	req2 := api.LoginRequest{LoginMethod: auth.ProviderTypeSecret, Username: "abc", Secret: "123"}

	mockAuthenticator := api_mocks.NewMockAuthenticator(mockCtrl)
	mockAuthenticator.EXPECT().LoginWithSecret(gomock.Any(), req1).Return(api.LoginDetails{User: model.User{AuthSecret: &model.AuthSecret{ExpiresAt: time.Now().UTC().Add(time.Hour * 24)}, EULAAccepted: true}, SessionToken: "imasession"}, nil)
	mockAuthenticator.EXPECT().LoginWithSecret(gomock.Any(), req2).Return(api.LoginDetails{User: model.User{AuthSecret: &model.AuthSecret{ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * -1)}, EULAAccepted: true}, SessionToken: "imasession"}, nil)
	mockDB.EXPECT().LookupUser(gomock.Any()).Return(model.User{EULAAccepted: false}, nil).Times(2)
	mockDB.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	resources := NewLoginResource(config.Configuration{}, mockAuthenticator, mockDB)

	type Input struct {
		Payload api.LoginRequest
	}

	type Expected struct {
		Code int
		Body map[string]any
	}

	var cases = []struct {
		Input    Input
		Expected Expected
	}{
		{
			Input{req1},
			Expected{
				http.StatusOK,
				map[string]any{
					"auth_expired":  false,
					"user_id":       "00000000-0000-0000-0000-000000000000",
					"session_token": "imasession",
				},
			},
		},
		{
			Input{req2},
			Expected{
				http.StatusOK,
				map[string]any{
					"auth_expired":  true,
					"user_id":       "00000000-0000-0000-0000-000000000000",
					"session_token": "imasession",
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

			if status := rr.Code; status != tc.Expected.Code {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.Expected.Code)
			}

			var body map[string]any
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatal("failed to unmarshal response body")
			}

			if !reflect.DeepEqual(body["data"], tc.Expected.Body) {
				t.Errorf("For input: %v, got %v, want %v", tc.Input, body["data"], tc.Expected.Body)
			}
		}
	}
}
