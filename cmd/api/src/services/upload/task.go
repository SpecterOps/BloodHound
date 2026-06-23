// Copyright 2025 Specter Ops, Inc.
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
	"context"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/metrics"
)

type IngestTaskParams struct {
	Filename         string
	ProvidedFileName string
	FileType         model.FileType
	RequestID        string
	JobID            int64
}

// fileFormatFromFileType converts a model.FileType to metrics.IngestFileFormat.
func fileFormatFromFileType(ft model.FileType) metrics.IngestFileFormat {
	switch ft.String() {
	case "json":
		return metrics.IngestFileFormatJSON
	case "zip":
		return metrics.IngestFileFormatZip
	default:
		return metrics.IngestFileFormatUnknown
	}
}

func CreateIngestTask(ctx context.Context, db UploadData, params IngestTaskParams) (model.IngestTask, error) {
	newIngestTask := model.IngestTask{
		StoredFileName:   params.Filename,
		OriginalFileName: params.ProvidedFileName,
		RequestGUID:      params.RequestID,
		JobId:            null.Int64From(params.JobID),
		FileType:         params.FileType,
	}

	if task, err := db.CreateIngestTask(ctx, newIngestTask); err != nil {
		// Record metric: file ingest task creation failed
		metrics.RecordIngestTask(
			metrics.IngestCollectorManual,
			fileFormatFromFileType(params.FileType),
			metrics.IngestTaskStatusFailed,
		)
		return task, err
	} else {
		// Record metric: file ingest task created and saved to disk
		metrics.RecordIngestTask(
			metrics.IngestCollectorManual,
			fileFormatFromFileType(params.FileType),
			metrics.IngestTaskStatusSuccess,
		)
		return task, nil
	}
}

func CreateCompositionInfo(ctx context.Context, db UploadData, nodes model.EdgeCompositionNodes, edges model.EdgeCompositionEdges) (model.EdgeCompositionNodes, model.EdgeCompositionEdges, error) {
	return db.CreateCompositionInfo(ctx, nodes, edges)
}
