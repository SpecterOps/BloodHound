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

// decision is the outcome of reconciling an incoming analysis request against
// any request already queued.
type decision struct {
	write   bool
	request model.AnalysisRequest
}

// reconcile applies precedence rules to an incoming request against the one
// already queued and returns what the caller should do next.
func reconcile(existing *model.AnalysisRequest, incoming model.AnalysisRequest) decision {
	// Nothing is queued: write the incoming request.
	if existing == nil {
		return decision{write: true, request: incoming}
	}
	// A deletion is already queued: drop the incoming request.
	// Nothing should override a deletion.
	if existing.RequestType == model.AnalysisRequestDeletion {
		return decision{write: false}
	}
	// The incoming request is a deletion.
	// A deletion request always beats a queued analysis request. Write it.
	if incoming.RequestType == model.AnalysisRequestDeletion {
		return decision{write: true, request: incoming}
	}
	// Both are analysis requests.
	// Check if the incoming request is a subset of what's already queued.
	// If there is overlap, skip the write.
	// If the incoming request has new steps, merge them in and write the updated request.
	var (
		unionedSteps = existing.AnalysisStep | incoming.AnalysisStep
	)
	if unionedSteps == existing.AnalysisStep {
		return decision{write: false}
	}
	incoming.AnalysisStep = unionedSteps
	return decision{write: true, request: incoming}
}
