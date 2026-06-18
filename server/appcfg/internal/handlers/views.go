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

package handlers

import (
	"encoding/json"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
)

// DatapipeStatusView is the JSON shape returned by the datapipe status handler.
// It is decoupled from services.DatapipeStatus so the wire format can evolve
// independently of the domain model.
type DatapipeStatusView struct {
	Status                  services.DatapipeStatusType `json:"status"`
	UpdatedAt               time.Time                   `json:"updated_at"`
	LastCompleteAnalysisAt  time.Time                   `json:"last_complete_analysis_at"`
	LastAnalysisRunAt       time.Time                   `json:"last_analysis_run_at"`
	NextScheduledAnalysisAt null.Time                   `json:"next_scheduled_analysis_at"`
}

// BuildDatapipeStatusView projects a services.DatapipeStatus into the
// view type the handlers return in their JSON envelope.
func BuildDatapipeStatusView(ds services.DatapipeStatus) DatapipeStatusView {
	return DatapipeStatusView{
		Status:                  ds.Status,
		UpdatedAt:               ds.UpdatedAt,
		LastCompleteAnalysisAt:  ds.LastCompleteAnalysisAt,
		LastAnalysisRunAt:       ds.LastAnalysisRunAt,
		NextScheduledAnalysisAt: ds.NextScheduledAnalysisAt,
	}
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s DatapipeStatusView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}
