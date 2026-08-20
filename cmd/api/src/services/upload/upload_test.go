// Copyright 2024 Specter Ops, Inc.
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

package upload

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/packages/go/storage"
	storagemocks "github.com/specterops/bloodhound/packages/go/storage/mocks"
	"github.com/specterops/chow/pkg/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func buildValidator(t *testing.T, expectedContent string, validationErr error) FileValidator {
	t.Helper()

	return func(src io.Reader, dst io.Writer) error {
		teeReader := io.TeeReader(src, dst)
		content, err := io.ReadAll(teeReader)
		if err != nil {
			return err
		}
		if string(content) != expectedContent {
			return fmt.Errorf("expected content %q, got %q", expectedContent, string(content))
		}

		return validationErr
	}
}

func TestAllowedFileUploadTypes(t *testing.T) {
	expected := []string{
		"application/json",
		"application/zip",
		"application/x-zip-compressed",
		"application/zip-compressed",
	}

	actual := AllowedFileUploadTypes()
	require.Equal(t, expected, actual)

	actual[0] = "mutated"
	require.Equal(t, expected, AllowedFileUploadTypes())
}

func TestAllowedZipFileUploadTypes(t *testing.T) {
	expected := []string{
		"application/zip",
		"application/x-zip-compressed",
		"application/zip-compressed",
	}

	actual := AllowedZipFileUploadTypes()
	require.Equal(t, expected, actual)

	actual[0] = "mutated"
	require.Equal(t, expected, AllowedZipFileUploadTypes())
}

func TestWriteAndValidateZip(t *testing.T) {
	t.Run("valid zip file is ok", func(t *testing.T) {
		var writer bytes.Buffer

		file, err := os.Open("../../test/fixtures/fixtures/goodzip.zip")
		require.NoError(t, err)
		t.Cleanup(func() { _ = file.Close() })

		err = WriteAndValidateZip(file, &writer)
		require.NoError(t, err)
		require.NotEmpty(t, writer.Bytes())
	})

	t.Run("invalid bytes causes error", func(t *testing.T) {
		var writer bytes.Buffer
		badZip := strings.NewReader("123123")

		err := WriteAndValidateZip(badZip, &writer)
		assert.ErrorIs(t, err, ErrInvalidZipFile)
	})
}

func TestWriteAndValidateJSON(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		expectedOutput []byte
		expectedError  error
	}{
		{
			name:           "UTF-8 without BOM",
			input:          []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`),
			expectedOutput: []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`),
		},
		{
			name:           "UTF-8 with BOM",
			input:          append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`)...),
			expectedOutput: []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`),
		},
		{
			name:           "UTF-16BE with BOM",
			input:          []byte{0xFE, 0xFF, 0x00, 0x7B, 0x00, 0x22, 0x00, 0x6D, 0x00, 0x65, 0x00, 0x74, 0x00, 0x61, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x7B, 0x00, 0x22, 0x00, 0x74, 0x00, 0x79, 0x00, 0x70, 0x00, 0x65, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x22, 0x00, 0x64, 0x00, 0x6F, 0x00, 0x6D, 0x00, 0x61, 0x00, 0x69, 0x00, 0x6E, 0x00, 0x73, 0x00, 0x22, 0x00, 0x2C, 0x00, 0x20, 0x00, 0x22, 0x00, 0x76, 0x00, 0x65, 0x00, 0x72, 0x00, 0x73, 0x00, 0x69, 0x00, 0x6F, 0x00, 0x6E, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x34, 0x00, 0x2C, 0x00, 0x20, 0x00, 0x22, 0x00, 0x63, 0x00, 0x6F, 0x00, 0x75, 0x00, 0x6E, 0x00, 0x74, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x31, 0x00, 0x7D, 0x00, 0x2C, 0x00, 0x20, 0x00, 0x22, 0x00, 0x64, 0x00, 0x61, 0x00, 0x74, 0x00, 0x61, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x5B, 0x00, 0x7B, 0x00, 0x22, 0x00, 0x64, 0x00, 0x6F, 0x00, 0x6D, 0x00, 0x61, 0x00, 0x69, 0x00, 0x6E, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x22, 0x00, 0x65, 0x00, 0x78, 0x00, 0x61, 0x00, 0x6D, 0x00, 0x70, 0x00, 0x6C, 0x00, 0x65, 0x00, 0x2E, 0x00, 0x63, 0x00, 0x6F, 0x00, 0x6D, 0x00, 0x22, 0x00, 0x7D, 0x00, 0x5D, 0x00, 0x7D},
			expectedOutput: []byte{0x7b, 0x22, 0x6d, 0x65, 0x74, 0x61, 0x22, 0x3a, 0x20, 0x7b, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x20, 0x22, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x22, 0x2c, 0x20, 0x22, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x3a, 0x20, 0x34, 0x2c, 0x20, 0x22, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x3a, 0x20, 0x31, 0x7d, 0x2c, 0x20, 0x22, 0x64, 0x61, 0x74, 0x61, 0x22, 0x3a, 0x20, 0x5b, 0x7b, 0x22, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x22, 0x3a, 0x20, 0x22, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x22, 0x7d, 0x5d, 0x7d},
		},
		{
			name:           "Missing meta tag",
			input:          []byte(`{"data": [{"domain": "example.com"}]}`),
			expectedOutput: []byte(`{"data": [{"domain": "example.com"}]}`),
			expectedError:  payload.ErrInvalidFileConfiguration,
		},
		{
			name:           "Missing data tag",
			input:          []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}}`),
			expectedOutput: []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}}`),
			expectedError:  payload.ErrInvalidFileConfiguration,
		},
	}

	schema, err := payload.LoadSchema()
	require.NoError(t, err)

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var destination bytes.Buffer
			src := bytes.NewReader(testCase.input)

			report, err := WriteAndValidateJSON(src, &destination, schema)
			if testCase.expectedError != nil {
				require.ErrorIs(t, err, testCase.expectedError)
				require.NotEmpty(t, report.CriticalErrors)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.expectedOutput, destination.Bytes())
		})
	}
}

