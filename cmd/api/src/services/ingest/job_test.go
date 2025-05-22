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

package ingest_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/src/model/ingest"
	ingestService "github.com/specterops/bloodhound/src/services/ingest"
	"github.com/stretchr/testify/assert"
)

func TestWriteAndValidateZip(t *testing.T) {
	t.Run("valid zip file is ok", func(t *testing.T) {
		writer := bytes.Buffer{}

		file, err := os.Open("../../test/fixtures/fixtures/goodzip.zip")
		assert.Nil(t, err)

		_, err = ingestService.WriteAndValidateZip(io.Reader(file), &writer)
		assert.Nil(t, err)
	})

	t.Run("invalid bytes causes error", func(t *testing.T) {
		var (
			writer = bytes.Buffer{}
			badZip = strings.NewReader("123123")
		)

		_, err := ingestService.WriteAndValidateZip(badZip, &writer)
		assert.Equal(t, err, ingest.ErrInvalidZipFile)
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
			expectedError:  nil,
		},
		{
			name:           "UTF-8 with BOM",
			input:          append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`)...),
			expectedOutput: []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"}]}`),
			expectedError:  nil,
		},
		{
			name:           "UTF-16BE with BOM",
			input:          []byte{0xFE, 0xFF, 0x00, 0x7B, 0x00, 0x22, 0x00, 0x6D, 0x00, 0x65, 0x00, 0x74, 0x00, 0x61, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x7B, 0x00, 0x22, 0x00, 0x74, 0x00, 0x79, 0x00, 0x70, 0x00, 0x65, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x22, 0x00, 0x64, 0x00, 0x6F, 0x00, 0x6D, 0x00, 0x61, 0x00, 0x69, 0x00, 0x6E, 0x00, 0x73, 0x00, 0x22, 0x00, 0x2C, 0x00, 0x20, 0x00, 0x22, 0x00, 0x76, 0x00, 0x65, 0x00, 0x72, 0x00, 0x73, 0x00, 0x69, 0x00, 0x6F, 0x00, 0x6E, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x34, 0x00, 0x2C, 0x00, 0x20, 0x00, 0x22, 0x00, 0x63, 0x00, 0x6F, 0x00, 0x75, 0x00, 0x6E, 0x00, 0x74, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x31, 0x00, 0x7D, 0x00, 0x2C, 0x00, 0x20, 0x00, 0x22, 0x00, 0x64, 0x00, 0x61, 0x00, 0x74, 0x00, 0x61, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x5B, 0x00, 0x7B, 0x00, 0x22, 0x00, 0x64, 0x00, 0x6F, 0x00, 0x6D, 0x00, 0x61, 0x00, 0x69, 0x00, 0x6E, 0x00, 0x22, 0x00, 0x3A, 0x00, 0x20, 0x00, 0x22, 0x00, 0x65, 0x00, 0x78, 0x00, 0x61, 0x00, 0x6D, 0x00, 0x70, 0x00, 0x6C, 0x00, 0x65, 0x00, 0x2E, 0x00, 0x63, 0x00, 0x6F, 0x00, 0x6D, 0x00, 0x22, 0x00, 0x7D, 0x00, 0x5D, 0x00, 0x7D},
			expectedOutput: []byte{0x7b, 0x22, 0x6d, 0x65, 0x74, 0x61, 0x22, 0x3a, 0x20, 0x7b, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x20, 0x22, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x73, 0x22, 0x2c, 0x20, 0x22, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x3a, 0x20, 0x34, 0x2c, 0x20, 0x22, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x3a, 0x20, 0x31, 0x7d, 0x2c, 0x20, 0x22, 0x64, 0x61, 0x74, 0x61, 0x22, 0x3a, 0x20, 0x5b, 0x7b, 0x22, 0x64, 0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x22, 0x3a, 0x20, 0x22, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x22, 0x7d, 0x5d, 0x7d},
			expectedError:  nil,
		},
		{
			name:           "Missing meta tag",
			input:          []byte(`{"data": [{"domain": "example.com"}]}`),
			expectedOutput: []byte(`{"data": [{"domain": "example.com"}]}`),
			expectedError:  ingest.ErrMetaTagNotFound,
		},
		{
			name:           "Missing data tag",
			input:          []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}}`),
			expectedOutput: []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}}`),
			expectedError:  ingest.ErrDataTagNotFound,
		},
		// NOTE: this test discovers a bug where invalid JSON files are not being invalidated due to the current
		// implemenation of ValidateMetaTag of decoding each token.
		// {
		// 	name:           "Invalid JSON",
		// 	input:          []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"`),
		// 	expectedOutput: []byte(`{"meta": {"type": "domains", "version": 4, "count": 1}, "data": [{"domain": "example.com"`),
		// 	expectedError:  ErrInvalidJSON,
		// },
	}

	schema, err := ingestService.LoadIngestSchema()
	if err != nil {
		assert.Fail(t, fmt.Sprintf("failed to load ingest schema: %s", err))
	}

	v := ingestService.NewIngestValidator(schema)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := bytes.NewReader(tt.input)
			dst := &bytes.Buffer{}

			_, err := v.WriteAndValidateJSON(src, dst)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedOutput, dst.Bytes())
		})
	}
}

func TestWriteAndValidateJSON_NormalizationError(t *testing.T) {
	src := &ErrorReader{err: errors.New("read error")}
	dst := &bytes.Buffer{}

	schema, err := ingestService.LoadIngestSchema()
	if err != nil {
		assert.Fail(t, fmt.Sprintf("failed to load ingest schema: %s", err))
	}

	v := ingestService.NewIngestValidator(schema)

	_, err = v.WriteAndValidateJSON(src, dst)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ingestService.ErrInvalidJSON)
}

