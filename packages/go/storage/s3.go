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
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
)

const (
	defaultUploadPartSize    int64 = 8 * 1024 * 1024
	defaultUploadConcurrency       = 4

	defaultMultipartCopyCutoff   int64 = 128 * 1024 * 1024
	defaultMultipartCopyPartSize int64 = 64 * 1024 * 1024
)

type Store struct {
	bucket string
	prefix string
	client *s3.Client
	// uploader has been depreciated, but its replacement, transfermanager, does not currently support copy
	// rather than having two different dependencies, for now we are just using the sdk until the replacement
	// can support all the features we need.
	uploader   *manager.Uploader
	partSize   int64
	copyCutoff int64
}

func NewS3Store(bucket, prefix string, client *s3.Client) *Store {
	// Used to help break up large uploads for Put and Copy
	uploader := manager.NewUploader(client, func(s *manager.Uploader) {
		s.PartSize = defaultUploadPartSize
		s.Concurrency = defaultUploadConcurrency
	})

	return &Store{
		bucket:     bucket,
		prefix:     strings.Trim(prefix, "/"),
		client:     client,
		uploader:   uploader,
		partSize:   defaultMultipartCopyPartSize,
		copyCutoff: defaultMultipartCopyCutoff,
	}
}

func normalizePath(name string) (string, error) {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "/")
	name = strings.ReplaceAll(name, "\\", "/")

	cleaned := path.Clean(name)
	if cleaned == "." || cleaned == "" {
		return "", errors.New("invalid empty path")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", errors.New("path escapes storage root")
	}

	return cleaned, nil
}

func s3DetectContentType(name string) string {
	ext := path.Ext(name)
	if ext == "" {
		return mediatypes.ApplicationOctetStream.String()
	}

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		return mediatypes.ApplicationOctetStream.String()
	}

	return contentType
}

func mapExistsError(err error) error {
	if err == nil {
		return nil
	}

	var httpErr *smithyhttp.ResponseError
	if errors.As(err, &httpErr) && httpErr.HTTPStatusCode() == http.StatusPreconditionFailed {
		return fmt.Errorf("%w: %w", fs.ErrExist, err)
	}
	return err
}

func mapNotFoundError(err error) error {
	if err == nil {
		return nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NotFound":
			return fmt.Errorf("%w: %w", os.ErrNotExist, err)
		}
	}

	var httpErr *smithyhttp.ResponseError
	if errors.As(err, &httpErr) && httpErr.HTTPStatusCode() == http.StatusNotFound {
		return fmt.Errorf("%w: %w", os.ErrNotExist, err)
	}

	return err
}

func (s *Store) key(name string) (string, error) {
	normalizedPath, err := normalizePath(name)
	if err != nil {
		return "", err
	}

	if s.prefix == "" {
		return normalizedPath, nil
	}
	return s.prefix + "/" + normalizedPath, nil
}

func (s *Store) Put(ctx context.Context, name string, reader io.Reader, options WriteOptions) error {
	key, err := s.key(name)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	}

	if options.FailIfExists {
		input.IfNoneMatch = aws.String("*")
	}

	if options.ContentType != "" {
		input.ContentType = aws.String(options.ContentType)
	}

	if len(options.Metadata) > 0 {
		input.Metadata = options.Metadata
	}

	_, err = s.uploader.Upload(ctx, input)
	return mapExistsError(err)
}

func (s *Store) Get(ctx context.Context, name string) (io.ReadCloser, FileInfo, error) {
	key, err := s.key(name)
	if err != nil {
		return nil, FileInfo{}, err
	}

	normalizedPath, err := normalizePath(name)
	if err != nil {
		return nil, FileInfo{}, err
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, FileInfo{}, mapNotFoundError(err)
	}

	info := FileInfo{
		Path:         normalizedPath,
		Size:         aws.ToInt64(out.ContentLength),
		ContentType:  aws.ToString(out.ContentType),
		ETag:         aws.ToString(out.ETag),
		LastModified: aws.ToTime(out.LastModified),
	}
	// Set in case S3 does not provide a content type
	if info.ContentType == "" {
		info.ContentType = s3DetectContentType(normalizedPath)
	}

	return out.Body, info, nil
}

func (s *Store) Stat(ctx context.Context, name string) (FileInfo, error) {
	key, err := s.key(name)
	if err != nil {
		return FileInfo{}, err
	}
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return FileInfo{}, mapNotFoundError(err)
	}

	normalizedPath, err := normalizePath(name)
	if err != nil {
		return FileInfo{}, err
	}

	info := FileInfo{
		Path:         normalizedPath,
		Size:         aws.ToInt64(out.ContentLength),
		ContentType:  aws.ToString(out.ContentType),
		ETag:         aws.ToString(out.ETag),
		LastModified: aws.ToTime(out.LastModified),
	}

	// Set in case S3 does not provide a content type
	if info.ContentType == "" {
		info.ContentType = s3DetectContentType(normalizedPath)
	}

	return info, nil
}

