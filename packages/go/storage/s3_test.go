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

package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/stretchr/testify/require"
)

type testHTTPResponse struct {
	statusCode int
	headers    map[string]string
	body       string
}

type testHTTPClient struct {
	requests  []*http.Request
	responses []testHTTPResponse
}

func (s *testHTTPClient) Do(request *http.Request) (*http.Response, error) {
	if len(s.responses) == 0 {
		return nil, fmt.Errorf("unexpected request: %s %s", request.Method, request.URL.String())
	}

	response := s.responses[0]
	s.responses = s.responses[1:]
	s.requests = append(s.requests, request)

	headers := http.Header{}
	for key, value := range response.headers {
		headers.Set(key, value)
	}

	return &http.Response{
		StatusCode:    response.statusCode,
		Header:        headers,
		Body:          io.NopCloser(strings.NewReader(response.body)),
		ContentLength: int64(len(response.body)),
		Request:       request,
	}, nil
}

func newTestStore(httpClient *testHTTPClient) *Store {
	cfg := aws.Config{
		Region: "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider(
			"test-access-key",
			"test-secret-key",
			"",
		),
		HTTPClient: httpClient,
		Retryer: func() aws.Retryer {
			return aws.NopRetryer{}
		},
	}

	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String("https://s3.test")
		options.UsePathStyle = true
	})

	return NewS3Store("test-bucket", "prefix", client)
}

func TestNormalizePath(t *testing.T) {
	t.Parallel()

	type expected struct {
		path  string
		error bool
	}

	type testData struct {
		name     string
		input    string
		expected expected
	}

	testCases := []testData{
		{
			name:  "trims whitespace",
			input: " file.json ",
			expected: expected{
				path: "file.json",
			},
		},
		{
			name:  "removes leading slash",
			input: "/dir/file.json",
			expected: expected{
				path: "dir/file.json",
			},
		},
		{
			name:  "normalizes windows separators",
			input: `dir\file.json`,
			expected: expected{
				path: "dir/file.json",
			},
		},
		{
			name:  "rejects empty path",
			input: " ",
			expected: expected{
				error: true,
			},
		},
		{
			name:  "rejects escaping path",
			input: "../file.json",
			expected: expected{
				error: true,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actualPath, err := normalizePath(testCase.input)
			if testCase.expected.error {
				require.Error(t, err)
				require.Empty(t, actualPath)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.expected.path, actualPath)
		})
	}
}

func TestDetectContentType(t *testing.T) {
	t.Parallel()

	type testData struct {
		name     string
		input    string
		expected string
	}

	testCases := []testData{
		{
			name:     "json extension",
			input:    "file.json",
			expected: "application/json",
		},
		{
			name:     "unknown extension",
			input:    "file.unknownext",
			expected: "application/octet-stream",
		},
		{
			name:     "no extension",
			input:    "file",
			expected: "application/octet-stream",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.expected, detectContentType(testCase.input))
		})
	}
}

func TestMapExistsError(t *testing.T) {
	t.Parallel()

	errWrapped := errors.New("wrapped")

	type expected struct {
		errIs error
		err   error
	}

	type testData struct {
		name     string
		input    error
		expected expected
	}

	testCases := []testData{
		{
			name: "nil error returns nil",
			expected: expected{
				err: nil,
			},
		},
		{
			name: "precondition failure maps to exists",
			input: &smithyhttp.ResponseError{
				Response: &smithyhttp.Response{Response: &http.Response{StatusCode: http.StatusPreconditionFailed}},
				Err:      errWrapped,
			},
			expected: expected{
				errIs: fs.ErrExist,
			},
		},
		{
			name: "other status returns original error",
			input: &smithyhttp.ResponseError{
				Response: &smithyhttp.Response{Response: &http.Response{StatusCode: http.StatusInternalServerError}},
				Err:      errWrapped,
			},
			expected: expected{
				errIs: errWrapped,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := mapExistsError(testCase.input)
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				return
			}

			require.Equal(t, testCase.expected.err, err)
		})
	}
}

