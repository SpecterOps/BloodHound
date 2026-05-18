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

package model

import (
	"time"

	"github.com/lib/pq"
)

type AnalysisRequestType string

const (
	AnalysisRequestAnalysis AnalysisRequestType = "analysis"
	AnalysisRequestDeletion AnalysisRequestType = "deletion"
)

// AnalysisSteps is a bitmask that selects which steps of the analysis pipeline
// should run. Each flag is independent, so callers may request any combination
// of steps. When steps are combined, they always execute in pipeline order:
// AD post-processing -> Azure post-processing -> Tagging -> Analysis.
type AnalysisSteps int

const (
	////////
	// Individual steps available to the bitmask
	////////
	// AnalysisStepADPostProcessing runs AD post-processing.
	AnalysisStepADPostProcessing AnalysisSteps = 1 << iota
	// AnalysisStepAzurePostProcessing runs Azure post-processing.
	AnalysisStepAzurePostProcessing
	// AnalysisStepTagging runs tagging of asset groups and tiers.
	AnalysisStepTagging
	// AnalysisStepGenerateFindings runs the analysis pipeline (BHE only).
	AnalysisStepGenerateFindings

	/////////
	// analysisStepSentinel MUST remain the last entry in this iota block;
	// AnalysisStepAll is derived from it. Add new steps above this line.
	////////
	analysisStepSentinel

	////////
	// Helpers available for ease of use throughout the codebase. If adding indvidual steps above, validate whether these need updating based on expected behavior.
	////////
	// AnalysisStepTaggingToCompletion runs the tagging step and every step that follows it.
	AnalysisStepTaggingToCompletion = AnalysisStepTagging | AnalysisStepGenerateFindings
	// AnalysisStepAll selects every step in the pipeline.
	AnalysisStepAll = analysisStepSentinel - 1
)

// Has reports whether all of the bits in step are set in s.
func (s AnalysisSteps) Has(step AnalysisSteps) bool {
	return s&step == step
}

func (s AnalysisSteps) Merge(steps AnalysisSteps) AnalysisSteps {
	return s | steps
}

type AnalysisRequest struct {
	RequestedBy   string              `json:"requested_by"`
	RequestType   AnalysisRequestType `json:"request_type"`
	RequestedAt   time.Time           `json:"requested_at"`
	AnalysisSteps AnalysisSteps       `json:"analysis_step"` // Bitmask indicating which analysis pipeline steps to run.

	DeleteAllGraph        bool           `json:"delete_all_graph"`                        // Deletes all nodes and edges in the graph
	DeleteSourcelessGraph bool           `json:"delete_sourceless_graph"`                 // Deletes all nodes and edges in the graph that have a type not registered in the source_kinds table
	DeleteSourceKinds     pq.StringArray `gorm:"type:text[];column:delete_source_kinds"`  // Deletes all nodes and edges per kind provided.
	DeleteRelationships   pq.StringArray `gorm:"type:text[];column:delete_relationships"` // Deletes all relationships by name
}
