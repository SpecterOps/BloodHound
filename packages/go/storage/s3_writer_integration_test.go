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

//go:build integration

package storage

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

func newIntegrationS3Store(t *testing.T) (context.Context, *Store, string) {
	t.Helper()

	bucket := strings.TrimSpace(os.Getenv("BLOODHOUND_TEST_S3_BUCKET"))
	if bucket == "" {
		t.Skip("set BLOODHOUND_TEST_S3_BUCKET to run real-S3 pending-file integration tests")
	}

	ctx := context.Background()
	awsConfig, err := config.LoadDefaultConfig(ctx)
	require.NoError(t, err)
	store := NewS3Store(bucket, "", s3.NewFromConfig(awsConfig))
	id, err := randomID()
	require.NoError(t, err)

	return ctx, store, "bed-7790-pending-file-test/" + id + "/"
}

func TestS3PendingFileSmallObjectAndConditionalCommit(t *testing.T) {
	ctx, store, prefix := newIntegrationS3Store(t)
	smallName := prefix + "small.json"
	t.Cleanup(func() { _ = store.Delete(context.Background(), smallName) })

	// Arrange
	smallFile, err := store.BeginWrite(ctx, smallName, WriteOptions{
		ContentType: "application/json",
		Metadata:    map[string]string{"purpose": "bed-7790-test"},
	})
	require.NoError(t, err)
	_, err = smallFile.Write([]byte(`{"ok":true}`))
	require.NoError(t, err)

	// Act
	_, err = smallFile.Commit()

	// Assert
	require.NoError(t, err)
	info, err := store.Stat(ctx, smallName)
	require.NoError(t, err)
	require.Equal(t, "application/json", info.ContentType)
	reader, _, err := store.Get(ctx, smallName)
	require.NoError(t, err)
	smallData, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	require.Equal(t, `{"ok":true}`, string(smallData))
	bucketName := store.bucket
	attributes, err := store.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: &bucketName, Key: &smallName})
	require.NoError(t, err)
	require.Equal(t, "bed-7790-test", attributes.Metadata["purpose"])

	conditionalFile, err := store.BeginWrite(ctx, smallName, WriteOptions{FailIfExists: true})
	require.NoError(t, err)
	_, err = conditionalFile.Commit()
	require.ErrorIs(t, err, fs.ErrExist)
	require.NoError(t, conditionalFile.Abort())
}

func TestS3PendingFileMultipartObjectRoundTrip(t *testing.T) {
	ctx, store, prefix := newIntegrationS3Store(t)
	largeName := prefix + "large.bin"
	t.Cleanup(func() { _ = store.Delete(context.Background(), largeName) })

	// Arrange
	largeData := bytes.Repeat([]byte("bed-7790"), 2*1024*1024)
	largeFile, err := store.BeginWrite(ctx, largeName, WriteOptions{SizeHint: int64(len(largeData))})
	require.NoError(t, err)
	_, err = largeFile.Write(largeData)
	require.NoError(t, err)

	// Act
	_, err = largeFile.Commit()

	// Assert
	require.NoError(t, err)
	info, err := store.Stat(ctx, largeName)
	require.NoError(t, err)
	require.Equal(t, int64(len(largeData)), info.Size)
	reader, _, err := store.Get(ctx, largeName)
	require.NoError(t, err)
	actualData, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	require.Equal(t, largeData, actualData)
}

func TestS3PendingFileAbortDoesNotPublishMultipartObject(t *testing.T) {
	ctx, store, prefix := newIntegrationS3Store(t)
	abortName := prefix + "abort.bin"
	t.Cleanup(func() { _ = store.Delete(context.Background(), abortName) })

	// Arrange
	pendingFile, err := store.BeginWrite(ctx, abortName, WriteOptions{})
	require.NoError(t, err)
	data := append(bytes.Repeat([]byte("x"), int(defaultUploadPartSize)), 'y')
	_, err = pendingFile.Write(data)
	require.NoError(t, err)

	// Act
	err = pendingFile.Abort()

	// Assert
	require.NoError(t, err)
	_, err = store.Stat(ctx, abortName)
	require.ErrorIs(t, err, fs.ErrNotExist)
}