func TestStore_PathHelpers(t *testing.T) {
	t.Parallel()

	type expected struct {
		key           string
		listPrefix    string
		logicalPath   string
		rootPath      bool
		strippedPath  string
		keyErr        bool
		listPrefixErr bool
	}

	type testData struct {
		name        string
		storePrefix string
		input       string
		keyInput    string
		stripPrefix string
		stripKey    string
		logicalKey  string
		expected    expected
	}

	testCases := []testData{
		{
			name:        "prefix is applied to keys",
			storePrefix: "prefix",
			input:       "dir/file.json",
			keyInput:    "dir/file.json",
			stripPrefix: "prefix",
			stripKey:    "prefix/dir/file.json",
			logicalKey:  "prefix/dir/file.json",
			expected: expected{
				key:          "prefix/dir/file.json",
				listPrefix:   "prefix/dir/file.json/",
				logicalPath:  "dir/file.json",
				strippedPath: "dir/file.json",
			},
		},
		{
			name:        "empty prefix uses normalized key",
			storePrefix: "",
			input:       "/dir/file.json",
			keyInput:    "/dir/file.json",
			stripKey:    "dir/file.json",
			logicalKey:  "dir/file.json",
			expected: expected{
				key:          "dir/file.json",
				listPrefix:   "dir/file.json/",
				logicalPath:  "dir/file.json",
				strippedPath: "dir/file.json",
			},
		},
		{
			name:        "root path is detected",
			storePrefix: "prefix",
			input:       "/",
			keyInput:    "/",
			expected: expected{
				rootPath:      true,
				keyErr:        true,
				listPrefixErr: true,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			store := &Store{prefix: testCase.storePrefix}

			actualKey, keyErr := store.key(testCase.keyInput)
			if testCase.expected.keyErr {
				require.Error(t, keyErr)
			} else {
				require.NoError(t, keyErr)
				require.Equal(t, testCase.expected.key, actualKey)
			}

			actualListPrefix, listPrefixErr := store.listPrefix(testCase.input)
			if testCase.expected.listPrefixErr {
				require.Error(t, listPrefixErr)
			} else {
				require.NoError(t, listPrefixErr)
				require.Equal(t, testCase.expected.listPrefix, actualListPrefix)
			}

			require.Equal(t, testCase.expected.logicalPath, store.logicalPathFromKey(testCase.logicalKey))
			require.Equal(t, testCase.expected.rootPath, isRootPath(testCase.input))
			require.Equal(t, testCase.expected.strippedPath, stripPrefix(testCase.stripPrefix, testCase.stripKey))
		})
	}
}

func TestStore_Stat(t *testing.T) {
	t.Parallel()

	lastModified := time.Date(2026, time.May, 20, 12, 30, 0, 0, time.UTC)

	type expected struct {
		errIs    error
		fileInfo FileInfo
		request  string
	}

	type testData struct {
		name      string
		fileName  string
		responses []testHTTPResponse
		expected  expected
	}

	testCases := []testData{
		{
			name:     "returns object metadata",
			fileName: "dir/file.json",
			responses: []testHTTPResponse{
				{
					statusCode: http.StatusOK,
					headers: map[string]string{
						"Content-Length": "12",
						"Content-Type":   "application/custom-json",
						"ETag":           `"etag-value"`,
						"Last-Modified":  lastModified.Format(http.TimeFormat),
					},
				},
			},
			expected: expected{
				fileInfo: FileInfo{
					Path:         "dir/file.json",
					Size:         12,
					ContentType:  "application/custom-json",
					ETag:         `"etag-value"`,
					LastModified: lastModified,
				},
				request: "/test-bucket/prefix/dir/file.json",
			},
		},
		{
			name:     "falls back to extension content type",
			fileName: "dir/file.json",
			responses: []testHTTPResponse{
				{
					statusCode: http.StatusOK,
					headers: map[string]string{
						"Content-Length": "12",
						"ETag":           `"etag-value"`,
						"Last-Modified":  lastModified.Format(http.TimeFormat),
					},
				},
			},
			expected: expected{
				fileInfo: FileInfo{
					Path:         "dir/file.json",
					Size:         12,
					ContentType:  "application/json",
					ETag:         `"etag-value"`,
					LastModified: lastModified,
				},
				request: "/test-bucket/prefix/dir/file.json",
			},
		},
		{
			name:     "not found error maps to os not exist",
			fileName: "missing.json",
			responses: []testHTTPResponse{
				{
					statusCode: http.StatusNotFound,
				},
			},
			expected: expected{
				errIs:   os.ErrNotExist,
				request: "/test-bucket/prefix/missing.json",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &testHTTPClient{responses: testCase.responses}
			store := newTestStore(httpClient)

			actualFileInfo, err := store.Stat(context.Background(), testCase.fileName)
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Empty(t, actualFileInfo)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expected.fileInfo, actualFileInfo)
			}

			require.Len(t, httpClient.requests, 1)
			require.Equal(t, http.MethodHead, httpClient.requests[0].Method)
			require.Equal(t, testCase.expected.request, httpClient.requests[0].URL.Path)
		})
	}
}

