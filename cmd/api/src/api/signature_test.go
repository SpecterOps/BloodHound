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

package api_test

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/src/api"

	"github.com/stretchr/testify/require"
)

func TestSignRequestAtInternalError(t *testing.T) {
	reader := strings.NewReader("Hello world")
	request, err := http.NewRequest("GET", "www.foo.bar", reader)
	require.Nil(t, err)

	now := time.Now()
	err = api.SignRequestAtTime(nil, "tokenID", "token", now, request)
	require.Error(t, err)
	require.ErrorContains(t, err, "hasher must not be nil")
}

func TestSignRequestSuccess(t *testing.T) {
	request, err := http.NewRequest("GET", "www.foo.bar", strings.NewReader("Hello world"))
	require.Nil(t, err)

	err = api.SignRequest("tokenID", "token", request)
	require.Nil(t, err)

	requestDate := request.Header.Get(headers.RequestDate.String())
	require.Equal(t, time.Now().Format(time.RFC3339), requestDate)

	authorization := request.Header[headers.Authorization.String()][0]
	require.Equal(t, "bhesignature tokenID", authorization)

	signature := request.Header[headers.Signature.String()][0]
	require.NotNil(t, signature)
}

func TestSelfDestructingTempFile(t *testing.T) {
	file, err := api.NewSelfDestructingTempFile("", "test-")
	require.NoError(t, err)

	// temp file should actually exist
	_, err = os.Stat(file.Name())
	require.False(t, os.IsNotExist(err))
	require.NoError(t, err)

	_, err = file.Write([]byte("I am a teapot"))
	require.NoError(t, err)

	// Seek to beginning for reading
	_, err = file.Seek(0, io.SeekStart)
	require.NoError(t, err)

	// Should return file content
	content, err := io.ReadAll(file)
	require.NoError(t, err)
	require.Equal(t, "I am a teapot", string(content))

	// Should be deleted
	_, err = os.Stat(file.Name())
	require.True(t, os.IsNotExist(err))

	// Should be closed
	require.ErrorContains(t, file.Close(), "file already closed")
}
