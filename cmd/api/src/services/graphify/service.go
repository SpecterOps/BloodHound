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
package graphify

import (
	"context"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/services/ingest"
)

// The IngestData interface is designed to manage the lifecycle of ingestion tasks and jobs in a system that processes graph-based data
type GraphifyData interface {
	// Task handlers
	GetAllIngestTasks(ctx context.Context) (model.IngestTasks, error)
	DeleteIngestTask(ctx context.Context, ingestTask model.IngestTask) error
	GetFlagByKey(context.Context, string) (appcfg.FeatureFlag, error)

	// THESE ARE BAD AND SHOULD NOT BE HERE.
	// we update the job in a bizzare place
	UpdateIngestJob(ctx context.Context, job model.IngestJob) error
	GetIngestJob(ctx context.Context, id int64) (model.IngestJob, error)
}

type GraphifyService struct {
	ctx     context.Context
	db      GraphifyData
	graphdb graph.Database
	cfg     config.Configuration
	schema  ingest.IngestSchema
}

func NewGraphifyService(ctx context.Context, db GraphifyData, graphDb graph.Database, cfg config.Configuration, schema ingest.IngestSchema) GraphifyService {
	return GraphifyService{
		ctx:     ctx,
		db:      db,
		graphdb: graphDb,
		cfg:     cfg,
		schema:  schema,
	}
}
