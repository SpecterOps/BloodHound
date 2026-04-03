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

package model

import "time"

// AnonymizeTranslationEntry stores a mapping between an original graph node
// property value and its anonymized replacement.
type AnonymizeTranslationEntry struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	NodeGraphID     int64     `json:"node_graph_id"`
	PropertyKey     string    `json:"property_key"`
	OriginalValue   string    `json:"original_value"`
	AnonymizedValue string    `json:"anonymized_value"`
	ObjectType      string    `json:"object_type"`
	CreatedAt       time.Time `json:"created_at"`
}

func (AnonymizeTranslationEntry) TableName() string {
	return "anonymize_translation_entries"
}
