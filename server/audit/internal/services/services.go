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
	"errors"
	"fmt"
	"strings"
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

type Status string

const (
	StatusIntent     Status = "intent"
	StatusSuccess    Status = "success"
	StatusFailure    Status = "failure"
	UnknownActorName        = "unknown"
)

var (
	ErrInvalidAuditRecord = errors.New("invalid audit record")
	sensitiveFieldNames   = []string{"password", "secret", "token", "apikey", "privatekey"}
	fieldNameReplacer     = strings.NewReplacer("_", "", "-", "")
)

type AuditRecord struct {
	Action          string
	ActorID         string
	ActorName       string
	ActorEmail      string
	RequestID       string
	SourceIPAddress string
	Status          Status
	CommitID        uuid.UUID
	Fields          map[string]any
}

type Entry struct {
	Action          string
	ActorID         string
	ActorName       string
	ActorEmail      string
	RequestID       string
	SourceIPAddress string
	Fields          map[string]any
}

// Database defines the persistence operations required by the audit service.
type Database interface {
	ListAuditLogs(ctx context.Context, query ListQuery) ([]Log, int, error)
	InsertAuditLog(ctx context.Context, record AuditRecord) error
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

func (s *Service) Intent(ctx context.Context, entry Entry) (uuid.UUID, error) {
	commitID, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generating audit commit id: %w", err)
	}

	return commitID, s.database.InsertAuditLog(ctx, s.toRecord(entry, commitID, StatusIntent))
}

func (s *Service) Success(ctx context.Context, commitID uuid.UUID, entry Entry) error {
	return s.database.InsertAuditLog(ctx, s.toRecord(entry, commitID, StatusSuccess))
}

func (s *Service) Failure(ctx context.Context, commitID uuid.UUID, entry Entry) error {
	return s.database.InsertAuditLog(ctx, s.toRecord(entry, commitID, StatusFailure))
}

func (s *Service) toRecord(entry Entry, commitID uuid.UUID, status Status) AuditRecord {
	var actorName = entry.ActorName
	if entry.ActorID == "" && entry.ActorName == "" && entry.ActorEmail == "" {
		actorName = UnknownActorName
	}

	return AuditRecord{
		Action:          entry.Action,
		ActorID:         entry.ActorID,
		ActorName:       actorName,
		ActorEmail:      entry.ActorEmail,
		RequestID:       entry.RequestID,
		SourceIPAddress: entry.SourceIPAddress,
		Status:          status,
		CommitID:        commitID,
		Fields:          redactSensitiveFields(entry.Fields),
	}
}

func redactSensitiveFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return fields
	}

	var redactedFields = make(map[string]any, len(fields))
	for key, value := range fields {
		var normalizedKey = fieldNameReplacer.Replace(strings.ToLower(key))
		redactedFields[key] = value

		for _, sensitiveFieldName := range sensitiveFieldNames {
			if strings.Contains(normalizedKey, sensitiveFieldName) {
				redactedFields[key] = "[REDACTED]"
				break
			}
		}
	}

	return redactedFields
}
