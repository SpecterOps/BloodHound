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
)

// Database describes the persistence capabilities the graphdb Service requires.
// Implementations are expected to translate driver-specific not-found errors into the
// sentinels defined in this package so that callers can map them without importing the
// persistence layer.
type Database interface {
	GetRelationship(ctx context.Context, id int64) (Relationship, error)
	GetNode(ctx context.Context, id int64) (Node, error)
	GetKindByName(ctx context.Context, name string) (Kind, error)
	GetNodeKindsByNames(ctx context.Context, names []string) ([]Kind, error)
}

// Kind is the domain representation of a relationship or node kind, pairing the kind name
// recorded on the graph with the integer identifier assigned to it in the schema_*_kinds table.
// ID is nil when the kind is not registered in the schema tables (best-effort resolution).
type Kind struct {
	ID   *int32
	Name string
}

// Service implements the graphdb use cases on top of a Database implementation.
type Service struct {
	db Database
}

// NewService constructs a Service backed by the supplied Database implementation.
func NewService(databaseInterface Database) *Service {
	return &Service{db: databaseInterface}
}
