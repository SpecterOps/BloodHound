// Copyright 2024 Specter Ops, Inc.
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

package apiclient

import (
	"net/http"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
)

func (s Client) RequestAnalysis() error {
	if response, err := s.Request(http.MethodPut, "api/v2/analysis", nil, nil); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}

		return nil
	}
}

func (s Client) GetAnalysisRequest() (model.AnalysisRequest, error) {
	var analReq model.AnalysisRequest

	if response, err := s.Request(http.MethodGet, "api/v2/analysis/status", nil, nil); err != nil {
		return analReq, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return analReq, ReadAPIError(response)
		}

		return analReq, api.ReadAPIV2ResponsePayload(&analReq, response)
	}
}
