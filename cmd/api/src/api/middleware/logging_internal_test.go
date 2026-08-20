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

package middleware

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/test/must"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/stretchr/testify/require"
)

func Test_signedRequestDate(t *testing.T) {
	var (
		expectedTime = time.Now()
		expectedID   = must.NewUUIDv4()
		request      = must.NewHTTPRequest(http.MethodGet, "http://example.com/", nil)
	)

	request.Header.Set(headers.Authorization.String(), "bhesignature "+expectedID.String())
	request.Header.Set(headers.RequestDate.String(), expectedTime.Format(time.RFC3339Nano))

	requestDate, hasHeader := getSignedRequestDate(request)

	require.True(t, hasHeader)
	require.Equal(t, expectedTime.Format(time.RFC3339Nano), requestDate)
}

// fullResponseWriter is a test delegate that implements http.Flusher,
// io.ReaderFrom, and http.Hijacker so the responseRecorder pass-through methods
// can be verified to forward to a capable delegate.
type fullResponseWriter struct {
	header      http.Header
	body        strings.Builder
	flushed     bool
	readFromN   int64
	hijacked    bool
	readFromErr error
}

func newFullResponseWriter() *fullResponseWriter {
	return &fullResponseWriter{header: http.Header{}}
}

func (s *fullResponseWriter) Header() http.Header         { return s.header }
func (s *fullResponseWriter) Write(p []byte) (int, error) { return s.body.Write(p) }
func (s *fullResponseWriter) WriteHeader(int)             {}
func (s *fullResponseWriter) Flush()                      { s.flushed = true }

func (s *fullResponseWriter) ReadFrom(source io.Reader) (int64, error) {
	if s.readFromErr != nil {
		return 0, s.readFromErr
	}
	written, err := io.Copy(&s.body, source)
	s.readFromN = written
	return written, err
}

func (s *fullResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	s.hijacked = true
	return nil, nil, nil
}

// plainResponseWriter implements only http.ResponseWriter so the fallback paths
// of the responseRecorder pass-through methods can be verified.
type plainResponseWriter struct {
	header http.Header
	body   strings.Builder
}

func newPlainResponseWriter() *plainResponseWriter {
	return &plainResponseWriter{header: http.Header{}}
}

func (s *plainResponseWriter) Header() http.Header         { return s.header }
func (s *plainResponseWriter) Write(p []byte) (int, error) { return s.body.Write(p) }
func (s *plainResponseWriter) WriteHeader(int)             {}

func Test_responseRecorder_Flush_ForwardsToDelegate(t *testing.T) {
	delegate := newFullResponseWriter()
	recorder := &responseRecorder{delegate: delegate}

	recorder.Flush()

	require.True(t, delegate.flushed)
}

func Test_responseRecorder_Flush_NoopWhenDelegateUnsupported(t *testing.T) {
	// A plain writer does not implement http.Flusher; Flush must not panic.
	recorder := &responseRecorder{delegate: newPlainResponseWriter()}

	require.NotPanics(t, recorder.Flush)
}

func Test_responseRecorder_ReadFrom_ForwardsToDelegate(t *testing.T) {
	var (
		delegate = newFullResponseWriter()
		recorder = &responseRecorder{delegate: delegate}
		payload  = "streamed-body"
	)

	written, err := recorder.ReadFrom(strings.NewReader(payload))

	require.NoError(t, err)
	require.Equal(t, int64(len(payload)), written)
	require.Equal(t, int64(len(payload)), delegate.readFromN)
	require.Equal(t, int64(len(payload)), recorder.bytesWritten)
	require.Equal(t, http.StatusOK, recorder.statusCode)
	require.Equal(t, payload, delegate.body.String())
}

func Test_responseRecorder_ReadFrom_FallsBackToCopy(t *testing.T) {
	var (
		delegate = newPlainResponseWriter()
		recorder = &responseRecorder{delegate: delegate}
		payload  = "copied-body"
	)

	written, err := recorder.ReadFrom(strings.NewReader(payload))

	require.NoError(t, err)
	require.Equal(t, int64(len(payload)), written)
	require.Equal(t, int64(len(payload)), recorder.bytesWritten)
	require.Equal(t, http.StatusOK, recorder.statusCode)
	require.Equal(t, payload, delegate.body.String())
}

func Test_responseRecorder_Hijack_ForwardsToDelegate(t *testing.T) {
	delegate := newFullResponseWriter()
	recorder := &responseRecorder{delegate: delegate}

	_, _, err := recorder.Hijack()

	require.NoError(t, err)
	require.True(t, delegate.hijacked)
}

func Test_responseRecorder_Hijack_UnsupportedWhenDelegateUnsupported(t *testing.T) {
	recorder := &responseRecorder{delegate: newPlainResponseWriter()}

	_, _, err := recorder.Hijack()

	require.ErrorIs(t, err, http.ErrNotSupported)
}

// Test_responseRecorder_Flush_ThroughHTTPTest verifies the recorder forwards
// Flush to a real http.ResponseWriter (httptest.ResponseRecorder implements
// http.Flusher).
func Test_responseRecorder_Flush_ThroughHTTPTest(t *testing.T) {
	httpRecorder := httptest.NewRecorder()
	recorder := &responseRecorder{delegate: httpRecorder}

	recorder.Flush()

	require.True(t, httpRecorder.Flushed)
}
