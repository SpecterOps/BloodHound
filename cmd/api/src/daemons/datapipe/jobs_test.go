package datapipe_test

import (
	"context"
	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestHasJobsWaitingForAnalysis(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		dbMock   = mocks.NewMockDatabase(mockCtrl)
	)

	defer mockCtrl.Finish()

	t.Run("Has Jobs Waiting for Analysis", func(t *testing.T) {
		dbMock.EXPECT().GetFileUploadJobsWithStatus(model.JobStatusAnalyzing).Return([]model.FileUploadJob{{}}, nil)

		hasJobs, err := datapipe.HasFileUploadJobsWaitingForAnalysis(dbMock)

		require.True(t, hasJobs)
		require.Nil(t, err)
	})

	t.Run("Has No Jobs Waiting for Analysis", func(t *testing.T) {
		dbMock.EXPECT().GetFileUploadJobsWithStatus(model.JobStatusAnalyzing).Return([]model.FileUploadJob{}, nil)

		hasJobs, err := datapipe.HasFileUploadJobsWaitingForAnalysis(dbMock)

		require.False(t, hasJobs)
		require.Nil(t, err)
	})
}

func TestFailAnalyzedFileUploadJobs(t *testing.T) {
	const jobID int64 = 1

	var (
		mockCtrl = gomock.NewController(t)
		dbMock   = mocks.NewMockDatabase(mockCtrl)
	)

	defer mockCtrl.Finish()

	t.Run("Fail Analyzed File Upload Jobs", func(t *testing.T) {
		dbMock.EXPECT().GetFileUploadJobsWithStatus(model.JobStatusAnalyzing).Return([]model.FileUploadJob{{
			BigSerial: model.BigSerial{
				ID: jobID,
			},
			Status: model.JobStatusAnalyzing,
		}}, nil)

		dbMock.EXPECT().UpdateFileUploadJob(gomock.Any()).DoAndReturn(func(fileUploadJob model.FileUploadJob) error {
			require.Equal(t, model.JobStatusFailed, fileUploadJob.Status)
			return nil
		})

		datapipe.FailAnalyzedFileUploadJobs(context.Background(), dbMock)
	})
}

func TestCompleteAnalyzedFileUploadJobs(t *testing.T) {
	const jobID int64 = 1

	var (
		mockCtrl = gomock.NewController(t)
		dbMock   = mocks.NewMockDatabase(mockCtrl)
	)

	defer mockCtrl.Finish()

	t.Run("Complete Analyzed File Upload Jobs", func(t *testing.T) {
		dbMock.EXPECT().GetFileUploadJobsWithStatus(model.JobStatusAnalyzing).Return([]model.FileUploadJob{{
			BigSerial: model.BigSerial{
				ID: jobID,
			},
			Status: model.JobStatusAnalyzing,
		}}, nil)

		dbMock.EXPECT().UpdateFileUploadJob(gomock.Any()).DoAndReturn(func(fileUploadJob model.FileUploadJob) error {
			require.Equal(t, model.JobStatusComplete, fileUploadJob.Status)
			return nil
		})

		datapipe.CompleteAnalyzedFileUploadJobs(context.Background(), dbMock)
	})
}

func TestProcessIngestedFileUploadJobs(t *testing.T) {
	const jobID int64 = 1

	var (
		mockCtrl = gomock.NewController(t)
		dbMock   = mocks.NewMockDatabase(mockCtrl)
	)

	defer mockCtrl.Finish()

	t.Run("Transition Jobs with No Remaining Ingest Tasks", func(t *testing.T) {
		dbMock.EXPECT().GetFileUploadJobsWithStatus(model.JobStatusIngesting).Return([]model.FileUploadJob{{
			BigSerial: model.BigSerial{
				ID: jobID,
			},
			Status: model.JobStatusIngesting,
		}}, nil)

		dbMock.EXPECT().GetIngestTasksForJob(jobID).Return([]model.IngestTask{}, nil)
		dbMock.EXPECT().UpdateFileUploadJob(gomock.Any()).DoAndReturn(func(fileUploadJob model.FileUploadJob) error {
			require.Equal(t, model.JobStatusAnalyzing, fileUploadJob.Status)
			return nil
		})

		datapipe.ProcessIngestedFileUploadJobs(context.Background(), dbMock)
	})

	t.Run("Don't Transition Jobs with Remaining Ingest Tasks", func(t *testing.T) {
		dbMock.EXPECT().GetFileUploadJobsWithStatus(model.JobStatusIngesting).Return([]model.FileUploadJob{{
			BigSerial: model.BigSerial{
				ID: jobID,
			},
			Status: model.JobStatusIngesting,
		}}, nil)

		dbMock.EXPECT().GetIngestTasksForJob(jobID).Return([]model.IngestTask{{}}, nil)

		datapipe.ProcessIngestedFileUploadJobs(context.Background(), dbMock)
	})
}
