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
	// AnalysisNoPostProcessing skips the post processing steps.
	analysisNoPostProcessing = analysisFull &^ analysisADPostProcessing &^ analysisAzurePostProcessing
	// AnalysisAll selects every step in the pipeline.
	analysisFull = analysisSentinel - 1
)

// AnalysisStep represents a single step in analysis
type AnalysisStep int32

var (
	analysisStepADPostProcessing    = AnalysisStep(analysisADPostProcessing)
	analysisStepAzurePostProcessing = AnalysisStep(analysisAzurePostProcessing)
	analysisStepTagging             = AnalysisStep(analysisTagging)
	analysisStepGenerateFindings    = AnalysisStep(analysisGenerateFindings)

	// Defines a name for a single step of analysis for string representation purposes.
	// If adding bits (steps) to the bitmask, this needs to be updated.
	// There is a test to make sure this stays in sync.
	analysisStepNames = map[AnalysisStep]string{
		analysisStepADPostProcessing:    "ad_post_processing",
		analysisStepAzurePostProcessing: "azure_post_processing",
		analysisStepTagging:             "tagging",
		analysisStepGenerateFindings:    "generate_findings",
	}
)

func AnalysisStepADPostProcessing() AnalysisStep {
	return analysisStepADPostProcessing
}

func AnalysisStepAzurePostProcessing() AnalysisStep {
	return analysisStepAzurePostProcessing
}

func AnalysisStepTagging() AnalysisStep {
	return analysisStepTagging
}

func AnalysisStepGenerateFindings() AnalysisStep {
	return analysisStepGenerateFindings
}

// Returns false if name is not found for AnalysisStep.
func AnalysisStepName(step AnalysisStep) (string, bool) {
	name, present := analysisStepNames[step]

	if present {
		return name, true
	}

	return fmt.Sprintf("unknown:%d", int(step)), false
}

// AnalysisSteps represents a set of steps in analysis using a bitmask.
// bits is an unexported field so callers cannot create arbitrary nonzero bitmasks.
type AnalysisSteps struct {
	bits int32
}

var (
	analysisStepsNoPostProcessing = AnalysisSteps{bits: analysisNoPostProcessing}
	analysisStepsFull             = AnalysisSteps{bits: analysisFull}
)

func AnalysisStepsNoPostProcessing() AnalysisSteps {
	return analysisStepsNoPostProcessing
}

func AnalysisStepsFull() AnalysisSteps {
	return analysisStepsFull
}

// Has reports whether all of the bits in step are set in s.
func (s AnalysisSteps) Has(step AnalysisStep) bool {
	i := int32(step)
	return s.bits&i == i
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

func analysisStepsFromDecodedBits(bits int) (AnalysisSteps, error) {
	if bits < 0 {
		return AnalysisSteps{}, fmt.Errorf("analysis steps cannot be negative: %d", bits)
	}

	return analysisStepsFromBits(bits), nil
}

// #region DB implementation methods
// AnalysisSteps implements these methods to easily Read/Write from the DB
func (s AnalysisSteps) Value() (driver.Value, error) {
	return int32(s.bits), nil
}

func (s *AnalysisSteps) Scan(value any) error {
	var bits int

	switch typedValue := value.(type) {
	case nil:
		*s = AnalysisSteps{}
		return nil
	case int64:
		bits = int(typedValue)
	case int32:
		bits = int(typedValue)
	case int:
		bits = typedValue
	case []byte:
		parsedBits, err := strconv.Atoi(string(typedValue))
		if err != nil {
			return err
		}
		bits = parsedBits
	case string:
		parsedBits, err := strconv.Atoi(typedValue)
		if err != nil {
			return err
		}
		bits = parsedBits
	default:
		return fmt.Errorf("unable to scan analysis steps from %T", value)
	}

	analysisSteps, err := analysisStepsFromDecodedBits(bits)
	if err != nil {
		return err
	}

	*s = analysisSteps
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

	analysisSteps, err := analysisStepsFromDecodedBits(bits)
	if err != nil {
		return err
	}

	*s = analysisSteps
	return nil
}

// #endregion

type AnalysisMode string

const (
	// This list needs to be updated when new modes are added.
	AnalysisModeNoPostProcessing AnalysisMode = "no_post_processing"
	AnalysisModeFull             AnalysisMode = "full"
)

// This function needs to be updated when new modes are added.
// If the mode is unknown, defaults to full analysis.
// With this function, and by making AnalysisMode a parameter for RequestAnalysis(),
// we ensure consumers of RequestAnalysis() cannot queue arbitrary bitmasks.
func (s AnalysisMode) AnalysisStepsFromMode() AnalysisSteps {
	switch s {
	case AnalysisModeNoPostProcessing:
		return analysisStepsNoPostProcessing
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
