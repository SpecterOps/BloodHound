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
package services

//go:generate go tool mockery

import (
	"context"
	"encoding/json"
	"time"
)

// Database describes the persistence capabilities the extensions Service requires.
// Implementations are expected to translate driver-specific errors into the
// sentinels defined in this package.
type Database interface {
	GetNodeKind(ctx context.Context, id int32) (NodeKind, error)
	GetKindInfosByNodeKindID(ctx context.Context, nodeKindID int32) ([]KindInfo, error)
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

// Service implements the extensions use cases on top of a Database implementation.
type Service struct {
	db Database
}

// NewService constructs a Service backed by the supplied Database implementation.
func NewService(databaseInterface Database) *Service {
	return &Service{db: databaseInterface}
}
