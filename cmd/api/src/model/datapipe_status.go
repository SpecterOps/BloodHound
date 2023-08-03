// Copyright 2023 Specter Ops, Inc.
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

import "time"

type DatapipeStatus string

const (
	DatapipeStatusIdle      DatapipeStatus = "idle"
	DatapipeStatusIngesting DatapipeStatus = "ingesting"
	DatapipeStatusAnalyzing DatapipeStatus = "analyzing"
)

type DatapipeStatusWrapper struct {
	Status                 DatapipeStatus `json:"status"`
	UpdatedAt              time.Time      `json:"updated_at"`
	LastCompleteAnalysisAt time.Time      `json:"last_complete_analysis_at"`
}

func (s *DatapipeStatusWrapper) Update(status DatapipeStatus, updateAnalysisTime bool) {
	s.Status = status
	s.UpdatedAt = time.Now().UTC()
	if updateAnalysisTime {
		s.LastCompleteAnalysisAt = time.Now().UTC()
	}
}
