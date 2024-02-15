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

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/src/api/v2/integration"
	"github.com/specterops/bloodhound/src/test/fixtures/fixtures"
	"github.com/stretchr/testify/assert"
)

func Test_FileUpload(t *testing.T) {
	testCtx := integration.NewFOSSContext(t)
	apiClient := testCtx.AdminClient()
	loader := testCtx.FixtureLoader

	uploadJob, err := apiClient.CreateFileUploadTask()
	assert.Nil(t, err)

	jobEndpoint := fmt.Sprintf("api/v2/file-upload/%d", uploadJob.ID)

	t.Run("JSON input with success", func(tx *testing.T) {
		jsonInput := loader.GetReader("v6/ingest/computers.json")
		defer jsonInput.Close()
		req, err := apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, jsonInput)
		assert.Nil(tx, err)
		resp, err := apiClient.Raw(req)
		assert.Nil(tx, err)
		assert.Equal(tx, http.StatusAccepted, resp.StatusCode)
	})

	t.Run("JSON input with incorrect compression header", func(tx *testing.T) {
		jsonInput := loader.GetReader("v6/ingest/containers.json")
		defer jsonInput.Close()
		req, err := apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, jsonInput)
		assert.Nil(tx, err)
		req.Header.Set(headers.ContentEncoding.String(), "gzip")
		resp, err := apiClient.Raw(req)
		assert.Nil(tx, err)
		assert.Equal(tx, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Gzip input with correct compression header", func(tx *testing.T) {
		var (
			body bytes.Buffer
			gw   = gzip.NewWriter(&body)
		)
		jsonInput := loader.GetReader("v6/ingest/domains.json")
		defer jsonInput.Close()
		n, err := io.Copy(gw, jsonInput)
		assert.Nil(tx, err)
		assert.NotEqual(tx, 0, n)
		assert.Nil(tx, gw.Close())
		req, err := apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, io.NopCloser(&body))
		assert.Nil(tx, err)
		req.Header.Set(headers.ContentEncoding.String(), "gzip")
		resp, err := apiClient.Raw(req)
		assert.Nil(tx, err)
		assert.Equal(tx, http.StatusAccepted, resp.StatusCode)
	})

	t.Run("Gzip input with incorrect compression header", func(tx *testing.T) {
		var (
			body bytes.Buffer
			gw   = gzip.NewWriter(&body)
		)
		jsonInput := loader.GetReader("v6/ingest/gpos.json")
		defer jsonInput.Close()
		n, err := io.Copy(gw, jsonInput)
		assert.Nil(tx, err)
		assert.NotEqual(tx, 0, n)
		assert.Nil(tx, gw.Close())
		req, err := apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, io.NopCloser(&body))
		assert.Nil(tx, err)
		req.Header.Set(headers.ContentEncoding.String(), "deflate")
		resp, err := apiClient.Raw(req)
		assert.Nil(tx, err)
		assert.Equal(tx, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("gzip input with missing compression header", func(tx *testing.T) {
		var (
			body bytes.Buffer
			gw   = gzip.NewWriter(&body)
		)
		jsonInput := loader.GetReader("v6/ingest/groups.json")
		defer jsonInput.Close()
		n, err := io.Copy(gw, jsonInput)
		assert.Nil(tx, err)
		assert.NotEqual(tx, 0, n)
		assert.Nil(tx, gw.Close())
		req, err := apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, io.NopCloser(&body))
		assert.Nil(tx, err)
		resp, err := apiClient.Raw(req)
		assert.Nil(tx, err)
		assert.Equal(tx, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("unsupported compression type", func(tx *testing.T) {
		jsonInput := loader.GetReader("v6/ingest/containers.json")
		defer jsonInput.Close()
		req, err := apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, jsonInput)
		assert.Nil(tx, err)
		req.Header.Set(headers.ContentEncoding.String(), "br")
		resp, err := apiClient.Raw(req)
		assert.Nil(tx, err)
		assert.Equal(tx, http.StatusUnsupportedMediaType, resp.StatusCode)
	})

	t.Run("not valid json", func(tx *testing.T) {
		jsonInput := loader.GetReader("v6/ingest/jker.png")
		defer jsonInput.Close()
		req, err := apiClient.NewRequest(http.MethodPost, jobEndpoint, nil, jsonInput)
		assert.Nil(tx, err)
		resp, err := apiClient.Raw(req)
		assert.Nil(tx, err)
		assert.Equal(tx, http.StatusBadRequest, resp.StatusCode)
	})
}

func Test_FileUploadWorkFlowVersion5(t *testing.T) {
	testCtx := integration.NewFOSSContext(t)

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
	testCtx := integration.NewFOSSContext(t)

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
	testCtx.AssertIngest(fixtures.IngestAssertionsv6)
}

func Test_FileUploadVersion6AllOptionADCS(t *testing.T) {
	testCtx := integration.NewFOSSContext(t)

	testCtx.SendFileIngest([]string{
		"v6/all/aiacas.json",
		"v6/all/certtemplates.json",
		"v6/all/computers.json",
		"v6/all/containers.json",
		"v6/all/domains.json",
		"v6/all/enterprisecas.json",
		"v6/all/gpos.json",
		"v6/all/groups.json",
		"v6/all/ntauthstores.json",
		"v6/all/ous.json",
		"v6/all/rootcas.json",
		"v6/all/users.json",
	})

	testCtx.AssertIngest(fixtures.IngestADCSAssertions)
}

func Test_CompressedFileUploadWorkFlowVersion5(t *testing.T) {
	testCtx := integration.NewFOSSContext(t)

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
	testCtx := integration.NewFOSSContext(t)

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
	testCtx.AssertIngest(fixtures.IngestAssertionsv6)
}
