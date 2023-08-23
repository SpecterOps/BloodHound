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

package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/api"
)

type ExpectedResponse struct {
	Code int
	Body any
}

func TestHandler(t *testing.T, methods []string, endpoint string, handler func(http.ResponseWriter, *http.Request), req http.Request, expected ExpectedResponse) {
	router := mux.NewRouter()
	router.HandleFunc(endpoint, handler).Methods(methods...)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, &req)

	if status := rr.Code; status != expected.Code {
		t.Errorf("handler returned wrong status code: got %v want %v", status, expected.Code)
	}

	if rr.Body.Bytes() != nil {
		var body any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatal("failed to unmarshal response body")
		}

		if !reflect.DeepEqual(body, expected.Body) {
			t.Errorf("For request: %s %v, got %v, want %v", req.Method, req.URL, body, expected.Body)
		}
	} else if expected.Body != nil {
		t.Errorf("For request: %v, got %v, want %v", req, rr.Body, expected.Body)
	}
}

func TestV2HandlerFailure(t *testing.T, methods []string, endpoint string, handler func(http.ResponseWriter, *http.Request), req http.Request, expected api.ErrorWrapper) {
	router := mux.NewRouter()
	router.HandleFunc(endpoint, handler).Methods(methods...)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, &req)

	if status := rr.Code; status != expected.HTTPStatus {
		t.Errorf("handler returned wrong status code: got %v want %v", status, expected.HTTPStatus)
	}

	if rr.Body.Bytes() != nil {
		var body any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to unmarshal response body: %s", err)
		}

		var actualErrorMsg any
		if body.(map[string]any)["errors"] != nil {
			actualErrorMsg = body.(map[string]any)["errors"].([]any)[0].(map[string]any)["message"]
		} else {
			// If there's no "errors" key this is probably a success, so just capture the whole body instead of panicking
			actualErrorMsg = body
		}
		if !reflect.DeepEqual(actualErrorMsg, expected.Errors[0].Message) {
			t.Errorf("For request: %s %v, got '%v', want '%v'", req.Method, req.URL, actualErrorMsg, expected.Errors[0].Message)
		}
	} else if expected.Errors[0].Message != "" {
		t.Errorf("For request: %v, got %v, want %v", req, rr.Body, expected.Errors[0].Message)
	}
}
