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
	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/model"
	"net/http"
)

func (s Client) ListSavedQueries() (model.SavedQueries, error) {
	var queries model.SavedQueries
	if response, err := s.Request(http.MethodGet, "api/v2/saved-queries", nil, nil); err != nil {
		return queries, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return queries, ReadAPIError(response)
		}

		return queries, api.ReadAPIV2ResponsePayload(&queries, response)
	}
}

func (s Client) CreateSavedQuery() (model.SavedQuery, error) {
	var query model.SavedQuery
	payload := v2.CreateSavedQueryRequest{
		Query: "Match(q:Question {life: 1, universe: 1, everything: 1}) return q",
		Name:  "AnswerToLifeUniverseEverything",
	}

	if response, err := s.Request(http.MethodPost, "api/v2/saved-queries", nil, payload); err != nil {
		return query, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return query, ReadAPIError(response)
		}

		return query, api.ReadAPIV2ResponsePayload(&query, response)
	}
}