func TestStore_Exists(t *testing.T) {
	t.Parallel()

	type expected struct {
		exists          bool
		responseErrorAs bool
	}

	type testData struct {
		name      string
		fileName  string
		responses []testHTTPResponse
		expected  expected
	}

	testCases := []testData{
		{
			name:     "head success returns true",
			fileName: "file.json",
			responses: []testHTTPResponse{
				{
					statusCode: http.StatusOK,
					headers: map[string]string{
						"Content-Length": "12",
					},
				},
			},
			expected: expected{
				exists: true,
			},
		},
		{
			name:     "head not found status returns false",
			fileName: "missing.json",
			responses: []testHTTPResponse{
				{
					statusCode: http.StatusNotFound,
				},
			},
			expected: expected{
				exists: false,
			},
		},
		{
			name:     "head no such key error returns false",
			fileName: "missing.json",
			responses: []testHTTPResponse{
				{
					statusCode: http.StatusNotFound,
					body:       `<Error><Code>NoSuchKey</Code><Message>missing</Message></Error>`,
				},
			},
			expected: expected{
				exists: false,
			},
		},
		{
			name:     "service error is returned",
			fileName: "file.json",
			responses: []testHTTPResponse{
				{
					statusCode: http.StatusInternalServerError,
				},
			},
			expected: expected{
				responseErrorAs: true,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &testHTTPClient{responses: testCase.responses}
			store := newTestStore(httpClient)

			actualExists, err := store.Exists(context.Background(), testCase.fileName)
			if testCase.expected.responseErrorAs {
				var responseError *smithyhttp.ResponseError
				require.ErrorAs(t, err, &responseError)
				require.False(t, actualExists)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.expected.exists, actualExists)
		})
	}
}

func TestStore_Get(t *testing.T) {
	t.Parallel()

	lastModified := time.Date(2026, time.May, 20, 12, 30, 0, 0, time.UTC)

	type expected struct {
		errIs    error
		fileInfo FileInfo
		content  string
	}

	type testData struct {
		name      string
		fileName  string
		responses []testHTTPResponse
		expected  expected
	}

	testCases := []testData{
		{
			name:     "returns object content and metadata",
			fileName: "dir/file.json",
			responses: []testHTTPResponse{
				{
					statusCode: http.StatusOK,
					headers: map[string]string{
						"Content-Length": "7",
						"Content-Type":   "application/json",
						"ETag":           `"etag-value"`,
						"Last-Modified":  lastModified.Format(http.TimeFormat),
					},
					body: `{"x":1}`,
				},
			},
			expected: expected{
				fileInfo: FileInfo{
					Path:         "dir/file.json",
					Size:         7,
					ContentType:  "application/json",
					ETag:         `"etag-value"`,
					LastModified: lastModified,
				},
				content: `{"x":1}`,
			},
		},
		{
			name:     "not found maps to os not exist",
			fileName: "missing.json",
			responses: []testHTTPResponse{
				{
					statusCode: http.StatusNotFound,
					body:       `<Error><Code>NoSuchKey</Code><Message>missing</Message></Error>`,
				},
			},
			expected: expected{
				errIs: os.ErrNotExist,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &testHTTPClient{responses: testCase.responses}
			store := newTestStore(httpClient)

			readCloser, actualFileInfo, err := store.Get(context.Background(), testCase.fileName)
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Nil(t, readCloser)
				require.Empty(t, actualFileInfo)
				return
			}
			defer readCloser.Close()

			actualContent, err := io.ReadAll(readCloser)
			require.NoError(t, err)
			require.Equal(t, testCase.expected.fileInfo, actualFileInfo)
			require.Equal(t, testCase.expected.content, string(actualContent))
		})
	}
}

