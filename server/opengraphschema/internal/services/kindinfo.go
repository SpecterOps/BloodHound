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

package services

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrKindInfoNotFound          = errors.New("kind info not found")
	ErrKindInfoKindNotFound      = errors.New("kind info references a kind that does not exist")
	ErrKindInfoDuplicatePosition = errors.New("kind info position already in use for this kind")
	ErrKindInfoDuplicateInfoKey  = errors.New("kind info key already in use for this kind")
)

// KindInfoInputs is the collection of entity panels supplied for a kind.
type KindInfoInputs []KindInfoInput

// KindInfoInput represents one entity panel definition as provided in a user's upload JSON
type KindInfoInput struct {
	InfoKey  string
	Title    string
	Position int32
	Content  json.RawMessage
}

// KindInfo holds the data associated with a single entity panel.
type KindInfo struct {
	ID int32

	KindID int32 // dawgs kind id
	// Only one of NodeKindID or RelationshipKindID will ever be populated
	// as the underlying table schema enforces that a KindInfo row must only
	// reference one or the other.
	NodeKindID         *int32
	RelationshipKindID *int32

	Name string

	InfoKey  string
	Title    string
	Position int32
	Content  json.RawMessage

	CreatedAt time.Time
	UpdatedAt time.Time
}
