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

//go:generate go tool mockery

import (
	"context"
)

// Database describes the persistence capabilities the opengraphschema Service requires.
type Database interface {
	GetEnvironmentsFiltered(ctx context.Context, onlyBuiltin bool) ([]Environment, error)
	GetNodeKind(ctx context.Context, id int32) (NodeKind, error)
	GetKindInfos(ctx context.Context, kindName string) ([]KindInfo, error)
	GetExtension(ctx context.Context, id int32) (Extension, error)
	GetRelationshipKind(ctx context.Context, id int32) (RelationshipKind, error)
}

// Service implements the OpenGraph schema use cases on top of a Database implementation.
type Service struct {
	db Database
}

// NewService constructs a Service backed by the supplied Database implementation.
func NewService(databaseInterface Database) *Service {
	return &Service{db: databaseInterface}
}
