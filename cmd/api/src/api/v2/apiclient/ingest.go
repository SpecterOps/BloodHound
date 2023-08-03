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
	"fmt"
	"net/http"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/headers"
)

func (s Client) SendIngestObject(payload map[string]any, userAgent string) error {
	header := http.Header{}
	header.Add(headers.UserAgent.String(), userAgent)

	if response, err := s.Request(http.MethodPost, "/api/v2/ingest", nil, payload, header); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}
	}

	return nil
}

func (s Client) CreateFileUploadTask() (model.FileUploadJob, error) {
	var job model.FileUploadJob
	if response, err := s.Request(http.MethodPost, "api/v2/file-upload/start", nil, nil); err != nil {
		return job, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return job, ReadAPIError(response)
		}

		return job, api.ReadAPIV2ResponsePayload(&job, response)
	}
}

func (s Client) SendFileUploadData(payload map[string]any, id int64) error {
	if response, err := s.Request(http.MethodPost, fmt.Sprintf("api/v2/file-upload/%d", id), nil, payload); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}

		return nil
	}
}

func (s Client) CompleteFileUpload(id int64) error {
	if response, err := s.Request(http.MethodPost, fmt.Sprintf("api/v2/file-upload/%d/end", id), nil, nil); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}

		return nil
	}
}

func (s Client) GetDatapipeStatus() (model.DatapipeStatusWrapper, error) {
	var status model.DatapipeStatusWrapper
	if response, err := s.Request(http.MethodGet, "/api/v2/datapipe/status", nil, nil); err != nil {
		return model.DatapipeStatusWrapper{}, err
	} else {
		defer response.Body.Close()

		return status, api.ReadAPIV2ResponsePayload(&status, response)
	}
}
