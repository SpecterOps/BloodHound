// Copyright 2025 Specter Ops, Inc.
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

package v2

import (
	"net/http"
	"slices"
	"strings"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/dawgs/graph"
)

type ListKindsResponse struct {
	Kinds graph.Kinds `json:"kinds"`
}

func (s Resources) ListKinds(response http.ResponseWriter, request *http.Request) {
	if kinds, err := s.Graph.FetchKinds(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		// Alpha sort
		slices.SortFunc(kinds, func(a, b graph.Kind) int {
			return strings.Compare(a.String(), b.String())
		})

		api.WriteBasicResponse(request.Context(), ListKindsResponse{Kinds: kinds}, http.StatusOK, response)
	}
}
