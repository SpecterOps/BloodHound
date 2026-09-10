// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"database/sql"
	"encoding/json"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
)

type DatapipeStatusView struct {
	Status                  services.DatapipeStatusType `json:"status"`
	UpdatedAt               time.Time                   `json:"updated_at"`
	LastCompleteAnalysisAt  time.Time                   `json:"last_complete_analysis_at"`
	LastAnalysisRunAt       time.Time                   `json:"last_analysis_run_at"`
	LastCompleteOptimizeAt  time.Time                   `json:"last_complete_optimize_at"`
	NextScheduledAnalysisAt null.Time                   `json:"next_scheduled_analysis_at"`
}

func BuildDatapipeStatusView(status services.DatapipeStatus) DatapipeStatusView {
	return DatapipeStatusView{
		Status:                  status.Status,
		UpdatedAt:               status.UpdatedAt,
		LastCompleteAnalysisAt:  status.LastCompleteAnalysisAt,
		LastAnalysisRunAt:       status.LastAnalysisRunAt,
		LastCompleteOptimizeAt:  status.LastCompleteOptimizeAt,
		NextScheduledAnalysisAt: status.NextScheduledAnalysisAt,
	}
}

// JSONView marshals the view to the byte slice expected by responses.WriteBasic,
// satisfying the responses.JSONViewer contract.
func (s DatapipeStatusView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

type ParameterView struct {
	ID          int32                 `json:"id"`
	Key         services.ParameterKey `json:"key"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Value       types.JSONBObject     `json:"value"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   sql.NullTime          `json:"deleted_at"`
}

type ParameterListView []ParameterView

func BuildParameterView(parameter services.Parameter) ParameterView {
	return ParameterView{
		ID:          parameter.ID,
		Key:         parameter.Key,
		Name:        parameter.Name,
		Description: parameter.Description,
		Value:       parameter.Value,
		CreatedAt:   parameter.CreatedAt,
		UpdatedAt:   parameter.UpdatedAt,
		DeletedAt:   parameter.DeletedAt,
	}
}

func BuildParameterListView(parameters services.Parameters) ParameterListView {
	var parametersView = []ParameterView{}

	for _, parameter := range parameters {
		parametersView = append(parametersView, BuildParameterView(parameter))
	}

	return parametersView
}

func (s ParameterListView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}
