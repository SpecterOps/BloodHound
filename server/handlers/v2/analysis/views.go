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

import (
	"time"

	"github.com/specterops/bloodhound/server/models"
)

type RequestedAnalysisView struct {
	RequestedBy string                       `json:"requested_by"`
	RequestType models.RequestedAnalysisType `json:"request_type"`
	RequestedAt time.Time                    `json:"requested_at"`
	// Deletes all nodes and edges in the graph
	DeleteAllGraph bool `json:"delete_all_graph"`
	// Deletes all nodes and edges in the graph that have a type not registered in the source_kinds table
	DeleteSourcelessGraph bool     `json:"delete_sourceless_graph"`
	DeleteSourceKinds     []string `json:"delete_source_kinds"`
	DeleteRelationships   []string `json:"delete_relationships"`
}

func BuildRequestedAnalysisView(ra models.RequestedAnalysis) RequestedAnalysisView {
	return RequestedAnalysisView{
		RequestedBy:           ra.RequestedBy,
		RequestType:           ra.RequestType,
		RequestedAt:           ra.RequestedAt,
		DeleteAllGraph:        ra.DeleteAllGraph,
		DeleteSourcelessGraph: ra.DeleteSourcelessGraph,
		DeleteSourceKinds:     ra.DeleteSourceKinds,
		DeleteRelationships:   ra.DeleteRelationships,
	}
}
