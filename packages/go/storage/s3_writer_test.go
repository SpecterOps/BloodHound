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
	"encoding/xml"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/retry"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/require"
)

type completedUploadPart struct {
	PartNumber    int32  `xml:"PartNumber"`
	ETag          string `xml:"ETag"`
	ChecksumCRC32 string `xml:"ChecksumCRC32"`
}

func readCompletedUploadParts(t *testing.T, body []byte) []completedUploadPart {
	t.Helper()

	var completion struct {
		Parts []completedUploadPart `xml:"Part"`
	}

	require.NoError(t, xml.Unmarshal(body, &completion))
	return completion.Parts
}

func TestS3PendingFileSingleRequestCommit(t *testing.T) {
	t.Parallel()

	type testData struct {
		name        string
		fileName    string
		content     []byte
		contentType string
		metadata    map[string]string
		checksum    string
	}

	tests := []testData{
		{
			name:        "writes content and headers",
			fileName:    "file.json",
			content:     []byte(`{"ok":true}`),
			contentType: "application/json",
			metadata:    map[string]string{"source": "test"},
			checksum:    "p9RfkA==",
		},
		{
			name:     "commits empty file",
			fileName: "empty",
			content:  []byte{},
			checksum: "AAAAAA==",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			client := &testHTTPClient{responses: []testHTTPResponse{{statusCode: http.StatusOK}}}
			store := newTestStore(client)
			options := WriteOptions{
				ContentType: testCase.contentType,
				Metadata:    testCase.metadata,
			}
			pendingFile, err := store.BeginWrite(context.Background(), testCase.fileName, options)
			require.NoError(t, err)
			require.Empty(t, client.requests)
			if len(testCase.content) > 0 {
				_, err = pendingFile.Write(testCase.content)
				require.NoError(t, err)
			}

			// Act
			committedName, err := pendingFile.Commit()

			// Assert
			require.NoError(t, err)
			require.Equal(t, testCase.fileName, committedName)
			require.Len(t, client.requests, 1)
			require.Equal(t, http.MethodPut, client.requests[0].Method)
			require.Equal(t, string(testCase.content), string(client.requestBodies[0]))
			require.Equal(t, "CRC32", client.requests[0].Header.Get("X-Amz-Sdk-Checksum-Algorithm"))
			require.Equal(t, testCase.checksum, client.requests[0].Header.Get("X-Amz-Checksum-Crc32"))
			if testCase.contentType != "" {
				require.Equal(t, testCase.contentType, client.requests[0].Header.Get("Content-Type"))
			}
			if source, exists := testCase.metadata["source"]; exists {
				require.Equal(t, source, client.requests[0].Header.Get("X-Amz-Meta-Source"))
			}

			repeatedName, err := pendingFile.Commit()
			require.NoError(t, err)
			require.Equal(t, committedName, repeatedName)
			_, err = pendingFile.Write([]byte("after commit"))
			require.ErrorIs(t, err, io.ErrClosedPipe)
			require.Len(t, client.requests, 1)
		})
	}
}

func TestS3PendingFileMultipartCommit(t *testing.T) {
	t.Parallel()

	// Arrange
	client := &testHTTPClient{
		responses: []testHTTPResponse{
			{
				statusCode: http.StatusOK,
				body:       `<InitiateMultipartUploadResult><UploadId>upload</UploadId></InitiateMultipartUploadResult>`,
			},
			{
				statusCode: http.StatusOK,
				headers:    map[string]string{"ETag": `"part-etag"`},
				body:       `<ETag>"part-etag"</ETag>`,
			},
			{
				statusCode: http.StatusOK,
				headers:    map[string]string{"ETag": `"last-part"`},
				body:       `<ETag>"last-part"</ETag>`,
			},
			{
				statusCode: http.StatusOK,
				body:       `<CompleteMultipartUploadResult><ETag>"final"</ETag></CompleteMultipartUploadResult>`,
			},
		},
	}
	store := newTestStore(client)
	pendingFile, err := store.BeginWrite(context.Background(), "large.zip", WriteOptions{
		ContentType: "application/zip",
		Metadata:    map[string]string{"source": "test"},
	})
	require.NoError(t, err)
	pendingFile.(*s3PendingFile).partSize = 5

	_, err = pendingFile.Write([]byte("12345"))
	require.NoError(t, err)
	require.Len(t, client.requests, 0)
	_, err = pendingFile.Write([]byte("6"))
	require.NoError(t, err)
	require.Len(t, client.requests, 2)

	// Act
	_, err = pendingFile.Commit()

	// Assert
	require.NoError(t, err)
	require.Len(t, client.requests, 4)
	require.Contains(t, client.requests[3].URL.RawQuery, "uploadId=upload")
	require.Equal(t, []completedUploadPart{
		{
			PartNumber:    1,
			ETag:          `"part-etag"`,
			ChecksumCRC32: "y/U6HA==",
		},
		{
			PartNumber:    2,
			ETag:          `"last-part"`,
			ChecksumCRC32: "Hbh6FA==",
		},
	}, readCompletedUploadParts(t, client.requestBodies[3]))
	require.Equal(t, "CRC32", client.requests[0].Header.Get("X-Amz-Checksum-Algorithm"))
	require.Equal(t, "y/U6HA==", client.requests[1].Header.Get("X-Amz-Checksum-Crc32"))
	require.Equal(t, "Hbh6FA==", client.requests[2].Header.Get("X-Amz-Checksum-Crc32"))
	require.Equal(t, []byte("12345"), client.requestBodies[1])
	require.Equal(t, []byte("6"), client.requestBodies[2])
	require.Equal(t, "application/zip", client.requests[0].Header.Get("Content-Type"))
	require.Equal(t, "test", client.requests[0].Header.Get("X-Amz-Meta-Source"))
}

