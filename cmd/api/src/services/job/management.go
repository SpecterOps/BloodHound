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

package job

import (
	"context"
	"fmt"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// NOTE: These methods are all called by the frontend/http handler to do stuff. We might want to consider moving
//       that to the service model too, but I'm not sure if that's important yet.

func GetIngestJobByID(ctx context.Context, db JobData, jobID int64) (model.IngestJob, error) {
	return db.GetIngestJob(ctx, jobID)
}

func GetAllIngestJobs(ctx context.Context, db JobData, skip int, limit int, order string, filter model.SQLFilter) ([]model.IngestJob, int, error) {
	return db.GetAllIngestJobs(ctx, skip, limit, order, filter)
}

func StartIngestJob(ctx context.Context, db JobData, user model.User) (model.IngestJob, error) {
	job := model.IngestJob{
		UserID:     user.ID,
		User:       user,
		Status:     model.JobStatusRunning,
		StartTime:  time.Now().UTC(),
		LastIngest: time.Now().UTC(),
	}
	return db.CreateIngestJob(ctx, job)
}

func TouchIngestJobLastIngest(ctx context.Context, db JobData, job model.IngestJob) error {
	job.LastIngest = time.Now().UTC()
	return db.UpdateIngestJob(ctx, job)
}

func EndIngestJob(ctx context.Context, db JobData, job model.IngestJob) error {
	job.Status = model.JobStatusIngesting

	if err := db.UpdateIngestJob(ctx, job); err != nil {
		return fmt.Errorf("error ending ingest job: %w", err)
	}

	return nil
}
