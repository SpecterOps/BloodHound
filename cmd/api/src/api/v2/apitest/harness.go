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

package apitest

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/src/test/must"

	"github.com/gorilla/mux"
)

type Harness struct {
	t           *testing.T
	handler     http.HandlerFunc
	commonInput InputFunc
}

func NewHarness(t *testing.T, handler http.HandlerFunc) *Harness {
	return &Harness{
		t:       t,
		handler: handler,
	}
}

func (s *Harness) WithCommonRequest(delegate InputFunc) *Harness {
	s.commonInput = delegate
	return s
}

func (s *Harness) Run(cases []Case) {
	for _, tcase := range cases {
		var (
			input = Input{
				query:   url.Values{},
				urlVars: make(map[string]string),
				request: must.NewHTTPRequest("GET", "/", nil),
			}
			recorder = httptest.NewRecorder()
		)
		if s.commonInput != nil {
			s.commonInput(&input)
		}
		if tcase.Input != nil {
			tcase.Input(&input)
		}
		if tcase.Setup != nil {
			tcase.Setup()
		}
		input.request.URL.RawQuery = input.query.Encode()
		input.request = mux.SetURLVars(input.request, input.urlVars)
		s.handler.ServeHTTP(recorder, input.request)
		s.t.Run(tcase.Name, func(t *testing.T) {
			defer func() {
				if recovery := recover(); recovery != nil {
					t.Fatalf("Panic during request execution against the handler: %v", recovery)
				}
			}()
			tcase.Test(Output{t: t, response: recorder})
		})
	}
}
