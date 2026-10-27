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
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/packages/go/params"
)

// Log is the audit domain representation shared by the service and its adapters.
type Log struct {
	ID              int64
	CreatedAt       time.Time
	ActorID         string
	ActorName       string
	ActorEmail      string
	Action          string
	Fields          map[string]any
	RequestID       string
	SourceIPAddress string
	Status          string
	CommitID        uuid.UUID
}

// ListQuery contains the validated constraints for an audit log request.
type ListQuery struct {
	Before  time.Time
	After   time.Time
	Skip    int
	Limit   int
	Filters params.Filters
	Sort    params.SortItems
}

// Database defines the persistence operations required by the audit service.
type Database interface {
	ListAuditLogs(ctx context.Context, query ListQuery) ([]Log, int, error)
}

// Service contains audit business logic.
type Service struct {
	database Database
}

// NewService returns an audit service backed by database.
func NewService(database Database) *Service {
	return &Service{database: database}
}

// ListAuditLogs returns audit logs matching the validated query.
func (s *Service) ListAuditLogs(ctx context.Context, query ListQuery) ([]Log, int, error) {
	return s.database.ListAuditLogs(ctx, query)
}
