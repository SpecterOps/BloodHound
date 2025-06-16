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
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSignRequestValuesUnsupportedHMACMethod(t *testing.T) {
	t.Parallel()
	datetime := time.Now().Format(time.RFC3339)
	reader := strings.NewReader("Hello world")

	_, err := NewRequestSignature(context.Background(), nil, "token", datetime, "GET", "www.foo.bar", reader)
	require.Error(t, err)
	require.ErrorContains(t, err, "hasher must not be nil")
}

func TestSignRequestTeeFailure(t *testing.T) {
	t.Parallel()
	testFailure := "tee failed to read"
	reader := iotest.ErrReader(errors.New(testFailure))
	datetime := time.Now().Format(time.RFC3339)

	_, err := NewRequestSignature(context.Background(), sha256.New, "token", datetime, "GET", "www.foo.bar", reader)
	require.Error(t, err)
	require.ErrorContains(t, err, testFailure)
}

func TestSignRequestValuesSuccess(t *testing.T) {
	t.Parallel()
	datetime := time.Now().Format(time.RFC3339)
	reader := strings.NewReader("Hello world")

	signature, err := NewRequestSignature(context.Background(), sha256.New, "token", datetime, "GET", "www.foo.bar", reader)
	require.Nil(t, err)
	require.NotNil(t, signature)
}
