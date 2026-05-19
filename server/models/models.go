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

package models

import (
	"time"
)

type RequestedAnalysisType string

const (
	RequestedAnalysisTypeAnalysis RequestedAnalysisType = "analysis"
	RequestedAnalysisTypeDeletion RequestedAnalysisType = "deletion"
)

type RequestedAnalysis struct {
	RequestedBy string
	RequestType RequestedAnalysisType
	RequestedAt time.Time
	// Deletes all nodes and edges in the graph
	DeleteAllGraph bool
	// Deletes all nodes and edges in the graph that have a type not registered in the source_kinds table
	DeleteSourcelessGraph bool
	// Deletes all nodes and edges per kind provided.
	DeleteSourceKinds []string
	// Deletes all relationships by name
	DeleteRelationships []string
}
