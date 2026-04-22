// Copyright 2025 Specter Ops, Inc.
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

package model

import (
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
)

type Event struct {
	ID          int                     `json:"id"`
	Type        string                  `json:"type"`
	Message     string                  `json:"message"`
	Data        types.JSONUntypedObject `json:"data"`
	CreatedAt   time.Time               `json:"created_at"`
	ProcessedAt null.Time               `json:"processed_at"`
}

type Events []Event

func (Event) TableName() string {
	return "events"
}
