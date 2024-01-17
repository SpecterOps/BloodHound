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

package ingest

import (
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
)

type IngestData interface {
	CreateIngestTask(task model.IngestTask) (model.IngestTask, error)
}

func CreateIngestTask(db IngestData, filename string, requestID string, jobID int64) (model.IngestTask, error) {
	newIngestTask := model.IngestTask{
		FileName:    filename,
		RequestGUID: requestID,
		TaskID:      null.Int64From(jobID),
	}

	return db.CreateIngestTask(newIngestTask)
}
