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
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type Status string

const (
	StatusIntent  Status = "intent"
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
)

type Source string

const (
	// SourceMiddleware marks rows written by the audit middleware/service.
	SourceMiddleware Source = "middleware"
	// SourceLegacy marks rows copied from the pre-partitioning audit_logs table by
	// the partitioning migration; the service never writes this value.
	SourceLegacy Source = "legacy"
)

// UnknownActorName is the actor name recorded for a completely unattributed
// (unauthenticated) request, so the row is kept and attributed by source IP
// rather than dropped. It is applied centrally in toRecord so individual callers
// (the middleware today, handlers later) need not handle this edge case.
const UnknownActorName = "unknown"

// AuditRecord is the persistence-facing representation the Store writes.
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
	Source          Source
}

// Entry is the domain input the middleware/public API hands to the service.
type Entry struct {
	Action          string
	ActorID         string
	ActorName       string
	ActorEmail      string
	RequestID       string
	SourceIPAddress string
	Fields          map[string]any
}
type Service struct {
	db Database
}

func NewService(db Database) *Service {
	return &Service{db: db}
}

// Database is the port the audit service requires; appdb.Store implements it.
type Database interface {
	InsertAuditLog(ctx context.Context, record AuditRecord) error
}

// Maintainer is the port the GC daemon requires to manage audit partitions.
// appdb.Store implements it.
type Maintainer interface {
	CreateNextPartition(ctx context.Context, asOf time.Time) error
	DropExpiredPartitions(ctx context.Context, asOf time.Time, retentionMonths int) error
}

var ErrInvalidAuditRecord = errors.New("invalid audit record")

// sensitivePatternsLower holds the lowercase, separator-free substrings that
// mark a field key as sensitive (matched after normalizeKey).
var sensitivePatternsLower = []string{
	"password", "secret", "token", "apikey", "privatekey",
}

// normalizeKey lowercases a field key and strips "_" and "-" so separator
// variants (api_key, api-key, apikey) match the same patterns.
func normalizeKey(key string) string {
	keyLower := strings.ToLower(key)
	keyLower = strings.ReplaceAll(keyLower, "_", "")
	keyLower = strings.ReplaceAll(keyLower, "-", "")
	return keyLower
}

func redactSensitiveFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return fields
	}
	redacted := make(map[string]any, len(fields))
	for key, value := range fields {
		normalized := normalizeKey(key)
		isSensitive := false
		for _, pattern := range sensitivePatternsLower {
			if strings.Contains(normalized, pattern) {
				isSensitive = true
				break
			}
		}
		if isSensitive {
			redacted[key] = "[REDACTED]"
		} else {
			redacted[key] = value
		}
	}
	return redacted
}

// Intent writes the pre-execution row synchronously and returns the commit id
// that links it to the eventual result.
func (s *Service) Intent(ctx context.Context, entry Entry) (uuid.UUID, error) {
	var (
		commitID uuid.UUID
		err      error
	)
	commitID, err = uuid.NewV4()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("generating commit id: %w", err)
	}
	if err = s.db.InsertAuditLog(ctx, s.toRecord(entry, commitID, StatusIntent)); err != nil {
		return commitID, err
	}
	return commitID, nil
}

// Success writes the post-execution success row synchronously.
func (s *Service) Success(ctx context.Context, commitID uuid.UUID, entry Entry) error {
	return s.db.InsertAuditLog(ctx, s.toRecord(entry, commitID, StatusSuccess))
}

// Failure writes the post-execution failure row synchronously.
func (s *Service) Failure(ctx context.Context, commitID uuid.UUID, entry Entry) error {
	return s.db.InsertAuditLog(ctx, s.toRecord(entry, commitID, StatusFailure))
}

func (s *Service) toRecord(entry Entry, commitID uuid.UUID, status Status) AuditRecord {
	var actorName = entry.ActorName

	// Default a completely unattributed actor (an unauthenticated request) to the
	// unknown marker here, so callers do not each have to handle this edge case.
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
		Source:          SourceMiddleware,
	}
}
