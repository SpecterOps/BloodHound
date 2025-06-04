// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package job

import (
	"context"

	"github.com/specterops/bloodhound/src/model"
)

// The JobData interface is designed to manage the lifecycle of jobs in a system that processes graph-based data
type JobData interface {
	// Task handlers
	CreateIngestTask(ctx context.Context, task model.IngestTask) (model.IngestTask, error)
	DeleteAllIngestTasks(ctx context.Context) error
	CreateCompositionInfo(ctx context.Context, nodes model.EdgeCompositionNodes, edges model.EdgeCompositionEdges) (model.EdgeCompositionNodes, model.EdgeCompositionEdges, error)

	GetIngestTasksForJob(ctx context.Context, jobID int64) (model.IngestTasks, error)
	// Job handlers
	CreateIngestJob(ctx context.Context, job model.IngestJob) (model.IngestJob, error)
	UpdateIngestJob(ctx context.Context, job model.IngestJob) error
	GetIngestJob(ctx context.Context, id int64) (model.IngestJob, error)
	GetAllIngestJobs(ctx context.Context, skip int, limit int, order string, filter model.SQLFilter) ([]model.IngestJob, int, error)
	GetIngestJobsWithStatus(ctx context.Context, status model.JobStatus) ([]model.IngestJob, error)
	DeleteAllIngestJobs(ctx context.Context) error
	CancelAllIngestJobs(ctx context.Context) error
}

type JobService struct {
	ctx context.Context
	db  JobData
}

func NewJobService(ctx context.Context, db JobData) JobService {
	return JobService{
		ctx: ctx,
		db:  db,
	}
}