func (s *Store) Exists(ctx context.Context, name string) (bool, error) {
	_, err := s.Stat(ctx, name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (s *Store) Delete(ctx context.Context, name string) error {
	key, err := s.key(name)
	if err != nil {
		return err
	}
	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	var noSuchKey *s3types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return nil
	}
	return err
}

func isRootPath(name string) bool {
	name = strings.TrimSpace(name)
	return name == "" || name == "/"
}

func stripPrefix(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return strings.TrimPrefix(key, prefix+"/")
}

func (s *Store) logicalPathFromKey(key string) string {
	return stripPrefix(s.prefix, key)
}

func (s *Store) listPrefix(name string) (string, error) {
	key, err := s.key(name)
	if err != nil {
		return "", err
	}

	if key == "" {
		return "", nil
	}

	return strings.TrimSuffix(key, "/") + "/", nil
}

func (s *Store) List(ctx context.Context, name string, options ListOptions) ([]FileInfo, error) {
	var (
		results        []FileInfo
		normalizedName string
		keyPrefix      string
		err            error
	)

	// This is done so items at the root path can be listed
	if isRootPath(name) {
		if s.prefix != "" {
			keyPrefix = s.prefix + "/"
		} else {
			keyPrefix = ""
		}
		normalizedName = ""
	} else {
		normalizedName, err = normalizePath(name)
		if err != nil {
			return nil, err
		}

		keyPrefix, err = s.listPrefix(normalizedName)
		if err != nil {
			return nil, err
		}
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(keyPrefix),
	}

	if !options.Recursive {
		input.Delimiter = aws.String("/")
	}

	paginator := s3.NewListObjectsV2Paginator(s.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			if err = ctx.Err(); err != nil {
				return []FileInfo{}, err
			}
			if options.Limit > 0 && len(results) >= options.Limit {
				return results, nil
			}
			logicalPath := s.logicalPathFromKey(aws.ToString(obj.Key))
			if logicalPath == normalizedName {
				continue
			}
			results = append(results, FileInfo{
				Path:         logicalPath,
				ContentType:  s3DetectContentType(logicalPath),
				Size:         aws.ToInt64(obj.Size),
				ETag:         aws.ToString(obj.ETag),
				LastModified: aws.ToTime(obj.LastModified),
			})
		}
	}

	return results, nil
}

func (s *Store) copySmall(ctx context.Context, srcKey, dstKey string, options WriteOptions) error {
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(url.PathEscape(s.bucket + "/" + srcKey)),
		Key:        aws.String(dstKey),
	}

	if options.FailIfExists {
		input.IfNoneMatch = aws.String("*")
	}
	_, err := s.client.CopyObject(ctx, input)
	return mapExistsError(err)
}

func (s *Store) copyMultipart(ctx context.Context, srcKey, dstKey string, sourceSize int64, options WriteOptions) error {
	createInput := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(dstKey),
	}

	// Cannot safely emulate FailIfExists, must do an explicit pre-check, which may lead to a race condition
	if options.FailIfExists {
		exists, err := s.Exists(ctx, s.logicalPathFromKey(dstKey))
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("destination already exists: %w", fs.ErrExist)
		}
	}

	createOutput, err := s.client.CreateMultipartUpload(ctx, createInput)
	if err != nil {
		return err
	}

	completedParts := make([]s3types.CompletedPart, 0)
	uploadID := aws.ToString(createOutput.UploadId)

	defer func() {
		if err != nil {
			_, _ = s.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(s.bucket),
				Key:      aws.String(dstKey),
				UploadId: aws.String(uploadID),
			})
		}
	}()

	var (
		partNumber int32 = 1
		start      int64 = 0
	)

	for start < sourceSize {
		end := start + s.partSize - 1
		if end >= sourceSize {
			end = sourceSize - 1
		}

		rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)

		partOutput, uploadErr := s.client.UploadPartCopy(ctx, &s3.UploadPartCopyInput{
			Bucket:          aws.String(s.bucket),
			Key:             aws.String(dstKey),
			UploadId:        aws.String(uploadID),
			PartNumber:      aws.Int32(partNumber),
			CopySource:      aws.String(url.PathEscape(s.bucket + "/" + srcKey)),
			CopySourceRange: aws.String(rangeHeader),
		})
		if uploadErr != nil {
			err = uploadErr
			return err
		}

		completedParts = append(completedParts, s3types.CompletedPart{
			ETag:       partOutput.CopyPartResult.ETag,
			PartNumber: aws.Int32(partNumber),
		})

		partNumber++
		start = end + 1
	}

	sort.Slice(completedParts, func(left, right int) bool {
		return aws.ToInt32(completedParts[left].PartNumber) < aws.ToInt32(completedParts[right].PartNumber)
	})

	_, err = s.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(dstKey),
		UploadId: aws.String(uploadID),
		MultipartUpload: &s3types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	return err
}

// Copy currently honors FailIfExists for the destination object.
// Metadata and ContentType are preserved from the source object.
// If a large copy is done (> copyCutoff) we use multipart copy, which cannot
// guarentee FailIfExists due to a possible race condition. This function does
// best effort by checking to see if the file exists before the copy starts.
func (s *Store) Copy(ctx context.Context, srcName, dstName string, options WriteOptions) error {
	srcKey, err := s.key(srcName)
	if err != nil {
		return err
	}

	dstKey, err := s.key(dstName)
	if err != nil {
		return err
	}

	head, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(srcKey),
	})
	if err != nil {
		return err
	}

	sourceSize := aws.ToInt64(head.ContentLength)
	if sourceSize < s.copyCutoff {
		return s.copySmall(ctx, srcKey, dstKey, options)
	}

	return s.copyMultipart(ctx, srcKey, dstKey, sourceSize, options)

}

func (s *Store) Move(ctx context.Context, srcPath, dstPath string, options WriteOptions) error {
	if err := s.Copy(ctx, srcPath, dstPath, options); err != nil {
		return err
	}
	return s.Delete(ctx, srcPath)
}
