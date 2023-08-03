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

package api

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/specterops/bloodhound/src/auth"
	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/errors"
)

func TestTeeNilBody(t *testing.T) {
	reader := io.Reader(nil)
	outA, outB := &bytes.Buffer{}, &bytes.Buffer{}

	err := tee(reader, outA, outB)
	require.Nil(t, err)
	require.Nil(t, outA.Bytes())
	require.Nil(t, outB.Bytes())
}

func TestTeeReadError(t *testing.T) {
	reader := iotest.ErrReader(errors.New("custom error"))
	outA, outB := &bytes.Buffer{}, &bytes.Buffer{}

	err := tee(reader, outA, outB)
	require.Error(t, err)
}

func TestTeeSuccess(t *testing.T) {
	reader := strings.NewReader("Hello world")
	outA, outB := &bytes.Buffer{}, &bytes.Buffer{}

	err := tee(reader, outA, outB)
	require.Nil(t, err)

	expected := "Hello world"
	require.Equal(t, expected, outA.String())
	require.Equal(t, expected, outB.String())
}

func TestSignRequestValuesUnsupportedHMACMethod(t *testing.T) {
	datetime := time.Now().Format(time.RFC3339)
	reader := strings.NewReader("Hello world")

	_, _, err := GenerateRequestSignature("token", datetime, "unsupportedHmacMethod", "GET", "www.foo.bar", reader)
	require.Error(t, err)
	require.Equal(t, "unsupported HMAC method: unsupportedHmacMethod", err.Error())
}

func TestSignRequestTeeFailure(t *testing.T) {
	testFailure := "tee failed to read"
	reader := iotest.ErrReader(errors.New(testFailure))
	datetime := time.Now().Format(time.RFC3339)

	_, _, err := GenerateRequestSignature("token", datetime, auth.HMAC_SHA2_256, "GET", "www.foo.bar", reader)
	require.Error(t, err)
	require.Equal(t, testFailure, err.Error())
}

func TestSignRequestValuesSuccess(t *testing.T) {
	datetime := time.Now().Format(time.RFC3339)
	reader := strings.NewReader("Hello world")

	readerBody, signature, err := GenerateRequestSignature("token", datetime, auth.HMAC_SHA2_256, "GET", "www.foo.bar", reader)
	require.Nil(t, err)
	require.Equal(t, readerBody, readerBody)
	require.NotNil(t, signature)
}
