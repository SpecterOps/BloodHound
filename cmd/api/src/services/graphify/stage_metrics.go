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

package graphify

import (
	"time"

	ingestmetrics "github.com/specterops/bloodhound/packages/go/metrics"
)

type graphifyStageMeasurement struct {
	activeDuration  time.Duration
	itemCount       int
	failedItemCount int
}

func (s *graphifyStageMeasurement) add(activeDuration time.Duration, itemCount int, failedItemCount int) {
	if activeDuration < 0 || itemCount <= 0 || failedItemCount < 0 || failedItemCount > itemCount {
		return
	}

	s.activeDuration += activeDuration
	s.itemCount += itemCount
	s.failedItemCount += failedItemCount
}

type graphifyChunkStageMeasurements struct {
	nodePrepare             graphifyStageMeasurement
	nodeDeduplicate         graphifyStageMeasurement
	nodeBatchUpdate         graphifyStageMeasurement
	relationshipResolution  graphifyStageMeasurement
	relationshipPrepare     graphifyStageMeasurement
	relationshipDeduplicate graphifyStageMeasurement
	relationshipBatchUpdate graphifyStageMeasurement
}

func publishIngestStage(stage ingestmetrics.IngestStage, measurement graphifyStageMeasurement) {
	if measurement.itemCount == 0 {
		return
	}

	ingestmetrics.RecordIngestStage(stage, measurement.activeDuration, measurement.itemCount, measurement.failedItemCount)
}

func (s *graphifyChunkStageMeasurements) publish() {
	publishIngestStage(ingestmetrics.IngestStageNodePrepare, s.nodePrepare)
	publishIngestStage(ingestmetrics.IngestStageNodeDeduplicate, s.nodeDeduplicate)
	publishIngestStage(ingestmetrics.IngestStageNodeBatchUpdate, s.nodeBatchUpdate)
	publishIngestStage(ingestmetrics.IngestStageRelationshipResolution, s.relationshipResolution)
	publishIngestStage(ingestmetrics.IngestStageRelationshipPrepare, s.relationshipPrepare)
	publishIngestStage(ingestmetrics.IngestStageRelationshipDeduplicate, s.relationshipDeduplicate)
	publishIngestStage(ingestmetrics.IngestStageRelationshipBatchUpdate, s.relationshipBatchUpdate)
}
