// Copyright 2026 Specter Ops, Inc.
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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponseRecorderPreservesFlusher(t *testing.T) {
	delegate := httptest.NewRecorder()
	recorder := &responseRecorder{delegate: delegate}

	flusher, ok := any(recorder).(http.Flusher)
	require.True(t, ok)

	flusher.Flush()

	require.True(t, delegate.Flushed)
}

func TestGzipResponseWriterPreservesFlusher(t *testing.T) {
	delegate := httptest.NewRecorder()
	writer := NewGzipResponseWriter(delegate)

	flusher, ok := any(writer).(http.Flusher)
	require.True(t, ok)

	flusher.Flush()

	require.True(t, delegate.Flushed)
}
