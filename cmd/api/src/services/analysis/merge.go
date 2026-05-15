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

package analysis

import "github.com/specterops/bloodhound/cmd/api/src/model"

// mergeDecision describes the outcome of reconciling an incoming analysis
// request against any request already queued. write is true when the caller
// should persist merged; when false the caller should no-op.
type mergeDecision struct {
	write  bool
	merged model.AnalysisRequest
}

// merge reconciles an incoming analysis request against an existing one and
// returns the decision the service should act on. The rules are:
//   - No existing request: the incoming one is written as-is.
//   - Existing deletion request: sticky; the incoming request is dropped.
//   - Incoming deletion request: overrides any pending analysis request.
//   - Both analysis: their AnalysisStep bitmasks are OR-ed so a wider scope
//     (e.g. AnalysisStepAll) absorbs a narrower one (e.g.
//     AnalysisStepTaggingToCompletion). If the merged step matches the
//     existing step, the existing request is left alone.
func merge(existing model.AnalysisRequest, hasExisting bool, incoming model.AnalysisRequest) mergeDecision {
	if !hasExisting {
		return mergeDecision{write: true, merged: incoming}
	}
	if existing.RequestType == model.AnalysisRequestDeletion {
		return mergeDecision{write: false}
	}
	if incoming.RequestType == model.AnalysisRequestDeletion {
		return mergeDecision{write: true, merged: incoming}
	}
	var mergedStep = existing.AnalysisStep | incoming.AnalysisStep
	if mergedStep == existing.AnalysisStep {
		return mergeDecision{write: false}
	}
	var merged = incoming
	merged.AnalysisStep = mergedStep
	return mergeDecision{write: true, merged: merged}
}
