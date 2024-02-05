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

package tools

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"runtime/pprof"

	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/src/api"
)

const (
	errUnknownProfile = errors.Error("unknown profile")

	pprofEnableDebugSymbols = 1

	profileQueryParameterName = "profile"

	pprofLookupGoroutines   = "goroutine"
	pprofLookupThreadCreate = "threadcreate"
	pprofLookupHeap         = "heap"
	pprofLookupAllocations  = "allocs"
	pprofLookupBlock        = "block"
	pprofLookupMutex        = "mutex"
)

func getProfileFromValues(values url.Values) (string, error) {
	// Default to looking up goroutine traces
	profile := pprofLookupGoroutines

	if passedProfile := values.Get(profileQueryParameterName); len(passedProfile) > 0 {
		switch passedProfile {
		case pprofLookupGoroutines, pprofLookupThreadCreate, pprofLookupHeap, pprofLookupAllocations, pprofLookupBlock, pprofLookupMutex:
			return passedProfile, nil
		default:
			return passedProfile, errUnknownProfile
		}
	}

	return profile, nil
}

type TraceHandlerResponse struct {
	Output string `json:"output"`
}

func NewTraceHandler() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		if profile, err := getProfileFromValues(request.URL.Query()); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("bad profile: %s", profile), request), response)
		} else {
			output := &bytes.Buffer{}

			if err := pprof.Lookup(profile).WriteTo(output, pprofEnableDebugSymbols); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("pprof error: %v", err), request), response)
			} else {
				api.WriteJSONResponse(request.Context(), TraceHandlerResponse{
					Output: output.String(),
				}, http.StatusOK, response)
			}
		}
	}
}