func TestStore_List(t *testing.T) {
	t.Parallel()

	lastModified := time.Date(2026, time.May, 20, 12, 30, 0, 0, time.UTC)
	listBody := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
	<Name>test-bucket</Name>
	<Prefix>prefix/dir/</Prefix>
	<KeyCount>2</KeyCount>
	<Contents>
		<Key>prefix/dir/file-a.json</Key>
		<LastModified>%s</LastModified>
		<ETag>"etag-a"</ETag>
		<Size>10</Size>
		<StorageClass>STANDARD</StorageClass>
	</Contents>
	<Contents>
		<Key>prefix/dir/file-b.zip</Key>
		<LastModified>%s</LastModified>
		<ETag>"etag-b"</ETag>
		<Size>20</Size>
		<StorageClass>STANDARD</StorageClass>
	</Contents>
</ListBucketResult>`, lastModified.Format(time.RFC3339), lastModified.Format(time.RFC3339))

	type expected struct {
		files []FileInfo
		query map[string]string
	}

	type testData struct {
		name     string
		path     string
		options  ListOptions
		expected expected
	}

	testCases := []testData{
		{
			name: "non recursive list sets delimiter",
			path: "dir",
			expected: expected{
				files: []FileInfo{
					{
						Path:         "dir/file-a.json",
						ContentType:  "application/json",
						Size:         10,
						ETag:         `"etag-a"`,
						LastModified: lastModified,
					},
					{
						Path:         "dir/file-b.zip",
						ContentType:  "application/zip",
						Size:         20,
						ETag:         `"etag-b"`,
						LastModified: lastModified,
					},
				},
				query: map[string]string{
					"delimiter": "/",
					"prefix":    "prefix/dir/",
				},
			},
		},
		{
			name: "recursive list omits delimiter",
			path: "dir",
			options: ListOptions{
				Recursive: true,
			},
			expected: expected{
				files: []FileInfo{
					{
						Path:         "dir/file-a.json",
						ContentType:  "application/json",
						Size:         10,
						ETag:         `"etag-a"`,
						LastModified: lastModified,
					},
					{
						Path:         "dir/file-b.zip",
						ContentType:  "application/zip",
						Size:         20,
						ETag:         `"etag-b"`,
						LastModified: lastModified,
					},
				},
				query: map[string]string{
					"prefix": "prefix/dir/",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &testHTTPClient{responses: []testHTTPResponse{
				{
					statusCode: http.StatusOK,
					body:       listBody,
				},
			}}
			store := newTestStore(httpClient)

			actualFiles, err := store.List(context.Background(), testCase.path, testCase.options)
			require.NoError(t, err)
			require.Equal(t, testCase.expected.files, actualFiles)
			require.Len(t, httpClient.requests, 1)
			require.Equal(t, http.MethodGet, httpClient.requests[0].Method)

			query := httpClient.requests[0].URL.Query()
			require.Equal(t, "2", query.Get("list-type"))
			require.Equal(t, testCase.expected.query["prefix"], query.Get("prefix"))
			require.Equal(t, testCase.expected.query["delimiter"], query.Get("delimiter"))
		})
	}
}