func TestS3PendingFileConditionalCommitRejectsExistingObject(t *testing.T) {
	t.Parallel()

	// Arrange
	client := &testHTTPClient{
		responses: []testHTTPResponse{
			{statusCode: http.StatusOK},
			{
				statusCode: http.StatusPreconditionFailed,
				body:       `<Error><Code>PreconditionFailed</Code></Error>`,
			},
		},
	}
	store := newTestStore(client)
	firstFile, err := store.BeginWrite(context.Background(), "same", WriteOptions{FailIfExists: true})
	require.NoError(t, err)
	secondFile, err := store.BeginWrite(context.Background(), "same", WriteOptions{FailIfExists: true})
	require.NoError(t, err)

	// Act
	_, err = firstFile.Commit()
	require.NoError(t, err)
	_, secondCommitErr := secondFile.Commit()

	// Assert
	require.ErrorIs(t, secondCommitErr, fs.ErrExist)
	require.Equal(t, "*", client.requests[0].Header.Get("If-None-Match"))
	require.Equal(t, "*", client.requests[1].Header.Get("If-None-Match"))
	require.NoError(t, secondFile.Abort())
}

func TestS3PendingFileConditionalMultipartCommitAbortsUpload(t *testing.T) {
	t.Parallel()

	// Arrange
	client := &testHTTPClient{
		responses: []testHTTPResponse{
			{
				statusCode: http.StatusOK,
				body:       `<InitiateMultipartUploadResult><UploadId>upload</UploadId></InitiateMultipartUploadResult>`,
			},
			{
				statusCode: http.StatusOK,
				headers:    map[string]string{"ETag": `"part-etag"`},
			},
			{
				statusCode: http.StatusOK,
				headers:    map[string]string{"ETag": `"last-part"`},
			},
			{
				statusCode: http.StatusPreconditionFailed,
				body:       `<Error><Code>PreconditionFailed</Code></Error>`,
			},
			{statusCode: http.StatusNoContent},
		},
	}
	store := newTestStore(client)
	pendingFile, err := store.BeginWrite(context.Background(), "appeared", WriteOptions{FailIfExists: true})
	require.NoError(t, err)
	pendingFile.(*s3PendingFile).partSize = 5
	_, err = pendingFile.Write([]byte("123456"))
	require.NoError(t, err)

	// Act
	_, err = pendingFile.Commit()

	// Assert
	require.ErrorIs(t, err, fs.ErrExist)
	require.Equal(t, "*", client.requests[3].Header.Get("If-None-Match"))
	require.Equal(t, []completedUploadPart{
		{
			PartNumber:    1,
			ETag:          `"part-etag"`,
			ChecksumCRC32: "y/U6HA==",
		},
		{
			PartNumber:    2,
			ETag:          `"last-part"`,
			ChecksumCRC32: "Hbh6FA==",
		},
	}, readCompletedUploadParts(t, client.requestBodies[3]))
	require.NoError(t, pendingFile.Abort())
	require.Contains(t, client.requests[4].URL.RawQuery, "uploadId=upload")
}

