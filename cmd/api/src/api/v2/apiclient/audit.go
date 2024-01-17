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

package apiclient

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/model"

	"github.com/specterops/bloodhound/src/api"
)

func (s Client) GetLatestAuditLogs() (v2.AuditLogsResponse, error) {
	var logs v2.AuditLogsResponse

	if response, err := s.Request(http.MethodGet, "api/v2/audit", nil, nil); err != nil {
		return logs, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return logs, ReadAPIError(response)
		}

		return logs, api.ReadAPIV2ResponsePayload(&logs, response)
	}
}

func (s Client) ListAuditLogs(after, before time.Time, offset, limit int) (v2.AuditLogsResponse, error) {
	var logs v2.AuditLogsResponse

	params := url.Values{
		model.PaginationQueryParameterAfter:  []string{after.Format(time.RFC3339Nano)},
		model.PaginationQueryParameterBefore: []string{before.Format(time.RFC3339Nano)},
		model.PaginationQueryParameterOffset: []string{strconv.Itoa(offset)},
		model.PaginationQueryParameterLimit:  []string{strconv.Itoa(limit)},
	}

	if response, err := s.Request(http.MethodGet, "api/v2/audit", params, nil); err != nil {
		return logs, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return logs, ReadAPIError(response)
		}

		return logs, api.ReadAPIV2ResponsePayload(&logs, response)
	}
}
