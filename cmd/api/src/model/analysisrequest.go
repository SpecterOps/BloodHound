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
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

type AnalysisRequestType string

const (
	AnalysisRequestAnalysis AnalysisRequestType = "analysis"
	AnalysisRequestDeletion AnalysisRequestType = "deletion"
)

type AnalysisEntrypoint string

const (
	AnalysisEntrypointFull    AnalysisEntrypoint = "full"
	AnalysisEntrypointTagging AnalysisEntrypoint = "tagging"
)

type AnalysisSteps struct {
	bits analysisStepsBitmask
}

// analysisStepsBitmask is the private bitmask that selects which steps of the
// analysis pipeline should run. Each flag is independent, so internal callers
// may represent any combination of steps. When steps are combined, they always
// execute in pipeline order: AD post-processing -> Azure post-processing ->
// Tagging -> Analysis.
type analysisStepsBitmask int

const (
	////////
	// Individual bits available to the internal bitmask
	////////
	// AnalysisStepADPostProcessing runs AD post-processing.
	analysisStepADPostProcessing analysisStepsBitmask = 1 << iota
	// AnalysisStepAzurePostProcessing runs Azure post-processing.
	analysisStepAzurePostProcessing
	// AnalysisStepTagging runs tagging of asset groups and tiers.
	analysisStepTagging
	// AnalysisStepGenerateFindings runs the analysis pipeline (BHE only).
	analysisStepGenerateFindings

	/////////
	// analysisStepSentinel MUST remain the last entry in this iota block;
	// analysisStepAll is derived from it. Add new steps above this line.
	////////
	analysisStepSentinel

	////////
	// Helpers available for ease of use throughout the codebase.
	// If adding individual steps above, validate whether these need updating based on expected behavior.
	////////
	// AnalysisStepTaggingToCompletion runs the tagging step and every step that follows it.
	analysisStepTaggingToCompletion = analysisStepTagging | analysisStepGenerateFindings
	// AnalysisStepAll selects every step in the pipeline.
	analysisStepAll = analysisStepSentinel - 1
)

var (
	AnalysisStepADPostProcessing    = AnalysisSteps{bits: analysisStepADPostProcessing}
	AnalysisStepAzurePostProcessing = AnalysisSteps{bits: analysisStepAzurePostProcessing}
	AnalysisStepTagging             = AnalysisSteps{bits: analysisStepTagging}
	AnalysisStepGenerateFindings    = AnalysisSteps{bits: analysisStepGenerateFindings}
	AnalysisStepTaggingToCompletion = AnalysisSteps{bits: analysisStepTaggingToCompletion}
	AnalysisStepAll                 = AnalysisSteps{bits: analysisStepAll}
)

// Has reports whether all of the bits in step are set in s.
func (s AnalysisSteps) Has(step AnalysisSteps) bool {
	return s.bits&step.bits == step.bits
}

func (s AnalysisSteps) Merge(steps AnalysisSteps) AnalysisSteps {
	return AnalysisSteps{bits: s.bits | steps.bits}
}

func (s AnalysisSteps) Bits() int {
	return int(s.bits)
}

func (s AnalysisSteps) IsEmpty() bool {
	return s.bits == 0
}

func (s AnalysisSteps) String() string {
	var (
		stepNames = []string{}
		steps     = []struct {
			step AnalysisSteps
			name string
		}{
			{
				step: AnalysisStepADPostProcessing,
				name: "ad_post_processing",
			},
			{
				step: AnalysisStepAzurePostProcessing,
				name: "azure_post_processing",
			},
			{
				step: AnalysisStepTagging,
				name: "tagging",
			},
			{
				step: AnalysisStepGenerateFindings,
				name: "analysis",
			},
		}
		unknownBits = s.bits &^ analysisStepAll
	)

	for _, step := range steps {
		if s.Has(step.step) {
			stepNames = append(stepNames, step.name)
		}
	}

	if unknownBits != 0 {
		stepNames = append(stepNames, fmt.Sprintf("unknown:%d", unknownBits))
	}

	if len(stepNames) == 0 {
		return "none"
	}

	return strings.Join(stepNames, ",")
}

func (s AnalysisSteps) Value() (driver.Value, error) {
	return int64(s.bits), nil
}

func (s *AnalysisSteps) Scan(value any) error {
	switch typedValue := value.(type) {
	case nil:
		*s = AnalysisSteps{}
	case int64:
		*s = AnalysisStepsFromBits(int(typedValue))
	case int32:
		*s = AnalysisStepsFromBits(int(typedValue))
	case int:
		*s = AnalysisStepsFromBits(typedValue)
	case []byte:
		bits, err := strconv.Atoi(string(typedValue))
		if err != nil {
			return err
		}
		*s = AnalysisStepsFromBits(bits)
	case string:
		bits, err := strconv.Atoi(typedValue)
		if err != nil {
			return err
		}
		*s = AnalysisStepsFromBits(bits)
	default:
		return fmt.Errorf("unable to scan analysis steps from %T", value)
	}

	return nil
}

func (s AnalysisSteps) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Bits())
}

func (s *AnalysisSteps) UnmarshalJSON(data []byte) error {
	var bits int

	if err := json.Unmarshal(data, &bits); err != nil {
		return err
	}

	*s = AnalysisStepsFromBits(bits)
	return nil
}

func AnalysisStepsFromEntrypoint(entrypoint AnalysisEntrypoint) AnalysisSteps {
	switch entrypoint {
	case AnalysisEntrypointTagging:
		return AnalysisStepTaggingToCompletion
	default:
		return FullAnalysisSteps()
	}
}

func AnalysisStepsFromBits(bits int) AnalysisSteps {
	return AnalysisSteps{bits: analysisStepsBitmask(bits)}
}

func FullAnalysisSteps() AnalysisSteps {
	return AnalysisStepAll
}

func (s AnalysisEntrypoint) AnalysisSteps() AnalysisSteps {
	return AnalysisStepsFromEntrypoint(s)
}

type AnalysisRequest struct {
	RequestedBy   string              `json:"requested_by"`
	RequestType   AnalysisRequestType `json:"request_type"`
	RequestedAt   time.Time           `json:"requested_at"`
	AnalysisSteps AnalysisSteps       `json:"analysis_step" gorm:"column:analysis_step"` // Bitmask indicating which analysis pipeline steps to run.

	DeleteAllGraph        bool           `json:"delete_all_graph"`                        // Deletes all nodes and edges in the graph
	DeleteSourcelessGraph bool           `json:"delete_sourceless_graph"`                 // Deletes all nodes and edges in the graph that have a type not registered in the source_kinds table
	DeleteSourceKinds     pq.StringArray `gorm:"type:text[];column:delete_source_kinds"`  // Deletes all nodes and edges per kind provided.
	DeleteRelationships   pq.StringArray `gorm:"type:text[];column:delete_relationships"` // Deletes all relationships by name
}