func TestS3PendingFileWriteFailureIsStickyAndAbortUsesDetachedContext(t *testing.T) {
	t.Parallel()

	// Arrange
	client := &testHTTPClient{
		responses: []testHTTPResponse{
			{
				statusCode: http.StatusOK,
				body:       `<InitiateMultipartUploadResult><UploadId>upload</UploadId></InitiateMultipartUploadResult>`,
			},
			{
				statusCode: http.StatusInternalServerError,
				body:       `<Error><Code>InternalError</Code></Error>`,
			},
			{statusCode: http.StatusNoContent},
		},
	}
	store := newTestStore(client)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pendingFile, err := store.BeginWrite(ctx, "failed", WriteOptions{})
	require.NoError(t, err)
	pendingFile.(*s3PendingFile).partSize = 5

	// Act
	_, firstErr := pendingFile.Write([]byte("123456"))
	cancel()
	_, repeatedErr := pendingFile.Write([]byte("different"))
	_, commitErr := pendingFile.Commit()

	// Assert
	require.ErrorIs(t, firstErr, ErrWriteFailed)
	require.ErrorIs(t, repeatedErr, firstErr)
	require.ErrorIs(t, commitErr, firstErr)
	require.Len(t, client.requests, 2)
	require.NoError(t, pendingFile.Abort())
	require.Len(t, client.requests, 3)
	require.Equal(t, http.MethodDelete, client.requests[2].Method)
	require.NoError(t, client.contextErrors[2])
	require.NoError(t, pendingFile.Abort())
	require.Len(t, client.requests, 3)
}

func TestS3PendingFileCancellationDuringPartUploadStillAborts(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := &testHTTPClient{
		responses: []testHTTPResponse{
			{
				statusCode: http.StatusOK,
				body:       `<InitiateMultipartUploadResult><UploadId>upload</UploadId></InitiateMultipartUploadResult>`,
			},
			{statusCode: http.StatusOK},
			{statusCode: http.StatusNoContent},
		},
		requestErrors: map[int]error{1: context.Canceled},
		onRequest: func(request *http.Request) {
			if request.Method == http.MethodPut {
				cancel()
			}
		},
	}
	store := newTestStore(client)
	pendingFile, err := store.BeginWrite(ctx, "cancel-upload", WriteOptions{})
	require.NoError(t, err)
	pendingFile.(*s3PendingFile).partSize = 5

	// Act
	_, writeErr := pendingFile.Write([]byte("123456"))

	// Assert
	require.ErrorIs(t, writeErr, context.Canceled)
	require.ErrorIs(t, writeErr, ErrWriteFailed)
	require.NoError(t, pendingFile.Abort())
	require.Len(t, client.requests, 3)
	require.Equal(t, http.MethodDelete, client.requests[2].Method)
	require.NoError(t, client.contextErrors[2])
}

func TestS3PendingFileRetriesPartWithIdenticalBytes(t *testing.T) {
	t.Parallel()

	// Arrange
	client := &testHTTPClient{
		responses: []testHTTPResponse{
			{
				statusCode: http.StatusOK,
				body:       `<InitiateMultipartUploadResult><UploadId>upload</UploadId></InitiateMultipartUploadResult>`,
			},
			{
				statusCode: http.StatusInternalServerError,
				body:       `<Error><Code>InternalError</Code></Error>`,
			},
			{
				statusCode: http.StatusOK,
				headers:    map[string]string{"ETag": `"part-etag"`},
			},
			{
				statusCode: http.StatusOK,
				headers:    map[string]string{"ETag": `"last-part"`},
			},
			{
				statusCode: http.StatusOK,
				body:       `<CompleteMultipartUploadResult/>`,
			},
		},
	}
	retryer := retry.NewStandard(func(options *retry.StandardOptions) {
		options.MaxAttempts = 2
	})
	store := newTestStoreWithRetryer(client, retryer)
	pendingFile, err := store.BeginWrite(context.Background(), "retry", WriteOptions{})
	require.NoError(t, err)
	pendingFile.(*s3PendingFile).partSize = 5
	_, err = pendingFile.Write([]byte("123456"))
	require.NoError(t, err)

	// Act
	_, err = pendingFile.Commit()

	// Assert
	require.NoError(t, err)
	require.Equal(t, []byte("12345"), client.requestBodies[1])
	require.Equal(t, client.requestBodies[1], client.requestBodies[2])
	require.Equal(t, []completedUploadPart{
		{
			PartNumber:    1,
			ETag:          `"part-etag"`,
			ChecksumCRC32: "y/U6HA==",
		},
		{
			PartNumber:    2,
			ETag:          `"last-part"`,
			ChecksumCRC32: "Hbh6FA==",
		},
	}, readCompletedUploadParts(t, client.requestBodies[4]))
}