func TestWriteAndValidateJSON_NormalizationError(t *testing.T) {
	var destination bytes.Buffer
	src := &ErrorReader{err: errors.New("read error")}

	schema, err := payload.LoadSchema()
	require.NoError(t, err)

	_, err = WriteAndValidateJSON(src, &destination, schema)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidJSON)
}

func TestSaveIngestFileReturnsValidationReport(t *testing.T) {
	var (
		ctx             = context.Background()
		mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
		request         = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"data": [{"domain": "example.com"}]}`))
	)
	request.Header.Set("Content-Type", "application/json")

	schema, err := payload.LoadSchema()
	require.NoError(t, err)

	mockFileService.EXPECT().
		WriteTempFile(ctx, ingestFileTempPrefix(1), gomock.Any(), storage.WriteOptions{}).
		DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ storage.WriteOptions) (string, error) {
			_, _ = io.ReadAll(reader)
			return "tmp-file", nil
		})
	mockFileService.EXPECT().DeleteFile(gomock.Any(), "tmp-file").Return(nil)

	ingestTaskParams, report, err := SaveIngestFile(ctx, mockFileService, request, schema, 1)

	require.ErrorIs(t, err, ErrInvalidJSON)
	require.ErrorIs(t, err, payload.ErrInvalidFileConfiguration)
	require.Empty(t, ingestTaskParams)
	require.NotEmpty(t, report.CriticalErrors)
}

func TestUpload_WriteAndValidateFile(t *testing.T) {
	t.Parallel()

	var (
		errValidation = errors.New("validation failed")
		errWrite      = errors.New("write failed")
	)

	type expected struct {
		errIs    error
		fileName string
	}

	type testData struct {
		name          string
		tempFileName  string
		writeErr      error
		validationErr error
		expected      expected
		expectDelete  bool
	}

	tests := []testData{
		{
			name:         "writes and validates file",
			tempFileName: "prefix/tmp-file",
			expected: expected{
				fileName: "prefix/tmp-file",
			},
		},
		{
			name:          "validation error deletes temp file",
			tempFileName:  "prefix/tmp-file",
			validationErr: errValidation,
			expected: expected{
				errIs: errValidation,
			},
			expectDelete: true,
		},
		{
			name:         "write error deletes temp file",
			tempFileName: "prefix/tmp-file",
			writeErr:     errWrite,
			expected: expected{
				errIs: errWrite,
			},
			expectDelete: true,
		},
		{
			name:          "validation error takes precedence over write error",
			tempFileName:  "prefix/tmp-file",
			writeErr:      errWrite,
			validationErr: errValidation,
			expected: expected{
				errIs: errValidation,
			},
			expectDelete: true,
		},
		{
			name:         "write error without temp path does not delete empty path",
			tempFileName: "",
			writeErr:     errWrite,
			expected: expected{
				errIs: errWrite,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				ctx             = context.Background()
				mockFileService = storagemocks.NewMockFileService(gomock.NewController(t))
				validator       = buildValidator(t, "content", testCase.validationErr)
			)

			mockFileService.EXPECT().
				WriteTempFile(ctx, "prefix", gomock.Any(), storage.WriteOptions{}).
				DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ storage.WriteOptions) (string, error) {
					content, err := io.ReadAll(reader)
					if testCase.validationErr != nil {
						require.Error(t, err, testCase.validationErr)
					} else {
						require.NoError(t, err)
					}
					require.Equal(t, "content", string(content))

					return testCase.tempFileName, testCase.writeErr
				})
			if testCase.expectDelete {
				mockFileService.EXPECT().DeleteFile(gomock.Any(), testCase.tempFileName).Return(nil)
			}

			actualFileName, err := WriteAndValidateFile(ctx, mockFileService, strings.NewReader("content"), "prefix", validator)

			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Empty(t, actualFileName)
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.expected.fileName, actualFileName)
		})
	}
}

// ErrorReader is a mock reader that always returns an error.
type ErrorReader struct {
	err error
}

func (s *ErrorReader) Read([]byte) (int, error) {
	return 0, s.err
}
