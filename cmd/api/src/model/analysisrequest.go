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
	"time"

	"github.com/lib/pq"
)

type AnalysisRequestType string

const (
	AnalysisRequestAnalysis AnalysisRequestType = "analysis"
	AnalysisRequestDeletion AnalysisRequestType = "deletion"
)

const (
	////////
	// Individual bits available to the internal bitmask
	////////
	// analysisADPostProcessing runs AD post-processing.
	analysisADPostProcessing int32 = 1 << iota
	// analysisAzurePostProcessing runs Azure post-processing.
	analysisAzurePostProcessing
	// analysisTagging runs tagging of asset groups and tiers.
	analysisTagging
	// analysisGenerateFindings runs the analysis pipeline (BHE only).
	analysisGenerateFindings

	/////////
	// analysisSentinel MUST remain the last entry in this iota block;
	// analysisFull is derived from it. Add new steps above this line.
	////////
	analysisSentinel

	////////
	// Helpers available for ease of use.
	// If adding individual steps above, validate whether these need updating based on expected behavior.
	////////
	// AnalysisTaggingOnwards runs the tagging step and every step that follows it.
	analysisTaggingOnwards = analysisTagging | analysisGenerateFindings
	// AnalysisAll selects every step in the pipeline.
	analysisFull = analysisSentinel - 1
)

// AnalysisSteps represents a set of steps or single step in analysis using a bitmask.
type AnalysisSteps struct {
	bits int32
}

var (
	analysisStepADPostProcessing    = AnalysisSteps{bits: analysisADPostProcessing}
	analysisStepAzurePostProcessing = AnalysisSteps{bits: analysisAzurePostProcessing}
	analysisStepTagging             = AnalysisSteps{bits: analysisTagging}
	analysisStepGenerateFindings    = AnalysisSteps{bits: analysisGenerateFindings}

	// Defines a name for a single step of analysis for string representation purposes.
	// If adding bits (steps) to the bitmask, this needs to be updated.
	analysisStepNames = map[AnalysisSteps]string{
		analysisStepADPostProcessing:    "ad_post_processing",
		analysisStepAzurePostProcessing: "azure_post_processing",
		analysisStepTagging:             "tagging",
		analysisStepGenerateFindings:    "generate_findings",
	}

	analysisStepsTaggingOnwards = AnalysisSteps{bits: analysisTaggingOnwards}
	analysisStepsFull           = AnalysisSteps{bits: analysisFull}
)

func (s AnalysisSteps) GetNameOfAnalysisStep() string {
	name, present := analysisStepNames[s]

	if present {
		return name
	}

	return "unknown"
}

func AnalysisStepADPostProcessing() AnalysisSteps {
	return analysisStepADPostProcessing
}

func AnalysisStepAzurePostProcessing() AnalysisSteps {
	return analysisStepAzurePostProcessing
}

func AnalysisStepTagging() AnalysisSteps {
	return analysisStepTagging
}

func AnalysisStepGenerateFindings() AnalysisSteps {
	return analysisStepGenerateFindings
}

func AnalysisStepsTaggingOnwards() AnalysisSteps {
	return AnalysisSteps{bits: analysisTaggingOnwards}
}

func AnalysisStepsFull() AnalysisSteps {
	return analysisStepsFull
}

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

func analysisStepsFromBits(bits int) AnalysisSteps {
	return AnalysisSteps{bits: int32(bits)}
}

// #region DB implementation methods
// AnalysisSteps implements these methods to easily Read/Write from the DB
func (s AnalysisSteps) Value() (driver.Value, error) {
	return int32(s.bits), nil
}

func (s *AnalysisSteps) Scan(value any) error {
	switch typedValue := value.(type) {
	case nil:
		*s = AnalysisSteps{}
	case int64:
		*s = analysisStepsFromBits(int(typedValue))
	case int32:
		*s = analysisStepsFromBits(int(typedValue))
	case int:
		*s = analysisStepsFromBits(typedValue)
	case []byte:
		bits, err := strconv.Atoi(string(typedValue))
		if err != nil {
			return err
		}
		*s = analysisStepsFromBits(bits)
	case string:
		bits, err := strconv.Atoi(typedValue)
		if err != nil {
			return err
		}
		*s = analysisStepsFromBits(bits)
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

	*s = analysisStepsFromBits(bits)
	return nil
}

// #endregion

type AnalysisMode string

const (
	AnalysisModeTaggingOnwards AnalysisMode = "tagging_onwards"
	AnalysisModeFull           AnalysisMode = "full"
)

// This function needs to be updated when new modes are added.
// If the mode is unknown, defaults to full analysis.
// With this function, and by making AnalysisMode a parameter for RequestAnalysis(),
// we ensure consumers of RequestAnalysis() cannot queue arbitrary bitmasks.
func (s AnalysisMode) AnalysisStepsFromMode() AnalysisSteps {
	switch s {
	case AnalysisModeTaggingOnwards:
		return analysisStepsTaggingOnwards
	case AnalysisModeFull:
		return analysisStepsFull
	default:
		return analysisStepsFull
	}
}

type AnalysisRequest struct {
	RequestedBy   string              `json:"requested_by"`
	RequestType   AnalysisRequestType `json:"request_type"`
	RequestedAt   time.Time           `json:"requested_at"`
	AnalysisSteps AnalysisSteps       `json:"analysis_step" gorm:"column:analysis_step"` // Internally bits indicating which analysis pipeline steps to run.

	DeleteAllGraph        bool           `json:"delete_all_graph"`                        // Deletes all nodes and edges in the graph
	DeleteSourcelessGraph bool           `json:"delete_sourceless_graph"`                 // Deletes all nodes and edges in the graph that have a type not registered in the source_kinds table
	DeleteSourceKinds     pq.StringArray `gorm:"type:text[];column:delete_source_kinds"`  // Deletes all nodes and edges per kind provided.
	DeleteRelationships   pq.StringArray `gorm:"type:text[];column:delete_relationships"` // Deletes all relationships by name
}