func TestS3PendingFileRejectsWritePastPartLimit(t *testing.T) {
	t.Parallel()

	// Arrange
	pendingFile := &s3PendingFile{
		ctx:      context.Background(),
		name:     "limited",
		uploadID: "upload",
		partSize: 5,
		parts:    make([]s3types.CompletedPart, maximumPendingParts),
	}
	_, err := pendingFile.buffer.Write([]byte("12345"))
	require.NoError(t, err)

	// Act
	_, err = pendingFile.Write([]byte("6"))

	// Assert
	require.ErrorIs(t, err, ErrWriteFailed)
	require.Contains(t, err.Error(), "10000 part limit")
}

func TestS3PendingFileAbortReturnsRequestFailure(t *testing.T) {
	t.Parallel()

	// Arrange
	client := &testHTTPClient{
		responses: []testHTTPResponse{
			{
				statusCode: http.StatusInternalServerError,
				body:       `<Error><Code>InternalError</Code></Error>`,
			},
		},
	}
	store := newTestStore(client)
	pendingFile, err := store.BeginWrite(context.Background(), "abort-failure", WriteOptions{})
	require.NoError(t, err)
	pendingFile.(*s3PendingFile).uploadID = "upload"

	// Act
	err = pendingFile.Abort()

	// Assert
	require.Error(t, err)
	require.Len(t, client.requests, 1)
	require.Equal(t, http.MethodDelete, client.requests[0].Method)
	require.Contains(t, client.requests[0].URL.RawQuery, "uploadId=upload")
}

func TestS3PendingFileUncertainCompletionIsTerminalAndCleansOwnedUpload(t *testing.T) {
	t.Parallel()

	// Arrange
	completionErr := errors.New("connection lost after publish")
	client := &testHTTPClient{
		responses: []testHTTPResponse{
			{
				statusCode: http.StatusOK,
				body:       `<InitiateMultipartUploadResult><UploadId>upload</UploadId></InitiateMultipartUploadResult>`,
			},
			{
				statusCode: http.StatusOK,
				headers:    map[string]string{"ETag": `"part-etag"`},
			},
			{
				statusCode: http.StatusOK,
				headers:    map[string]string{"ETag": `"last-part"`},
			},
			{statusCode: http.StatusOK},
			{statusCode: http.StatusNoContent},
		},
		requestErrors: map[int]error{3: completionErr},
	}
	store := newTestStore(client)
	pendingFile, err := store.BeginWrite(context.Background(), "uncertain", WriteOptions{})
	require.NoError(t, err)
	pendingFile.(*s3PendingFile).partSize = 5
	_, err = pendingFile.Write([]byte("123456"))
	require.NoError(t, err)

	// Act
	_, firstCommitErr := pendingFile.Commit()
	_, repeatedCommitErr := pendingFile.Commit()

	// Assert
	require.ErrorIs(t, firstCommitErr, completionErr)
	require.ErrorIs(t, repeatedCommitErr, firstCommitErr)
	require.Len(t, client.requests, 4)
	require.NoError(t, pendingFile.Abort())
	require.Len(t, client.requests, 5)
	require.Equal(t, http.MethodDelete, client.requests[4].Method)
	require.Contains(t, client.requests[4].URL.RawQuery, "uploadId=upload")
	for _, request := range client.requests {
		require.False(t, request.Method == http.MethodDelete && request.URL.RawQuery == "")
	}
}

func TestUploadPartSize(t *testing.T) {
	t.Parallel()

	type testData struct {
		name      string
		size      int64
		wantSize  int64
		wantError bool
	}

	tests := []testData{
		{
			name:     "unknown size uses minimum",
			size:     0,
			wantSize: 8 * 1024 * 1024,
		},
		{
			name:     "large object increases part size",
			size:     100 * 1024 * 1024 * 1024,
			wantSize: 16 * 1024 * 1024,
		},
		{
			name:      "too many parts is rejected",
			size:      5 * 1024 * 1024 * 1024 * 10001,
			wantError: true,
		},
		{
			name:      "maximum integer is rejected",
			size:      int64(^uint64(0) >> 1),
			wantError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Act
			actualSize, err := uploadPartSize(testCase.size)

			// Assert
			if testCase.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), "part")
				return
			}

			require.NoError(t, err)
			require.Equal(t, testCase.wantSize, actualSize)
		})
	}
}