// ErrorReader is a mock reader that always returns an error
type ErrorReader struct {
	err error
}

func (er *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, er.err
}

// func TestHasJobsWaitingForAnalysis(t *testing.T) {
// 	var (
// 		mockCtrl = gomock.NewController(t)
// 		dbMock   = mocks.NewMockDatabase(mockCtrl)
// 	)

// 	defer mockCtrl.Finish()

// 	t.Run("Has Jobs Waiting for Analysis", func(t *testing.T) {
// 		dbMock.EXPECT().GetIngestJobsWithStatus(gomock.Any(), model.JobStatusAnalyzing).Return([]model.IngestJob{{}}, nil)

// 		hasJobs, err := ingestService.HasIngestJobsWaitingForAnalysis()

// 		require.True(t, hasJobs)
// 		require.Nil(t, err)
// 	})

// 	t.Run("Has No Jobs Waiting for Analysis", func(t *testing.T) {
// 		dbMock.EXPECT().GetIngestJobsWithStatus(gomock.Any(), model.JobStatusAnalyzing).Return([]model.IngestJob{}, nil)

// 		hasJobs, err := ingestService.HasIngestJobsWaitingForAnalysis(context.Background(), dbMock)

// 		require.False(t, hasJobs)
// 		require.Nil(t, err)
// 	})
// }

// func TestFailAnalyzedIngestJobs(t *testing.T) {
// 	const jobID int64 = 1

// 	var (
// 		mockCtrl = gomock.NewController(t)
// 		dbMock   = mocks.NewMockDatabase(mockCtrl)
// 	)

// 	defer mockCtrl.Finish()

// 	t.Run("Fail Analyzed Ingest Jobs", func(t *testing.T) {
// 		dbMock.EXPECT().GetIngestJobsWithStatus(gomock.Any(), model.JobStatusAnalyzing).Return([]model.IngestJob{{
// 			BigSerial: model.BigSerial{
// 				ID: jobID,
// 			},
// 			Status: model.JobStatusAnalyzing,
// 		}}, nil)

// 		dbMock.EXPECT().UpdateIngestJob(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, job model.IngestJob) error {
// 			require.Equal(t, model.JobStatusFailed, job.Status)
// 			return nil
// 		})

// 		ingestService.FailAnalyzedIngestJobs(context.Background(), dbMock)
// 	})
// }

// func TestCompleteAnalyzedIngestJobs(t *testing.T) {
// 	const jobID int64 = 1

// 	var (
// 		mockCtrl = gomock.NewController(t)
// 		dbMock   = mocks.NewMockDatabase(mockCtrl)
// 	)

// 	defer mockCtrl.Finish()

// 	t.Run("Complete Analyzed Ingest Jobs", func(t *testing.T) {
// 		dbMock.EXPECT().GetIngestJobsWithStatus(gomock.Any(), model.JobStatusAnalyzing).Return([]model.IngestJob{{
// 			BigSerial: model.BigSerial{
// 				ID: jobID,
// 			},
// 			Status: model.JobStatusAnalyzing,
// 		}}, nil)

// 		dbMock.EXPECT().UpdateIngestJob(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, IngestJob model.IngestJob) error {
// 			require.Equal(t, model.JobStatusComplete, IngestJob.Status)
// 			return nil
// 		})

// 		ingestService.CompleteAnalyzedIngestJobs(context.Background(), dbMock)
// 	})
// }

// func TestProcessFinishedIngestJobs(t *testing.T) {
// 	const jobID int64 = 1

// 	var (
// 		mockCtrl = gomock.NewController(t)
// 		dbMock   = mocks.NewMockDatabase(mockCtrl)
// 	)

// 	defer mockCtrl.Finish()

// 	t.Run("Transition Jobs with No Remaining Ingest Tasks", func(t *testing.T) {
// 		dbMock.EXPECT().GetIngestJobsWithStatus(gomock.Any(), model.JobStatusIngesting).Return([]model.IngestJob{{
// 			BigSerial: model.BigSerial{
// 				ID: jobID,
// 			},
// 			Status: model.JobStatusIngesting,
// 		}}, nil)

// 		dbMock.EXPECT().GetIngestTasksForJob(gomock.Any(), jobID).Return([]model.IngestTask{}, nil)
// 		dbMock.EXPECT().UpdateIngestJob(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, IngestJob model.IngestJob) error {
// 			require.Equal(t, model.JobStatusAnalyzing, IngestJob.Status)
// 			return nil
// 		})

// 		ingestService.ProcessFinishedIngestJobs(context.Background(), dbMock)
// 	})

// 	t.Run("Don't Transition Jobs with Remaining Ingest Tasks", func(t *testing.T) {
// 		dbMock.EXPECT().GetIngestJobsWithStatus(gomock.Any(), model.JobStatusIngesting).Return([]model.IngestJob{{
// 			BigSerial: model.BigSerial{
// 				ID: jobID,
// 			},
// 			Status: model.JobStatusIngesting,
// 		}}, nil)

// 		dbMock.EXPECT().GetIngestTasksForJob(gomock.Any(), jobID).Return([]model.IngestTask{{}}, nil)

// 		ingestService.ProcessFinishedIngestJobs(context.Background(), dbMock)
// 	})
// }
