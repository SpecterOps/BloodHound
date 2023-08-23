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

//go:build serial_integration
// +build serial_integration

package v2_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/src/api/v2/integration"
	"github.com/specterops/bloodhound/src/test/fixtures/fixtures"
	"github.com/stretchr/testify/assert"
)

func Test_FileUpload(t *testing.T) {
	testCtx := integration.NewContext(t, integration.StartBHServer)
	apiClient := testCtx.AdminClient()
	loader := testCtx.FixtureLoader

	uploadJob, err := apiClient.CreateFileUploadTask()
	assert.Nil(t, err)

	jobEndpoint := fmt.Sprintf("api/v2/file-upload/%d", uploadJob.ID)

	// JSON input success
	jsonInput := loader.GetReader("v6/ingest/computers.json")
	defer jsonInput.Close()
	req, err := apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, jsonInput)
	assert.Nil(t, err)
	resp, err := apiClient.Raw(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	// JSON input with incorrect compression header
	jsonInput2 := loader.GetReader("v6/ingest/containers.json")
	defer jsonInput2.Close()
	req, err = apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, jsonInput2)
	assert.Nil(t, err)
	req.Header.Set("Content-Encoding", "gzip")
	resp, err = apiClient.Raw(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// gzip input with correct compression header
	var (
		body bytes.Buffer
		gw   = gzip.NewWriter(&body)
	)
	jsonInput3 := loader.GetReader("v6/ingest/domains.json")
	defer jsonInput3.Close()
	n, err := io.Copy(gw, jsonInput3)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, n)
	assert.Nil(t, gw.Close())
	req, err = apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, io.NopCloser(&body))
	assert.Nil(t, err)
	req.Header.Set("Content-Encoding", "gzip")
	resp, err = apiClient.Raw(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	// gzip input with incorrect compression header
	var (
		body2 bytes.Buffer
		gw2   = gzip.NewWriter(&body2)
	)
	jsonInput4 := loader.GetReader("v6/ingest/gpos.json")
	defer jsonInput3.Close()
	n, err = io.Copy(gw2, jsonInput4)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, n)
	assert.Nil(t, gw.Close())
	req, err = apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, io.NopCloser(&body2))
	assert.Nil(t, err)
	req.Header.Set("Content-Encoding", "deflate")
	resp, err = apiClient.Raw(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// gzip input with missing compression header
	var (
		body3 bytes.Buffer
		gw3   = gzip.NewWriter(&body3)
	)
	jsonInput5 := loader.GetReader("v6/ingest/groups.json")
	defer jsonInput3.Close()
	n, err = io.Copy(gw3, jsonInput5)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, n)
	assert.Nil(t, gw.Close())
	req, err = apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, io.NopCloser(&body3))
	assert.Nil(t, err)
	resp, err = apiClient.Raw(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func Test_FileUploadWorkFlowVersion5(t *testing.T) {
	testCtx := integration.NewContext(t, integration.StartBHServer)

	testCtx.SendFileIngest([]string{
		"v5/ingest/domains.json",
		"v5/ingest/computers.json",
		"v5/ingest/containers.json",
		"v5/ingest/gpos.json",
		"v5/ingest/groups.json",
		"v5/ingest/ous.json",
		"v5/ingest/users.json",
		"v5/ingest/deleted.json",
		"v5/ingest/sessions.json",
	})

	//Assert that we created stuff we expected
	testCtx.AssertIngest(fixtures.IngestAssertions)
}

func Test_FileUploadWorkFlowVersion6(t *testing.T) {
	testCtx := integration.NewContext(t, integration.StartBHServer)

	testCtx.SendFileIngest([]string{
		"v6/ingest/domains.json",
		"v6/ingest/computers.json",
		"v6/ingest/containers.json",
		"v6/ingest/gpos.json",
		"v6/ingest/groups.json",
		"v6/ingest/ous.json",
		"v6/ingest/users.json",
		"v6/ingest/deleted.json",
		"v6/ingest/sessions.json",
	})

	//Assert that we created stuff we expected
	testCtx.AssertIngest(fixtures.IngestAssertions)
}

func Test_CompressedFileUploadWorkFlowVersion5(t *testing.T) {
	testCtx := integration.NewContext(t, integration.StartBHServer)

	testCtx.SendCompressedFileIngest([]string{
		"v5/ingest/domains.json",
		"v5/ingest/computers.json",
		"v5/ingest/containers.json",
		"v5/ingest/gpos.json",
		"v5/ingest/groups.json",
		"v5/ingest/ous.json",
		"v5/ingest/users.json",
		"v5/ingest/deleted.json",
		"v5/ingest/sessions.json",
	})

	//Assert that we created stuff we expected
	testCtx.AssertIngest(fixtures.IngestAssertions)
}

func Test_CompressedFileUploadWorkFlowVersion6(t *testing.T) {
	testCtx := integration.NewContext(t, integration.StartBHServer)

	testCtx.SendCompressedFileIngest([]string{
		"v6/ingest/domains.json",
		"v6/ingest/computers.json",
		"v6/ingest/containers.json",
		"v6/ingest/gpos.json",
		"v6/ingest/groups.json",
		"v6/ingest/ous.json",
		"v6/ingest/users.json",
		"v6/ingest/deleted.json",
		"v6/ingest/sessions.json",
	})

	//Assert that we created stuff we expected
	testCtx.AssertIngest(fixtures.IngestAssertions)
}
